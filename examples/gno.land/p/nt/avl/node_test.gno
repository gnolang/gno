package avl

import (
	"sort"
	"strings"
	"testing"

	"gno.land/p/nt/ufmt"
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
		asc  bool
	}{
		{"ascending", true},
		{"descending", false},
	}

	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			// use sl to insert the values, and reversed to match the values
			// we do this to ensure that the order of TraverseByOffset is independent
			// from the insertion order
			sl := strings.Split(testStrings, "\n")
			sort.Strings(sl)
			reversed := append([]string{}, sl...)
			reverseSlice(reversed)

			if !tt.asc {
				sl, reversed = reversed, sl
			}

			r := NewNode(reversed[0], nil)
			for _, v := range reversed[1:] {
				r, _ = r.Set(v, nil)
			}

			var result []string
			for i := 0; i < len(sl); i++ {
				r.TraverseByOffset(i, 1, tt.asc, true, func(n *Node) bool {
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

					r.TraverseByOffset(i, l, tt.asc, true, func(tr *Node) bool {
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

			t.Run("early termination", func(t *testing.T) {
				if len(tt.input) == 0 {
					return // Skip for empty tree
				}

				var result []string
				var count int
				tree.Iterate("", "", func(n *Node) bool {
					count++
					result = append(result, n.Key())
					return true // Stop after first item
				})

				if count != 1 {
					t.Errorf("Expected callback to be called exactly once, got %d calls", count)
				}
				if len(result) != 1 {
					t.Errorf("Expected exactly one result, got %d items", len(result))
				}
				if len(result) > 0 && result[0] != tt.expected[0] {
					t.Errorf("Expected first item to be %v, got %v", tt.expected[0], result[0])
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
			[]string{"A", "B", "C", "D", "E"},
		},
		{
			"left rotation when right subtree is higher",
			[]string{"A", "C", "E", "D", "F"},
			[]string{"A", "C", "D", "E", "F"},
		},
		{
			"left-right rotation",
			[]string{"E", "A", "C", "B", "D"},
			[]string{"A", "B", "C", "D", "E"},
		},
		{
			"right-left rotation",
			[]string{"A", "E", "C", "B", "D"},
			[]string{"A", "B", "C", "D", "E"},
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

func TestRemoveFromEmptyTree(t *testing.T) {
	var tree *Node
	newTree, _, val, removed := tree.Remove("NonExistent")
	if newTree != nil {
		t.Errorf("Removing from an empty tree should still be nil tree.")
	}
	if val != nil || removed {
		t.Errorf("Expected no value and removed=false when removing from empty tree.")
	}
}

func TestBalanceAfterRemoval(t *testing.T) {
	tests := []struct {
		name            string
		insertKeys      []string
		removeKey       string
		expectedBalance int
	}{
		{
			name:            "balance after removing right node",
			insertKeys:      []string{"B", "A", "D", "C", "E"},
			removeKey:       "E",
			expectedBalance: 0,
		},
		{
			name:            "balance after removing left node",
			insertKeys:      []string{"D", "B", "E", "A", "C"},
			removeKey:       "A",
			expectedBalance: 0,
		},
		{
			name:            "ensure no lean after removal",
			insertKeys:      []string{"C", "B", "E", "A", "D", "F"},
			removeKey:       "F",
			expectedBalance: -1,
		},
		{
			name:            "descending order insert, remove middle node",
			insertKeys:      []string{"E", "D", "C", "B", "A"},
			removeKey:       "C",
			expectedBalance: 0,
		},
		{
			name:            "ascending order insert, remove middle node",
			insertKeys:      []string{"A", "B", "C", "D", "E"},
			removeKey:       "C",
			expectedBalance: 0,
		},
		{
			name:            "duplicate key insert, remove the duplicated key",
			insertKeys:      []string{"C", "B", "C", "A", "D"},
			removeKey:       "C",
			expectedBalance: 1,
		},
		{
			name:            "complex rotation case",
			insertKeys:      []string{"H", "B", "A", "C", "E", "D", "F", "G"},
			removeKey:       "B",
			expectedBalance: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tree *Node
			for _, key := range tt.insertKeys {
				tree, _ = tree.Set(key, nil)
			}

			tree, _, _, _ = tree.Remove(tt.removeKey)

			balance := tree.calcBalance()
			if balance != tt.expectedBalance {
				t.Errorf("Expected balance factor %d, got %d", tt.expectedBalance, balance)
			}

			if balance < -1 || balance > 1 {
				t.Errorf("Tree is unbalanced with factor %d", balance)
			}

			if errMsg := checkSubtreeBalance(t, tree); errMsg != "" {
				t.Errorf("AVL property violation after removal: %s", errMsg)
			}
		})
	}
}

func TestBSTProperty(t *testing.T) {
	var tree *Node
	keys := []string{"D", "B", "F", "A", "C", "E", "G"}
	for _, key := range keys {
		tree, _ = tree.Set(key, nil)
	}

	var result []string
	inorderTraversal(t, tree, &result)

	for i := 1; i < len(result); i++ {
		if result[i] < result[i-1] {
			t.Errorf("BST property violated: %s < %s (index %d)",
				result[i], result[i-1], i)
		}
	}
}

// inorderTraversal performs an inorder traversal of the tree and returns the keys in a list.
func inorderTraversal(t *testing.T, node *Node, result *[]string) {
	t.Helper()

	if node == nil {
		return
	}
	// leaf
	if node.height == 0 {
		*result = append(*result, node.key)
		return
	}
	inorderTraversal(t, node.leftNode, result)
	inorderTraversal(t, node.rightNode, result)
}

// checkSubtreeBalance checks if all nodes under the given node satisfy the AVL tree conditions.
// The balance factor of all nodes must be ∈ [-1, +1]
func checkSubtreeBalance(t *testing.T, node *Node) string {
	t.Helper()

	if node == nil {
		return ""
	}

	if node.IsLeaf() {
		// leaf node must be height=0, size=1
		if node.height != 0 {
			return ufmt.Sprintf("Leaf node %s has height %d, expected 0", node.Key(), node.height)
		}
		if node.size != 1 {
			return ufmt.Sprintf("Leaf node %s has size %d, expected 1", node.Key(), node.size)
		}
		return ""
	}

	// check balance factor for current node
	balanceFactor := node.calcBalance()
	if balanceFactor < -1 || balanceFactor > 1 {
		return ufmt.Sprintf("Node %s is unbalanced: balanceFactor=%d", node.Key(), balanceFactor)
	}

	// check height / size relationship for children
	left, right := node.getLeftNode(), node.getRightNode()
	expectedHeight := maxInt8(left.height, right.height) + 1
	if node.height != expectedHeight {
		return ufmt.Sprintf("Node %s has incorrect height %d, expected %d", node.Key(), node.height, expectedHeight)
	}
	expectedSize := left.Size() + right.Size()
	if node.size != expectedSize {
		return ufmt.Sprintf("Node %s has incorrect size %d, expected %d", node.Key(), node.size, expectedSize)
	}

	// recursively check the left/right subtree
	if errMsg := checkSubtreeBalance(t, left); errMsg != "" {
		return errMsg
	}
	if errMsg := checkSubtreeBalance(t, right); errMsg != "" {
		return errMsg
	}

	return ""
}

func slicesEqual(w1, w2 []string) bool {
	if len(w1) != len(w2) {
		return false
	}
	for i := 0; i < len(w1); i++ {
		if w1[i] != w2[i] {
			return false
		}
	}
	return true
}

func reverseSlice(ss []string) {
	for i := 0; i < len(ss)/2; i++ {
		j := len(ss) - 1 - i
		ss[i], ss[j] = ss[j], ss[i]
	}
}
