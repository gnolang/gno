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

func (u user) Type() string {
	return "user"
}

type product struct {
	id    string
	name  string
	price uint64
}

func (p product) Value(fieldName string) string {
	switch fieldName {
	case "id":
		return p.id
	case "name":
		return p.name
	case "price":
		return ufmt.Sprintf("%d", p.price)
	default:
		return ""
	}
}

func (p product) Type() string {
	return "product"
}

// Helper function to create an IndexedTree with some initial indexes
func setupIndexedTree() *IndexedTree {
	tree := NewIndexedTree()
	tree.NewIndex("user", "name")
	tree.NewIndex("user", "age")
	tree.NewIndex("user", "email")
	tree.NewIndex("user", "male")
	tree.NewIndex("product", "id")
	tree.NewIndex("product", "price")
	return tree
}

// TestMultipleObjects tests insertion, querying, updating, and removal of multiple object types
func TestMultipleObjects(t *testing.T) {
	tree := setupIndexedTree()

	// Define and insert objects
	users := []user{
		{name: "Alice", age: 30, male: true, email: "alice@example.com"},
		{name: "Bob", age: 25, male: false, email: "bob@example.com"},
	}
	products := []product{
		{id: "prod1", name: "Laptop", price: 1000},
		{id: "prod2", name: "Smartphone", price: 500},
	}

	for i, u := range users {
		tree.Set(ufmt.Sprintf("userKey%d", i+1), u)
	}
	for i, p := range products {
		tree.Set(ufmt.Sprintf("productKey%d", i+1), p)
	}

	// Verify insertion
	verifyObject(t, tree, "userKey1", users[0])
	verifyObject(t, tree, "productKey1", products[0])

	// Query and verify
	verifyQueryByField(t, tree, "user", "age", "30", 1)
	verifyQueryByField(t, tree, "product", "price", "500", 1)

	// Update values
	users[0].age = 31
	users[0].email = "alice@newexample.com"
	tree.Set("userKey1", users[0])
	verifyObject(t, tree, "userKey1", users[0])

	products[0].name = "Macbook"
	products[0].price = 1200
	tree.Set("productKey1", products[0])
	verifyObject(t, tree, "productKey1", products[0])

	// Remove values
	verifyRemoval(t, tree, "userKey1")
	verifyRemoval(t, tree, "productKey1")
}

// TestSet tests insertion of a single value
func TestSet(t *testing.T) {
	tree := setupIndexedTree()

	obj := user{
		name:  "Alice",
		age:   30,
		male:  true,
		email: "alice@example.com",
	}
	tree.Set("key1", obj)

	// Verify insertion
	verifyObject(t, tree, "key1", obj)

	// Verify QueryByField for each indexed field
	verifyQueryByField(t, tree, "user", "name", "Alice", 1)
}

// TestUpdate tests updating an existing value
func TestUpdate(t *testing.T) {
	tree := setupIndexedTree()

	initial := user{
		name:  "Bob",
		age:   25,
		male:  false,
		email: "bob@example.com",
	}
	tree.Set("key1", initial)

	// Update the value
	updated := user{
		name:  "Bob",
		age:   26,
		male:  true,
		email: "bob@example.org",
	}
	tree.Set("key1", updated)

	// Verify update
	verifyObject(t, tree, "key1", updated)

	// Verify QueryByField after update
	verifyQueryByField(t, tree, "user", "age", "26", 1)
}

// TestRemove tests removing a value
func TestRemove(t *testing.T) {
	tree := setupIndexedTree()

	obj := user{name: "Charlie", age: 28, male: true, email: "charlie@example.com"}
	tree.Set("key3", obj)

	verifyRemoval(t, tree, "key3")
}

// TestQueryByField tests querying by a field
func TestQueryByField(t *testing.T) {
	tree := setupIndexedTree()

	users := []user{
		{name: "Alice", age: 30, male: true, email: "alice@example.com"},
		{name: "Bob", age: 30, male: false, email: "bob@example.com"},
	}

	for i, u := range users {
		tree.Set(ufmt.Sprintf("key%d", i+1), u)
	}

	verifyQueryByField(t, tree, "user", "age", "30", 2)
	verifyQueryByField(t, tree, "user", "male", "true", 1)
}

// TestQueryByNonExistentField tests querying by a non-existent field
func TestQueryByNonExistentField(t *testing.T) {
	tree := setupIndexedTree()

	obj := user{
		name:  "Alice",
		age:   30,
		male:  true,
		email: "alice@example.com",
	}
	tree.Set("key1", obj)

	// Query by non-existent field
	values, found := tree.QueryByField("user", "nonExistentField", "someValue")
	urequire.False(t, found, "Expected field 'nonExistentField' to not be found, but it was found")
	urequire.Equal(t, 0, len(values), ufmt.Sprintf("Expected 0 values for non-existent field, but got %d", len(values)))
}

// Benchmark tests for IndexedTree
func BenchmarkIndexedTreeQuery(b *testing.B) {
	tree := setupIndexedTree()
	for i := 0; i < 10000; i++ {
		obj := user{
			name:  ufmt.Sprintf("User%d", i),
			age:   i % 100,
			male:  i%2 == 0,
			email: ufmt.Sprintf("user%d@example.com", i),
		}
		tree.Set(ufmt.Sprintf("key%d", i), obj)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.QueryByField("user", "name", "User5000")
	}
}

func BenchmarkIndexedTreeRemove(b *testing.B) {
	tree := setupIndexedTree()
	for i := 0; i < 10000; i++ {
		obj := user{
			name:  ufmt.Sprintf("User%d", i),
			age:   i % 100,
			male:  i%2 == 0,
			email: ufmt.Sprintf("user%d@example.com", i),
		}
		tree.Set(ufmt.Sprintf("key%d", i), obj)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Remove(ufmt.Sprintf("key%d", i%10000))
	}
}

// Helper function to verify object retrieval
func verifyObject(t *testing.T, tree *IndexedTree, key string, expected interface{}) {
	value, exists := tree.Get(key)
	urequire.True(t, exists, "Expected value to be found, but it does not exist")

	switch v := expected.(type) {
	case user:
		u, ok := value.(user)
		urequire.True(t, ok, "Expected value to be of type user")
		urequire.Equal(t, v.name, u.name, "Expected name does not match")
		urequire.Equal(t, v.age, u.age, "Expected age does not match")
		urequire.Equal(t, v.male, u.male, "Expected male does not match")
		urequire.Equal(t, v.email, u.email, "Expected email does not match")
	case product:
		p, ok := value.(product)
		urequire.True(t, ok, "Expected value to be of type product")
		urequire.Equal(t, v.id, p.id, "Expected id does not match")
		urequire.Equal(t, v.name, p.name, "Expected name does not match")
		urequire.Equal(t, v.price, p.price, "Expected price does not match")
	default:
		t.Fatalf("Unsupported type: %T", expected)
	}
}

// Helper function to verify querying by field
func verifyQueryByField(t *testing.T, tree *IndexedTree, typeName, fieldName, fieldValue string, expectedCount int) {
	values, found := tree.QueryByField(typeName, fieldName, fieldValue)
	urequire.True(t, found, "Expected field '%s' to be found, but it was not found", fieldName)
	urequire.Equal(t, expectedCount, len(values), ufmt.Sprintf("Expected %d values for %s '%s', but got %d", expectedCount, fieldName, fieldValue, len(values)))
}

// Helper function to verify removal
func verifyRemoval(t *testing.T, tree *IndexedTree, key string) {
	_, removed := tree.Remove(key)
	urequire.True(t, removed, "Expected value to be removed, but it was not removed")
	_, exists := tree.Get(key)
	urequire.False(t, exists, "Expected value to be removed, but it still exists")
}
