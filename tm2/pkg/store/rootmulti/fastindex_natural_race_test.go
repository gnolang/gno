package rootmulti_test

// Natural (hook-free) reproduction of #6011: hammer the query path
// (MultiImmutableCacheWrapWithVersion, as handleQueryCustom does on the
// query connection's independent mutex) while blocks commit. Each block
// updates a DISJOINT slice of pre-populated keys exactly once, so a stale
// rebuild racing block N leaves permanently stale entries for block N's keys
// that the final audit detects.

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	bp "github.com/gnolang/gno/tm2/pkg/bptree"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	_ "github.com/gnolang/gno/tm2/pkg/db/pebbledb"
	storebptree "github.com/gnolang/gno/tm2/pkg/store/bptree"
	"github.com/gnolang/gno/tm2/pkg/store/rootmulti"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

func TestFastIndex_NaturalQueryCommitRace(t *testing.T) {
	if testing.Short() {
		t.Skip("timing-dependent race hunt")
	}
	db, err := dbm.NewDB("gnolang", dbm.PebbleDBBackend, t.TempDir())
	if err != nil {
		t.Fatalf("pebble: %v", err)
	}
	defer db.Close()

	mainKey := types.NewStoreKey("main")
	ms := rootmulti.NewMultiStore(db)
	ms.SetStoreOptions(types.StoreOptions{PruningOptions: types.NewPruningOptions(0, 1)})
	ms.MountStoreWithDB(mainKey, storebptree.FastStoreConstructor, db)
	if err := ms.LoadLatestVersion(); err != nil {
		t.Fatalf("load: %v", err)
	}
	defer ms.Close() // release the query snapshot before db.Close

	kname := func(i int) []byte { return fmt.Appendf(nil, "acct%05d", i) }
	oracle := map[string]string{}

	commit := func(from, to int, blk int) {
		cms := ms.MultiCacheWrap()
		st := cms.GetStore(mainKey)
		for i := from; i < to; i++ {
			v := fmt.Sprintf("v%d/%d", blk, i)
			st.Set(nil, kname(i), []byte(v))
			oracle[string(kname(i))] = v
		}
		cms.MultiWrite()
		ms.Commit()
	}

	// Pre-populate 20000 keys across 10 blocks.
	const total = 20000
	blk := 0
	for i := 0; i < total; i += 2000 {
		blk++
		commit(i, i+2000, blk)
	}

	var stop atomic.Bool
	var height atomic.Int64
	var queryLoads atomic.Int64 // successful query-view loads, for the race floor
	height.Store(int64(blk))
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for !stop.Load() {
				h := height.Load()
				if h < 2 {
					continue
				}
				// handleQueryCustom's exact call, racing the commit drain.
				cacheMS, release, err := ms.MultiImmutableCacheWrapWithVersion(h - 1)
				if err != nil {
					continue
				}
				cacheMS.GetStore(mainKey) // construction already did the damage, if any
				release()
				queryLoads.Add(1)
			}
		})
	}

	// Update blocks: each block touches a disjoint 10-key slice exactly once.
	const updateBlocks = 1900
	for u := range updateBlocks {
		blk++
		from := (u * 10) % total
		commit(from, from+10, blk)
		height.Store(int64(blk))
	}
	stop.Store(true)
	wg.Wait()

	// Race floor: if the query goroutines never actually interleaved with the
	// committing writer (a fully-serialized scheduler), a zero-stale result
	// below would be vacuous. Require that they did substantial concurrent
	// work, so a genuine regression can't hide behind a quiet runner.
	if n := queryLoads.Load(); n < int64(updateBlocks) {
		t.Fatalf("query goroutines barely ran (%d loads over %d blocks); race not exercised", n, updateBlocks)
	}

	// Audit persisted fast index against the tree, from fresh handles.
	pdb := dbm.NewPrefixDB(db, []byte("s/_/"))
	plain := bp.NewMutableTreeWithDB(pdb, 256, bp.NewNopLogger())
	if _, err := plain.Load(); err != nil {
		t.Fatalf("plain load: %v", err)
	}
	stale := 0
	for i := range total {
		k := kname(i)
		want := oracle[string(k)]
		got, err := plain.Get(k)
		if err != nil {
			t.Fatalf("walk get %q: %v", k, err)
		}
		if string(got) != want {
			t.Fatalf("tree diverged at %q: %q != %q", k, got, want)
		}
		raw, err := pdb.Get(append([]byte{bp.PrefixFast}, k...))
		if err != nil {
			t.Fatalf("F get %q: %v", k, err)
		}
		if raw == nil {
			continue // missing entry is advisory-benign
		}
		payload, cerr := reproVerifyChecksum(raw)
		if cerr != nil || len(payload) < 8 {
			t.Fatalf("corrupt F entry %q: %v", k, cerr)
		}
		if !bytes.Equal(payload[8:], got) {
			stale++
			if stale <= 5 {
				t.Logf("STALE F entry key=%q fast=%q tree=%q", k, payload[8:], got)
			}
		}
	}
	if stale > 0 {
		t.Fatalf("NATURAL RACE REPRODUCED #6011: %d stale fast-index entries persisted", stale)
	}
}
