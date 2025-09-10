package typeutil

import (
	"strings"
	"testing"
	"time"
)

type testStringer struct {
	value string
}

func (t testStringer) String() string {
	return "test:" + t.value
}

func TestToString(t *testing.T) {
	// setup test data
	str := "hello"
	num := 42
	b := true
	now := time.Now()
	addr := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	stringer := testStringer{value: "hello"}

	type testCase struct {
		name     string
		input    any
		expected string
	}

	tests := []testCase{
		// basic types
		{"string", "hello", "hello"},
		{"empty_string", "", ""},
		{"nil", nil, ""},

		// integer types
		{"int", 42, "42"},
		{"int8", int8(8), "8"},
		{"int16", int16(16), "16"},
		{"int32", int32(32), "32"},
		{"int64", int64(64), "64"},
		{"uint", uint(42), "42"},
		{"uint8", uint8(8), "8"},
		{"uint16", uint16(16), "16"},
		{"uint32", uint32(32), "32"},
		{"uint64", uint64(64), "64"},

		// float types
		{"float32", float32(3.14), "3.14"},
		{"float64", 3.14159, "3.14159"},

		// boolean
		{"bool_true", true, "true"},
		{"bool_false", false, "false"},

		// special types
		{"time", now, now.String()},
		{"address", addr, string(addr)},
		{"bytes", []byte("hello"), "hello"},
		{"stringer", stringer, "test:hello"},

		// slices
		{"empty_slice", []string{}, "[]"},
		{"string_slice", []string{"a", "b"}, "[a b]"},
		{"int_slice", []int{1, 2}, "[1 2]"},
		{"int32_slice", []int32{1, 2}, "[1 2]"},
		{"int64_slice", []int64{1, 2}, "[1 2]"},
		{"float32_slice", []float32{1.1, 2.2}, "[1.1 2.2]"},
		{"float64_slice", []float64{1.1, 2.2}, "[1.1 2.2]"},
		{"bytes_slice", [][]byte{[]byte("a"), []byte("b")}, "[a b]"},
		{"time_slice", []time.Time{now, now}, "[" + now.String() + " " + now.String() + "]"},
		{"address_slice", []address{addr, addr}, "[" + string(addr) + " " + string(addr) + "]"},
		{"interface_slice", []any{1, "a", true}, "[1 a true]"},

		// empty slices
		{"empty_string_slice", []string{}, "[]"},
		{"empty_int_slice", []int{}, "[]"},
		{"empty_int32_slice", []int32{}, "[]"},
		{"empty_int64_slice", []int64{}, "[]"},
		{"empty_float32_slice", []float32{}, "[]"},
		{"empty_float64_slice", []float64{}, "[]"},
		{"empty_bytes_slice", [][]byte{}, "[]"},
		{"empty_time_slice", []time.Time{}, "[]"},
		{"empty_address_slice", []address{}, "[]"},
		{"empty_interface_slice", []any{}, "[]"},

		// maps
		{"empty_string_map", map[string]string{}, "map[]"},
		{"string_map", map[string]string{"a": "1", "b": "2"}, "map[a:1 b:2]"},
		{"empty_interface_map", map[string]any{}, "map[]"},
		{"interface_map", map[string]any{"a": 1, "b": "2"}, "map[a:1 b:2]"},

		// edge cases
		{"empty_bytes", []byte{}, ""},
		{"nil_interface", any(nil), ""},
		{"empty_struct", struct{}{}, "{}"},
		{"unknown_type", struct{ foo string }{}, "<unknown>"},

		// pointer types
		{"nil_string_ptr", (*string)(nil), ""},
		{"string_ptr", &str, "hello"},
		{"nil_int_ptr", (*int)(nil), ""},
		{"int_ptr", &num, "42"},
		{"nil_bool_ptr", (*bool)(nil), ""},
		{"bool_ptr", &b, "true"},
		// {"nil_time_ptr", (*time.Time)(nil), ""}, // TODO: fix this
		{"time_ptr", &now, now.String()},
		// {"nil_address_ptr", (*address)(nil), ""}, // TODO: fix this
		{"address_ptr", &addr, string(addr)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToString(tt.input)
			if got != tt.expected {
				t.Errorf("%s: ToString(%v) = %q, want %q", tt.name, tt.input, got, tt.expected)
			}
		})
	}
}

func TestToBool(t *testing.T) {
	str := "true"
	num := 42
	b := true
	now := time.Now()
	addr := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	zero := 0
	empty := ""
	falseVal := false

	type testCase struct {
		name     string
		input    any
		expected bool
	}

	tests := []testCase{
		// basic types
		{"true", true, true},
		{"false", false, false},
		{"nil", nil, false},

		// strings
		{"empty_string", "", false},
		{"zero_string", "0", false},
		{"false_string", "false", false},
		{"f_string", "f", false},
		{"no_string", "no", false},
		{"n_string", "n", false},
		{"off_string", "off", false},
		{"space_string", " ", false},
		{"true_string", "true", true},
		{"yes_string", "yes", true},
		{"random_string", "hello", true},

		// numbers
		{"zero_int", 0, false},
		{"positive_int", 1, true},
		{"negative_int", -1, true},
		{"zero_float", 0.0, false},
		{"positive_float", 0.1, true},
		{"negative_float", -0.1, true},

		// special types
		{"empty_bytes", []byte{}, false},
		{"non_empty_bytes", []byte{1}, true},
		/*{"zero_time", time.Time{}, false},*/ // TODO: fix this
		{"empty_address", address(""), false},

		// slices
		{"empty_slice", []string{}, false},
		{"non_empty_slice", []string{"a"}, true},

		// maps
		{"empty_map", map[string]string{}, false},
		{"non_empty_map", map[string]string{"a": "b"}, true},

		// pointer types
		{"nil_bool_ptr", (*bool)(nil), false},
		{"true_ptr", &b, true},
		{"false_ptr", &falseVal, false},
		{"nil_string_ptr", (*string)(nil), false},
		{"string_ptr", &str, true},
		{"empty_string_ptr", &empty, false},
		{"nil_int_ptr", (*int)(nil), false},
		{"int_ptr", &num, true},
		{"zero_int_ptr", &zero, false},
		// {"nil_time_ptr", (*time.Time)(nil), false}, // TODO: fix this
		{"time_ptr", &now, true},
		// {"nil_address_ptr", (*address)(nil), false}, // TODO: fix this
		{"address_ptr", &addr, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToBool(tt.input)
			if got != tt.expected {
				t.Errorf("%s: ToBool(%v) = %v, want %v", tt.name, tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsZero(t *testing.T) {
	str := "hello"
	num := 42
	b := true
	now := time.Now()
	addr := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	zero := 0
	empty := ""
	falseVal := false

	type testCase struct {
		name     string
		input    any
		expected bool
	}

	tests := []testCase{
		// basic types
		{"true", true, false},
		{"false", false, true},
		{"nil", nil, true},

		// strings
		{"empty_string", "", true},
		{"non_empty_string", "hello", false},

		// numbers
		{"zero_int", 0, true},
		{"non_zero_int", 1, false},
		{"zero_float", 0.0, true},
		{"non_zero_float", 0.1, false},

		// special types
		{"empty_bytes", []byte{}, true},
		{"non_empty_bytes", []byte{1}, false},
		/*{"zero_time", time.Time{}, true},*/ // TODO: fix this
		{"empty_address", address(""), true},

		// slices
		{"empty_slice", []string{}, true},
		{"non_empty_slice", []string{"a"}, false},

		// maps
		{"empty_map", map[string]string{}, true},
		{"non_empty_map", map[string]string{"a": "b"}, false},

		// pointer types
		{"nil_bool_ptr", (*bool)(nil), true},
		{"false_ptr", &falseVal, true},
		{"true_ptr", &b, false},
		{"nil_string_ptr", (*string)(nil), true},
		{"empty_string_ptr", &empty, true},
		{"string_ptr", &str, false},
		{"nil_int_ptr", (*int)(nil), true},
		{"zero_int_ptr", &zero, true},
		{"int_ptr", &num, false},
		// {"nil_time_ptr", (*time.Time)(nil), true}, // TODO: fix this
		{"time_ptr", &now, false},
		// {"nil_address_ptr", (*address)(nil), true}, // TODO: fix this
		{"address_ptr", &addr, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsZero(tt.input)
			if got != tt.expected {
				t.Errorf("%s: IsZero(%v) = %v, want %v", tt.name, tt.input, got, tt.expected)
			}
		})
	}
}

func TestToInterfaceSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected []any
		compare  func([]any, []any) bool
	}{
		{
			name:     "nil",
			input:    nil,
			expected: nil,
			compare:  compareNil,
		},
		{
			name:     "empty_interface_slice",
			input:    []any{},
			expected: []any{},
			compare:  compareEmpty,
		},
		{
			name:     "interface_slice",
			input:    []any{1, "two", true},
			expected: []any{1, "two", true},
			compare:  compareInterfaces,
		},
		{
			name:     "string_slice",
			input:    []string{"a", "b", "c"},
			expected: []any{"a", "b", "c"},
			compare:  compareStrings,
		},
		{
			name:     "int_slice",
			input:    []int{1, 2, 3},
			expected: []any{1, 2, 3},
			compare:  compareInts,
		},
		{
			name:     "int32_slice",
			input:    []int32{1, 2, 3},
			expected: []any{int32(1), int32(2), int32(3)},
			compare:  compareInt32s,
		},
		{
			name:     "int64_slice",
			input:    []int64{1, 2, 3},
			expected: []any{int64(1), int64(2), int64(3)},
			compare:  compareInt64s,
		},
		{
			name:     "float32_slice",
			input:    []float32{1.1, 2.2, 3.3},
			expected: []any{float32(1.1), float32(2.2), float32(3.3)},
			compare:  compareFloat32s,
		},
		{
			name:     "float64_slice",
			input:    []float64{1.1, 2.2, 3.3},
			expected: []any{1.1, 2.2, 3.3},
			compare:  compareFloat64s,
		},
		{
			name:     "bool_slice",
			input:    []bool{true, false, true},
			expected: []any{true, false, true},
			compare:  compareBools,
		},
		/* {
			name:     "time_slice",
			input:    []time.Time{now},
			expected: []any{now},
			compare:  compareTimes,
		}, */ // TODO: fix this
		/* {
			name:     "address_slice",
			input:    []address{addr},
			expected: []any{addr},
			compare:  compareAddresses,
		},*/ // TODO: fix this
		/* {
			name:     "bytes_slice",
			input:    [][]byte{[]byte("hello"), []byte("world")},
			expected: []any{[]byte("hello"), []byte("world")},
			compare:  compareBytes,
		},*/ // TODO: fix this
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToInterfaceSlice(tt.input)
			if !tt.compare(got, tt.expected) {
				t.Errorf("ToInterfaceSlice() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func compareNil(a, b []any) bool {
	return a == nil && b == nil
}

func compareEmpty(a, b []any) bool {
	return len(a) == 0 && len(b) == 0
}

func compareInterfaces(a, b []any) bool {
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

func compareStrings(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		as, ok1 := a[i].(string)
		bs, ok2 := b[i].(string)
		if !ok1 || !ok2 || as != bs {
			return false
		}
	}
	return true
}

func compareInts(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ai, ok1 := a[i].(int)
		bi, ok2 := b[i].(int)
		if !ok1 || !ok2 || ai != bi {
			return false
		}
	}
	return true
}

func compareInt32s(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ai, ok1 := a[i].(int32)
		bi, ok2 := b[i].(int32)
		if !ok1 || !ok2 || ai != bi {
			return false
		}
	}
	return true
}

func compareInt64s(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ai, ok1 := a[i].(int64)
		bi, ok2 := b[i].(int64)
		if !ok1 || !ok2 || ai != bi {
			return false
		}
	}
	return true
}

func compareFloat32s(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ai, ok1 := a[i].(float32)
		bi, ok2 := b[i].(float32)
		if !ok1 || !ok2 || ai != bi {
			return false
		}
	}
	return true
}

func compareFloat64s(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ai, ok1 := a[i].(float64)
		bi, ok2 := b[i].(float64)
		if !ok1 || !ok2 || ai != bi {
			return false
		}
	}
	return true
}

func compareBools(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ab, ok1 := a[i].(bool)
		bb, ok2 := b[i].(bool)
		if !ok1 || !ok2 || ab != bb {
			return false
		}
	}
	return true
}

func compareTimes(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		at, ok1 := a[i].(time.Time)
		bt, ok2 := b[i].(time.Time)
		if !ok1 || !ok2 || !at.Equal(bt) {
			return false
		}
	}
	return true
}

func compareAddresses(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		aa, ok1 := a[i].(address)
		ba, ok2 := b[i].(address)
		if !ok1 || !ok2 || aa != ba {
			return false
		}
	}
	return true
}

func compareBytes(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ab, ok1 := a[i].([]byte)
		bb, ok2 := b[i].([]byte)
		if !ok1 || !ok2 || string(ab) != string(bb) {
			return false
		}
	}
	return true
}

// compareStringInterfaceMaps compares two map[string]any for equality
func compareStringInterfaceMaps(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v1 := range a {
		v2, ok := b[k]
		if !ok {
			return false
		}
		// Compare values based on their type
		switch val1 := v1.(type) {
		case string:
			val2, ok := v2.(string)
			if !ok || val1 != val2 {
				return false
			}
		case int:
			val2, ok := v2.(int)
			if !ok || val1 != val2 {
				return false
			}
		case float64:
			val2, ok := v2.(float64)
			if !ok || val1 != val2 {
				return false
			}
		case bool:
			val2, ok := v2.(bool)
			if !ok || val1 != val2 {
				return false
			}
		case []any:
			val2, ok := v2.([]any)
			if !ok || len(val1) != len(val2) {
				return false
			}
			for i := range val1 {
				if val1[i] != val2[i] {
					return false
				}
			}
		case map[string]any:
			val2, ok := v2.(map[string]any)
			if !ok || !compareStringInterfaceMaps(val1, val2) {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func TestToMapStringInterface(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected map[string]any
		wantErr  bool
	}{
		{
			name: "map[string]any",
			input: map[string]any{
				"key1": "value1",
				"key2": 42,
			},
			expected: map[string]any{
				"key1": "value1",
				"key2": 42,
			},
			wantErr: false,
		},
		{
			name: "map[string]string",
			input: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			expected: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr: false,
		},
		{
			name: "map[string]int",
			input: map[string]int{
				"key1": 1,
				"key2": 2,
			},
			expected: map[string]any{
				"key1": 1,
				"key2": 2,
			},
			wantErr: false,
		},
		{
			name: "map[string]float64",
			input: map[string]float64{
				"key1": 1.1,
				"key2": 2.2,
			},
			expected: map[string]any{
				"key1": 1.1,
				"key2": 2.2,
			},
			wantErr: false,
		},
		{
			name: "map[string]bool",
			input: map[string]bool{
				"key1": true,
				"key2": false,
			},
			expected: map[string]any{
				"key1": true,
				"key2": false,
			},
			wantErr: false,
		},
		{
			name: "map[string][]string",
			input: map[string][]string{
				"key1": {"a", "b"},
				"key2": {"c", "d"},
			},
			expected: map[string]any{
				"key1": []any{"a", "b"},
				"key2": []any{"c", "d"},
			},
			wantErr: false,
		},
		{
			name: "nested map[string]map[string]string",
			input: map[string]map[string]string{
				"key1": {"nested1": "value1"},
				"key2": {"nested2": "value2"},
			},
			expected: map[string]any{
				"key1": map[string]any{"nested1": "value1"},
				"key2": map[string]any{"nested2": "value2"},
			},
			wantErr: false,
		},
		{
			name:     "unsupported type",
			input:    42, // not a map
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToMapStringInterface(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToMapStringInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !compareStringInterfaceMaps(got, tt.expected) {
					t.Errorf("ToMapStringInterface() = %v, expected %v", got, tt.expected)
				}
			}
		})
	}
}

// Test error messages
func TestToMapStringInterfaceErrors(t *testing.T) {
	_, err := ToMapStringInterface(42)
	if err == nil || !strings.Contains(err.Error(), "unsupported map type") {
		t.Errorf("Expected error containing 'unsupported map type', got %v", err)
	}
}

// compareIntInterfaceMaps compares two map[int]any for equality
func compareIntInterfaceMaps(a, b map[int]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v1 := range a {
		v2, ok := b[k]
		if !ok {
			return false
		}
		// Compare values based on their type
		switch val1 := v1.(type) {
		case string:
			val2, ok := v2.(string)
			if !ok || val1 != val2 {
				return false
			}
		case int:
			val2, ok := v2.(int)
			if !ok || val1 != val2 {
				return false
			}
		case float64:
			val2, ok := v2.(float64)
			if !ok || val1 != val2 {
				return false
			}
		case bool:
			val2, ok := v2.(bool)
			if !ok || val1 != val2 {
				return false
			}
		case []any:
			val2, ok := v2.([]any)
			if !ok || len(val1) != len(val2) {
				return false
			}
			for i := range val1 {
				if val1[i] != val2[i] {
					return false
				}
			}
		case map[string]any:
			val2, ok := v2.(map[string]any)
			if !ok || !compareStringInterfaceMaps(val1, val2) {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func TestToMapIntInterface(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected map[int]any
		wantErr  bool
	}{
		{
			name: "map[int]any",
			input: map[int]any{
				1: "value1",
				2: 42,
			},
			expected: map[int]any{
				1: "value1",
				2: 42,
			},
			wantErr: false,
		},
		{
			name: "map[int]string",
			input: map[int]string{
				1: "value1",
				2: "value2",
			},
			expected: map[int]any{
				1: "value1",
				2: "value2",
			},
			wantErr: false,
		},
		{
			name: "map[int]int",
			input: map[int]int{
				1: 10,
				2: 20,
			},
			expected: map[int]any{
				1: 10,
				2: 20,
			},
			wantErr: false,
		},
		{
			name: "map[int]float64",
			input: map[int]float64{
				1: 1.1,
				2: 2.2,
			},
			expected: map[int]any{
				1: 1.1,
				2: 2.2,
			},
			wantErr: false,
		},
		{
			name: "map[int]bool",
			input: map[int]bool{
				1: true,
				2: false,
			},
			expected: map[int]any{
				1: true,
				2: false,
			},
			wantErr: false,
		},
		{
			name: "map[int][]string",
			input: map[int][]string{
				1: {"a", "b"},
				2: {"c", "d"},
			},
			expected: map[int]any{
				1: []any{"a", "b"},
				2: []any{"c", "d"},
			},
			wantErr: false,
		},
		{
			name: "map[int]map[string]any",
			input: map[int]map[string]any{
				1: {"nested1": "value1"},
				2: {"nested2": "value2"},
			},
			expected: map[int]any{
				1: map[string]any{"nested1": "value1"},
				2: map[string]any{"nested2": "value2"},
			},
			wantErr: false,
		},
		{
			name:     "unsupported type",
			input:    42, // not a map
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToMapIntInterface(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToMapIntInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !compareIntInterfaceMaps(got, tt.expected) {
					t.Errorf("ToMapIntInterface() = %v, expected %v", got, tt.expected)
				}
			}
		})
	}
}

func TestToStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected []string
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "string slice",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "int slice",
			input:    []int{1, 2, 3},
			expected: []string{"1", "2", "3"},
		},
		{
			name:     "int32 slice",
			input:    []int32{1, 2, 3},
			expected: []string{"1", "2", "3"},
		},
		{
			name:     "int64 slice",
			input:    []int64{1, 2, 3},
			expected: []string{"1", "2", "3"},
		},
		{
			name:     "uint slice",
			input:    []uint{1, 2, 3},
			expected: []string{"1", "2", "3"},
		},
		{
			name:     "uint8 slice",
			input:    []uint8{1, 2, 3},
			expected: []string{"1", "2", "3"},
		},
		{
			name:     "uint16 slice",
			input:    []uint16{1, 2, 3},
			expected: []string{"1", "2", "3"},
		},
		{
			name:     "uint32 slice",
			input:    []uint32{1, 2, 3},
			expected: []string{"1", "2", "3"},
		},
		{
			name:     "uint64 slice",
			input:    []uint64{1, 2, 3},
			expected: []string{"1", "2", "3"},
		},
		{
			name:     "float32 slice",
			input:    []float32{1.1, 2.2, 3.3},
			expected: []string{"1.1", "2.2", "3.3"},
		},
		{
			name:     "float64 slice",
			input:    []float64{1.1, 2.2, 3.3},
			expected: []string{"1.1", "2.2", "3.3"},
		},
		{
			name:     "bool slice",
			input:    []bool{true, false, true},
			expected: []string{"true", "false", "true"},
		},
		{
			name:     "[]byte slice",
			input:    [][]byte{[]byte("hello"), []byte("world")},
			expected: []string{"hello", "world"},
		},
		{
			name:     "interface slice",
			input:    []any{1, "hello", true},
			expected: []string{"1", "hello", "true"},
		},
		{
			name:     "time slice",
			input:    []time.Time{{}, {}},
			expected: []string{"0001-01-01 00:00:00 +0000 UTC", "0001-01-01 00:00:00 +0000 UTC"},
		},
		{
			name:     "address slice",
			input:    []address{"addr1", "addr2"},
			expected: []string{"addr1", "addr2"},
		},
		{
			name:     "non-slice input",
			input:    42,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToStringSlice(tt.input)
			if !slicesEqual(result, tt.expected) {
				t.Errorf("ToStringSlice(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// Helper function to compare string slices
func slicesEqual(a, b []string) bool {
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

func TestToStringAdvanced(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name: "slice with mixed basic types",
			input: []any{
				42,
				"hello",
				true,
				3.14,
			},
			expected: "[42 hello true 3.14]",
		},
		{
			name: "map with basic types",
			input: map[string]any{
				"int":   42,
				"str":   "hello",
				"bool":  true,
				"float": 3.14,
			},
			expected: "map[bool:true float:3.14 int:42 str:hello]",
		},
		{
			name: "mixed types map",
			input: map[any]any{
				42:         "number",
				"string":   123,
				true:       []int{1, 2, 3},
				struct{}{}: "empty",
			},
			expected: "map[42:number string:123 true:[1 2 3] {}:empty]",
		},
		{
			name: "nested maps",
			input: map[string]any{
				"a": map[string]int{
					"x": 1,
					"y": 2,
				},
				"b": []any{1, "two", true},
			},
			expected: "map[a:map[x:1 y:2] b:[1 two true]]",
		},
		{
			name:     "empty struct",
			input:    struct{}{},
			expected: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToString(tt.input)
			if result != tt.expected {
				t.Errorf("\nToString(%v) =\n%v\nwant:\n%v", tt.input, result, tt.expected)
			}
		})
	}
}
