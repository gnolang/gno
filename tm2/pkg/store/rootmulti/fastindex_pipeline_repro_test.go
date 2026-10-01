package rootmulti_test

// Single-threaded pipeline guard for the gno#6011 fix set. It drives the
// bptree store through the REAL production write pipeline — rootmulti
// multiStore -> CollectingDB (shared BatchCollector, buffered sub-batches) ->
// PrefixDB("s/_/") -> bptree — across blocks, clean/crash restarts,
// fast-index upgrades (off->on rebuild), and pruning, asserting after every
// commit that the persisted fast index matches an oracle and the
// authoritative tree, and that each load re-seeds the query snapshot.
//
// This does NOT reproduce #6011 itself — that is a concurrency race between
// the query connection and consensus commits, covered by the natural-race
// and concurrent tests. It guards the pieces the fix depends on that are
// observable single-threaded: buffered-collector batch semantics, LoadVersion
// query-snapshot seeding, and index parity across write/restart/upgrade/prune
// schedules. The existing store tests drive Store directly over a plain
// memdb, bypassing the collector pipeline entirely.

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"testing"

	bp "github.com/gnolang/gno/tm2/pkg/bptree"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	storebptree "github.com/gnolang/gno/tm2/pkg/store/bptree"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/rootmulti"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

func TestFastIndex_RootmultiPipelineFuzz(t *testing.T) {
	if testing.Short() {
		t.Skip("30-seed x 3-mode pipeline fuzz (~20s); CI runs it (no -short)")
	}
	for seed := int64(1); seed <= 30; seed++ {
		t.Run(fmt.Sprintf("seed=%d", seed), func(t *testing.T) {
			runPipelineFuzz(t, seed, false, false)
		})
		t.Run(fmt.Sprintf("seed=%d/prune", seed), func(t *testing.T) {
			runPipelineFuzz(t, seed, true, false)
		})
		t.Run(fmt.Sprintf("seed=%d/upgrade", seed), func(t *testing.T) {
			runPipelineFuzz(t, seed, false, true)
		})
	}
}

func runPipelineFuzz(t *testing.T, seed int64, prune, upgrade bool) {
	t.Helper()
	rng := rand.New(rand.NewSource(seed))
	db := memdb.NewMemDB()
	mainKey := types.NewStoreKey("main")
	baseKey := types.NewStoreKey("base")

	// newMS mirrors gnoland: main = bptree (fast index per `fast`), base =
	// dbadapter, both mounted over the same cfg.DB (=> prefix "s/_/"), and the
	// rootmulti metadata in the same DB.
	newMS := func(fast bool) types.CommitMultiStore {
		ms := rootmulti.NewMultiStore(db)
		if prune {
			ms.SetStoreOptions(types.StoreOptions{
				PruningOptions: types.NewPruningOptions(2, 0),
			})
		}
		ctor := storebptree.StoreConstructor
		if fast {
			ctor = storebptree.FastStoreConstructor
		}
		ms.MountStoreWithDB(mainKey, ctor, db)
		ms.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, db)
		if err := ms.LoadLatestVersion(); err != nil {
			t.Fatalf("LoadLatestVersion: %v", err)
		}
		// Loading must seed the query snapshot (snapshot-capable backend), so
		// a restarted node has query isolation before its first commit.
		if ms.QuerySnapshot() == nil {
			t.Fatal("LoadLatestVersion did not seed the query snapshot")
		}
		return ms
	}

	// Upgrade mode: first blocks run with the fast index OFF, then a restart
	// switches the mount to FastStoreConstructor (Load rebuilds the index
	// through the CollectingDB) — the PR #5937 deployment path.
	fastOn := !upgrade
	ms := newMS(fastOn)

	oracle := map[string]string{} // expected committed state of the main store
	kname := func(i int) []byte { return fmt.Appendf(nil, "key%04d", i) }
	// Large enough to force leaf splits and multi-level inner nodes (B=32):
	// 4096 keys => height 3, exercising COW split/merge on every block.
	const keyspace = 4096

	stageBlock := func(cms types.MultiStore, commitOracle bool, blk int) {
		st := cms.GetStore(mainKey)
		bst := cms.GetStore(baseKey)
		bst.Set(nil, []byte("_lastHeader"), fmt.Appendf(nil, "hdr%d", blk))
		n := 1 + rng.Intn(120)
		for i := range n {
			k := kname(rng.Intn(keyspace))
			if rng.Intn(5) == 0 {
				st.Delete(nil, k)
				if commitOracle {
					delete(oracle, string(k))
				}
			} else {
				v := fmt.Sprintf("v%d.%d.%d", blk, i, rng.Intn(1000))
				st.Set(nil, k, []byte(v))
				if commitOracle {
					oracle[string(k)] = v
				}
			}
		}
	}

	const blocks = 60
	for blk := 1; blk <= blocks; blk++ {
		if upgrade && blk == blocks/2 {
			fastOn = true
			ms = newMS(true) // upgrade restart: triggers index rebuild at Load
		}
		switch rng.Intn(8) {
		case 0: // clean restart
			ms = newMS(fastOn)
		case 1: // crash restart: block staged through MultiWrite, never committed
			cms := ms.MultiCacheWrap()
			stageBlock(cms, false, blk)
			cms.MultiWrite()
			ms = newMS(fastOn) // collector (and tree state) dropped; db unchanged
		}
		cms := ms.MultiCacheWrap()
		stageBlock(cms, true, blk)
		cms.MultiWrite()
		ms.Commit()
		if fastOn {
			checkFastIndexParity(t, db, oracle, blk, keyspace)
		}
	}
}

// checkFastIndexParity audits the RAW persisted state (what a restarted node
// would see) after a commit:
//  1. the fast-index stamp must be current (else the next Load would rebuild,
//     masking any staleness),
//  2. Get with the fast index on and off must agree with the oracle for every
//     key in the keyspace,
//  3. every persisted 'F' entry must match the authoritative tree: no
//     stale-present entries (key removed but entry remains), no stale values,
//     no entries newer than the committed version.
func checkFastIndexParity(t *testing.T, db dbm.DB, oracle map[string]string, blk, keyspace int) {
	t.Helper()
	pdb := dbm.NewPrefixDB(db, []byte("s/_/"))

	plain := bp.NewMutableTreeWithDB(pdb, 256, bp.NewNopLogger())
	if _, err := plain.Load(); err != nil {
		t.Fatalf("blk %d: plain load: %v", blk, err)
	}
	latest := plain.Version()

	// Stamp must exist and equal latest BEFORE constructing the fast tree —
	// a stale stamp would make fast.Load() rebuild the index and hide the bug.
	stampRaw, err := db.Get(append(append([]byte("s/_/"), bp.PrefixMeta), "fastidx"...))
	if err != nil || stampRaw == nil {
		t.Fatalf("blk %d: fast index stamp missing (err=%v)", blk, err)
	}
	payload, cerr := reproVerifyChecksum(stampRaw)
	if cerr != nil || len(payload) != 8 {
		t.Fatalf("blk %d: bad stamp record: %v", blk, cerr)
	}
	if stamp := int64(binary.BigEndian.Uint64(payload)); stamp != latest {
		t.Fatalf("blk %d: stamp %d != latest %d (Load would rebuild)", blk, stamp, latest)
	}

	fast := bp.NewMutableTreeWithDB(pdb, 256, bp.NewNopLogger(), bp.FastIndexOption(true))
	if _, err := fast.Load(); err != nil {
		t.Fatalf("blk %d: fast load: %v", blk, err)
	}

	// Whole keyspace (plus a margin of never-written keys): oracle vs plain
	// walk vs fast-index read.
	for i := 0; i < keyspace+8; i++ {
		k := fmt.Appendf(nil, "key%04d", i)
		want, present := oracle[string(k)]
		pv, err := plain.Get(k)
		if err != nil {
			t.Fatalf("blk %d: plain get %q: %v", blk, k, err)
		}
		fv, err := fast.Get(k)
		if err != nil {
			t.Fatalf("blk %d: fast get %q: %v", blk, k, err)
		}
		if present != (pv != nil) || (present && string(pv) != want) {
			t.Fatalf("blk %d: TREE DIVERGES FROM ORACLE key=%q tree=%q oracle=%q,%v",
				blk, k, pv, want, present)
		}
		if !bytes.Equal(pv, fv) {
			t.Fatalf("blk %d: FAST INDEX DIVERGES key=%q walk=%q fast=%q (issue #6011 shape)",
				blk, k, pv, fv)
		}
	}

	// Every persisted 'F' entry vs the authoritative tree.
	itr, err := pdb.Iterator([]byte{bp.PrefixFast}, []byte{bp.PrefixFast + 1})
	if err != nil {
		t.Fatalf("blk %d: F iterator: %v", blk, err)
	}
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		key := append([]byte(nil), itr.Key()[1:]...)
		payload, cerr := reproVerifyChecksum(itr.Value())
		if cerr != nil || len(payload) < 8 {
			t.Fatalf("blk %d: corrupt F entry %q: %v", blk, key, cerr)
		}
		ver := int64(binary.BigEndian.Uint64(payload[:8]))
		val := payload[8:]
		want, present := oracle[string(key)]
		if !present {
			t.Fatalf("blk %d: STALE-PRESENT F entry key=%q ver=%d val=%q (key removed from tree)",
				blk, key, ver, val)
		}
		if string(val) != want {
			t.Fatalf("blk %d: STALE F VALUE key=%q ver=%d fast=%q tree=%q (issue #6011 shape)",
				blk, key, ver, val, want)
		}
		if ver > latest {
			t.Fatalf("blk %d: F entry from the future key=%q ver=%d latest=%d", blk, key, ver, latest)
		}
	}
	if err := itr.Error(); err != nil {
		t.Fatalf("blk %d: F iteration: %v", blk, err)
	}
}
