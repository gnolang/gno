package immut

import (
	"github.com/gnolang/gno/tm2/pkg/store/cache"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

var (
	_ types.Store          = immutStore{}
	_ types.DepthEstimator = immutStore{}
)

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

// Forward DepthEstimator to parent so cache.New(immutStore) charges depth-based
// gas in simulation (same as deliver). Flat-parent fallback of 100 (depth 1.0)
// equals ReadCostFlat, so non-tree stores are unaffected.
func (is immutStore) ExpectedGetReadDepth100() int64 {
	if de, ok := is.parent.(types.DepthEstimator); ok {
		return de.ExpectedGetReadDepth100()
	}
	return 100
}

func (is immutStore) ExpectedSetReadDepth100() int64 {
	if de, ok := is.parent.(types.DepthEstimator); ok {
		return de.ExpectedSetReadDepth100()
	}
	return 100
}

func (is immutStore) ExpectedWriteDepth100() int64 {
	if de, ok := is.parent.(types.DepthEstimator); ok {
		return de.ExpectedWriteDepth100()
	}
	return 100
}
