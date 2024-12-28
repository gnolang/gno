// Package memo provides a simple memoization utility to cache function results.
//
// The package offers a Memoizer type that can cache function results based on keys,
// with optional validation of cached values. This is useful for expensive computations
// that need to be cached and potentially invalidated based on custom conditions.
//
// ⚠️ Important Warning for Gno Usage:
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
//	result := m.Memoize("key", func() interface{} {
//	    // expensive operation
//	    return "computed-value"
//	})
//
//	// Subsequent calls with same key return cached result
//	result = m.Memoize("key", func() interface{} {
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
//	    func() interface{} {
//	        return TimestampedValue{
//	            Value:     "data",
//	            Timestamp: time.Now(),
//	        }
//	    },
//	    func(cached interface{}) bool {
//	        // Validate that the cached value is not older than 1 hour
//	        if tv, ok := cached.(TimestampedValue); ok {
//	            return time.Since(tv.Timestamp) < time.Hour
//	        }
//	        return false
//	    },
//	)
package memo

// Memoizer is a structure to handle memoization of function results.
type Memoizer struct {
	cache map[interface{}]interface{}
}

// New creates a new Memoizer instance.
func New() *Memoizer {
	return &Memoizer{
		cache: make(map[interface{}]interface{}),
	}
}

// Memoize ensures the result of the given function is cached for the specified key.
func (m *Memoizer) Memoize(key interface{}, fn func() interface{}) interface{} {
	if cached, exists := m.cache[key]; exists {
		return cached
	}

	value := fn()
	m.cache[key] = value
	return value
}

// MemoizeWithValidator ensures the result is cached and valid according to the validator function.
func (m *Memoizer) MemoizeWithValidator(key interface{}, fn func() interface{}, isValid func(interface{}) bool) interface{} {
	if cached, exists := m.cache[key]; exists && isValid(cached) {
		return cached
	}

	value := fn()
	m.cache[key] = value
	return value
}

// Invalidate removes the cached value for the specified key.
func (m *Memoizer) Invalidate(key interface{}) {
	delete(m.cache, key)
}

// Clear clears all cached values.
func (m *Memoizer) Clear() {
	m.cache = make(map[interface{}]interface{})
}
