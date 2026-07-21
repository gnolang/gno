package immut

import (
	"github.com/gnolang/gno/tm2/pkg/store/cache"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

var _ types.Store = immutStore{}

type immutStore struct {
	parent types.Store
}

// immutStoreDE wraps a depth-estimating parent (e.g. IAVL). Flat stores (e.g.
// dbadapter) use plain immutStore so FixedGetReadDepth100 in VM gas contexts
// does not override their 1× ReadCostFlat rate during simulation.
type immutStoreDE struct {
	immutStore
	de types.DepthEstimator
}

var (
	_ types.Store          = immutStoreDE{}
	_ types.DepthEstimator = immutStoreDE{}
)

// New wraps parent as immutable, forwarding DepthEstimator only if parent has it.
func New(parent types.Store) types.Store {
	is := immutStore{parent: parent}
	if de, ok := parent.(types.DepthEstimator); ok {
		return immutStoreDE{immutStore: is, de: de}
	}
	return is
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

func (i immutStoreDE) CacheWrap() types.Store {
	return cache.New(i)
}

func (i immutStoreDE) ExpectedGetReadDepth100() int64 { return i.de.ExpectedGetReadDepth100() }
func (i immutStoreDE) ExpectedSetReadDepth100() int64 { return i.de.ExpectedSetReadDepth100() }
func (i immutStoreDE) ExpectedWriteDepth100() int64   { return i.de.ExpectedWriteDepth100() }
