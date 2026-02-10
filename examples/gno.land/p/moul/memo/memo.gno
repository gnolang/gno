// Package memo provides a simple memoization utility to cache function results.
//
// The package offers a Memoizer type that can cache function results based on keys,
// with optional validation of cached values. This is useful for expensive computations
// that need to be cached and potentially invalidated based on custom conditions.
//
// /!\ Important Warning for Gno Usage:
// In Gno, storage updates only persist during transactions. This means:
//   - Cache entries created during queries will NOT persist
//   - Creating cache entries during queries will actually decrease performance
//     as it wastes resources trying to save data that won't be saved
//
// Best Practices:
// - Use this pattern in transaction-driven contexts rather than query/render scenarios
// - Consider controlled cache updates, e.g., by specific accounts (like oracles)
// - Ideal for cases where cache updates happen every N blocks or on specific events
// - Carefully evaluate if caching will actually improve performance in your use case
//
// Basic usage example:
//
//	m := memo.New()
//
//	// Cache expensive computation
//	result := m.Memoize("key", func() any {
//	    // expensive operation
//	    return "computed-value"
//	})
//
//	// Subsequent calls with same key return cached result
//	result = m.Memoize("key", func() any {
//	    // function won't be called, cached value is returned
//	    return "computed-value"
//	})
//
// Example with validation:
//
//	type TimestampedValue struct {
//	    Value     string
//	    Timestamp time.Time
//	}
//
//	m := memo.New()
//
//	// Cache value with timestamp
//	result := m.MemoizeWithValidator(
//	    "key",
//	    func() any {
//	        return TimestampedValue{
//	            Value:     "data",
//	            Timestamp: time.Now(),
//	        }
//	    },
//	    func(cached any) bool {
//	        // Validate that the cached value is not older than 1 hour
//	        if tv, ok := cached.(TimestampedValue); ok {
//	            return time.Since(tv.Timestamp) < time.Hour
//	        }
//	        return false
//	    },
//	)
package memo

import (
	"gno.land/p/demo/btree"
	"gno.land/p/nt/ufmt"
)

// Record implements the btree.Record interface for our cache entries
type cacheEntry struct {
	key   any
	value any
}

// Less implements btree.Record interface
func (e cacheEntry) Less(than btree.Record) bool {
	// Convert the other record to cacheEntry
	other := than.(cacheEntry)
	// Compare string representations of keys for consistent ordering
	return ufmt.Sprintf("%v", e.key) < ufmt.Sprintf("%v", other.key)
}

// Memoizer is a structure to handle memoization of function results.
type Memoizer struct {
	cache *btree.BTree
}

// New creates a new Memoizer instance.
func New() *Memoizer {
	return &Memoizer{
		cache: btree.New(),
	}
}

// Memoize ensures the result of the given function is cached for the specified key.
func (m *Memoizer) Memoize(key any, fn func() any) any {
	entry := cacheEntry{key: key}
	if found := m.cache.Get(entry); found != nil {
		return found.(cacheEntry).value
	}

	value := fn()
	m.cache.Insert(cacheEntry{key: key, value: value})
	return value
}

// MemoizeWithValidator ensures the result is cached and valid according to the validator function.
func (m *Memoizer) MemoizeWithValidator(key any, fn func() any, isValid func(any) bool) any {
	entry := cacheEntry{key: key}
	if found := m.cache.Get(entry); found != nil {
		cachedEntry := found.(cacheEntry)
		if isValid(cachedEntry.value) {
			return cachedEntry.value
		}
	}

	value := fn()
	m.cache.Insert(cacheEntry{key: key, value: value})
	return value
}

// Invalidate removes the cached value for the specified key.
func (m *Memoizer) Invalidate(key any) {
	m.cache.Delete(cacheEntry{key: key})
}

// Clear clears all cached values.
func (m *Memoizer) Clear() {
	m.cache.Clear(true)
}

// Size returns the number of items currently in the cache.
func (m *Memoizer) Size() int {
	return m.cache.Len()
}
