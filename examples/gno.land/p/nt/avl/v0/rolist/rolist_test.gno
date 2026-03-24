package rolist

import (
	"testing"

	"gno.land/p/nt/avl/v0/list"
)

func TestExample(t *testing.T) {
	// User represents our internal data structure
	type User struct {
		ID       string
		Name     string
		Balance  int
		Internal string // sensitive internal data
	}

	// Create and populate the original list
	l := &list.List{}
	l.Append(
		&User{
			ID:       "1",
			Name:     "Alice",
			Balance:  100,
			Internal: "sensitive_data_1",
		},
		&User{
			ID:       "2",
			Name:     "Bob",
			Balance:  200,
			Internal: "sensitive_data_2",
		},
	)

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

	// Create a read-only view of the list
	roList := Wrap(l, makeEntrySafeFn)

	// Test retrieving and verifying a user
	t.Run("Get User", func(t *testing.T) {
		// Get user from read-only list
		value := roList.Get(0)
		if value == nil {
			t.Fatal("User at index 0 not found")
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
		originalUser := l.Get(0).(*User)
		if user == originalUser {
			t.Error("Read-only list should return a copy, not the original pointer")
		}
	})

	// Test slice functionality
	t.Run("Slice Users", func(t *testing.T) {
		users := roList.Slice(0, 2)
		if len(users) != 2 {
			t.Fatalf("Expected 2 users, got %d", len(users))
		}

		for _, v := range users {
			user := v.(*User)
			if user.Internal != "" {
				t.Error("Sensitive data exposed in slice")
			}
		}
	})

	// Test ForEach functionality
	t.Run("ForEach Users", func(t *testing.T) {
		count := 0
		roList.ForEach(func(index int, value any) bool {
			user := value.(*User)
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
}

func TestNilMakeEntrySafeFn(t *testing.T) {
	// Create a list with some test data
	l := &list.List{}
	originalValue := []int{1, 2, 3}
	l.Append(originalValue)

	// Create a ReadOnlyList with nil makeEntrySafeFn
	roList := Wrap(l, nil)

	// Test that we get back the original value
	value := roList.Get(0)
	if value == nil {
		t.Fatal("Value not found")
	}

	// Verify it's the exact same slice (not a copy)
	retrievedSlice := value.([]int)
	if &retrievedSlice[0] != &originalValue[0] {
		t.Error("Expected to get back the original slice reference")
	}
}

func TestReadOnlyList(t *testing.T) {
	// Example of a makeEntrySafeFn that appends "_readonly" to demonstrate transformation
	makeEntrySafeFn := func(value any) any {
		return value.(string) + "_readonly"
	}

	l := &list.List{}
	l.Append("value1", "value2", "value3")

	roList := Wrap(l, makeEntrySafeFn)

	tests := []struct {
		name     string
		index    int
		expected any
	}{
		{"ExistingIndex0", 0, "value1_readonly"},
		{"ExistingIndex1", 1, "value2_readonly"},
		{"NonExistingIndex", 3, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := roList.Get(tt.index)
			if value != tt.expected {
				t.Errorf("For index %d, expected %v, got %v", tt.index, tt.expected, value)
			}
		})
	}
}
