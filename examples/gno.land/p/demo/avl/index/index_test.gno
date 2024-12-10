package index

import (
	"strconv"
	"testing"
)

type Person struct {
	ID   string
	Name string
	Age  int
}

type InvalidPerson struct {
	ID string
}

func TestIndexedTreeComprehensive(t *testing.T) {
	// Test 1: Basic operations without any indexes
	t.Run("NoIndexes", func(t *testing.T) {
		tree := NewIndexedTree()
		p1 := &Person{ID: "1", Name: "Alice", Age: 30}

		// Test Set and Get using primary tree directly
		tree.Set("1", p1)
		val, exists := tree.GetPrimary().Get("1")
		if !exists || val.(*Person).Name != "Alice" {
			t.Error("Basic Get failed without indexes")
		}

		// Test direct tree iteration
		count := 0
		tree.GetPrimary().Iterate("", "", func(key string, value interface{}) bool {
			count++
			return false
		})
		if count != 1 {
			t.Error("Basic iteration failed")
		}
	})

	// Test 2: Multiple indexes on same field
	t.Run("DuplicateIndexes", func(t *testing.T) {
		tree := NewIndexedTree()
		tree.AddIndex("age1", func(v interface{}) string {
			return strconv.Itoa(v.(*Person).Age)
		})
		tree.AddIndex("age2", func(v interface{}) string {
			return strconv.Itoa(v.(*Person).Age)
		})

		p1 := &Person{ID: "1", Name: "Alice", Age: 30}
		p2 := &Person{ID: "2", Name: "Bob", Age: 30}

		tree.Set("1", p1)
		tree.Set("2", p2)

		// Both indexes should return the same results
		results1 := tree.GetByIndex("age1", "30")
		results2 := tree.GetByIndex("age2", "30")

		if len(results1) != 2 || len(results2) != 2 {
			t.Error("Duplicate indexes returned different results")
		}
	})

	// Test 3: Invalid extractor
	t.Run("InvalidExtractor", func(t *testing.T) {
		didPanic := false

		func() {
			defer func() {
				if r := recover(); r != nil {
					didPanic = true
				}
			}()

			tree := NewIndexedTree()
			tree.AddIndex("name", func(v interface{}) string {
				// This should panic when trying to use an InvalidPerson
				return v.(*Person).Name // Intentionally wrong type assertion
			})

			invalid := &InvalidPerson{ID: "1"}
			tree.Set("1", invalid) // This should trigger the panic
		}()

		if !didPanic {
			t.Error("Expected panic from invalid type")
		}
	})

	// Test 4: Mixed usage of indexed and direct access
	t.Run("MixedUsage", func(t *testing.T) {
		tree := NewIndexedTree()
		tree.AddIndex("age", func(v interface{}) string {
			return strconv.Itoa(v.(*Person).Age)
		})

		p1 := &Person{ID: "1", Name: "Alice", Age: 30}

		// Use Set instead of direct tree access to ensure indexes are updated
		tree.Set("1", p1)

		// Index should work
		results := tree.GetByIndex("age", "30")
		if len(results) != 1 {
			t.Error("Index failed after direct tree usage")
		}
	})

	// Test 5: Using index as TreeInterface
	t.Run("IndexAsTreeInterface", func(t *testing.T) {
		tree := NewIndexedTree()
		tree.AddIndex("age", func(v interface{}) string {
			return strconv.Itoa(v.(*Person).Age)
		})

		p1 := &Person{ID: "1", Name: "Alice", Age: 30}
		p2 := &Person{ID: "2", Name: "Bob", Age: 30}

		tree.Set("1", p1)
		tree.Set("2", p2)

		// Get the index as TreeInterface
		ageIndex := tree.GetIndexTree("age")
		if ageIndex == nil {
			t.Error("Failed to get index as TreeInterface")
		}

		// Use the interface methods
		val, exists := ageIndex.Get("30")
		if !exists {
			t.Error("Failed to get value through index interface")
		}

		// The value should be a []string of primary keys
		primaryKeys := val.([]string)
		if len(primaryKeys) != 2 {
			t.Error("Wrong number of primary keys in index")
		}
	})

	// Test 6: Remove operations
	t.Run("RemoveOperations", func(t *testing.T) {
		tree := NewIndexedTree()
		tree.AddIndex("age", func(v interface{}) string {
			return strconv.Itoa(v.(*Person).Age)
		})

		p1 := &Person{ID: "1", Name: "Alice", Age: 30}
		tree.Set("1", p1)

		// Remove and verify both primary and index
		tree.Remove("1")

		if _, exists := tree.GetPrimary().Get("1"); exists {
			t.Error("Entry still exists in primary after remove")
		}

		results := tree.GetByIndex("age", "30")
		if len(results) != 0 {
			t.Error("Entry still exists in index after remove")
		}
	})

	// Test 7: Update operations
	t.Run("UpdateOperations", func(t *testing.T) {
		tree := NewIndexedTree()
		tree.AddIndex("age", func(v interface{}) string {
			return strconv.Itoa(v.(*Person).Age)
		})

		p1 := &Person{ID: "1", Name: "Alice", Age: 30}
		tree.Set("1", p1)

		// Update age using the new Update method
		p1New := &Person{ID: "1", Name: "Alice", Age: 31}
		tree.Update("1", p1, p1New)

		// Check old index is removed
		results30 := tree.GetByIndex("age", "30")
		if len(results30) != 0 {
			t.Error("Old index entry still exists")
		}

		// Check new index is added
		results31 := tree.GetByIndex("age", "31")
		if len(results31) != 1 {
			t.Error("New index entry not found")
		}
	})
}
