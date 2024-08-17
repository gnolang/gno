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

	// Test inserting a value
	obj1 := user{
		name:  "Alice",
		age:   30,
		male:  true,
		email: "alice@example.com",
	}
	tree.Set("key1", obj1)

	// Verify insertion
	value, exists := tree.Get("key1")
	urequire.True(t, exists, "Expected value to be found, but it does not exist")

	u := value.(user)
	urequire.Equal(t, "Alice", u.name, "Expected name to be 'Alice'")
	urequire.Equal(t, 30, u.age, "Expected age to be 30")
	urequire.Equal(t, "alice@example.com", u.email, "Expected email to be 'alice@example.com'")
	urequire.Equal(t, true, u.male, "Expected male to be true")

	// Verify QueryByField for each indexed field
	values, found := tree.QueryByField("name", "Alice")
	urequire.True(t, found, "Expected field 'name' to be found, but it was not found")
	urequire.Equal(t, 1, len(values), ufmt.Sprintf("Expected 1 value for name 'Alice', but got %d", len(values)))
	queriedUser := values[0].(user)
	urequire.Equal(t, "Alice", queriedUser.name, "Expected queried name to be 'Alice'")
}

func TestUpdate(t *testing.T) {
	tree := NewIndexedTree()

	// Specify which fields to index
	tree.NewIndex("name")
	tree.NewIndex("age")
	tree.NewIndex("email")
	tree.NewIndex("male")

	// Insert initial value
	obj1 := user{
		name:  "Bob",
		age:   25,
		male:  false,
		email: "bob@example.com",
	}
	tree.Set("key1", obj1)

	// Update the value
	obj2 := user{
		name:  "Bob",
		age:   26,
		male:  true,
		email: "bob@example.org",
	}
	updated := tree.Set("key1", obj2)
	urequire.True(t, updated, "Expected value to be updated")

	// Verify update
	value, exists := tree.Get("key1")
	urequire.True(t, exists, "Expected value to be found, but it does not exist")

	u, ok := value.(user)
	urequire.True(t, ok, "Expected value to be of type user")
	urequire.Equal(t, "Bob", u.name, "Expected name to be 'Bob'")
	urequire.Equal(t, 26, u.age, "Expected age to be 26")
	urequire.Equal(t, "bob@example.org", u.email, "Expected email to be 'bob@example.org'")
	urequire.Equal(t, true, u.male, "Expected male to be true")

	// Verify QueryByField after update
	values, found := tree.QueryByField("name", "Bob")
	urequire.True(t, found, "Expected field 'name' to be found, but it was not found")
	urequire.Equal(t, 1, len(values), ufmt.Sprintf("Expected 1 value for name 'Bob', but got %d", len(values)))

	// Verify returned value from QueryByField
	queriedUser, ok := values[0].(user)
	urequire.True(t, ok, "Expected queried value to be of type user")
	urequire.Equal(t, "Bob", queriedUser.name, "Expected queried name to be 'Bob'")
	urequire.Equal(t, 26, queriedUser.age, "Expected queried age to be 26")
	urequire.Equal(t, "bob@example.org", queriedUser.email, "Expected queried email to be 'bob@example.org'")
	urequire.Equal(t, true, queriedUser.male, "Expected queried male to be true")

	// Additional QueryByField checks for other fields
	values, found = tree.QueryByField("age", "26")
	urequire.True(t, found, "Expected field 'age' to be found, but it was not found")
	urequire.Equal(t, 1, len(values), ufmt.Sprintf("Expected 1 value for age 26, but got %d", len(values)))

	values, found = tree.QueryByField("email", "bob@example.org")
	urequire.True(t, found, "Expected field 'email' to be found, but it was not found")
	urequire.Equal(t, 1, len(values), ufmt.Sprintf("Expected 1 value for email 'bob@example.org', but got %d", len(values)))

	values, found = tree.QueryByField("male", "true")
	urequire.True(t, found, "Expected field 'male' to be found, but it was not found")
	urequire.Equal(t, 1, len(values), ufmt.Sprintf("Expected 1 value for male 'true', but got %d", len(values)))
}

func TestRemove(t *testing.T) {
	tree := NewIndexedTree()

	// Specify which fields to index
	tree.NewIndex("name")
	tree.NewIndex("age")
	tree.NewIndex("email")
	tree.NewIndex("male")

	// Insert value to be removed
	obj1 := user{name: "Charlie", age: 28, male: true, email: "charlie@example.com"}
	tree.Set("key3", obj1)

	// Remove the value
	value, removed := tree.Remove("key3")
	urequire.True(t, removed, "Expected value to be removed, but it was not removed")

	// Verify removal
	_, exists := tree.Get("key3")
	urequire.False(t, exists, "Expected value to be removed, but it still exists")
}

func TestQueryByField(t *testing.T) {
	tree := NewIndexedTree()

	// Specify which fields to index
	tree.NewIndex("age")
	tree.NewIndex("male")

	// Insert values
	obj1 := user{
		name:  "Alice",
		age:   30,
		male:  true,
		email: "alice@example.com",
	}

	obj2 := user{
		name:  "Bob",
		age:   30,
		male:  false,
		email: "bob@example.com",
	}

	tree.Set("key1", obj1)
	tree.Set("key2", obj2)

	// Query by age
	values, found := tree.QueryByField("age", "30")
	urequire.True(t, found, "Expected field 'age' to be found, but it was not found")
	urequire.Equal(t, 2, len(values), ufmt.Sprintf("Expected 2 values for age 30, but got %d", len(values)))

	// Verify returned values
	expectedUsers := []user{
		{name: "Alice", age: 30, male: true, email: "alice@example.com"},
		{name: "Bob", age: 30, male: false, email: "bob@example.com"},
	}

	for i, v := range values {
		u, ok := v.(user)
		urequire.True(t, ok, "Expected value to be of type user")
		urequire.Equal(t, expectedUsers[i].name, u.name)
		urequire.Equal(t, expectedUsers[i].age, u.age)
		urequire.Equal(t, expectedUsers[i].male, u.male)
		urequire.Equal(t, expectedUsers[i].email, u.email)
	}

	// Query by male
	values, found = tree.QueryByField("male", "true")
	urequire.True(t, found, "Expected field 'male' to be found, but it was not found")
	urequire.Equal(t, 1, len(values), ufmt.Sprintf("Expected 1 value for male true, but got %d", len(values)))

	u := values[0].(user)
	urequire.Equal(t, "Alice", u.name, "Expected name to be 'Alice'")
	urequire.Equal(t, 30, u.age, "Expected age to be 30")
	urequire.Equal(t, true, u.male, "Expected male to be true")
	urequire.Equal(t, "alice@example.com", u.email, "Expected email to be 'alice@example.com'")
}

func TestQueryByNonExistentField(t *testing.T) {
	tree := NewIndexedTree()

	// Specify which fields to index
	tree.NewIndex("name")

	// Insert values
	obj1 := user{
		name:  "Alice",
		age:   30,
		male:  true,
		email: "alice@example.com",
	}
	tree.Set("key1", obj1)

	// Query by non-existent field
	values, found := tree.QueryByField("nonExistentField", "someValue")
	urequire.False(t, found, "Expected field 'nonExistentField' to not be found, but it was found")
	urequire.Equal(t, 0, len(values), ufmt.Sprintf("Expected 0 values for non-existent field, but got %d", len(values)))
}

func BenchmarkIndexedTreeQuery(b *testing.B) {
	tree := NewIndexedTree()

	// Specify which fields to index
	tree.NewIndex("name")
	tree.NewIndex("age")

	// Create a large number of values
	for i := 0; i < 10000; i++ {
		user := user{
			name:  ufmt.Sprintf("User%d", i),
			age:   i % 100,
			male:  i%2 == 0,
			email: ufmt.Sprintf("user%d@example.com", i),
		}

		tree.Set(ufmt.Sprintf("key%d", i), user)
	}

	// Benchmark querying by name
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.QueryByField("name", "User5000")
	}
}

func BenchmarkIndexedTreeRemove(b *testing.B) {
	tree := NewIndexedTree()

	// Specify which fields to index
	tree.NewIndex("name")

	// Create a large number of values
	for i := 0; i < 10000; i++ {
		user := user{
			name:  ufmt.Sprintf("User%d", i),
			age:   i % 100,
			male:  i%2 == 0,
			email: ufmt.Sprintf("user%d@example.com", i),
		}

		tree.Set(ufmt.Sprintf("key%d", i), user)
	}

	// Benchmark removing values
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Remove(ufmt.Sprintf("key%d", i%10000))
	}
}

// TestRender tests rendering a tree with multiple indexes
func TestRender(t *testing.T) {
	tree := NewIndexedTree()
	tree.NewIndex("name")
	tree.NewIndex("age")

	user1 := user{
		name:  "Alice",
		age:   30,
		male:  true,
		email: "alice@example.com",
	}

	user2 := user{
		name:  "Bob",
		age:   25,
		male:  false,
		email: "bob@example.com",
	}

	tree.Set("key1", user1)
	tree.Set("key2", user2)

	expected := "IndexedTree Status:\nMain Tree Size: 2\nIndex Count: 2\nIndexes:\n  Field: name, Count: 2\n  Field: age, Count: 2\n"
	urequire.Equal(t, expected, tree.Render())
}
