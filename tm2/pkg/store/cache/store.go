package cache

import (
	"bytes"
	"container/list"
	"fmt"
	"maps"
	"reflect"
	"sort"
	"sync"

	"github.com/gnolang/gno/tm2/pkg/colors"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/overflow"
	"github.com/gnolang/gno/tm2/pkg/std"

	"github.com/gnolang/gno/tm2/pkg/store/trace"
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
	mtx           sync.Mutex
	cache         map[string]*cValue
	unsortedCache map[string]struct{}
	sortedCache   *list.List // always ascending sorted
	parent        types.Store
	chargedGas    map[string]types.Gas // write/delete gas deduplication per key

	// Checkpoint for rollback support. When set, WriteCheckpoint()
	// restores these snapshots and flushes only the checkpointed state.
	checkpointCache      map[string]*cValue
	checkpointChargedGas map[string]types.Gas

	// Depth estimation for gas. Cached at construction time from
	// DepthEstimator (IAVL/B+tree). 100x fixed-point (300 = 3.0).
	// hasEstimator is false for flat stores (dbadapter).
	hasEstimator    bool
	getReadDepth100 int64
	setReadDepth100 int64
	writeDepth100   int64
}

var (
	_ types.Store          = (*cacheStore)(nil)
	_ types.Checkpointable = (*cacheStore)(nil)
)

func New(parent types.Store) *cacheStore {
	cs := &cacheStore{
		cache:         make(map[string]*cValue),
		unsortedCache: make(map[string]struct{}),
		sortedCache:   list.New(),
		parent:        parent,
		chargedGas:    make(map[string]types.Gas),
	}
	// Auto-detect DepthEstimator from parent and cache depths.
	if de, ok := parent.(types.DepthEstimator); ok {
		cs.hasEstimator = true
		cs.getReadDepth100 = de.ExpectedGetReadDepth100()
		cs.setReadDepth100 = de.ExpectedSetReadDepth100()
		cs.writeDepth100 = de.ExpectedWriteDepth100()
	}
	return cs
}

// effectiveGetReadDepth100 returns the GET read depth.
// If FixedGetReadDepth100 is set, uses that exactly.
// Otherwise uses the tree estimate, floored by MinGetReadDepth100.
func (store *cacheStore) effectiveGetReadDepth100(gctx *types.GasContext) int64 {
	if gctx != nil && gctx.Config.FixedGetReadDepth100 > 0 {
		return gctx.Config.FixedGetReadDepth100
	}
	d := store.getReadDepth100
	if gctx != nil && gctx.Config.MinGetReadDepth100 > 0 && d < gctx.Config.MinGetReadDepth100 {
		d = gctx.Config.MinGetReadDepth100
	}
	return d
}

// effectiveSetReadDepth100 returns the SET read depth.
// If FixedSetReadDepth100 is set, uses that exactly.
// Otherwise uses the tree estimate, floored by MinSetReadDepth100.
func (store *cacheStore) effectiveSetReadDepth100(gctx *types.GasContext) int64 {
	if gctx != nil && gctx.Config.FixedSetReadDepth100 > 0 {
		return gctx.Config.FixedSetReadDepth100
	}
	d := store.setReadDepth100
	if gctx != nil && gctx.Config.MinSetReadDepth100 > 0 && d < gctx.Config.MinSetReadDepth100 {
		d = gctx.Config.MinSetReadDepth100
	}
	return d
}

// effectiveWriteDepth100 returns the write depth.
// If FixedWriteDepth100 is set, uses that exactly.
// Otherwise uses the tree estimate, floored by MinWriteDepth100.
func (store *cacheStore) effectiveWriteDepth100(gctx *types.GasContext) int64 {
	if gctx != nil && gctx.Config.FixedWriteDepth100 > 0 {
		return gctx.Config.FixedWriteDepth100
	}
	d := store.writeDepth100
	if gctx != nil && gctx.Config.MinWriteDepth100 > 0 && d < gctx.Config.MinWriteDepth100 {
		d = gctx.Config.MinWriteDepth100
	}
	return d
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
			var gas types.Gas
			if store.hasEstimator {
				d := store.effectiveGetReadDepth100(gctx)
				gas = overflow.Mulp(d, gctx.Config.ReadCostFlat) / 100
				gctx.ConsumeGas(gas, "DepthReadFlat")
			} else {
				gctx.WillGet() // flat ReadCostFlat (non-depth store)
				gas = gctx.Config.ReadCostFlat
			}
			value = store.parent.Get(nil, key)
			perByte := overflow.Mulp(gctx.Config.ReadCostPerByte, types.Gas(len(value)))
			gctx.DidGet(value) // ReadCostPerByte (nil-safe)
			if trace.StoreGasEnabled {
				trace.Store("GET", overflow.Addp(gas, perByte), key, len(value),
					fmt.Sprintf("depth=%v", store.hasEstimator))
			}
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
			if trace.StoreGasEnabled {
				trace.Store("REFUND", prev, key, 0, "dedup")
			}
		}
		var gas types.Gas
		if store.hasEstimator {
			rd := store.effectiveSetReadDepth100(gctx)
			wd := store.effectiveWriteDepth100(gctx)
			rdGas := overflow.Mulp(rd, gctx.Config.ReadCostFlat) / 100
			wdGas := overflow.Mulp(wd, gctx.Config.WriteCostFlat) / 100
			pbGas := overflow.Mulp(gctx.Config.WriteCostPerByte, types.Gas(len(value)))
			gas = overflow.Addp(overflow.Addp(rdGas, wdGas), pbGas)
			gctx.ConsumeGas(gas, "DepthSet")
		} else {
			gas = gctx.WillSet(value)
		}
		store.chargedGas[k] = gas
		if trace.StoreGasEnabled {
			trace.Store("SET", gas, key, len(value),
				fmt.Sprintf("depth=%v", store.hasEstimator))
		}
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
			if trace.StoreGasEnabled {
				trace.Store("REFUND", prev, key, 0, "dedup")
			}
		}
		var gas types.Gas
		if store.hasEstimator {
			rd := store.effectiveSetReadDepth100(gctx)
			wd := store.effectiveWriteDepth100(gctx)
			rdGas := overflow.Mulp(rd, gctx.Config.ReadCostFlat) / 100
			wdGas := overflow.Mulp(wd, gctx.Config.WriteCostFlat) / 100
			gas = overflow.Addp(rdGas, wdGas)
			gctx.ConsumeGas(gas, "DepthDelete")
		} else {
			gas = gctx.WillDelete() // DeleteCost
		}
		store.chargedGas[k] = gas
		if trace.StoreGasEnabled {
			trace.Store("DELETE", gas, key, 0,
				fmt.Sprintf("depth=%v", store.hasEstimator))
		}
	}

	store.setCacheValue(key, nil, true, true)
}

// Implements types.Store.
func (store *cacheStore) Write() {
	store.mtx.Lock()
	defer store.mtx.Unlock()
	store.writeLocked()
}

// writeLocked flushes dirty cache entries to the parent and clears the cache.
// Caller must hold store.mtx.
func (store *cacheStore) writeLocked() {
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
	store.checkpointCache = nil
	store.checkpointChargedGas = nil
}

// ----------------------------------------
// Checkpoint/rollback support.

// Checkpoint saves a shallow clone of the cache and chargedGas maps.
// Used by BaseApp to snapshot ante handler state before msg execution.
// setCacheValue always allocates a new *cValue, so the cloned map's
// pointers remain valid after subsequent Set/Delete calls.
//
// The GasMeter's consumed counter is intentionally NOT snapshotted.
// On msg failure/OOG, WriteCheckpoint rewinds writes but the meter
// keeps everything charged during the msg — the SDK "failed tx burns
// gas" invariant. Rewinding the meter would refund gas for a
// rolled-back attempt and let an attacker retry expensive operations
// for the cost of the ante alone. chargedGas being tx-local (one
// cacheStore per tx) means no later tx sees the restored map, so
// there's no cross-tx write-dedup inconsistency from the asymmetry.
func (store *cacheStore) Checkpoint() {
	store.mtx.Lock()
	defer store.mtx.Unlock()
	store.checkpointCache = maps.Clone(store.cache)
	store.checkpointChargedGas = maps.Clone(store.chargedGas)
}

// HasCheckpoint returns true if a checkpoint is active.
func (store *cacheStore) HasCheckpoint() bool {
	store.mtx.Lock()
	defer store.mtx.Unlock()
	return store.checkpointCache != nil
}

// WriteCheckpoint restores the checkpoint snapshot, then flushes only
// the checkpointed (ante) entries to the parent store.
func (store *cacheStore) WriteCheckpoint() {
	store.mtx.Lock()
	defer store.mtx.Unlock()
	if store.checkpointCache == nil {
		panic("WriteCheckpoint called without Checkpoint")
	}
	store.cache = store.checkpointCache
	store.chargedGas = store.checkpointChargedGas
	store.checkpointCache = nil
	store.checkpointChargedGas = nil
	store.writeLocked()
}

// ----------------------------------------
// To cache-wrap this Store further.

// Implements Store.
func (store *cacheStore) CacheWrap() types.Store {
	cs := New(store)
	// Propagate cached depths to nested cache layers.
	cs.hasEstimator = store.hasEstimator
	cs.getReadDepth100 = store.getReadDepth100
	cs.setReadDepth100 = store.setReadDepth100
	cs.writeDepth100 = store.writeDepth100
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

	return newGasIterator(gctx, newCacheMergeIterator(parent, cache, ascending))
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
