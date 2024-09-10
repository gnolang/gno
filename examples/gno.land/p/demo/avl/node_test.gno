package avl

import (
	"sort"
	"strings"
	"testing"
)

func TestTraverseByOffset(t *testing.T) {
	const testStrings = `Alfa
Alfred
Alpha
Alphabet
Beta
Beth
Book
Browser`
	tt := []struct {
		name string
		desc bool
	}{
		{"ascending", false},
		{"descending", true},
	}

	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			sl := strings.Split(testStrings, "\n")

			// sort a first time in the order opposite to how we'll be traversing
			// the tree, to ensure that we are not just iterating through with
			// insertion order.
			sort.Strings(sl)
			if !tt.desc {
				reverseSlice(sl)
			}

			r := NewNode(sl[0], nil)
			for _, v := range sl[1:] {
				r, _ = r.Set(v, nil)
			}

			// then sort sl in the order we'll be traversing it, so that we can
			// compare the result with sl.
			reverseSlice(sl)

			var result []string
			for i := 0; i < len(sl); i++ {
				r.TraverseByOffset(i, 1, tt.desc, true, func(n *Node) bool {
					result = append(result, n.Key())
					return false
				})
			}

			if !slicesEqual(sl, result) {
				t.Errorf("want %v got %v", sl, result)
			}

			for l := 2; l <= len(sl); l++ {
				// "slices"
				for i := 0; i <= len(sl); i++ {
					max := i + l
					if max > len(sl) {
						max = len(sl)
					}
					exp := sl[i:max]
					actual := []string{}

					r.TraverseByOffset(i, l, tt.desc, true, func(tr *Node) bool {
						actual = append(actual, tr.Key())
						return false
					})
					if !slicesEqual(exp, actual) {
						t.Errorf("want %v got %v", exp, actual)
					}
				}
			}
		})
	}
}

func TestHas(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		hasKey   string
		expected bool
	}{
		{
			"has key in non-empty tree",
			[]string{"C", "A", "B", "E", "D"},
			"B",
			true,
		},
		{
			"does not have key in non-empty tree",
			[]string{"C", "A", "B", "E", "D"},
			"F",
			false,
		},
		{
			"has key in single-node tree",
			[]string{"A"},
			"A",
			true,
		},
		{
			"does not have key in single-node tree",
			[]string{"A"},
			"B",
			false,
		},
		{
			"does not have key in empty tree",
			[]string{},
			"A",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tree *Node
			for _, key := range tt.input {
				tree, _ = tree.Set(key, nil)
			}

			result := tree.Has(tt.hasKey)

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name         string
		input        []string
		getKey       string
		expectIdx    int
		expectVal    interface{}
		expectExists bool
	}{
		{
			"get existing key",
			[]string{"C", "A", "B", "E", "D"},
			"B",
			1,
			nil,
			true,
		},
		{
			"get non-existent key (smaller)",
			[]string{"C", "A", "B", "E", "D"},
			"@",
			0,
			nil,
			false,
		},
		{
			"get non-existent key (larger)",
			[]string{"C", "A", "B", "E", "D"},
			"F",
			5,
			nil,
			false,
		},
		{
			"get from empty tree",
			[]string{},
			"A",
			0,
			nil,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tree *Node
			for _, key := range tt.input {
				tree, _ = tree.Set(key, nil)
			}

			idx, val, exists := tree.Get(tt.getKey)

			if idx != tt.expectIdx {
				t.Errorf("Expected index %d, got %d", tt.expectIdx, idx)
			}

			if val != tt.expectVal {
				t.Errorf("Expected value %v, got %v", tt.expectVal, val)
			}

			if exists != tt.expectExists {
				t.Errorf("Expected exists %t, got %t", tt.expectExists, exists)
			}
		})
	}
}

func TestGetByIndex(t *testing.T) {
	tests := []struct {
		name        string
		input       []string
		idx         int
		expectKey   string
		expectVal   interface{}
		expectPanic bool
	}{
		{
			"get by valid index",
			[]string{"C", "A", "B", "E", "D"},
			2,
			"C",
			nil,
			false,
		},
		{
			"get by valid index (smallest)",
			[]string{"C", "A", "B", "E", "D"},
			0,
			"A",
			nil,
			false,
		},
		{
			"get by valid index (largest)",
			[]string{"C", "A", "B", "E", "D"},
			4,
			"E",
			nil,
			false,
		},
		{
			"get by invalid index (negative)",
			[]string{"C", "A", "B", "E", "D"},
			-1,
			"",
			nil,
			true,
		},
		{
			"get by invalid index (out of range)",
			[]string{"C", "A", "B", "E", "D"},
			5,
			"",
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tree *Node
			for _, key := range tt.input {
				tree, _ = tree.Set(key, nil)
			}

			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected a panic but didn't get one")
					}
				}()
			}

			key, val := tree.GetByIndex(tt.idx)

			if !tt.expectPanic {
				if key != tt.expectKey {
					t.Errorf("Expected key %s, got %s", tt.expectKey, key)
				}

				if val != tt.expectVal {
					t.Errorf("Expected value %v, got %v", tt.expectVal, val)
				}
			}
		})
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name      string
		input     []string
		removeKey string
		expected  []string
	}{
		{
			"remove leaf node",
			[]string{"C", "A", "B", "D"},
			"B",
			[]string{"A", "C", "D"},
		},
		{
			"remove node with one child",
			[]string{"C", "A", "B", "D"},
			"A",
			[]string{"B", "C", "D"},
		},
		{
			"remove node with two children",
			[]string{"C", "A", "B", "E", "D"},
			"C",
			[]string{"A", "B", "D", "E"},
		},
		{
			"remove root node",
			[]string{"C", "A", "B", "E", "D"},
			"C",
			[]string{"A", "B", "D", "E"},
		},
		{
			"remove non-existent key",
			[]string{"C", "A", "B", "E", "D"},
			"F",
			[]string{"A", "B", "C", "D", "E"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tree *Node
			for _, key := range tt.input {
				tree, _ = tree.Set(key, nil)
			}

			tree, _, _, _ = tree.Remove(tt.removeKey)

			result := make([]string, 0)
			tree.Iterate("", "", func(n *Node) bool {
				result = append(result, n.Key())
				return false
			})

			if !slicesEqual(tt.expected, result) {
				t.Errorf("want %v got %v", tt.expected, result)
			}
		})
	}
}

func TestTraverse(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			"empty tree",
			[]string{},
			[]string{},
		},
		{
			"single node tree",
			[]string{"A"},
			[]string{"A"},
		},
		{
			"small tree",
			[]string{"C", "A", "B", "E", "D"},
			[]string{"A", "B", "C", "D", "E"},
		},
		{
			"large tree",
			[]string{"H", "D", "L", "B", "F", "J", "N", "A", "C", "E", "G", "I", "K", "M", "O"},
			[]string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tree *Node
			for _, key := range tt.input {
				tree, _ = tree.Set(key, nil)
			}

			t.Run("iterate", func(t *testing.T) {
				var result []string
				tree.Iterate("", "", func(n *Node) bool {
					result = append(result, n.Key())
					return false
				})
				if !slicesEqual(tt.expected, result) {
					t.Errorf("want %v got %v", tt.expected, result)
				}
			})

			t.Run("ReverseIterate", func(t *testing.T) {
				var result []string
				tree.ReverseIterate("", "", func(n *Node) bool {
					result = append(result, n.Key())
					return false
				})
				expected := make([]string, len(tt.expected))
				copy(expected, tt.expected)
				for i, j := 0, len(expected)-1; i < j; i, j = i+1, j-1 {
					expected[i], expected[j] = expected[j], expected[i]
				}
				if !slicesEqual(expected, result) {
					t.Errorf("want %v got %v", expected, result)
				}
			})

			t.Run("TraverseInRange", func(t *testing.T) {
				var result []string
				start, end := "C", "M"
				tree.TraverseInRange(start, end, true, true, func(n *Node) bool {
					result = append(result, n.Key())
					return false
				})
				expected := make([]string, 0)
				for _, key := range tt.expected {
					if key >= start && key < end {
						expected = append(expected, key)
					}
				}
				if !slicesEqual(expected, result) {
					t.Errorf("want %v got %v", expected, result)
				}
			})
		})
	}
}

func TestRotateWhenHeightDiffers(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			"right rotation when left subtree is higher",
			[]string{"E", "C", "A", "B", "D"},
			[]string{"A", "B", "C", "E", "D"},
		},
		{
			"left rotation when right subtree is higher",
			[]string{"A", "C", "E", "D", "F"},
			[]string{"A", "C", "D", "E", "F"},
		},
		{
			"left-right rotation",
			[]string{"E", "A", "C", "B", "D"},
			[]string{"A", "B", "C", "E", "D"},
		},
		{
			"right-left rotation",
			[]string{"A", "E", "C", "B", "D"},
			[]string{"A", "B", "C", "E", "D"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tree *Node
			for _, key := range tt.input {
				tree, _ = tree.Set(key, nil)
			}

			// perform rotation or balance
			tree = tree.balance()

			// check tree structure
			var result []string
			tree.Iterate("", "", func(n *Node) bool {
				result = append(result, n.Key())
				return false
			})

			if !slicesEqual(tt.expected, result) {
				t.Errorf("want %v got %v", tt.expected, result)
			}
		})
	}
}

func TestRotateAndBalance(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			"right rotation",
			[]string{"A", "B", "C", "D", "E"},
			[]string{"A", "B", "C", "D", "E"},
		},
		{
			"left rotation",
			[]string{"E", "D", "C", "B", "A"},
			[]string{"A", "B", "C", "D", "E"},
		},
		{
			"left-right rotation",
			[]string{"C", "A", "E", "B", "D"},
			[]string{"A", "B", "C", "D", "E"},
		},
		{
			"right-left rotation",
			[]string{"C", "E", "A", "D", "B"},
			[]string{"A", "B", "C", "D", "E"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tree *Node
			for _, key := range tt.input {
				tree, _ = tree.Set(key, nil)
			}

			tree = tree.balance()

			var result []string
			tree.Iterate("", "", func(n *Node) bool {
				result = append(result, n.Key())
				return false
			})

			if !slicesEqual(tt.expected, result) {
				t.Errorf("want %v got %v", tt.expected, result)
			}
		})
	}
}

func slicesEqual(w1, w2 []string) bool {
	if len(w1) != len(w2) {
		return false
	}
	for i := 0; i < len(w1); i++ {
		if w1[0] != w2[0] {
			return false
		}
	}
	return true
}

func maxint8(a, b int8) int8 {
	if a > b {
		return a
	}
	return b
}

func reverseSlice(ss []string) {
	for i := 0; i < len(ss)/2; i++ {
		j := len(ss) - 1 - i
		ss[i], ss[j] = ss[j], ss[i]
	}
}
