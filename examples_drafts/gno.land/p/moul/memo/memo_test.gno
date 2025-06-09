package memo

import (
	"testing"
	"time"
)

type timestampedValue struct {
	value     any
	timestamp time.Time
}

// complexKey is used to test struct keys
type complexKey struct {
	ID   int
	Name string
}

func TestMemoize(t *testing.T) {
	tests := []struct {
		name      string
		key       any
		value     any
		callCount *int
	}{
		{
			name:      "string key and value",
			key:       "test-key",
			value:     "test-value",
			callCount: new(int),
		},
		{
			name:      "int key and value",
			key:       42,
			value:     123,
			callCount: new(int),
		},
		{
			name:      "mixed types",
			key:       "number",
			value:     42,
			callCount: new(int),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			if m.Size() != 0 {
				t.Errorf("Initial size = %d, want 0", m.Size())
			}

			fn := func() any {
				*tt.callCount++
				return tt.value
			}

			// First call should compute
			result := m.Memoize(tt.key, fn)
			if result != tt.value {
				t.Errorf("Memoize() = %v, want %v", result, tt.value)
			}
			if *tt.callCount != 1 {
				t.Errorf("Function called %d times, want 1", *tt.callCount)
			}
			if m.Size() != 1 {
				t.Errorf("Size after first call = %d, want 1", m.Size())
			}

			// Second call should use cache
			result = m.Memoize(tt.key, fn)
			if result != tt.value {
				t.Errorf("Memoize() second call = %v, want %v", result, tt.value)
			}
			if *tt.callCount != 1 {
				t.Errorf("Function called %d times, want 1", *tt.callCount)
			}
			if m.Size() != 1 {
				t.Errorf("Size after second call = %d, want 1", m.Size())
			}
		})
	}
}

func TestMemoizeWithValidator(t *testing.T) {
	tests := []struct {
		name            string
		key             any
		value           any
		validDuration   time.Duration
		waitDuration    time.Duration
		expectedCalls   int
		shouldRecompute bool
	}{
		{
			name:            "valid cache",
			key:             "key1",
			value:           "value1",
			validDuration:   time.Hour,
			waitDuration:    time.Millisecond,
			expectedCalls:   1,
			shouldRecompute: false,
		},
		{
			name:            "expired cache",
			key:             "key2",
			value:           "value2",
			validDuration:   time.Millisecond,
			waitDuration:    time.Millisecond * 2,
			expectedCalls:   2,
			shouldRecompute: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			callCount := 0

			fn := func() any {
				callCount++
				return timestampedValue{
					value:     tt.value,
					timestamp: time.Now(),
				}
			}

			isValid := func(cached any) bool {
				if tv, ok := cached.(timestampedValue); ok {
					return time.Since(tv.timestamp) < tt.validDuration
				}
				return false
			}

			// First call
			result := m.MemoizeWithValidator(tt.key, fn, isValid)
			if tv, ok := result.(timestampedValue); !ok || tv.value != tt.value {
				t.Errorf("MemoizeWithValidator() = %v, want value %v", result, tt.value)
			}

			// Wait
			testing.SkipHeights(10)

			// Second call
			result = m.MemoizeWithValidator(tt.key, fn, isValid)
			if tv, ok := result.(timestampedValue); !ok || tv.value != tt.value {
				t.Errorf("MemoizeWithValidator() second call = %v, want value %v", result, tt.value)
			}

			if callCount != tt.expectedCalls {
				t.Errorf("Function called %d times, want %d", callCount, tt.expectedCalls)
			}
		})
	}
}

func TestInvalidate(t *testing.T) {
	tests := []struct {
		name      string
		key       any
		value     any
		callCount *int
	}{
		{
			name:      "invalidate existing key",
			key:       "test-key",
			value:     "test-value",
			callCount: new(int),
		},
		{
			name:      "invalidate non-existing key",
			key:       "missing-key",
			value:     "test-value",
			callCount: new(int),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			fn := func() any {
				*tt.callCount++
				return tt.value
			}

			// First call
			m.Memoize(tt.key, fn)
			if m.Size() != 1 {
				t.Errorf("Size after first call = %d, want 1", m.Size())
			}

			// Invalidate
			m.Invalidate(tt.key)
			if m.Size() != 0 {
				t.Errorf("Size after invalidate = %d, want 0", m.Size())
			}

			// Call again should recompute
			result := m.Memoize(tt.key, fn)
			if result != tt.value {
				t.Errorf("Memoize() after invalidate = %v, want %v", result, tt.value)
			}
			if *tt.callCount != 2 {
				t.Errorf("Function called %d times, want 2", *tt.callCount)
			}
			if m.Size() != 1 {
				t.Errorf("Size after recompute = %d, want 1", m.Size())
			}
		})
	}
}

func TestClear(t *testing.T) {
	m := New()
	callCount := 0

	fn := func() any {
		callCount++
		return "value"
	}

	// Cache some values
	m.Memoize("key1", fn)
	m.Memoize("key2", fn)

	if callCount != 2 {
		t.Errorf("Initial calls = %d, want 2", callCount)
	}
	if m.Size() != 2 {
		t.Errorf("Size after initial calls = %d, want 2", m.Size())
	}

	// Clear cache
	m.Clear()
	if m.Size() != 0 {
		t.Errorf("Size after clear = %d, want 0", m.Size())
	}

	// Recompute values
	m.Memoize("key1", fn)
	m.Memoize("key2", fn)

	if callCount != 4 {
		t.Errorf("Calls after clear = %d, want 4", callCount)
	}
	if m.Size() != 2 {
		t.Errorf("Size after recompute = %d, want 2", m.Size())
	}
}

func TestSize(t *testing.T) {
	m := New()

	if m.Size() != 0 {
		t.Errorf("Initial size = %d, want 0", m.Size())
	}

	callCount := 0
	fn := func() any {
		callCount++
		return "value"
	}

	// Add items
	m.Memoize("key1", fn)
	if m.Size() != 1 {
		t.Errorf("Size after first insert = %d, want 1", m.Size())
	}

	m.Memoize("key2", fn)
	if m.Size() != 2 {
		t.Errorf("Size after second insert = %d, want 2", m.Size())
	}

	// Duplicate key should not increase size
	m.Memoize("key1", fn)
	if m.Size() != 2 {
		t.Errorf("Size after duplicate insert = %d, want 2", m.Size())
	}

	// Remove item
	m.Invalidate("key1")
	if m.Size() != 1 {
		t.Errorf("Size after invalidate = %d, want 1", m.Size())
	}

	// Clear all
	m.Clear()
	if m.Size() != 0 {
		t.Errorf("Size after clear = %d, want 0", m.Size())
	}
}

func TestMemoizeWithDifferentKeyTypes(t *testing.T) {
	tests := []struct {
		name      string
		keys      []any    // Now an array of keys
		values    []string // Corresponding values
		callCount *int
	}{
		{
			name:      "integer keys",
			keys:      []any{42, 43},
			values:    []string{"value-for-42", "value-for-43"},
			callCount: new(int),
		},
		{
			name:      "float keys",
			keys:      []any{3.14, 2.718},
			values:    []string{"value-for-pi", "value-for-e"},
			callCount: new(int),
		},
		{
			name:      "bool keys",
			keys:      []any{true, false},
			values:    []string{"value-for-true", "value-for-false"},
			callCount: new(int),
		},
		/*
			{
				name: "struct keys",
				keys: []any{
					complexKey{ID: 1, Name: "test1"},
					complexKey{ID: 2, Name: "test2"},
				},
				values:    []string{"value-for-struct1", "value-for-struct2"},
				callCount: new(int),
			},
			{
				name:      "nil and empty interface keys",
				keys:      []any{nil, any(nil)},
				values:    []string{"value-for-nil", "value-for-empty-interface"},
				callCount: new(int),
			},
		*/
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()

			// Test both keys
			for i, key := range tt.keys {
				value := tt.values[i]
				fn := func() any {
					*tt.callCount++
					return value
				}

				// First call should compute
				result := m.Memoize(key, fn)
				if result != value {
					t.Errorf("Memoize() for key %v = %v, want %v", key, result, value)
				}
				if *tt.callCount != i+1 {
					t.Errorf("Function called %d times, want %d", *tt.callCount, i+1)
				}
			}

			// Verify size includes both entries
			if m.Size() != 2 {
				t.Errorf("Size after both inserts = %d, want 2", m.Size())
			}

			// Second call for each key should use cache
			for i, key := range tt.keys {
				initialCount := *tt.callCount
				result := m.Memoize(key, func() any {
					*tt.callCount++
					return "should-not-be-called"
				})

				if result != tt.values[i] {
					t.Errorf("Memoize() second call for key %v = %v, want %v", key, result, tt.values[i])
				}
				if *tt.callCount != initialCount {
					t.Errorf("Cache miss for key %v", key)
				}
			}

			// Test invalidate for each key
			for i, key := range tt.keys {
				m.Invalidate(key)
				if m.Size() != 1-i {
					t.Errorf("Size after invalidate %d = %d, want %d", i+1, m.Size(), 1-i)
				}
			}
		})
	}
}

func TestMultipleKeyTypes(t *testing.T) {
	m := New()
	callCount := 0

	// Insert different key types simultaneously (two of each type)
	keys := []any{
		42, 43, // ints
		"string-key1", "string-key2", // strings
		3.14, 2.718, // floats
		true, false, // bools
	}

	for i, key := range keys {
		value := i
		m.Memoize(key, func() any {
			callCount++
			return value
		})
	}

	// Verify size includes all entries
	if m.Size() != len(keys) {
		t.Errorf("Size = %d, want %d", m.Size(), len(keys))
	}

	// Verify all values are cached correctly
	for i, key := range keys {
		initialCount := callCount
		result := m.Memoize(key, func() any {
			callCount++
			return -1 // Should never be returned if cache works
		})

		if result != i {
			t.Errorf("Memoize(%v) = %v, want %v", key, result, i)
		}
		if callCount != initialCount {
			t.Errorf("Cache miss for key %v", key)
		}
	}

	// Test invalidation of pairs
	for i := 0; i < len(keys); i += 2 {
		m.Invalidate(keys[i])
		m.Invalidate(keys[i+1])
		expectedSize := len(keys) - (i + 2)
		if m.Size() != expectedSize {
			t.Errorf("Size after invalidating pair %d = %d, want %d", i/2, m.Size(), expectedSize)
		}
	}

	// Clear and verify
	m.Clear()
	if m.Size() != 0 {
		t.Errorf("Size after clear = %d, want 0", m.Size())
	}
}
