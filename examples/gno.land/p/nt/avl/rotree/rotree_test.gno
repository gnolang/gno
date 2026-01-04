package rotree

import (
	"testing"

	"gno.land/p/nt/avl"
)

func TestExample(t *testing.T) {
	// User represents our internal data structure
	type User struct {
		ID       string
		Name     string
		Balance  int
		Internal string // sensitive internal data
	}

	// Create and populate the original tree with user pointers
	tree := avl.NewTree()
	tree.Set("alice", &User{
		ID:       "1",
		Name:     "Alice",
		Balance:  100,
		Internal: "sensitive_data_1",
	})
	tree.Set("bob", &User{
		ID:       "2",
		Name:     "Bob",
		Balance:  200,
		Internal: "sensitive_data_2",
	})

	// Define a makeEntrySafeFn that:
	// 1. Creates a defensive copy of the User struct
	// 2. Omits sensitive internal data
	makeEntrySafeFn := func(v any) any {
		originalUser := v.(*User)
		return &User{
			ID:       originalUser.ID,
			Name:     originalUser.Name,
			Balance:  originalUser.Balance,
			Internal: "", // Omit sensitive data
		}
	}

	// Create a read-only view of the tree
	roTree := Wrap(tree, makeEntrySafeFn)

	// Test retrieving and verifying a user
	t.Run("Get User", func(t *testing.T) {
		// Get user from read-only tree
		value, exists := roTree.Get("alice")
		if !exists {
			t.Fatal("User 'alice' not found")
		}

		user := value.(*User)

		// Verify user data is correct
		if user.Name != "Alice" || user.Balance != 100 {
			t.Errorf("Unexpected user data: got name=%s balance=%d", user.Name, user.Balance)
		}

		// Verify sensitive data is not exposed
		if user.Internal != "" {
			t.Error("Sensitive data should not be exposed")
		}

		// Verify it's a different instance than the original
		originalValue, _ := tree.Get("alice")
		originalUser := originalValue.(*User)
		if user == originalUser {
			t.Error("Read-only tree should return a copy, not the original pointer")
		}
	})

	// Test iterating over users
	t.Run("Iterate Users", func(t *testing.T) {
		count := 0
		roTree.Iterate("", "", func(key string, value any) bool {
			user := value.(*User)
			// Verify each user has empty Internal field
			if user.Internal != "" {
				t.Error("Sensitive data exposed during iteration")
			}
			count++
			return false
		})

		if count != 2 {
			t.Errorf("Expected 2 users, got %d", count)
		}
	})

	// Verify that modifications to the returned user don't affect the original
	t.Run("Modification Safety", func(t *testing.T) {
		value, _ := roTree.Get("alice")
		user := value.(*User)

		// Try to modify the returned user
		user.Balance = 999
		user.Internal = "hacked"

		// Verify original is unchanged
		originalValue, _ := tree.Get("alice")
		originalUser := originalValue.(*User)
		if originalUser.Balance != 100 || originalUser.Internal != "sensitive_data_1" {
			t.Error("Original user data was modified")
		}
	})
}

func TestReadOnlyTree(t *testing.T) {
	// Example of a makeEntrySafeFn that appends "_readonly" to demonstrate transformation
	makeEntrySafeFn := func(value any) any {
		return value.(string) + "_readonly"
	}

	tree := avl.NewTree()
	tree.Set("key1", "value1")
	tree.Set("key2", "value2")
	tree.Set("key3", "value3")

	roTree := Wrap(tree, makeEntrySafeFn)

	tests := []struct {
		name     string
		key      string
		expected any
		exists   bool
	}{
		{"ExistingKey1", "key1", "value1_readonly", true},
		{"ExistingKey2", "key2", "value2_readonly", true},
		{"NonExistingKey", "key4", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, exists := roTree.Get(tt.key)
			if exists != tt.exists || value != tt.expected {
				t.Errorf("For key %s, expected %v (exists: %v), got %v (exists: %v)", tt.key, tt.expected, tt.exists, value, exists)
			}
		})
	}
}

// Add example tests showing different makeEntrySafeFn implementations
func TestMakeEntrySafeFnVariants(t *testing.T) {
	tree := avl.NewTree()
	tree.Set("slice", []int{1, 2, 3})
	tree.Set("map", map[string]int{"a": 1})

	tests := []struct {
		name            string
		makeEntrySafeFn func(any) any
		key             string
		validate        func(t *testing.T, value any)
	}{
		{
			name: "Defensive Copy Slice",
			makeEntrySafeFn: func(v any) any {
				original := v.([]int)
				return append([]int{}, original...)
			},
			key: "slice",
			validate: func(t *testing.T, value any) {
				slice := value.([]int)
				// Modify the returned slice
				slice[0] = 999
				// Verify original is unchanged
				originalValue, _ := tree.Get("slice")
				original := originalValue.([]int)
				if original[0] != 1 {
					t.Error("Original slice was modified")
				}
			},
		},
		// Add more test cases for different makeEntrySafeFn implementations
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roTree := Wrap(tree, tt.makeEntrySafeFn)
			value, exists := roTree.Get(tt.key)
			if !exists {
				t.Fatal("Key not found")
			}
			tt.validate(t, value)
		})
	}
}

func TestNilMakeEntrySafeFn(t *testing.T) {
	// Create a tree with some test data
	tree := avl.NewTree()
	originalValue := []int{1, 2, 3}
	tree.Set("test", originalValue)

	// Create a ReadOnlyTree with nil makeEntrySafeFn
	roTree := Wrap(tree, nil)

	// Test that we get back the original value
	value, exists := roTree.Get("test")
	if !exists {
		t.Fatal("Key not found")
	}

	// Verify it's the exact same slice (not a copy)
	retrievedSlice := value.([]int)
	if &retrievedSlice[0] != &originalValue[0] {
		t.Error("Expected to get back the original slice reference")
	}

	// Test through iteration as well
	roTree.Iterate("", "", func(key string, value any) bool {
		retrievedSlice := value.([]int)
		if &retrievedSlice[0] != &originalValue[0] {
			t.Error("Expected to get back the original slice reference in iteration")
		}
		return false
	})
}
