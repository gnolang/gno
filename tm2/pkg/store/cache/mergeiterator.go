package cache

import (
	"bytes"

	"github.com/gnolang/gno/tm2/ordering"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// cacheMergeIterator merges a parent Iterator and a cache Iterator.
// The cache iterator may return nil keys to signal that an item
// had been deleted (but not deleted in the parent).
// If the cache iterator has the same key as the parent, the
// cache shadows (overrides) the parent.
//
// TODO: Optimize by memoizing.
type cacheMergeIterator struct {
	parent    types.Iterator
	cache     types.Iterator
	ascending bool
}

var _ types.Iterator = (*cacheMergeIterator)(nil)

func newCacheMergeIterator(parent, cache types.Iterator, ascending bool) *cacheMergeIterator {
	iter := &cacheMergeIterator{
		parent:    parent,
		cache:     cache,
		ascending: ascending,
	}
	return iter
}

// Domain implements Iterator.
// If the domains are different, returns the union.
func (iter *cacheMergeIterator) Domain() (start, end []byte) {
	startP, endP := iter.parent.Domain()
	startC, endC := iter.cache.Domain()
	if iter.compare(startP, startC).IsLess() {
		start = startP
	} else {
		start = startC
	}
	if iter.compare(endP, endC).IsLess() {
		end = endC
	} else {
		end = endP
	}
	return start, end
}

// Valid implements Iterator.
func (iter *cacheMergeIterator) Valid() bool {
	return iter.skipUntilExistsOrInvalid()
}

// Next implements Iterator
func (iter *cacheMergeIterator) Next() {
	iter.skipUntilExistsOrInvalid()
	iter.assertValid()

	// If parent is invalid, get the next cache item.
	if !iter.parent.Valid() {
		iter.cache.Next()
		return
	}

	// If cache is invalid, get the next parent item.
	if !iter.cache.Valid() {
		iter.parent.Next()
		return
	}

	// Both are valid.  Compare keys.
	keyP, keyC := iter.parent.Key(), iter.cache.Key()
	switch iter.compare(keyP, keyC) {
	case ordering.Less: // parent < cache
		iter.parent.Next()
	case ordering.Equal: // parent == cache
		iter.parent.Next()
		iter.cache.Next()
	case ordering.Greater: // parent > cache
		iter.cache.Next()
	}
}

// Key implements Iterator
func (iter *cacheMergeIterator) Key() []byte {
	iter.skipUntilExistsOrInvalid()
	iter.assertValid()

	// If parent is invalid, get the cache key.
	if !iter.parent.Valid() {
		return iter.cache.Key()
	}

	// If cache is invalid, get the parent key.
	if !iter.cache.Valid() {
		return iter.parent.Key()
	}

	// Both are valid.  Compare keys.
	keyP, keyC := iter.parent.Key(), iter.cache.Key()
	cmp := iter.compare(keyP, keyC)
	switch cmp {
	case ordering.Less: // parent < cache
		return keyP
	case ordering.Equal: // parent == cache
		return keyP
	case ordering.Greater: // parent > cache
		return keyC
	default:
		panic("invalid compare result")
	}
}

// Value implements Iterator
func (iter *cacheMergeIterator) Value() []byte {
	iter.skipUntilExistsOrInvalid()
	iter.assertValid()

	// If parent is invalid, get the cache value.
	if !iter.parent.Valid() {
		return iter.cache.Value()
	}

	// If cache is invalid, get the parent value.
	if !iter.cache.Valid() {
		return iter.parent.Value()
	}

	// Both are valid.  Compare keys.
	keyP, keyC := iter.parent.Key(), iter.cache.Key()
	cmp := iter.compare(keyP, keyC)
	switch cmp {
	case ordering.Less: // parent < cache
		return iter.parent.Value()
	case ordering.Equal: // parent == cache
		return iter.cache.Value()
	case ordering.Greater: // parent > cache
		return iter.cache.Value()
	default:
		panic("invalid comparison result")
	}
}

// Close implements Iterator
func (iter *cacheMergeIterator) Close() {
	iter.parent.Close()
	iter.cache.Close()
}

// Like bytes.Compare but opposite if not ascending.
func (iter *cacheMergeIterator) compare(a, b []byte) ordering.Ordering {
	if iter.ascending {
		return ordering.NewOrdering(ordering.Order(bytes.Compare(a, b)))
	}
	return ordering.NewOrdering(ordering.Order(bytes.Compare(a, b) * -1))
}

// Skip all delete-items from the cache w/ `key < until`.  After this function,
// current cache item is a non-delete-item, or `until <= key`.
// If the current cache item is not a delete item, does nothing.
// If `until` is nil, there is no limit, and cache may end up invalid.
// CONTRACT: cache is valid.
func (iter *cacheMergeIterator) skipCacheDeletes(until []byte) {
	for iter.cache.Valid() &&
		iter.cache.Value() == nil &&
		(until == nil || iter.compare(iter.cache.Key(), until).IsLess()) {
		iter.cache.Next()
	}
}

// Fast forwards cache (or parent+cache in case of deleted items) until current
// item exists, or until iterator becomes invalid.
// Returns whether the iterator is valid.
func (iter *cacheMergeIterator) skipUntilExistsOrInvalid() bool {
	for {
		// If parent is invalid, fast-forward cache.
		if !iter.parent.Valid() {
			iter.skipCacheDeletes(nil)
			return iter.cache.Valid()
		}
		// Parent is valid.

		if !iter.cache.Valid() {
			return true
		}
		// Parent is valid, cache is valid.

		// Compare parent and cache.
		keyP := iter.parent.Key()
		keyC := iter.cache.Key()
		switch iter.compare(keyP, keyC) {
		case ordering.Less: // parent < cache.
			return true

		case ordering.Equal: // parent == cache.

			// Skip over if cache item is a delete.
			valueC := iter.cache.Value()
			if valueC == nil {
				iter.parent.Next()
				iter.cache.Next()
				continue
			}
			// Cache is not a delete.

			return true // cache exists.

		case ordering.Greater: // cache < parent

			// Skip over if cache item is a delete.
			valueC := iter.cache.Value()
			if valueC == nil {
				iter.skipCacheDeletes(keyP)
				continue
			}
			// Cache is not a delete.

			return true // cache exists.
		}
	}
}

// If not valid, panics.
// NOTE: May have side-effect of iterating over cache.
func (iter *cacheMergeIterator) assertValid() {
	if !iter.Valid() {
		panic("iterator is invalid")
	}
}
