package cachemulti

import (
	dbm "github.com/gnolang/gno/pkgs/db"

	"github.com/gnolang/gno/pkgs/store/cache"
	"github.com/gnolang/gno/pkgs/store/dbadapter"
	"github.com/gnolang/gno/pkgs/store/types"
)

//----------------------------------------
// Store

// Store holds many cache-wrapped stores.
// Implements MultiStore.
// NOTE: a Store (and MultiStores in general) should never expose the
// keys for the substores.
type Store struct {
	main   types.Store
	stores map[types.StoreKey]types.Store
	keys   map[string]types.StoreKey
}

var _ types.MultiStore = Store{}

func NewFromStores(
	main types.Store,
	stores map[types.StoreKey]types.Store,
	keys map[string]types.StoreKey,
) Store {
	cms := Store{
		main:   cache.New(main),
		stores: make(map[types.StoreKey]types.Store, len(stores)),
		keys:   keys,
	}

	for key, store := range stores {
		cms.stores[key] = store.CacheWrap()
	}

	return cms
}

func New(
	db dbm.DB,
	stores map[types.StoreKey]types.Store,
	keys map[string]types.StoreKey,
) Store {
	return NewFromStores(dbadapter.Store{db}, stores, keys)
}

func newStoreFromCMS(cms Store) Store {
	stores := make(map[types.StoreKey]types.Store)
	for k, v := range cms.stores {
		stores[k] = v
	}
	return NewFromStores(cms.main, stores, nil)
}

// MultiWrite calls Write on each underlying store.
func (cms Store) MultiWrite() {
	cms.main.Write()
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
	return cms.stores[key].(types.Store)
}
