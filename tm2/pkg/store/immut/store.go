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
func (is immutStore) Get(key []byte) []byte {
	return is.parent.Get(key)
}

// Implements Store
func (is immutStore) Has(key []byte) bool {
	return is.parent.Has(key)
}

// Implements Store
func (is immutStore) Set(key, value []byte) {
	panic("unexpected .Set() on immutStore")
}

// Implements Store
func (is immutStore) Delete(key []byte) {
	panic("unexpected .Delete() on immutStore")
}

// Implements Store
func (is immutStore) Iterator(start, end []byte) types.Iterator {
	return is.parent.Iterator(start, end)
}

// Implements Store
func (is immutStore) ReverseIterator(start, end []byte) types.Iterator {
	return is.parent.ReverseIterator(start, end)
}

// Implements Store
func (is immutStore) CacheWrap() types.Store {
	return cache.New(is)
}

// Implements Store
func (is immutStore) Write() {
	panic("unexpected .Write() on immutStore")
}
