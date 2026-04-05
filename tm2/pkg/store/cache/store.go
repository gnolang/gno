package cache

import (
	"bytes"
	"container/list"
	"fmt"
	"reflect"
	"sort"
	"sync"

	"github.com/gnolang/gno/tm2/pkg/colors"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/std"

	"github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/gnolang/gno/tm2/pkg/store/utils"
)

// dbadapterStore is implemented by dbadapter.Store. Used by Write()
// to batch writes to the backing DB for atomicity and performance.
type dbadapterStore interface {
	GetDB() dbm.DB
}

// If value is nil but deleted is false, it means the parent doesn't have the
// key.  (No need to delete upon Write())
type cValue struct {
	value   []byte
	deleted bool
	dirty   bool
}

func (cv cValue) String() string {
	return fmt.Sprintf("cValue{%s,%v,%v}",
		colors.DefaultColoredBytes(cv.value),
		cv.deleted, cv.dirty)
}

// cacheStore wraps an in-memory cache around an underlying types.Store.
type cacheStore struct {
	mtx            sync.Mutex
	cache          map[string]*cValue
	unsortedCache  map[string]struct{}
	sortedCache    *list.List // always ascending sorted
	parent         types.Store
	depthEstimator types.DepthEstimator // nil for flat stores (e.g. dbadapter)
	chargedGas     map[string]types.Gas // write/delete gas deduplication per key
}

var _ types.Store = (*cacheStore)(nil)

func New(parent types.Store) *cacheStore {
	cs := &cacheStore{
		cache:         make(map[string]*cValue),
		unsortedCache: make(map[string]struct{}),
		sortedCache:   list.New(),
		parent:        parent,
		chargedGas:    make(map[string]types.Gas),
	}
	// Auto-detect DepthEstimator from parent.
	if de, ok := parent.(types.DepthEstimator); ok {
		cs.depthEstimator = de
	}
	return cs
}

// SetDepthEstimator sets the depth estimator for IAVL-backed stores.
func (store *cacheStore) SetDepthEstimator(de types.DepthEstimator) {
	store.depthEstimator = de
}

// expectedDepth returns the estimated IAVL tree depth, floored by
// GasConfig.MinDepth. Returns 1 for non-IAVL stores (no estimator).
func (store *cacheStore) expectedDepth(gctx *types.GasContext) int64 {
	if store.depthEstimator == nil {
		return 1 // flat store (dbadapter), no depth
	}
	depth := store.depthEstimator.ExpectedDepth()
	if gctx != nil && gctx.Config.MinDepth > 0 && depth < gctx.Config.MinDepth {
		return gctx.Config.MinDepth
	}
	return depth
}

// Implements types.Store.
func (store *cacheStore) Get(gctx *types.GasContext, key []byte) (value []byte) {
	store.mtx.Lock()
	defer store.mtx.Unlock()
	types.AssertValidKey(key)

	cacheValue, ok := store.cache[string(key)]
	if !ok {
		// Cache miss — charge depth-based I/O gas, then fetch.
		if gctx != nil {
			depth := store.expectedDepth(gctx)
			if depth > 1 {
				gctx.ConsumeGas(types.Gas(depth)*gctx.Config.ReadCostFlat, "DepthReadFlat")
			} else {
				gctx.WillGet() // flat ReadCostFlat (non-depth store)
			}
			value = store.parent.Get(nil, key)
			gctx.DidGet(value) // ReadCostPerByte (nil-safe)
		} else {
			value = store.parent.Get(nil, key)
		}
		store.setCacheValue(key, value, false, false)
	} else {
		// Cache hit — no gas.
		value = cacheValue.value
	}

	return value
}

// Implements types.Store.
func (store *cacheStore) Set(gctx *types.GasContext, key []byte, value []byte) {
	store.mtx.Lock()
	defer store.mtx.Unlock()
	types.AssertValidKey(key)
	types.AssertValidValue(value)

	// Write gas deduplication: refund previous charge for this key,
	// then charge for the new operation. Last operation wins.
	if gctx != nil {
		k := string(key)
		if prev, exists := store.chargedGas[k]; exists && prev > 0 {
			gctx.RefundGas(prev)
		}
		var gas types.Gas
		depth := store.expectedDepth(gctx)
		if depth > 1 {
			depthGas := types.Gas(depth) * (gctx.Config.ReadCostFlat + gctx.Config.WriteCostFlat)
			depthGas += gctx.Config.WriteCostPerByte * types.Gas(len(value))
			gctx.ConsumeGas(depthGas, "IavlSet")
			gas = depthGas
		} else {
			gas = gctx.WillSet(value)
		}
		store.chargedGas[k] = gas
	}

	store.setCacheValue(key, value, false, true)
}

// Implements types.Store.
func (store *cacheStore) Has(gctx *types.GasContext, key []byte) bool {
	value := store.Get(gctx, key)
	return value != nil
}

// Implements types.Store.
func (store *cacheStore) Delete(gctx *types.GasContext, key []byte) {
	store.mtx.Lock()
	defer store.mtx.Unlock()
	types.AssertValidKey(key)

	// Write gas deduplication: refund previous charge, charge delete.
	if gctx != nil {
		k := string(key)
		if prev, exists := store.chargedGas[k]; exists && prev > 0 {
			gctx.RefundGas(prev)
		}
		var gas types.Gas
		depth := store.expectedDepth(gctx)
		if depth > 1 {
			// IAVL: depth reads + depth writes to remove and rebalance
			depthGas := types.Gas(depth) * (gctx.Config.ReadCostFlat + gctx.Config.WriteCostFlat)
			gctx.ConsumeGas(depthGas, "IavlDelete")
			gas = depthGas
		} else {
			gas = gctx.WillDelete() // DeleteCost
		}
		store.chargedGas[k] = gas
	}

	store.setCacheValue(key, nil, true, true)
}

// Implements types.Store.
func (store *cacheStore) Write() {
	store.mtx.Lock()
	defer store.mtx.Unlock()

	// We need a copy of all of the keys.
	// Not the best, but probably not a bottleneck depending.
	keys := make([]string, 0, len(store.cache))
	for key, dbValue := range store.cache {
		if dbValue.dirty {
			keys = append(keys, key)
		}
	}

	sort.Strings(keys)

	// Use Batch if the parent is a dbadapter with a backing DB,
	// for atomicity and to amortize fsync cost (critical for LMDB).
	if dba, ok := store.parent.(dbadapterStore); ok {
		db := dba.GetDB()
		batch := db.NewBatch()
		defer batch.Close()
		for _, key := range keys {
			cacheValue := store.cache[key]
			if cacheValue.deleted {
				if err := batch.Delete([]byte(key)); err != nil {
					panic(err)
				}
			} else if cacheValue.value == nil {
				// Skip, it already doesn't exist in parent.
			} else {
				if err := batch.Set([]byte(key), cacheValue.value); err != nil {
					panic(err)
				}
			}
		}
		if err := batch.Write(); err != nil {
			panic(err)
		}
	} else {
		for _, key := range keys {
			cacheValue := store.cache[key]
			if cacheValue.deleted {
				store.parent.Delete(nil, []byte(key))
			} else if cacheValue.value == nil {
				// Skip, it already doesn't exist in parent.
			} else {
				store.parent.Set(nil, []byte(key), cacheValue.value)
			}
		}
	}

	// Clear the cache
	store.clear()
}

func (store *cacheStore) Flush() {
	store.Write()
	if fs, ok := store.parent.(types.Flusher); ok {
		fs.Flush()
	}
}

func (store *cacheStore) clear() {
	store.cache = make(map[string]*cValue)
	store.unsortedCache = make(map[string]struct{})
	store.sortedCache = list.New()
	store.chargedGas = make(map[string]types.Gas)
}

// ----------------------------------------
// To cache-wrap this Store further.

// Implements Store.
func (store *cacheStore) CacheWrap() types.Store {
	cs := New(store)
	// Propagate depth estimator to nested cache layers.
	cs.depthEstimator = store.depthEstimator
	return cs
}

// ----------------------------------------
// Iteration

// Implements types.Store.
func (store *cacheStore) Iterator(gctx *types.GasContext, start, end []byte) types.Iterator {
	return store.iterator(gctx, start, end, true)
}

// Implements types.Store.
func (store *cacheStore) ReverseIterator(gctx *types.GasContext, start, end []byte) types.Iterator {
	return store.iterator(gctx, start, end, false)
}

func (store *cacheStore) iterator(gctx *types.GasContext, start, end []byte, ascending bool) types.Iterator {
	store.mtx.Lock()
	defer store.mtx.Unlock()

	var parent, cache types.Iterator

	if ascending {
		parent = store.parent.Iterator(nil, start, end)
	} else {
		parent = store.parent.ReverseIterator(nil, start, end)
	}

	store.dirtyItems(start, end)
	cache = newMemIterator(start, end, store.sortedCache, ascending)

	return newCacheMergeIterator(parent, cache, ascending)
}

// Constructs a slice of dirty items, to use w/ memIterator.
func (store *cacheStore) dirtyItems(start, end []byte) {
	unsorted := make([]*std.KVPair, 0)

	for key := range store.unsortedCache {
		cacheValue := store.cache[key]
		if dbm.IsKeyInDomain([]byte(key), start, end) {
			unsorted = append(unsorted, &std.KVPair{Key: []byte(key), Value: cacheValue.value})
			delete(store.unsortedCache, key)
		}
	}

	sort.Slice(unsorted, func(i, j int) bool {
		return bytes.Compare(unsorted[i].Key, unsorted[j].Key) < 0
	})

	// #nosec G602
	for e := store.sortedCache.Front(); e != nil && len(unsorted) != 0; {
		uitem := unsorted[0]
		sitem := e.Value.(*std.KVPair)
		comp := bytes.Compare(uitem.Key, sitem.Key)
		switch comp {
		case -1:
			unsorted = unsorted[1:]
			store.sortedCache.InsertBefore(uitem, e)
		case 1:
			e = e.Next()
		case 0:
			unsorted = unsorted[1:]
			e.Value = uitem
			e = e.Next()
		}
	}

	for _, kvp := range unsorted {
		store.sortedCache.PushBack(kvp)
	}
}

// ----------------------------------------
// etc

// Only entrypoint to mutate store.cache.
func (store *cacheStore) setCacheValue(key, value []byte, deleted bool, dirty bool) {
	store.cache[string(key)] = &cValue{
		value:   value,
		deleted: deleted,
		dirty:   dirty,
	}
	if dirty {
		store.unsortedCache[string(key)] = struct{}{}
	}
}

func (store *cacheStore) Print() {
	fmt.Println(colors.Cyan("cacheStore.Print"), fmt.Sprintf("%p", store))
	for key, value := range store.cache {
		fmt.Println(
			colors.DefaultColoredBytesN([]byte(key), 50),
			colors.DefaultColoredBytesN(value.value, 100),
			"deleted", value.deleted,
			"dirty", value.dirty,
		)
	}
	fmt.Println(colors.Cyan("cacheStore.Print"), fmt.Sprintf("%p", store),
		"print parent", fmt.Sprintf("%p", store.parent), reflect.TypeOf(store.parent))
	if ps, ok := store.parent.(types.Printer); ok {
		ps.Print()
	} else {
		utils.Print(store.parent)
	}
	fmt.Println(colors.Cyan("cacheStore.Print END"), fmt.Sprintf("%p", store))
}
