package avl

import (
	"testing"

	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
)

// user struct implementing Indexable interface
type user struct {
	name  string
	age   int
	male  bool
	email string
}

func (u user) Value(fieldName string) string {
	switch fieldName {
	case "name":
		return u.name
	case "age":
		return ufmt.Sprintf("%d", u.age)
	case "email":
		return u.email
	case "male":
		return ufmt.Sprintf("%t", u.male)
	default:
		return ""
	}
}

func TestSet(t *testing.T) {
	tree := NewIndexedTree()

	// Specify which fields to index
	tree.NewIndex("name")
	tree.NewIndex("age")

	// Test inserting a value
	obj1 := user{name: "Alice", age: 30, email: "alice@example.com"}
	tree.Set("key1", obj1)

	// Verify insertion
	value, exists := tree.Get("key1")
	urequire.True(t, exists, "Expected value to be found, but it does not exist")

	u := value.(user)
	urequire.Equal(t, "Alice", u.Name, "Expected Name to be 'Alice'")
	urequire.Equal(t, 30, u.age, "Expected age to be 30")
	urequire.Equal(t, "alice@example.com", u.email, "Expected email to be 'alice@example.com'")
}

func TestUpdate(t *testing.T) {
	tree := NewIndexedTree()

	// Specify which fields to index
	tree.NewIndex("Name")
	tree.NewIndex("age")

	// Insert initial value
	obj1 := user{Name: "Bob", age: 25, email: "bob@example.com"}
	tree.Set("key2", obj1)

	// Update the value
	obj2 := user{Name: "Bob", age: 26, email: "bob@example.org"}
	updated := tree.Set("key2", obj2)
	urequire.True(t, updated, "Expected value to be updated")

	// Verify update
	value, exists := tree.Get("key2")
	urequire.True(t, exists, "Expected value to be found, but it does not exist")

	u := value.(user)
	urequire.Equal(t, 26, u.age, "Expected age to be 26")
	urequire.Equal(t, "bob@example.org", u.email, "Expected email to be 'bob@example.org'")
}

func TestRemove(t *testing.T) {
	tree := NewIndexedTree()

	// Specify which fields to index
	tree.NewIndex("Name")

	// Insert value to be removed
	obj1 := User{Name: "Charlie", age: 28, email: "charlie@example.com"}
	tree.Set("key3", obj1)

	// Remove the value
	value, removed := tree.Remove("key3")
	urequire.True(t, removed, "Expected value to be removed, but it was not removed")
	// urequire.NotNil(t, value, "Expected removed value to be non-nil")

	// Verify removal
	_, exists := tree.Get("key3")
	urequire.False(t, exists, "Expected value to be removed, but it still exists")
}

func TestQueryByField(t *testing.T) {
	tree := NewIndexedTree()

	// Specify which fields to index
	tree.NewIndex("age")

	// Insert values
	obj1 := user{Name: "Alice", age: 30, email: "alice@example.com"}
	obj2 := user{Name: "Bob", age: 30, email: "bob@example.com"}
	tree.Set("key1", obj1)
	tree.Set("key2", obj2)

	// Query by age
	values, found := tree.QueryByField("age", "30")
	urequire.True(t, found, "Expected field 'age' to be found, but it was not found")
	urequire.Equal(t, 2, len(values), ufmt.Sprintf("Expected 2 values for age 30, but got %d", len(values)))

	// Verify returned values
	expectedNames := []user{
		{
			Name: "Alice",
		},
	}
	for i, v := range values {
		u, ok := v.(User)
		urequire.True(t, ok, "Expected value to be of type User")
		urequire.Equal(t, expectedNames[i].Name, u.name)
	}
}

func TestQueryByNonExistentField(t *testing.T) {
	tree := NewIndexedTree()

	// Specify which fields to index
	tree.NewIndex("Name")

	// Insert values
	obj1 := User{Name: "Alice", age: 30, email: "alice@example.com"}
	tree.Set("key1", obj1)

	// Query by non-existent field
	values, found := tree.QueryByField("NonExistentField", "someValue")
	urequire.False(t, found, "Expected field 'NonExistentField' to not be found, but it was found")
	urequire.Equal(t, 0, len(values), ufmt.Sprintf("Expected 0 values for non-existent field, but got %d", len(values)))
}

func BenchmarkIndexedTree(b *testing.B) {
	tree := NewIndexedTree()

	// Specify which fields to index
	tree.NewIndex("Name")

	// Create a large number of values
	for i := 0; i < 10000; i++ {
		user := User{Name: ufmt.Sprintf("User%d", i), age: i % 100, email: ufmt.Sprintf("user%d@example.com", i)}
		tree.Set(ufmt.Sprintf("key%d", i), user)
	}

	// Benchmark querying by Name
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.QueryByField("Name", "User5000")
	}
}

// TestPerformanceRemove benchmarks the performance of removing values from the IndexedTree.
func BenchmarkIndexedTreeRemove(b *testing.B) {
	tree := NewIndexedTree()

	// Specify which fields to index
	tree.NewIndex("Name")

	// Create a large number of values
	for i := 0; i < 10000; i++ {
		user := User{Name: ufmt.Sprintf("User%d", i), age: i % 100, email: ufmt.Sprintf("user%d@example.com", i)}
		tree.Set(ufmt.Sprintf("key%d", i), user)
	}

	// Benchmark removing values
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Remove(ufmt.Sprintf("key%d", i%10000))
	}
}
