package cow

import (
	"testing"
)

func TestNewTree(t *testing.T) {
	tree := NewTree()
	if tree.node != nil {
		t.Error("Expected tree.node to be nil")
	}
}

func TestTreeSize(t *testing.T) {
	tree := NewTree()
	if tree.Size() != 0 {
		t.Error("Expected empty tree size to be 0")
	}

	tree.Set("key1", "value1")
	tree.Set("key2", "value2")
	if tree.Size() != 2 {
		t.Error("Expected tree size to be 2")
	}
}

func TestTreeHas(t *testing.T) {
	tree := NewTree()
	tree.Set("key1", "value1")

	if !tree.Has("key1") {
		t.Error("Expected tree to have key1")
	}

	if tree.Has("key2") {
		t.Error("Expected tree to not have key2")
	}
}

func TestTreeGet(t *testing.T) {
	tree := NewTree()
	tree.Set("key1", "value1")

	value, exists := tree.Get("key1")
	if !exists || value != "value1" {
		t.Error("Expected Get to return value1 and true")
	}

	_, exists = tree.Get("key2")
	if exists {
		t.Error("Expected Get to return false for non-existent key")
	}
}

func TestTreeGetByIndex(t *testing.T) {
	tree := NewTree()
	tree.Set("key1", "value1")
	tree.Set("key2", "value2")

	key, value := tree.GetByIndex(0)
	if key != "key1" || value != "value1" {
		t.Error("Expected GetByIndex(0) to return key1 and value1")
	}

	key, value = tree.GetByIndex(1)
	if key != "key2" || value != "value2" {
		t.Error("Expected GetByIndex(1) to return key2 and value2")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected GetByIndex to panic for out-of-range index")
		}
	}()
	tree.GetByIndex(2)
}

func TestTreeRemove(t *testing.T) {
	tree := NewTree()
	tree.Set("key1", "value1")

	value, removed := tree.Remove("key1")
	if !removed || value != "value1" || tree.Size() != 0 {
		t.Error("Expected Remove to remove key-value pair")
	}

	_, removed = tree.Remove("key2")
	if removed {
		t.Error("Expected Remove to return false for non-existent key")
	}
}

func TestTreeIterate(t *testing.T) {
	tree := NewTree()
	tree.Set("key1", "value1")
	tree.Set("key2", "value2")
	tree.Set("key3", "value3")

	var keys []string
	tree.Iterate("", "", func(key string, value any) bool {
		keys = append(keys, key)
		return false
	})

	expectedKeys := []string{"key1", "key2", "key3"}
	if !slicesEqual(keys, expectedKeys) {
		t.Errorf("Expected keys %v, got %v", expectedKeys, keys)
	}
}

func TestTreeReverseIterate(t *testing.T) {
	tree := NewTree()
	tree.Set("key1", "value1")
	tree.Set("key2", "value2")
	tree.Set("key3", "value3")

	var keys []string
	tree.ReverseIterate("", "", func(key string, value any) bool {
		keys = append(keys, key)
		return false
	})

	expectedKeys := []string{"key3", "key2", "key1"}
	if !slicesEqual(keys, expectedKeys) {
		t.Errorf("Expected keys %v, got %v", expectedKeys, keys)
	}
}

func TestTreeIterateByOffset(t *testing.T) {
	tree := NewTree()
	tree.Set("key1", "value1")
	tree.Set("key2", "value2")
	tree.Set("key3", "value3")

	var keys []string
	tree.IterateByOffset(1, 2, func(key string, value any) bool {
		keys = append(keys, key)
		return false
	})

	expectedKeys := []string{"key2", "key3"}
	if !slicesEqual(keys, expectedKeys) {
		t.Errorf("Expected keys %v, got %v", expectedKeys, keys)
	}
}

func TestTreeReverseIterateByOffset(t *testing.T) {
	tree := NewTree()
	tree.Set("key1", "value1")
	tree.Set("key2", "value2")
	tree.Set("key3", "value3")

	var keys []string
	tree.ReverseIterateByOffset(1, 2, func(key string, value any) bool {
		keys = append(keys, key)
		return false
	})

	expectedKeys := []string{"key2", "key1"}
	if !slicesEqual(keys, expectedKeys) {
		t.Errorf("Expected keys %v, got %v", expectedKeys, keys)
	}
}

// Verify that Tree implements avl.ITree
// var _ avl.ITree = (*Tree)(nil) // TODO: fix gnovm bug: ./examples/gno.land/p/moul/cow: test pkg: panic: gno.land/p/moul/cow/tree_test.gno:166:5: name avl not defined in fileset with files [node.gno tree.gno node_test.gno tree_test.gno]:

func TestCopyOnWrite(t *testing.T) {
	// Create original tree
	original := NewTree()
	original.Set("A", 1)
	original.Set("B", 2)
	original.Set("C", 3)

	// Create a clone
	clone := original.Clone()

	// Modify clone
	clone.Set("B", 20)
	clone.Set("D", 4)

	// Verify original is unchanged
	if val, _ := original.Get("B"); val != 2 {
		t.Errorf("Original tree was modified: expected B=2, got B=%v", val)
	}
	if original.Has("D") {
		t.Error("Original tree was modified: found key D")
	}

	// Verify clone has new values
	if val, _ := clone.Get("B"); val != 20 {
		t.Errorf("Clone not updated: expected B=20, got B=%v", val)
	}
	if val, _ := clone.Get("D"); val != 4 {
		t.Errorf("Clone not updated: expected D=4, got D=%v", val)
	}
}

func TestCopyOnWriteEdgeCases(t *testing.T) {
	t.Run("nil tree clone", func(t *testing.T) {
		var original *Tree
		clone := original.Clone()
		if clone != nil {
			t.Error("Expected nil clone from nil tree")
		}
	})

	t.Run("empty tree clone", func(t *testing.T) {
		original := NewTree()
		clone := original.Clone()

		// Modify clone
		clone.Set("A", 1)

		if original.Size() != 0 {
			t.Error("Original empty tree was modified")
		}
		if clone.Size() != 1 {
			t.Error("Clone was not modified")
		}
	})

	t.Run("multiple clones", func(t *testing.T) {
		original := NewTree()
		original.Set("A", 1)
		original.Set("B", 2)

		// Create multiple clones
		clone1 := original.Clone()
		clone2 := original.Clone()
		clone3 := clone1.Clone()

		// Modify each clone differently
		clone1.Set("A", 10)
		clone2.Set("B", 20)
		clone3.Set("C", 30)

		// Check original remains unchanged
		if val, _ := original.Get("A"); val != 1 {
			t.Errorf("Original modified: expected A=1, got A=%v", val)
		}
		if val, _ := original.Get("B"); val != 2 {
			t.Errorf("Original modified: expected B=2, got B=%v", val)
		}

		// Verify each clone has correct values
		if val, _ := clone1.Get("A"); val != 10 {
			t.Errorf("Clone1 incorrect: expected A=10, got A=%v", val)
		}
		if val, _ := clone2.Get("B"); val != 20 {
			t.Errorf("Clone2 incorrect: expected B=20, got B=%v", val)
		}
		if val, _ := clone3.Get("C"); val != 30 {
			t.Errorf("Clone3 incorrect: expected C=30, got C=%v", val)
		}
	})

	t.Run("clone after removal", func(t *testing.T) {
		original := NewTree()
		original.Set("A", 1)
		original.Set("B", 2)
		original.Set("C", 3)

		// Remove a node and then clone
		original.Remove("B")
		clone := original.Clone()

		// Modify clone
		clone.Set("B", 20)

		// Verify original state
		if original.Has("B") {
			t.Error("Original tree should not have key B")
		}

		// Verify clone state
		if val, _ := clone.Get("B"); val != 20 {
			t.Errorf("Clone incorrect: expected B=20, got B=%v", val)
		}
	})

	t.Run("concurrent modifications", func(t *testing.T) {
		original := NewTree()
		original.Set("A", 1)
		original.Set("B", 2)

		clone1 := original.Clone()
		clone2 := original.Clone()

		// Modify same key in different clones
		clone1.Set("B", 20)
		clone2.Set("B", 30)

		// Each clone should have its own value
		if val, _ := clone1.Get("B"); val != 20 {
			t.Errorf("Clone1 incorrect: expected B=20, got B=%v", val)
		}
		if val, _ := clone2.Get("B"); val != 30 {
			t.Errorf("Clone2 incorrect: expected B=30, got B=%v", val)
		}
	})

	t.Run("deep tree modifications", func(t *testing.T) {
		original := NewTree()
		// Create a deeper tree
		keys := []string{"M", "F", "T", "B", "H", "P", "Z"}
		for _, k := range keys {
			original.Set(k, k)
		}

		clone := original.Clone()

		// Modify a deep node
		clone.Set("H", "modified")

		// Check original remains unchanged
		if val, _ := original.Get("H"); val != "H" {
			t.Errorf("Original modified: expected H='H', got H=%v", val)
		}

		// Verify clone modification
		if val, _ := clone.Get("H"); val != "modified" {
			t.Errorf("Clone incorrect: expected H='modified', got H=%v", val)
		}
	})

	t.Run("rebalancing test", func(t *testing.T) {
		original := NewTree()
		// Insert nodes that will cause rotations
		keys := []string{"A", "B", "C", "D", "E"}
		for _, k := range keys {
			original.Set(k, k)
		}

		clone := original.Clone()

		// Add more nodes to clone to trigger rebalancing
		clone.Set("F", "F")
		clone.Set("G", "G")

		// Verify original structure remains unchanged
		originalKeys := collectKeys(original)
		expectedOriginal := []string{"A", "B", "C", "D", "E"}
		if !slicesEqual(originalKeys, expectedOriginal) {
			t.Errorf("Original tree structure changed: got %v, want %v", originalKeys, expectedOriginal)
		}

		// Verify clone has all keys
		cloneKeys := collectKeys(clone)
		expectedClone := []string{"A", "B", "C", "D", "E", "F", "G"}
		if !slicesEqual(cloneKeys, expectedClone) {
			t.Errorf("Clone tree structure incorrect: got %v, want %v", cloneKeys, expectedClone)
		}
	})

	t.Run("value mutation test", func(t *testing.T) {
		type MutableValue struct {
			Data string
		}

		original := NewTree()
		mutable := &MutableValue{Data: "original"}
		original.Set("key", mutable)

		clone := original.Clone()

		// Modify the mutable value
		mutable.Data = "modified"

		// Both original and clone should see the modification
		// because we're not deep copying values
		origVal, _ := original.Get("key")
		cloneVal, _ := clone.Get("key")

		if origVal.(*MutableValue).Data != "modified" {
			t.Error("Original value not modified as expected")
		}
		if cloneVal.(*MutableValue).Data != "modified" {
			t.Error("Clone value not modified as expected")
		}
	})
}

// Helper function to collect all keys in order
func collectKeys(tree *Tree) []string {
	var keys []string
	tree.Iterate("", "", func(key string, _ any) bool {
		keys = append(keys, key)
		return false
	})
	return keys
}
