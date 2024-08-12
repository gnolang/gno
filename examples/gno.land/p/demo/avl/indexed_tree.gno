package avl

import (
	"gno.land/p/demo/ufmt"
)

// Indexable interface for various types
type Indexable interface {
	GetFields() map[string]string
}

// FieldIndex keeps track of field values to keys
type FieldIndex struct {
	index *Tree // AVL tree for field values to keys
}

// IndexedTree manages the main tree and field indexes
type IndexedTree struct {
	mainTree     *Tree
	fieldIndexes map[string]*FieldIndex
}

// NewIndexedTree creates a new empty indexed tree
func NewIndexedTree() *IndexedTree {
	return &IndexedTree{
		mainTree:     NewTree(),
		fieldIndexes: make(map[string]*FieldIndex),
	}
}

// Set inserts or updates a key-value pair
func (tree *IndexedTree) Set(key string, value Indexable) (updated bool) {
	// Remove the old value from field indexes
	oldValue, exists := tree.mainTree.Get(key)
	if exists && oldValue != nil {
		oldValueIndexable, ok := oldValue.(Indexable)
		if ok {
			tree.updateFieldIndexes(key, oldValueIndexable, value, false)
		}
	}

	// Set the new value in the main tree
	updated = tree.mainTree.Set(key, value)

	// Add the new value to field indexes
	if value != nil {
		tree.updateFieldIndexes(key, nil, value, true)
	}

	return updated
}

// Remove removes a key-value pair
func (tree *IndexedTree) Remove(key string) (value Indexable, removed bool) {
	val, removed := tree.mainTree.Remove(key)
	if removed {
		if val != nil {
			value, _ = val.(Indexable)
			if value != nil {
				tree.updateFieldIndexes(key, value, nil, false)
			}
		}
	}
	return value, removed
}

// updateFieldIndexes updates or removes the field indexes based on the change in value
func (tree *IndexedTree) updateFieldIndexes(key string, oldValue, newValue Indexable, add bool) {
	for field, fieldValue := range extractFields(newValue) {
		if _, indexed := tree.fieldIndexes[field]; !indexed {
			continue
		}

		index, _ := tree.fieldIndexes[field]

		if oldValue != nil {
			oldFieldValue := oldValue.GetFields()[field]
			if oldFieldValue != fieldValue {
				// Remove the old key if it exists
				index.index.Remove(ufmt.Sprintf("%s|%s", oldFieldValue, key))
			}
		}

		if add {
			// Add or update the key
			index.index.Set(ufmt.Sprintf("%s|%s", fieldValue, key), key)
		}
	}
}

// extractFields extracts fields from an Indexable and returns them as a map of strings.
func extractFields(value Indexable) map[string]string {
	if value == nil {
		return nil
	}
	return value.GetFields()
}

// Get retrieves the value associated with the given key
func (tree *IndexedTree) Get(key string) (value Indexable, exists bool) {
	v, exists := tree.mainTree.Get(key)
	if exists && v != nil {
		value, _ = v.(Indexable)
	}
	return value, exists
}

// QueryByField retrieves objects based on the field value
func (tree *IndexedTree) QueryByField(field string, value string) ([]Indexable, bool) {
	index, ok := tree.fieldIndexes[field]
	if !ok {
		return nil, false
	}

	var values []Indexable

	// Use Iterate to collect keys
	index.index.Iterate("", "", func(key string, obj interface{}) bool {
		// Retrieve the actual object from the main tree using the key
		if obj, exists := tree.mainTree.Get(obj.(string)); exists {
			if idxObj, ok := obj.(Indexable); ok {
				if idxObj.GetFields()[field] == value {
					values = append(values, idxObj)
				}
			}
		}
		return false
	})

	return values, len(values) > 0
}

// AddFieldToIndex adds a field to be indexed
func (tree *IndexedTree) AddFieldToIndex(field string) {
	if _, exists := tree.fieldIndexes[field]; !exists {
		tree.fieldIndexes[field] = &FieldIndex{index: NewTree()}
	}
}
