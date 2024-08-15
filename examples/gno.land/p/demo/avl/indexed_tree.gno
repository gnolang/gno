package avl

import (
	"gno.land/p/demo/ufmt"
)

// Indexable interface for various types.
//
// The `Indexable` interface is designed to be implemented by objects that
// will be stored and indexed within the `IndexedTree`. This interface
// facilitates field-based querying by exposing a method to retrieve the
// value of a specified field as a string.
//
// Restrictions:
//   - Only primitive data types (string, int, bool, ...) are supported for indexing.
//   - The `Value` method must return the value of the specified field as a string.
//     For fields of type int or bool, the method should convert these values to their
//     string representations using appropriate formatting (e.g., using `ufmt.Sprintf`).
//   - The `Value` method should return an empty string ("") for unsupported or
//     non-primitive field types, signaling that such fields cannot be indexed or queried.
//
// Example Implementation:
//
//	type MyStruct struct {
//	    Name  string
//	    Age   int
//	    IsMale bool
//	}
//
//	func (m MyStruct) Value(fieldName string) string {
//	    switch fieldName {
//	    case "Name":
//	        return m.Name
//	    case "Age":
//	        return ufmt.Sprintf("%d", m.Age)
//	    case "IsMale":
//	        return ufmt.Sprintf("%t", m.IsMale)
//	    default:
//	        return "" // Unsupported field or non-primitive type
//	    }
//	}
type Indexable interface {
	Value(fieldName string) string
}

// index is a wrapper around the AVL tree for indexing field values to keys
type index struct {
	*Tree // AVL tree for storing keys based on field values
}

// IndexedTree manages the main AVL tree and additional field-based indexes
type IndexedTree struct {
	mainTree *Tree
	indexes  map[string]*index
}

// NewIndexedTree creates a new empty IndexedTree
func NewIndexedTree() *IndexedTree {
	return &IndexedTree{
		mainTree: NewTree(),
		indexes:  make(map[string]*index),
	}
}

// IndexCount returns the number of indexes currently in the IndexedTree.
// This method can be used to check how many fields have been indexed.
func (tree *IndexedTree) IndexCount() int {
	return len(tree.indexes)
}

// Size returns the number of key-value pair in the main tree.
func (tree *IndexedTree) Size() int {
	return tree.mainTree.Size()
}

// Has checks whether a key exists in the main tree.
// It returns true if the key exists, otherwise false.
func (tree *IndexedTree) Has(key string) (has bool) {
	return tree.mainTree.Has(key)
}

// Set inserts or updates a key-value pair in the main tree and updates indexes
func (tree *IndexedTree) Set(key string, value Indexable) (updated bool) {
	// Set the new value in the main tree
	updated = tree.mainTree.Set(key, value)

	// Update indexes with the new value
	tree.updateIndexes(key, value)

	return updated
}

// Remove deletes a key-value pair from the main tree and updates indexes
func (tree *IndexedTree) Remove(key string) (value Indexable, removed bool) {
	val, removed := tree.mainTree.Remove(key)
	if removed {
		value, _ = val.(Indexable)
		tree.updateIndexes(key, nil)
	}

	return value, removed
}

// updateIndexes handles adding or removing entries from the indexes
func (tree *IndexedTree) updateIndexes(key string, newValue Indexable) {
	for fieldName, index := range tree.indexes {
		if newValue != nil {
			newFieldValue := newValue.Value(fieldName)
			if newFieldValue != "" {
				index.Set(ufmt.Sprintf("%s|%s", newFieldValue, key), key)
			}
		} else {
			// If newValue is nil, remove the key from the indexes
			index.Remove(ufmt.Sprintf("%s|%s", newValue.Value(fieldName), key))
		}
	}
}

// Get retrieves the value associated with the given key from the main tree
func (tree *IndexedTree) Get(key string) (value Indexable, exists bool) {
	v, exists := tree.mainTree.Get(key)
	if exists && v != nil {
		value, _ = v.(Indexable)
	}
	return value, exists
}

// QueryByField retrieves objects based on a specific field value
func (tree *IndexedTree) QueryByField(field string, value string) ([]Indexable, bool) {
	index, ok := tree.indexes[field]
	if !ok {
		return nil, false
	}

	var values []Indexable

	// Iterate through the index tree to find matching objects
	index.Iterate("", "", func(key string, val interface{}) bool {
		// Retrieve the actual object from the main tree using the key
		if obj, exists := tree.mainTree.Get(val.(string)); exists {
			if idxObj, ok := obj.(Indexable); ok {
				if idxObj.Value(field) == value {
					values = append(values, idxObj)
				}
			}
		}
		return false
	})

	return values, len(values) > 0
}

// AddFieldToIndex creates an index for a specific field if it doesn't already exist
func (tree *IndexedTree) NewIndex(field string) {
	if _, exists := tree.indexes[field]; !exists {
		tree.indexes[field] = &index{Tree: NewTree()}
	}
}

// Iterate performs an in-order traversal of the main tree within the specified key range.
// It calls the provided callback function for each key-value pair encountered.
// If the callback returns true, the iteration is stopped.
func (tree *IndexedTree) Iterate(start, end string, cb IterCbFn) bool {
	return tree.mainTree.Iterate(start, end, cb)
}

// ReverseIterate performs a reverse in-order traversal of the main tree within the specified key range.
// It calls the provided callback function for each key-value pair encountered.
// If the callback returns true, the iteration is stopped.
func (tree *IndexedTree) ReverseIterate(start, end string, cb IterCbFn) bool {
	return tree.mainTree.ReverseIterate(start, end, cb)
}

// IterateByOffset performs an in-order traversal of the main tree starting from the specified offset.
// It calls the provided callback function for each key-value pair encountered, up to the specified count.
// If the callback returns true, the iteration is stopped.
func (tree *IndexedTree) IterateByOffset(offset int, count int, cb IterCbFn) bool {
	return tree.mainTree.IterateByOffset(offset, count, cb)
}

// ReverseIterateByOffset performs a reverse in-order traversal of the main tree starting from the specified offset.
// It calls the provided callback function for each key-value pair encountered, up to the specified count.
// If the callback returns true, the iteration is stopped.
func (tree *IndexedTree) ReverseIterateByOffset(offset int, count int, cb IterCbFn) bool {
	return tree.mainTree.ReverseIterateByOffset(offset, count, cb)
}

// Render provides a string representation of the overall status of the IndexedTree
func (tree *IndexedTree) Render() string {
	result := "IndexedTree Status:\n"
	result += "Main Tree Size: " + ufmt.Sprintf("%d", tree.Size()) + "\n"
	result += "Index Count: " + ufmt.Sprintf("%d", tree.IndexCount()) + "\n"

	result += "Indexes:\n"
	for field, index := range tree.indexes {
		count := index.Size()
		result += "  Field: " + field + ", Count: " + ufmt.Sprintf("%d", count) + "\n"
	}

	return result
}
