package rootmulti_test

// Concurrency regression test for https://github.com/gnolang/gno/issues/6011:
// the query ABCI connection has an INDEPENDENT mutex (tm2/pkg/bft/proxy), so
// query-path reads run concurrently with consensus commits. This hammers the
// production query surfaces — custom-query snapshot loads (handleQueryCustom /
// Simulate shape) and snapshot-backed .store queries (QueryImmutable) — while
// blocks commit, then audits the persisted fast index for staleness. Any
// panic in a query goroutine fails the test (no recover): post-fix these
// surfaces read frozen snapshots and must be crash-free.

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	_ "github.com/gnolang/gno/tm2/pkg/db/pebbledb"
	storebptree "github.com/gnolang/gno/tm2/pkg/store/bptree"
	"github.com/gnolang/gno/tm2/pkg/store/rootmulti"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

func TestFastIndex_ConcurrentQueryCommit(t *testing.T) {
	db, err := dbm.NewDB("gnolang", dbm.PebbleDBBackend, t.TempDir())
	if err != nil {
		t.Fatalf("pebble: %v", err)
	}
	defer db.Close()

	mainKey := types.NewStoreKey("main")
	ms := rootmulti.NewMultiStore(db)
	// Production-like retention (topaz runs syncable; pruning never fired in
	// the incident window).
	ms.SetStoreOptions(types.StoreOptions{PruningOptions: types.NewPruningOptions(0, 1)})
	ms.MountStoreWithDB(mainKey, storebptree.FastStoreConstructor, db)
	if err := ms.LoadLatestVersion(); err != nil {
		t.Fatalf("load: %v", err)
	}
	defer ms.Close() // release the query snapshot before db.Close

	rng := rand.New(rand.NewSource(1))
	oracle := map[string]string{}
	kname := func(i int) []byte { return fmt.Appendf(nil, "key%04d", i) }
	const keyspace = 2048

	var stop atomic.Bool
	var height atomic.Int64
	var wg sync.WaitGroup

	// Query hammer goroutines: the surfaces the RPC/query connection exercises
	// concurrently with consensus in production.
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			qrng := rand.New(rand.NewSource(int64(100 + g)))
			for !stop.Load() {
				h := height.Load()
				if h < 2 {
					continue
				}
				switch qrng.Intn(3) {
				case 0:
					// handleQueryCustom shape: snapshot multistore at a
					// recent height.
					qh := h - qrng.Int63n(2)
					cacheMS, release, err := ms.MultiImmutableCacheWrapWithVersion(qh)
					if err != nil {
						continue // height pruned/not yet visible — fine
					}
					st := cacheMS.GetStore(mainKey)
					for i := 0; i < 8; i++ {
						st.Get(nil, kname(qrng.Intn(keyspace)))
					}
					release()
				case 1:
					// handleQueryStore shape: snapshot-backed .store query.
					req := abci.RequestQuery{
						Path:   "/main/key",
						Data:   kname(qrng.Intn(keyspace)),
						Height: h - qrng.Int63n(2),
					}
					if _, err := ms.QueryImmutable(req); err != nil {
						continue // no snapshot view for that height — fine
					}
				case 2:
					// Simulate shape: snapshot multistore at the latest height.
					cacheMS, release, err := ms.MultiImmutableCacheWrapWithVersion(h)
					if err != nil {
						continue
					}
					st := cacheMS.GetStore(mainKey)
					for i := 0; i < 8; i++ {
						st.Get(nil, kname(qrng.Intn(keyspace)))
					}
					release()
				}
			}
		}(g)
	}

	const blocks = 400
	for blk := 1; blk <= blocks; blk++ {
		cms := ms.MultiCacheWrap()
		st := cms.GetStore(mainKey)
		n := 1 + rng.Intn(60)
		for i := 0; i < n; i++ {
			k := kname(rng.Intn(keyspace))
			if rng.Intn(6) == 0 {
				st.Delete(nil, k)
				delete(oracle, string(k))
			} else {
				v := fmt.Sprintf("v%d.%d", blk, i)
				st.Set(nil, k, []byte(v))
				oracle[string(k)] = v
			}
		}
		cms.MultiWrite()
		ms.Commit()
		height.Store(int64(blk))

		if blk%20 == 0 || blk == blocks {
			checkFastIndexParity(t, db, oracle, blk, keyspace)
			if t.Failed() {
				break
			}
		}
	}
	stop.Store(true)
	wg.Wait()
	checkFastIndexParity(t, db, oracle, blocks, keyspace)
}
