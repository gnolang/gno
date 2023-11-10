package cache

import (
	"bytes"
	"container/list"
	"sort"
	"sync"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/std"

	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// If value is nil but deleted is false, it means the parent doesn't have the
// key.  (No need to delete upon Write())
type cValue struct {
	value   []byte
	deleted bool
	dirty   bool
}

// cacheStore wraps an in-memory cache around an underlying types.Store.
type cacheStore struct {
	mtx           sync.Mutex
	cache         map[string]*cValue
	unsortedCache map[string]struct{}
	sortedCache   *list.List // always ascending sorted
	parent        types.Store
}

var _ types.Store = (*cacheStore)(nil)

func New(parent types.Store) *cacheStore {
	return &cacheStore{
		cache:         make(map[string]*cValue),
		unsortedCache: make(map[string]struct{}),
		sortedCache:   list.New(),
		parent:        parent,
	}
}

// Implements types.Store.
func (store *cacheStore) Get(key []byte) (value []byte) {
	store.mtx.Lock()
	defer store.mtx.Unlock()
	types.AssertValidKey(key)

	cacheValue, ok := store.cache[string(key)]
	if !ok {
		value = store.parent.Get(key)
		store.setCacheValue(key, value, false, false)
	} else {
		value = cacheValue.value
	}

	return value
}

// Implements types.Store.
func (store *cacheStore) Set(key []byte, value []byte) {
	store.mtx.Lock()
	defer store.mtx.Unlock()
	types.AssertValidKey(key)
	types.AssertValidValue(value)

	store.setCacheValue(key, value, false, true)
}

// Implements types.Store.
func (store *cacheStore) Has(key []byte) bool {
	value := store.Get(key)
	return value != nil
}

// Implements types.Store.
func (store *cacheStore) Delete(key []byte) {
	store.mtx.Lock()
	defer store.mtx.Unlock()
	types.AssertValidKey(key)

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

	// TODO: Consider allowing usage of Batch, which would allow the write to
	// at least happen atomically.
	for _, key := range keys {
		cacheValue := store.cache[key]
		if cacheValue.deleted {
			store.parent.Delete([]byte(key))
		} else if cacheValue.value == nil {
			// Skip, it already doesn't exist in parent.
		} else {
			store.parent.Set([]byte(key), cacheValue.value)
		}
	}

	// Clear the cache
	store.cache = make(map[string]*cValue)
	store.unsortedCache = make(map[string]struct{})
	store.sortedCache = list.New()
}

// ----------------------------------------
// To cache-wrap this Store further.

// Implements Store.
func (store *cacheStore) CacheWrap() types.Store {
	return New(store)
}

// ----------------------------------------
// Iteration

// Implements types.Store.
func (store *cacheStore) Iterator(start, end []byte) types.Iterator {
	return store.iterator(start, end, true)
}

// Implements types.Store.
func (store *cacheStore) ReverseIterator(start, end []byte) types.Iterator {
	return store.iterator(start, end, false)
}

func (store *cacheStore) iterator(start, end []byte, ascending bool) types.Iterator {
	store.mtx.Lock()
	defer store.mtx.Unlock()

	var parent, cache types.Iterator

	if ascending {
		parent = store.parent.Iterator(start, end)
	} else {
		parent = store.parent.ReverseIterator(start, end)
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
