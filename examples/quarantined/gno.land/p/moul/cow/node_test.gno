package cow

import (
	"fmt"
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
		expectVal    any
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
		expectVal   any
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

func TestNodeStructuralSharing(t *testing.T) {
	t.Run("unmodified paths remain shared", func(t *testing.T) {
		root := NewNode("B", 2)
		root, _ = root.Set("A", 1)
		root, _ = root.Set("C", 3)

		originalRight := root.rightNode
		newRoot, _ := root.Set("A", 10)

		if newRoot.rightNode != originalRight {
			t.Error("Unmodified right subtree should remain shared")
		}
	})

	t.Run("multiple modifications reuse shared structure", func(t *testing.T) {
		// Create initial tree
		root := NewNode("B", 2)
		root, _ = root.Set("A", 1)
		root, _ = root.Set("C", 3)

		// Store original nodes
		originalRight := root.rightNode

		// First modification
		mod1, _ := root.Set("A", 10)

		// Second modification
		mod2, _ := mod1.Set("C", 30)

		// Check sharing in first modification
		if mod1.rightNode != originalRight {
			t.Error("First modification should share unmodified right subtree")
		}

		// Check that second modification creates new right node
		if mod2.rightNode == originalRight {
			t.Error("Second modification should create new right node")
		}
	})
}

func TestNodeCopyOnWrite(t *testing.T) {
	t.Run("copy preserves structure", func(t *testing.T) {
		root := NewNode("B", 2)
		root, _ = root.Set("A", 1)
		root, _ = root.Set("C", 3)

		// Only copy non-leaf nodes
		if !root.IsLeaf() {
			copied := root._copy()
			if copied == root {
				t.Error("Copy should create new instance")
			}

			// Create temporary trees to use Equal method
			original := &Tree{node: root}
			copiedTree := &Tree{node: copied}
			if !original.Equal(copiedTree) {
				t.Error("Copied node should preserve structure")
			}
		}
	})

	t.Run("removal copy pattern", func(t *testing.T) {
		// Create a more complex tree to test removal
		root := NewNode("B", 2)
		root, _ = root.Set("A", 1)
		root, _ = root.Set("C", 3)
		root, _ = root.Set("D", 4) // Add this to ensure proper tree structure

		// Store references to original nodes
		originalRight := root.rightNode
		originalRightRight := originalRight.rightNode

		// Remove "A" which should only affect the left subtree
		newRoot, _, _, _ := root.Remove("A")

		// Verify right subtree remains unchanged and shared
		if newRoot.rightNode != originalRight {
			t.Error("Right subtree should remain shared during removal of left node")
		}

		// Also verify deeper nodes remain shared
		if newRoot.rightNode.rightNode != originalRightRight {
			t.Error("Deep right subtree should remain shared during removal")
		}

		// Verify original tree is unchanged
		if _, _, exists := root.Get("A"); !exists {
			t.Error("Original tree should remain unchanged")
		}
	})

	t.Run("copy leaf node panic", func(t *testing.T) {
		leaf := NewNode("A", 1)

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when copying leaf node")
			}
		}()

		// This should panic with our specific message
		leaf._copy()
	})
}

func TestNodeEqual(t *testing.T) {
	tests := []struct {
		name     string
		node1    func() *Node
		node2    func() *Node
		expected bool
	}{
		{
			name:     "nil nodes",
			node1:    func() *Node { return nil },
			node2:    func() *Node { return nil },
			expected: true,
		},
		{
			name:     "one nil node",
			node1:    func() *Node { return NewNode("A", 1) },
			node2:    func() *Node { return nil },
			expected: false,
		},
		{
			name:     "single leaf nodes equal",
			node1:    func() *Node { return NewNode("A", 1) },
			node2:    func() *Node { return NewNode("A", 1) },
			expected: true,
		},
		{
			name:     "single leaf nodes different key",
			node1:    func() *Node { return NewNode("A", 1) },
			node2:    func() *Node { return NewNode("B", 1) },
			expected: false,
		},
		{
			name:     "single leaf nodes different value",
			node1:    func() *Node { return NewNode("A", 1) },
			node2:    func() *Node { return NewNode("A", 2) },
			expected: false,
		},
		{
			name: "complex trees equal",
			node1: func() *Node {
				n, _ := NewNode("B", 2).Set("A", 1)
				n, _ = n.Set("C", 3)
				return n
			},
			node2: func() *Node {
				n, _ := NewNode("B", 2).Set("A", 1)
				n, _ = n.Set("C", 3)
				return n
			},
			expected: true,
		},
		{
			name: "complex trees different structure",
			node1: func() *Node {
				// Create a tree with structure:
				//     B
				//    / \
				//   A   D
				n := NewNode("B", 2)
				n, _ = n.Set("A", 1)
				n, _ = n.Set("D", 4)
				return n
			},
			node2: func() *Node {
				// Create a tree with structure:
				//     C
				//    / \
				//   A   D
				n := NewNode("C", 3)
				n, _ = n.Set("A", 1)
				n, _ = n.Set("D", 4)
				return n
			},
			expected: false, // These trees should be different
		},
		{
			name: "complex trees same structure despite different insertion order",
			node1: func() *Node {
				n, _ := NewNode("B", 2).Set("A", 1)
				n, _ = n.Set("C", 3)
				return n
			},
			node2: func() *Node {
				n, _ := NewNode("A", 1).Set("B", 2)
				n, _ = n.Set("C", 3)
				return n
			},
			expected: true,
		},
		{
			name: "truly different structures",
			node1: func() *Node {
				n, _ := NewNode("B", 2).Set("A", 1)
				return n // Tree with just two nodes
			},
			node2: func() *Node {
				n, _ := NewNode("B", 2).Set("C", 3)
				return n // Different two-node tree
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node1 := tt.node1()
			node2 := tt.node2()
			result := node1.Equal(node2)
			if result != tt.expected {
				t.Errorf("Expected Equal to return %v for %s", tt.expected, tt.name)
				println("\nComparison failed:")
				println("Tree 1:")
				printTree(node1, 0)
				println("Tree 2:")
				printTree(node2, 0)
			}
		})
	}
}

// Helper function to print tree structure
func printTree(node *Node, level int) {
	if node == nil {
		return
	}
	indent := strings.Repeat("  ", level)
	println(fmt.Sprintf("%sKey: %s, Value: %v, Height: %d, Size: %d",
		indent, node.key, node.value, node.height, node.size))
	printTree(node.leftNode, level+1)
	printTree(node.rightNode, level+1)
}
