package immut

import (
	"github.com/gnolang/gno/tm2/pkg/store/cache"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

var _ types.Store = immutStore{}

type immutStore struct {
	parent types.Store
}

func New(parent types.Store) immutStore {
	return immutStore{
		parent: parent,
	}
}

// Implements Store
func (is immutStore) Get(gctx *types.GasContext, key []byte) []byte {
	return is.parent.Get(gctx, key)
}

// Implements Store
func (is immutStore) Has(gctx *types.GasContext, key []byte) bool {
	return is.parent.Has(gctx, key)
}

// Implements Store
func (is immutStore) Set(gctx *types.GasContext, key, value []byte) {
	panic("unexpected .Set() on immutStore")
}

// Implements Store
func (is immutStore) Delete(gctx *types.GasContext, key []byte) {
	panic("unexpected .Delete() on immutStore")
}

// Implements Store
func (is immutStore) Iterator(gctx *types.GasContext, start, end []byte) types.Iterator {
	return is.parent.Iterator(gctx, start, end)
}

// Implements Store
func (is immutStore) ReverseIterator(gctx *types.GasContext, start, end []byte) types.Iterator {
	return is.parent.ReverseIterator(gctx, start, end)
}

// Implements Store
func (is immutStore) CacheWrap() types.Store {
	return cache.New(is)
}

// Implements Store
func (is immutStore) Write() {
	panic("unexpected .Write() on immutStore")
}
