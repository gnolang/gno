package rootmulti_test

// Pins the constructStore snapshot routing (gno#6011 fix layer: immutable
// query multistores read ms.db — the frozen post-commit snapshot — instead of
// the live dedicated-db mount): a query view held across commits must keep
// serving its version even after the LIVE DB prunes it. With live-DB routing
// the held view's lazy node loads hit pruned records and fail; with snapshot
// routing they read the frozen state. Deterministic (single-threaded).

import (
	"fmt"
	"testing"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	_ "github.com/gnolang/gno/tm2/pkg/db/pebbledb"
	storebptree "github.com/gnolang/gno/tm2/pkg/store/bptree"
	"github.com/gnolang/gno/tm2/pkg/store/rootmulti"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

func TestQueryView_SnapshotIsolationUnderPruning(t *testing.T) {
	db, err := dbm.NewDB("gnolang", dbm.PebbleDBBackend, t.TempDir())
	if err != nil {
		t.Fatalf("pebble: %v", err)
	}
	defer db.Close()

	mainKey := types.NewStoreKey("main")
	ms := rootmulti.NewMultiStore(db)
	// Aggressive retention: Store.Commit prunes everything below the previous
	// version on every commit.
	ms.SetStoreOptions(types.StoreOptions{PruningOptions: types.NewPruningOptions(1, 0)})
	ms.MountStoreWithDB(mainKey, storebptree.FastStoreConstructor, db)
	if err := ms.LoadLatestVersion(); err != nil {
		t.Fatalf("load: %v", err)
	}
	defer ms.Close()

	kname := func(i int) []byte { return fmt.Appendf(nil, "key%04d", i) }
	const keyspace = 2000
	commit := func(blk int, from, to int) {
		cms := ms.MultiCacheWrap()
		st := cms.GetStore(mainKey)
		for i := from; i < to; i++ {
			st.Set(nil, kname(i), fmt.Appendf(nil, "v%d/%d", blk, i))
		}
		cms.MultiWrite()
		ms.Commit()
	}

	// Blocks 1..5 populate the keyspace; the view is taken at height 5.
	for blk := 1; blk <= 5; blk++ {
		commit(blk, (blk-1)*keyspace/5, blk*keyspace/5)
	}
	view, release, err := ms.MultiImmutableCacheWrapWithVersion(5)
	if err != nil {
		t.Fatalf("view at 5: %v", err)
	}
	defer release()

	// Blocks 6..10 overwrite everything; pruning removes the version-5 records
	// (nodes and orphaned values) from the LIVE DB while the view is held. The
	// view is deliberately unregistered (long-lived query views must not pin
	// versions against pruning), so only snapshot isolation protects it.
	for blk := 6; blk <= 10; blk++ {
		commit(blk, (blk-6)*keyspace/5, (blk-5)*keyspace/5)
	}

	st := view.GetStore(mainKey)
	for i := 0; i < keyspace; i++ {
		wantBlk := i*5/keyspace + 1 // the block (1..5) that last wrote key i
		want := fmt.Sprintf("v%d/%d", wantBlk, i)
		got := st.Get(nil, kname(i)) // panics on pruned nodes under live routing
		if string(got) != want {
			t.Fatalf("held view diverged at %q: got %q want %q (snapshot isolation broken)",
				kname(i), got, want)
		}
	}
}
