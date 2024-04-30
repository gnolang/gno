package avl

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
	tree.Iterate("", "", func(key string, value interface{}) bool {
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
	tree.ReverseIterate("", "", func(key string, value interface{}) bool {
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
	tree.IterateByOffset(1, 2, func(key string, value interface{}) bool {
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
	tree.ReverseIterateByOffset(1, 2, func(key string, value interface{}) bool {
		keys = append(keys, key)
		return false
	})

	expectedKeys := []string{"key2", "key1"}
	if !slicesEqual(keys, expectedKeys) {
		t.Errorf("Expected keys %v, got %v", expectedKeys, keys)
	}
}
