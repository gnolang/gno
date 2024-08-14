package avl

import (
	"gno.land/p/demo/ufmt"
)

// Indexable interface for various types
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

// Set inserts or updates a key-value pair in the main tree and updates indexes
func (tree *IndexedTree) Set(key string, value Indexable) (updated bool) {
	// Remove the old value from field indexes if it exists
	oldValue, exists := tree.mainTree.Get(key)
	if exists && oldValue != nil {
		oldValueIndexable, ok := oldValue.(Indexable)
		if ok {
			tree.updateIndexes(key, oldValueIndexable, value, false)
		}
	}

	// Set the new value in the main tree
	updated = tree.mainTree.Set(key, value)

	// Add the new value to field indexes
	if value != nil {
		tree.updateIndexes(key, nil, value, true)
	}

	return updated
}

// Remove deletes a key-value pair from the main tree and updates indexes
func (tree *IndexedTree) Remove(key string) (value Indexable, removed bool) {
	val, removed := tree.mainTree.Remove(key)
	if removed {
		if val != nil {
			value, _ = val.(Indexable)
			if value != nil {
				tree.updateIndexes(key, value, nil, false)
			}
		}
	}
	return value, removed
}

// updateIndexes updates or removes the field indexes based on changes in the value
func (tree *IndexedTree) updateIndexes(key string, oldValue, newValue Indexable, add bool) {
	for fieldName, index := range tree.indexes {

		// If there's an old value, remove it from the index if the field value changed
		if oldValue != nil {
			oldFieldValue := oldValue.Value(fieldName)
			if oldFieldValue != "" {
				index.Remove(ufmt.Sprintf("%s|%s", oldFieldValue, key))
			}
		}

		// Add the new value to the index if specified
		if add && newValue != nil {
			newFieldValue := newValue.Value(fieldName)
			if newFieldValue != "" {
				index.Set(ufmt.Sprintf("%s|%s", newFieldValue, key), key)
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

// QueryByField retrieves objects based on a specific field value
func (tree *IndexedTree) QueryByField(field string, value string) ([]Indexable, bool) {
	index, ok := tree.indexes[field]
	if !ok {
		return nil, false
	}

	var values []Indexable

	// Iterate through the index tree to find matching objects
	index.Iterate("", "", func(key string, obj interface{}) bool {
		// Retrieve the actual object from the main tree using the key
		if obj, exists := tree.mainTree.Get(obj.(string)); exists {
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
