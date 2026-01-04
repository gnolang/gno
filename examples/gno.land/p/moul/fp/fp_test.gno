package fp

import (
	"fmt"
	"testing"
)

func TestMap(t *testing.T) {
	tests := []struct {
		name     string
		input    []any
		fn       func(any) any
		expected []any
	}{
		{
			name:     "multiply numbers by 2",
			input:    []any{1, 2, 3},
			fn:       func(v any) any { return v.(int) * 2 },
			expected: []any{2, 4, 6},
		},
		{
			name:     "empty slice",
			input:    []any{},
			fn:       func(v any) any { return v.(int) * 2 },
			expected: []any{},
		},
		{
			name:     "convert numbers to strings",
			input:    []any{1, 2, 3},
			fn:       func(v any) any { return fmt.Sprintf("%d", v.(int)) },
			expected: []any{"1", "2", "3"},
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
		input    []any
		fn       func(any) bool
		expected []any
	}{
		{
			name:     "filter even numbers",
			input:    []any{1, 2, 3, 4},
			fn:       func(v any) bool { return v.(int)%2 == 0 },
			expected: []any{2, 4},
		},
		{
			name:     "empty slice",
			input:    []any{},
			fn:       func(v any) bool { return v.(int)%2 == 0 },
			expected: []any{},
		},
		{
			name:     "no matches",
			input:    []any{1, 3, 5},
			fn:       func(v any) bool { return v.(int)%2 == 0 },
			expected: []any{},
		},
		{
			name:     "all matches",
			input:    []any{2, 4, 6},
			fn:       func(v any) bool { return v.(int)%2 == 0 },
			expected: []any{2, 4, 6},
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
		input    []any
		fn       func(any, any) any
		initial  any
		expected any
	}{
		{
			name:     "sum numbers",
			input:    []any{1, 2, 3},
			fn:       func(a, b any) any { return a.(int) + b.(int) },
			initial:  0,
			expected: 6,
		},
		{
			name:     "empty slice",
			input:    []any{},
			fn:       func(a, b any) any { return a.(int) + b.(int) },
			initial:  0,
			expected: 0,
		},
		{
			name:     "concatenate strings",
			input:    []any{"a", "b", "c"},
			fn:       func(a, b any) any { return a.(string) + b.(string) },
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
		input    []any
		fn       func(any) any
		expected []any
	}{
		{
			name:  "split words into chars",
			input: []any{"go", "fn"},
			fn: func(word any) any {
				chars := []any{}
				for _, c := range word.(string) {
					chars = append(chars, string(c))
				}
				return chars
			},
			expected: []any{"g", "o", "f", "n"},
		},
		{
			name:  "empty string handling",
			input: []any{"", "a", ""},
			fn: func(word any) any {
				chars := []any{}
				for _, c := range word.(string) {
					chars = append(chars, string(c))
				}
				return chars
			},
			expected: []any{"a"},
		},
		{
			name:  "nil handling",
			input: []any{nil, "a", nil},
			fn: func(word any) any {
				if word == nil {
					return []any{}
				}
				return []any{word}
			},
			expected: []any{"a"},
		},
		{
			name:  "empty slice result",
			input: []any{"", "", ""},
			fn: func(word any) any {
				return []any{}
			},
			expected: []any{},
		},
		{
			name:  "nested array flattening",
			input: []any{1, 2, 3},
			fn: func(n any) any {
				return []any{n, n}
			},
			expected: []any{1, 1, 2, 2, 3, 3},
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
		input        []any
		fn           func(any) bool
		expectedAll  bool
		expectedAny  bool
		expectedNone bool
	}{
		{
			name:         "all even numbers",
			input:        []any{2, 4, 6, 8},
			fn:           func(x any) bool { return x.(int)%2 == 0 },
			expectedAll:  true,
			expectedAny:  true,
			expectedNone: false,
		},
		{
			name:         "no even numbers",
			input:        []any{1, 3, 5, 7},
			fn:           func(x any) bool { return x.(int)%2 == 0 },
			expectedAll:  false,
			expectedAny:  false,
			expectedNone: true,
		},
		{
			name:         "mixed even/odd numbers",
			input:        []any{1, 2, 3, 4},
			fn:           func(x any) bool { return x.(int)%2 == 0 },
			expectedAll:  false,
			expectedAny:  true,
			expectedNone: false,
		},
		{
			name:         "empty slice",
			input:        []any{},
			fn:           func(x any) bool { return x.(int)%2 == 0 },
			expectedAll:  true,  // vacuously true
			expectedAny:  false, // vacuously false
			expectedNone: true,  // vacuously true
		},
		{
			name:         "nil predicate handling",
			input:        []any{nil, nil, nil},
			fn:           func(x any) bool { return x == nil },
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
		input    []any
		size     int
		expected [][]any
	}{
		{
			name:     "normal chunks",
			input:    []any{1, 2, 3, 4, 5},
			size:     2,
			expected: [][]any{{1, 2}, {3, 4}, {5}},
		},
		{
			name:     "empty slice",
			input:    []any{},
			size:     2,
			expected: [][]any{},
		},
		{
			name:     "chunk size equals length",
			input:    []any{1, 2, 3},
			size:     3,
			expected: [][]any{{1, 2, 3}},
		},
		{
			name:     "chunk size larger than length",
			input:    []any{1, 2},
			size:     3,
			expected: [][]any{{1, 2}},
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
		input       []any
		fn          func(any) bool
		expected    any
		shouldFound bool
	}{
		{
			name:        "find first number greater than 2",
			input:       []any{1, 2, 3, 4},
			fn:          func(v any) bool { return v.(int) > 2 },
			expected:    3,
			shouldFound: true,
		},
		{
			name:        "empty slice",
			input:       []any{},
			fn:          func(v any) bool { return v.(int) > 2 },
			expected:    nil,
			shouldFound: false,
		},
		{
			name:        "no match",
			input:       []any{1, 2},
			fn:          func(v any) bool { return v.(int) > 10 },
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
		input    []any
		expected []any
	}{
		{
			name:     "normal sequence",
			input:    []any{1, 2, 3, 4},
			expected: []any{4, 3, 2, 1},
		},
		{
			name:     "empty slice",
			input:    []any{},
			expected: []any{},
		},
		{
			name:     "single element",
			input:    []any{1},
			expected: []any{1},
		},
		{
			name:     "mixed types",
			input:    []any{1, "a", true, 2.5},
			expected: []any{2.5, true, "a", 1},
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
		a           []any
		b           []any
		expectedZip [][2]any
		expectedA   []any
		expectedB   []any
	}{
		{
			name:        "normal case",
			a:           []any{1, 2, 3},
			b:           []any{"a", "b", "c"},
			expectedZip: [][2]any{{1, "a"}, {2, "b"}, {3, "c"}},
			expectedA:   []any{1, 2, 3},
			expectedB:   []any{"a", "b", "c"},
		},
		{
			name:        "empty slices",
			a:           []any{},
			b:           []any{},
			expectedZip: [][2]any{},
			expectedA:   []any{},
			expectedB:   []any{},
		},
		{
			name:        "different lengths - a shorter",
			a:           []any{1, 2},
			b:           []any{"a", "b", "c"},
			expectedZip: [][2]any{{1, "a"}, {2, "b"}},
			expectedA:   []any{1, 2},
			expectedB:   []any{"a", "b"},
		},
		{
			name:        "different lengths - b shorter",
			a:           []any{1, 2, 3},
			b:           []any{"a"},
			expectedZip: [][2]any{{1, "a"}},
			expectedA:   []any{1},
			expectedB:   []any{"a"},
		},
		{
			name:        "mixed types",
			a:           []any{1, true, "x"},
			b:           []any{2.5, false, "y"},
			expectedZip: [][2]any{{1, 2.5}, {true, false}, {"x", "y"}},
			expectedA:   []any{1, true, "x"},
			expectedB:   []any{2.5, false, "y"},
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
		input    []any
		fn       func(any) any
		expected map[any][]any
	}{
		{
			name:  "group by even/odd",
			input: []any{1, 2, 3, 4, 5, 6},
			fn:    func(v any) any { return v.(int) % 2 },
			expected: map[any][]any{
				0: {2, 4, 6},
				1: {1, 3, 5},
			},
		},
		{
			name:     "empty slice",
			input:    []any{},
			fn:       func(v any) any { return v.(int) % 2 },
			expected: map[any][]any{},
		},
		{
			name:  "single group",
			input: []any{2, 4, 6},
			fn:    func(v any) any { return v.(int) % 2 },
			expected: map[any][]any{
				0: {2, 4, 6},
			},
		},
		{
			name:  "group by type",
			input: []any{1, "a", 2, "b", true},
			fn: func(v any) any {
				switch v.(type) {
				case int:
					return "int"
				case string:
					return "string"
				default:
					return "other"
				}
			},
			expected: map[any][]any{
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
		input    [][]any
		expected []any
	}{
		{
			name:     "normal nested slices",
			input:    [][]any{{1, 2}, {3, 4}, {5}},
			expected: []any{1, 2, 3, 4, 5},
		},
		{
			name:     "empty outer slice",
			input:    [][]any{},
			expected: []any{},
		},
		{
			name:     "empty inner slices",
			input:    [][]any{{}, {}, {}},
			expected: []any{},
		},
		{
			name:     "mixed types",
			input:    [][]any{{1, "a"}, {true, 2.5}, {nil}},
			expected: []any{1, "a", true, 2.5, nil},
		},
		{
			name:     "single element slices",
			input:    [][]any{{1}, {2}, {3}},
			expected: []any{1, 2, 3},
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
		slice    []any
		item     any
		expected bool
	}{
		{
			name:     "contains integer",
			slice:    []any{1, 2, 3},
			item:     2,
			expected: true,
		},
		{
			name:     "does not contain integer",
			slice:    []any{1, 2, 3},
			item:     4,
			expected: false,
		},
		{
			name:     "contains string",
			slice:    []any{"a", "b", "c"},
			item:     "b",
			expected: true,
		},
		{
			name:     "empty slice",
			slice:    []any{},
			item:     1,
			expected: false,
		},
		{
			name:     "contains nil",
			slice:    []any{1, nil, 3},
			item:     nil,
			expected: true,
		},
		{
			name:     "mixed types",
			slice:    []any{1, "a", true},
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
func contains(slice []any, item any) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// Helper functions for comparing slices
func equalSlices(a, b []any) bool {
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

func equalNestedSlices(a, b [][]any) bool {
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
