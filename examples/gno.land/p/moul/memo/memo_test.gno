package memo

import (
	"std"
	"testing"
	"time"
)

type timestampedValue struct {
	value     interface{}
	timestamp time.Time
}

func TestMemoize(t *testing.T) {
	tests := []struct {
		name      string
		key       interface{}
		value     interface{}
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
			fn := func() interface{} {
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

			// Second call should use cache
			result = m.Memoize(tt.key, fn)
			if result != tt.value {
				t.Errorf("Memoize() second call = %v, want %v", result, tt.value)
			}
			if *tt.callCount != 1 {
				t.Errorf("Function called %d times, want 1", *tt.callCount)
			}
		})
	}
}

func TestMemoizeWithValidator(t *testing.T) {
	tests := []struct {
		name            string
		key             interface{}
		value           interface{}
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

			fn := func() interface{} {
				callCount++
				return timestampedValue{
					value:     tt.value,
					timestamp: time.Now(),
				}
			}

			isValid := func(cached interface{}) bool {
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
			std.TestSkipHeights(10)

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
		key       interface{}
		value     interface{}
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
			fn := func() interface{} {
				*tt.callCount++
				return tt.value
			}

			// First call
			m.Memoize(tt.key, fn)

			// Invalidate
			m.Invalidate(tt.key)

			// Call again should recompute
			result := m.Memoize(tt.key, fn)
			if result != tt.value {
				t.Errorf("Memoize() after invalidate = %v, want %v", result, tt.value)
			}
			if *tt.callCount != 2 {
				t.Errorf("Function called %d times, want 2", *tt.callCount)
			}
		})
	}
}

func TestClear(t *testing.T) {
	m := New()
	callCount := 0

	fn := func() interface{} {
		callCount++
		return "value"
	}

	// Cache some values
	m.Memoize("key1", fn)
	m.Memoize("key2", fn)

	if callCount != 2 {
		t.Errorf("Initial calls = %d, want 2", callCount)
	}

	// Clear cache
	m.Clear()

	// Recompute values
	m.Memoize("key1", fn)
	m.Memoize("key2", fn)

	if callCount != 4 {
		t.Errorf("Calls after clear = %d, want 4", callCount)
	}
}
