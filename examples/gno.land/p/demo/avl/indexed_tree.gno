package avl

import (
	"gno.land/p/demo/ufmt"
)

// Indexable interface for various types.
//
// The `Indexable` interface is designed to be implemented by objects that
// will be stored and indexed within the `IndexedTree`. This interface
// facilitates field-based querying by exposing methods to retrieve the
// value of a specified field as a string and to identify the type of the struct.
//
// Methods:
//
// Value(fieldName string) string:
//   - This method is responsible for returning the value of the specified field
//     in the struct as a string. The purpose of this method is to allow the
//     `IndexedTree` to index and query based on the values of different fields.
//   - Parameters:
//   - fieldName: The name of the field whose value is to be retrieved.
//   - Returns:
//   - A string representation of the field's value. For fields of primitive
//     types (e.g., string, int, bool), the method should convert these values
//     to their string representations (e.g., using `ufmt.Sprintf` for formatting).
//   - If the field is unsupported or not a primitive type, the method should
//     return an empty string (""). This signals that the field cannot be indexed
//     or queried.
//
// Type() string:
//   - This method returns a unique type identifier for the struct. It is used
//     by the `IndexedTree` to distinguish between different types of structs
//     that may share the same field names.
//   - Returns:
//   - A string that uniquely identifies the type of the struct. This is important
//     when indexing and querying, as it ensures that different struct types are
//     handled separately even if they have fields with the same names.
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
//
//	func (m MyStruct) Type() string {
//	    return "MyStruct"
//	}
type Indexable interface {
	Value(fieldName string) string
	Type() string
}

// index is a wrapper around the AVL tree for indexing field values to keys
type index struct {
	field string // The field name associated with this index
	*Tree        // AVL tree for storing keys based on field values
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
func (tree *IndexedTree) IndexCount() int {
	return len(tree.indexes)
}

// Size returns the number of key-value pairs in the main tree.
func (tree *IndexedTree) Size() int {
	return tree.mainTree.Size()
}

// Has checks whether a key exists in the main tree.
func (tree *IndexedTree) Has(key string) (has bool) {
	return tree.mainTree.Has(key)
}

func (tree *IndexedTree) Set(key string, value Indexable) (updated bool) {
	// Set the new value in the main tree
	updated = tree.mainTree.Set(key, value)

	// Update indexes with the new value
	tree.updateIndexes(key, value, false)

	return updated
}

// Remove deletes a key-value pair from the main tree and updates indexes
func (tree *IndexedTree) Remove(key string) (value Indexable, removed bool) {
	// Retrieve the value associated with the key
	val, removed := tree.mainTree.Remove(key)
	if removed && val != nil {
		value, _ = val.(Indexable)
		// Remove the key's associated indexes
		tree.updateIndexes(key, value, true)
	}

	return value, removed
}

// updateIndexes handles adding or removing entries from the indexes
func (tree *IndexedTree) updateIndexes(key string, value Indexable, remove bool) {
	for indexkey, idx := range tree.indexes {
		// Extract the field name from the index key
		fieldName := idx.field
		fieldValue := value.Value(fieldName)
		if fieldValue != "" {
			// Use both the struct type and the field value for indexing
			compositeKey := indexkey + "|" + fieldValue + "|" + key
			if remove {
				idx.Remove(compositeKey)
			} else {
				idx.Set(compositeKey, key)
			}
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

// QueryByField retrieves objects based on a specific field value.
// It limits the iteration to keys that start with a specific prefix
// to improve the performance by reducing the number of keys scanned.
//
// Parameters:
// - typeName: The type of the object being queried.
// - field: The name of the field to query.
// - value: The value of the field to search for.
//
// Returns:
// - A slice of Indexable objects that match the query.
// - A boolean indicating if any matches were found.
func (tree *IndexedTree) QueryByField(typeName, field, value string) ([]Indexable, bool) {
	indexKey := getIndexKey(typeName, field)
	index, ok := tree.indexes[indexKey]
	if !ok {
		return nil, false
	}

	var values []Indexable
	// Use prefix-based iteration to limit the range of keys scanned
	start := indexKey + "|" + value
	end := start + string('|'+1)

	index.Iterate(start, end, func(key string, val interface{}) bool {
		// Retrieve the actual object from the main tree using the key
		if obj, exists := tree.mainTree.Get(val.(string)); exists {
			if idxObj, ok := obj.(Indexable); ok {
				values = append(values, idxObj)
			}
		}
		return false
	})

	return values, len(values) > 0
}

// NewIndex creates an index for a specific type and field if it doesn't already exist.
func (tree *IndexedTree) NewIndex(typeName, field string) {
	indexKey := getIndexKey(typeName, field)
	if _, exists := tree.indexes[indexKey]; !exists {
		tree.indexes[indexKey] = &index{
			field: field,
			Tree:  NewTree(),
		}
	}
}

// Iterate performs an in-order traversal of the main tree within the specified key range.
func (tree *IndexedTree) Iterate(start, end string, cb IterCbFn) bool {
	return tree.mainTree.Iterate(start, end, cb)
}

// ReverseIterate performs a reverse in-order traversal of the main tree within the specified key range.
func (tree *IndexedTree) ReverseIterate(start, end string, cb IterCbFn) bool {
	return tree.mainTree.ReverseIterate(start, end, cb)
}

// IterateByOffset performs an in-order traversal of the main tree starting from the specified offset.
func (tree *IndexedTree) IterateByOffset(offset int, count int, cb IterCbFn) bool {
	return tree.mainTree.IterateByOffset(offset, count, cb)
}

// ReverseIterateByOffset performs a reverse in-order traversal of the main tree starting from the specified offset.
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
		result += "  IndexKey: " + field + ", Count: " + ufmt.Sprintf("%d", count) + "\n"
	}

	return result
}

func getIndexKey(typeName, field string) string {
	return typeName + "|" + field
}
