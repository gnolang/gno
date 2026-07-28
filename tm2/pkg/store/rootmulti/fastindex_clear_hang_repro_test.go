package rootmulti_test

// Regression test for the clearFastIndex/CollectingDB incompatibility: a
// fast-index rebuild whose clear phase must delete >= 65536 existing 'F'
// entries used to loop forever when the tree's DB is CollectingDB-wrapped
// (the production wiring), because the chunked ndb.Commit() only moves
// deletes into the collector (not the DB) and CollectingDB.Iterator reads the
// underlying DB — so a restart-from-start scan re-saw the same keys forever.
// clearFastIndex now resumes each chunk after the last staged key, making
// progress independent of whether the deletes have been applied.
//
// Production trigger: run with the fast index ON (>= 65536 entries
// persisted), commit at least one version with it OFF (stamp goes stale),
// then restart with it ON — Load() rebuilds at startup.

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"

	bp "github.com/gnolang/gno/tm2/pkg/bptree"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

func TestFastIndex_RebuildOverCollectingDB(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a 70k-entry index")
	}
	db := memdb.NewMemDB()
	const entries = 70_000 // > fastRebuildFlush (65536)

	key := func(i int) []byte {
		return []byte{'k', byte(i >> 16), byte(i >> 8), byte(i)}
	}

	// Step 1: fast-ON tree over the raw db — persist `entries` F entries and
	// the stamp at version 1.
	on := bp.NewMutableTreeWithDB(db, 1024, bp.NewNopLogger(), bp.FastIndexOption(true))
	for i := range entries {
		if _, err := on.Set(key(i), []byte("v1")); err != nil {
			t.Fatalf("set: %v", err)
		}
	}
	if _, _, err := on.SaveVersion(); err != nil {
		t.Fatalf("save v1: %v", err)
	}

	// Step 2: fast-OFF tree over the same db commits version 2 — no index
	// maintenance, no stamp update. Disk now: stamp=1, latest=2, and the
	// version-2 write leaves a stale 'F' entry for its key.
	off := bp.NewMutableTreeWithDB(db, 1024, bp.NewNopLogger())
	if _, err := off.Load(); err != nil {
		t.Fatalf("off load: %v", err)
	}
	if _, err := off.Set(key(3), []byte("v2")); err != nil {
		t.Fatalf("off set: %v", err)
	}
	if _, _, err := off.SaveVersion(); err != nil {
		t.Fatalf("save v2: %v", err)
	}

	// Step 3: fast-ON tree over the PRODUCTION wiring (CollectingDB). Load()
	// sees stamp(1) < latest(2) and rebuilds; the clear phase must remove the
	// 70k stale entries first — and must terminate.
	collector := dbm.NewBatchCollector()
	cdb := dbm.NewCollectingDB(db, collector)
	prod := bp.NewMutableTreeWithDB(cdb, 1024, bp.NewNopLogger(), bp.FastIndexOption(true))

	type loadResult struct {
		latest int64
		err    error
	}
	done := make(chan loadResult, 1)
	go func() {
		latest, err := prod.Load()
		done <- loadResult{latest, err}
	}()

	var res loadResult
	select {
	case res = <-done:
	case <-time.After(30 * time.Second):
		t.Fatalf("REGRESSED: rebuild did not terminate (clearFastIndex loop); %d ops staged", collector.Len())
	}
	if res.err != nil {
		t.Fatalf("rebuild load: %v", res.err)
	}
	if res.latest != 2 {
		t.Fatalf("loaded latest = %d, want 2", res.latest)
	}
	// Single-pass clear (~entries deletes) + rebuild (~entries sets) + stamp:
	// far below the re-scan blowup (the old loop staged the first 65536 keys
	// over and over).
	if n := collector.Len(); n > 3*entries {
		t.Fatalf("REGRESSED: %d ops staged for %d entries (chunked clear re-scanned)", n, entries)
	}

	// The rebuild is durable only after the collector drains (in production:
	// the next block's rootmulti Commit). Drain and verify the result.
	realBatch := db.NewBatch()
	defer realBatch.Close()
	if err := collector.Drain(realBatch); err != nil {
		t.Fatalf("drain: %v", err)
	}
	if err := realBatch.Write(); err != nil {
		t.Fatalf("apply drain: %v", err)
	}

	stampRaw, err := db.Get(append([]byte{bp.PrefixMeta}, "fastidx"...))
	if err != nil || stampRaw == nil {
		t.Fatalf("stamp missing after drain (err=%v)", err)
	}
	payload, cerr := reproVerifyChecksum(stampRaw)
	if cerr != nil || len(payload) != 8 {
		t.Fatalf("bad stamp record: %v", cerr)
	}
	if stamp := int64(binary.BigEndian.Uint64(payload)); stamp != 2 {
		t.Fatalf("stamp = %d, want 2", stamp)
	}

	// Spot-check rebuilt entries against the authoritative tree.
	check := bp.NewMutableTreeWithDB(db, 1024, bp.NewNopLogger())
	if _, err := check.Load(); err != nil {
		t.Fatalf("check load: %v", err)
	}
	for _, i := range []int{0, 3, 65535, 65536, entries - 1} {
		k := key(i)
		want, err := check.Get(k)
		if err != nil {
			t.Fatalf("walk get %x: %v", k, err)
		}
		raw, err := db.Get(append([]byte{bp.PrefixFast}, k...))
		if err != nil || raw == nil {
			t.Fatalf("F entry missing for %x (err=%v)", k, err)
		}
		payload, cerr := reproVerifyChecksum(raw)
		if cerr != nil || len(payload) < 8 {
			t.Fatalf("corrupt F entry %x: %v", k, cerr)
		}
		if !bytes.Equal(payload[8:], want) {
			t.Fatalf("rebuilt F entry stale for %x: fast=%q tree=%q", k, payload[8:], want)
		}
	}
}
