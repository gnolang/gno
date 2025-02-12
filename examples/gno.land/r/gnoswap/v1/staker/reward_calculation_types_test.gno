package staker

import "testing"

func TestEncodeUint(t *testing.T) {
	tests := []struct {
		input    uint64
		expected string
	}{
		{0, "00000000000000000000"},                    // minimum
		{12345, "00000000000000012345"},                // normal value
		{18446744073709551615, "18446744073709551615"}, // maximum (uint64 max)
	}

	for _, tt := range tests {
		result := EncodeUint(tt.input)
		if result != tt.expected {
			t.Errorf("EncodeUint(%d) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestDecodeUint(t *testing.T) {
	tests := []struct {
		input       string
		expected    uint64
		shouldPanic bool
	}{
		{"00000000000000000000", 0, false},
		{"00000000000000012345", 12345, false},
		{"18446744073709551615", 18446744073709551615, false},
		{"invalid", 0, true},
		{"18446744073709551616", 0, true},
	}

	for _, tt := range tests {
		if tt.shouldPanic {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("DecodeUint(%s) did not panic as expected", tt.input)
				}
			}()
			_ = DecodeUint(tt.input)
		} else {
			result := DecodeUint(tt.input)
			if result != tt.expected {
				t.Errorf("DecodeUint(%s) = %d; want %d", tt.input, result, tt.expected)
			}
		}
	}
}
func TestUintTree(t *testing.T) {
	tree := NewUintTree()

	// Test Set and Get
	tree.Set(12345, "testValue")
	value, ok := tree.Get(12345)
	if !ok || value != "testValue" {
		t.Errorf("UintTree.Get(12345) = %v, %v; want testValue, true", value, ok)
	}

	// Test Has
	if !tree.Has(12345) {
		t.Errorf("UintTree.Has(12345) = false; want true")
	}

	// Test Remove
	tree.Remove(12345)
	if tree.Has(12345) {
		t.Errorf("UintTree.Has(12345) after Remove = true; want false")
	}

	// Test Iterate
	tree.Set(100, "a")
	tree.Set(200, "b")
	tree.Set(300, "c")

	var keys []uint64
	var values []interface{}

	tree.Iterate(100, 300, func(key int64, value interface{}) bool {
		keys = append(keys, uint64(key))
		values = append(values, value)
		return false
	})

	// Verify results
	expectedKeys := []uint64{100, 200}
	expectedValues := []interface{}{"a", "b"}

	if !compareUintSlices(t, keys, expectedKeys) || !compareInterfaces(t, values, expectedValues) {
		t.Errorf("UintTree.Iterate() keys = %v, values = %v; want keys = %v, values = %v", keys, values, expectedKeys, expectedValues)
	}
}

// Helper function to compare slices of uint64
func compareUintSlices(t *testing.T, a, b []uint64) bool {
	t.Helper()
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Helper function to compare slices of interfaces
func compareInterfaces(t *testing.T, a, b []interface{}) bool {
	t.Helper()
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
