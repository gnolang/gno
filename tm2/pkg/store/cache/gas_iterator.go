package cache

import (
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// gasIterator wraps a types.Iterator and charges gas for seek (once
// on creation) and per-step (including per-byte on the value returned
// by the current position). Wrapping happens only at the cache.Store
// layer; parent stores (iavl, dbadapter) return un-wrapped iterators
// and rely on the cache wrapper above them to charge.
type gasIterator struct {
	types.Iterator
	gctx *types.GasContext
}

// newGasIterator charges the seek cost immediately and, if the
// iterator is already positioned on a valid item, charges the first
// step. Returns the input iterator unchanged when gctx is nil so
// non-gas-metered callers pay no overhead.
func newGasIterator(gctx *types.GasContext, it types.Iterator) types.Iterator {
	if gctx == nil {
		return it
	}
	gctx.WillIterator()
	if it.Valid() {
		gctx.WillIterNext(it.Value())
	}
	return &gasIterator{Iterator: it, gctx: gctx}
}

// Next advances and charges for the new position (if valid). The
// charge fires eagerly on advance, not lazily on Value(), matching
// the physical DB cost: the backend has already fetched the page.
func (gi *gasIterator) Next() {
	gi.Iterator.Next()
	if gi.Iterator.Valid() {
		gi.gctx.WillIterNext(gi.Iterator.Value())
	}
}
