package fp

import (
	"fmt"
	"testing"
)

func TestMap(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		fn       func(interface{}) interface{}
		expected []interface{}
	}{
		{
			name:     "multiply numbers by 2",
			input:    []interface{}{1, 2, 3},
			fn:       func(v interface{}) interface{} { return v.(int) * 2 },
			expected: []interface{}{2, 4, 6},
		},
		{
			name:     "empty slice",
			input:    []interface{}{},
			fn:       func(v interface{}) interface{} { return v.(int) * 2 },
			expected: []interface{}{},
		},
		{
			name:     "convert numbers to strings",
			input:    []interface{}{1, 2, 3},
			fn:       func(v interface{}) interface{} { return fmt.Sprintf("%d", v.(int)) },
			expected: []interface{}{"1", "2", "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Map(tt.input, tt.fn)
			if !equalSlices(result, tt.expected) {
				t.Errorf("Map failed, expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		fn       func(interface{}) bool
		expected []interface{}
	}{
		{
			name:     "filter even numbers",
			input:    []interface{}{1, 2, 3, 4},
			fn:       func(v interface{}) bool { return v.(int)%2 == 0 },
			expected: []interface{}{2, 4},
		},
		{
			name:     "empty slice",
			input:    []interface{}{},
			fn:       func(v interface{}) bool { return v.(int)%2 == 0 },
			expected: []interface{}{},
		},
		{
			name:     "no matches",
			input:    []interface{}{1, 3, 5},
			fn:       func(v interface{}) bool { return v.(int)%2 == 0 },
			expected: []interface{}{},
		},
		{
			name:     "all matches",
			input:    []interface{}{2, 4, 6},
			fn:       func(v interface{}) bool { return v.(int)%2 == 0 },
			expected: []interface{}{2, 4, 6},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Filter(tt.input, tt.fn)
			if !equalSlices(result, tt.expected) {
				t.Errorf("Filter failed, expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestReduce(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		fn       func(interface{}, interface{}) interface{}
		initial  interface{}
		expected interface{}
	}{
		{
			name:     "sum numbers",
			input:    []interface{}{1, 2, 3},
			fn:       func(a, b interface{}) interface{} { return a.(int) + b.(int) },
			initial:  0,
			expected: 6,
		},
		{
			name:     "empty slice",
			input:    []interface{}{},
			fn:       func(a, b interface{}) interface{} { return a.(int) + b.(int) },
			initial:  0,
			expected: 0,
		},
		{
			name:     "concatenate strings",
			input:    []interface{}{"a", "b", "c"},
			fn:       func(a, b interface{}) interface{} { return a.(string) + b.(string) },
			initial:  "",
			expected: "abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Reduce(tt.input, tt.fn, tt.initial)
			if result != tt.expected {
				t.Errorf("Reduce failed, expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFlatMap(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		fn       func(interface{}) interface{}
		expected []interface{}
	}{
		{
			name:  "split words into chars",
			input: []interface{}{"go", "fn"},
			fn: func(word interface{}) interface{} {
				chars := []interface{}{}
				for _, c := range word.(string) {
					chars = append(chars, string(c))
				}
				return chars
			},
			expected: []interface{}{"g", "o", "f", "n"},
		},
		{
			name:  "empty string handling",
			input: []interface{}{"", "a", ""},
			fn: func(word interface{}) interface{} {
				chars := []interface{}{}
				for _, c := range word.(string) {
					chars = append(chars, string(c))
				}
				return chars
			},
			expected: []interface{}{"a"},
		},
		{
			name:  "nil handling",
			input: []interface{}{nil, "a", nil},
			fn: func(word interface{}) interface{} {
				if word == nil {
					return []interface{}{}
				}
				return []interface{}{word}
			},
			expected: []interface{}{"a"},
		},
		{
			name:  "empty slice result",
			input: []interface{}{"", "", ""},
			fn: func(word interface{}) interface{} {
				return []interface{}{}
			},
			expected: []interface{}{},
		},
		{
			name:  "nested array flattening",
			input: []interface{}{1, 2, 3},
			fn: func(n interface{}) interface{} {
				return []interface{}{n, n}
			},
			expected: []interface{}{1, 1, 2, 2, 3, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FlatMap(tt.input, tt.fn)
			if !equalSlices(result, tt.expected) {
				t.Errorf("FlatMap failed, expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestAllAnyNone(t *testing.T) {
	tests := []struct {
		name         string
		input        []interface{}
		fn           func(interface{}) bool
		expectedAll  bool
		expectedAny  bool
		expectedNone bool
	}{
		{
			name:         "all even numbers",
			input:        []interface{}{2, 4, 6, 8},
			fn:           func(x interface{}) bool { return x.(int)%2 == 0 },
			expectedAll:  true,
			expectedAny:  true,
			expectedNone: false,
		},
		{
			name:         "no even numbers",
			input:        []interface{}{1, 3, 5, 7},
			fn:           func(x interface{}) bool { return x.(int)%2 == 0 },
			expectedAll:  false,
			expectedAny:  false,
			expectedNone: true,
		},
		{
			name:         "mixed even/odd numbers",
			input:        []interface{}{1, 2, 3, 4},
			fn:           func(x interface{}) bool { return x.(int)%2 == 0 },
			expectedAll:  false,
			expectedAny:  true,
			expectedNone: false,
		},
		{
			name:         "empty slice",
			input:        []interface{}{},
			fn:           func(x interface{}) bool { return x.(int)%2 == 0 },
			expectedAll:  true,  // vacuously true
			expectedAny:  false, // vacuously false
			expectedNone: true,  // vacuously true
		},
		{
			name:         "nil predicate handling",
			input:        []interface{}{nil, nil, nil},
			fn:           func(x interface{}) bool { return x == nil },
			expectedAll:  true,
			expectedAny:  true,
			expectedNone: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultAll := All(tt.input, tt.fn)
			if resultAll != tt.expectedAll {
				t.Errorf("All failed, expected %v, got %v", tt.expectedAll, resultAll)
			}

			resultAny := Any(tt.input, tt.fn)
			if resultAny != tt.expectedAny {
				t.Errorf("Any failed, expected %v, got %v", tt.expectedAny, resultAny)
			}

			resultNone := None(tt.input, tt.fn)
			if resultNone != tt.expectedNone {
				t.Errorf("None failed, expected %v, got %v", tt.expectedNone, resultNone)
			}
		})
	}
}

func TestChunk(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		size     int
		expected [][]interface{}
	}{
		{
			name:     "normal chunks",
			input:    []interface{}{1, 2, 3, 4, 5},
			size:     2,
			expected: [][]interface{}{{1, 2}, {3, 4}, {5}},
		},
		{
			name:     "empty slice",
			input:    []interface{}{},
			size:     2,
			expected: [][]interface{}{},
		},
		{
			name:     "chunk size equals length",
			input:    []interface{}{1, 2, 3},
			size:     3,
			expected: [][]interface{}{{1, 2, 3}},
		},
		{
			name:     "chunk size larger than length",
			input:    []interface{}{1, 2},
			size:     3,
			expected: [][]interface{}{{1, 2}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Chunk(tt.input, tt.size)
			if !equalNestedSlices(result, tt.expected) {
				t.Errorf("Chunk failed, expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFind(t *testing.T) {
	tests := []struct {
		name        string
		input       []interface{}
		fn          func(interface{}) bool
		expected    interface{}
		shouldFound bool
	}{
		{
			name:        "find first number greater than 2",
			input:       []interface{}{1, 2, 3, 4},
			fn:          func(v interface{}) bool { return v.(int) > 2 },
			expected:    3,
			shouldFound: true,
		},
		{
			name:        "empty slice",
			input:       []interface{}{},
			fn:          func(v interface{}) bool { return v.(int) > 2 },
			expected:    nil,
			shouldFound: false,
		},
		{
			name:        "no match",
			input:       []interface{}{1, 2},
			fn:          func(v interface{}) bool { return v.(int) > 10 },
			expected:    nil,
			shouldFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, found := Find(tt.input, tt.fn)
			if found != tt.shouldFound {
				t.Errorf("Find failed, expected found=%v, got found=%v", tt.shouldFound, found)
			}
			if found && result != tt.expected {
				t.Errorf("Find failed, expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []interface{}
	}{
		{
			name:     "normal sequence",
			input:    []interface{}{1, 2, 3, 4},
			expected: []interface{}{4, 3, 2, 1},
		},
		{
			name:     "empty slice",
			input:    []interface{}{},
			expected: []interface{}{},
		},
		{
			name:     "single element",
			input:    []interface{}{1},
			expected: []interface{}{1},
		},
		{
			name:     "mixed types",
			input:    []interface{}{1, "a", true, 2.5},
			expected: []interface{}{2.5, true, "a", 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Reverse(tt.input)
			if !equalSlices(result, tt.expected) {
				t.Errorf("Reverse failed, expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestZipUnzip(t *testing.T) {
	tests := []struct {
		name        string
		a           []interface{}
		b           []interface{}
		expectedZip [][2]interface{}
		expectedA   []interface{}
		expectedB   []interface{}
	}{
		{
			name:        "normal case",
			a:           []interface{}{1, 2, 3},
			b:           []interface{}{"a", "b", "c"},
			expectedZip: [][2]interface{}{{1, "a"}, {2, "b"}, {3, "c"}},
			expectedA:   []interface{}{1, 2, 3},
			expectedB:   []interface{}{"a", "b", "c"},
		},
		{
			name:        "empty slices",
			a:           []interface{}{},
			b:           []interface{}{},
			expectedZip: [][2]interface{}{},
			expectedA:   []interface{}{},
			expectedB:   []interface{}{},
		},
		{
			name:        "different lengths - a shorter",
			a:           []interface{}{1, 2},
			b:           []interface{}{"a", "b", "c"},
			expectedZip: [][2]interface{}{{1, "a"}, {2, "b"}},
			expectedA:   []interface{}{1, 2},
			expectedB:   []interface{}{"a", "b"},
		},
		{
			name:        "different lengths - b shorter",
			a:           []interface{}{1, 2, 3},
			b:           []interface{}{"a"},
			expectedZip: [][2]interface{}{{1, "a"}},
			expectedA:   []interface{}{1},
			expectedB:   []interface{}{"a"},
		},
		{
			name:        "mixed types",
			a:           []interface{}{1, true, "x"},
			b:           []interface{}{2.5, false, "y"},
			expectedZip: [][2]interface{}{{1, 2.5}, {true, false}, {"x", "y"}},
			expectedA:   []interface{}{1, true, "x"},
			expectedB:   []interface{}{2.5, false, "y"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zipped := Zip(tt.a, tt.b)
			if len(zipped) != len(tt.expectedZip) {
				t.Errorf("Zip failed, expected length %v, got %v", len(tt.expectedZip), len(zipped))
			}
			for i, pair := range zipped {
				if pair[0] != tt.expectedZip[i][0] || pair[1] != tt.expectedZip[i][1] {
					t.Errorf("Zip failed at index %d, expected %v, got %v", i, tt.expectedZip[i], pair)
				}
			}

			unzippedA, unzippedB := Unzip(zipped)
			if !equalSlices(unzippedA, tt.expectedA) {
				t.Errorf("Unzip failed for slice A, expected %v, got %v", tt.expectedA, unzippedA)
			}
			if !equalSlices(unzippedB, tt.expectedB) {
				t.Errorf("Unzip failed for slice B, expected %v, got %v", tt.expectedB, unzippedB)
			}
		})
	}
}

func TestGroupBy(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		fn       func(interface{}) interface{}
		expected map[interface{}][]interface{}
	}{
		{
			name:  "group by even/odd",
			input: []interface{}{1, 2, 3, 4, 5, 6},
			fn:    func(v interface{}) interface{} { return v.(int) % 2 },
			expected: map[interface{}][]interface{}{
				0: {2, 4, 6},
				1: {1, 3, 5},
			},
		},
		{
			name:     "empty slice",
			input:    []interface{}{},
			fn:       func(v interface{}) interface{} { return v.(int) % 2 },
			expected: map[interface{}][]interface{}{},
		},
		{
			name:  "single group",
			input: []interface{}{2, 4, 6},
			fn:    func(v interface{}) interface{} { return v.(int) % 2 },
			expected: map[interface{}][]interface{}{
				0: {2, 4, 6},
			},
		},
		{
			name:  "group by type",
			input: []interface{}{1, "a", 2, "b", true},
			fn: func(v interface{}) interface{} {
				switch v.(type) {
				case int:
					return "int"
				case string:
					return "string"
				default:
					return "other"
				}
			},
			expected: map[interface{}][]interface{}{
				"int":    {1, 2},
				"string": {"a", "b"},
				"other":  {true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GroupBy(tt.input, tt.fn)
			if len(result) != len(tt.expected) {
				t.Errorf("GroupBy failed, expected %d groups, got %d", len(tt.expected), len(result))
			}
			for k, v := range tt.expected {
				if !equalSlices(result[k], v) {
					t.Errorf("GroupBy failed for key %v, expected %v, got %v", k, v, result[k])
				}
			}
		})
	}
}

func TestFlatten(t *testing.T) {
	tests := []struct {
		name     string
		input    [][]interface{}
		expected []interface{}
	}{
		{
			name:     "normal nested slices",
			input:    [][]interface{}{{1, 2}, {3, 4}, {5}},
			expected: []interface{}{1, 2, 3, 4, 5},
		},
		{
			name:     "empty outer slice",
			input:    [][]interface{}{},
			expected: []interface{}{},
		},
		{
			name:     "empty inner slices",
			input:    [][]interface{}{{}, {}, {}},
			expected: []interface{}{},
		},
		{
			name:     "mixed types",
			input:    [][]interface{}{{1, "a"}, {true, 2.5}, {nil}},
			expected: []interface{}{1, "a", true, 2.5, nil},
		},
		{
			name:     "single element slices",
			input:    [][]interface{}{{1}, {2}, {3}},
			expected: []interface{}{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Flatten(tt.input)
			if !equalSlices(result, tt.expected) {
				t.Errorf("Flatten failed, expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []interface{}
		item     interface{}
		expected bool
	}{
		{
			name:     "contains integer",
			slice:    []interface{}{1, 2, 3},
			item:     2,
			expected: true,
		},
		{
			name:     "does not contain integer",
			slice:    []interface{}{1, 2, 3},
			item:     4,
			expected: false,
		},
		{
			name:     "contains string",
			slice:    []interface{}{"a", "b", "c"},
			item:     "b",
			expected: true,
		},
		{
			name:     "empty slice",
			slice:    []interface{}{},
			item:     1,
			expected: false,
		},
		{
			name:     "contains nil",
			slice:    []interface{}{1, nil, 3},
			item:     nil,
			expected: true,
		},
		{
			name:     "mixed types",
			slice:    []interface{}{1, "a", true},
			item:     true,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			if result != tt.expected {
				t.Errorf("contains failed, expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Helper function for testing
func contains(slice []interface{}, item interface{}) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// Helper functions for comparing slices
func equalSlices(a, b []interface{}) bool {
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

func equalNestedSlices(a, b [][]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !equalSlices(a[i], b[i]) {
			return false
		}
	}
	return true
}
