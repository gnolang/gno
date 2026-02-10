package cachemulti

import (
	"maps"

	"github.com/gnolang/gno/tm2/pkg/store/types"
)

//----------------------------------------
// Store

// Store holds many cache-wrapped stores.
// Implements MultiStore.
// NOTE: a Store (and MultiStores in general) should never expose the
// keys for the substores.
type Store struct {
	stores map[types.StoreKey]types.Store
	keys   map[string]types.StoreKey
}

var _ types.MultiStore = Store{}

func NewFromStores(
	stores map[types.StoreKey]types.Store,
	keys map[string]types.StoreKey,
) Store {
	cms := Store{
		stores: make(map[types.StoreKey]types.Store, len(stores)),
		keys:   keys,
	}

	for key, store := range stores {
		cms.stores[key] = store.CacheWrap()
	}

	return cms
}

func New(
	stores map[types.StoreKey]types.Store,
	keys map[string]types.StoreKey,
) Store {
	return NewFromStores(stores, keys)
}

func newStoreFromCMS(cms Store) Store {
	stores := make(map[types.StoreKey]types.Store)
	maps.Copy(stores, cms.stores)
	return NewFromStores(stores, nil)
}

// MultiWrite calls Write on each underlying store.
func (cms Store) MultiWrite() {
	for _, store := range cms.stores {
		store.Write()
	}
}

// Implements MultiStore.
func (cms Store) MultiCacheWrap() types.MultiStore {
	return newStoreFromCMS(cms)
}

// GetStore returns an underlying Store by key.
func (cms Store) GetStore(key types.StoreKey) types.Store {
	return cms.stores[key]
}
