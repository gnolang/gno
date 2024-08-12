package avl

import (
	"testing"

	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
)

// User struct implementing Indexable interface
type User struct {
	Name  string
	Age   int
	Email string
}

func (u User) GetFields() map[string]string {
	return map[string]string{
		"Name":  u.Name,
		"Age":   ufmt.Sprintf("%d", u.Age),
		"Email": u.Email,
	}
}

func TestSet(t *testing.T) {
	tree := NewIndexedTree()

	// Specify which fields to index
	tree.AddFieldToIndex("Name")
	tree.AddFieldToIndex("Age")

	// Test inserting a value
	obj1 := User{Name: "Alice", Age: 30, Email: "alice@example.com"}
	tree.Set("key1", obj1)

	// Verify insertion
	value, exists := tree.Get("key1")
	urequire.True(t, exists, "Expected value to be found, but it does not exist")
	// urequire.NotNil(t, value, "Expected value to be non-nil, but got nil")

	user := value.(User)
	urequire.Equal(t, "Alice", user.Name, "Expected Name to be 'Alice'")
	urequire.Equal(t, 30, user.Age, "Expected Age to be 30")
	urequire.Equal(t, "alice@example.com", user.Email, "Expected Email to be 'alice@example.com'")
}

func TestUpdate(t *testing.T) {
	tree := NewIndexedTree()

	// Specify which fields to index
	tree.AddFieldToIndex("Name")
	tree.AddFieldToIndex("Age")

	// Insert initial value
	obj1 := User{Name: "Bob", Age: 25, Email: "bob@example.com"}
	tree.Set("key2", obj1)

	// Update the value
	obj2 := User{Name: "Bob", Age: 26, Email: "bob@example.org"}
	updated := tree.Set("key2", obj2)
	urequire.True(t, updated, "Expected value to be updated")

	// Verify update
	value, exists := tree.Get("key2")
	urequire.True(t, exists, "Expected value to be found, but it does not exist")

	user := value.(User)
	urequire.Equal(t, 26, user.Age, "Expected Age to be 26")
	urequire.Equal(t, "bob@example.org", user.Email, "Expected Email to be 'bob@example.org'")
}

func TestRemove(t *testing.T) {
	tree := NewIndexedTree()

	// Specify which fields to index
	tree.AddFieldToIndex("Name")

	// Insert value to be removed
	obj1 := User{Name: "Charlie", Age: 28, Email: "charlie@example.com"}
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
	tree.AddFieldToIndex("Age")

	// Insert values
	obj1 := User{Name: "Alice", Age: 30, Email: "alice@example.com"}
	obj2 := User{Name: "Bob", Age: 30, Email: "bob@example.com"}
	tree.Set("key1", obj1)
	tree.Set("key2", obj2)

	// Query by Age
	values, found := tree.QueryByField("Age", "30")
	urequire.True(t, found, "Expected field 'Age' to be found, but it was not found")
	urequire.Equal(t, 2, len(values), ufmt.Sprintf("Expected 2 values for Age 30, but got %d", len(values)))

	// Verify returned values
	expectedNames := map[string]bool{"Alice": true, "Bob": true}
	for _, v := range values {
		user, ok := v.(User)
		urequire.True(t, ok, "Expected value to be of type User")
		// urequire.True(t, expectedNames[user.Name], "Unexpected user: %v", user)
	}
}

func TestQueryByNonExistentField(t *testing.T) {
	tree := NewIndexedTree()

	// Specify which fields to index
	tree.AddFieldToIndex("Name")

	// Insert values
	obj1 := User{Name: "Alice", Age: 30, Email: "alice@example.com"}
	tree.Set("key1", obj1)

	// Query by non-existent field
	values, found := tree.QueryByField("NonExistentField", "someValue")
	urequire.False(t, found, "Expected field 'NonExistentField' to not be found, but it was found")
	urequire.Equal(t, 0, len(values), ufmt.Sprintf("Expected 0 values for non-existent field, but got %d", len(values)))
}
