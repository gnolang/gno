package index

import (
	"gno.land/p/demo/avl"
)

// KeyExtractor is a function that extracts an index key from a value
type KeyExtractor func(value interface{}) string

// IndexedTree wraps an AVL tree and maintains secondary indexes
type IndexedTree struct {
	primary *avl.Tree
	indexes map[string]*Index
}

// Index represents a secondary index
type Index struct {
	tree    *avl.Tree
	extract KeyExtractor
}

// NewIndexedTree creates a new indexed tree wrapper
func NewIndexedTree() *IndexedTree {
	return &IndexedTree{
		primary: avl.NewTree(),
		indexes: make(map[string]*Index),
	}
}

// AddIndex creates a new secondary index with the given name and key extractor
func (it *IndexedTree) AddIndex(name string, extract KeyExtractor) {
	it.indexes[name] = &Index{
		tree:    avl.NewTree(),
		extract: extract,
	}
}

// Set adds or updates a value in the primary tree and all indexes
func (it *IndexedTree) Set(key string, value interface{}) bool {
	// First, if this is an update, we need to remove old index entries
	if oldValue, exists := it.primary.Get(key); exists {
		it.removeFromIndexes(key, oldValue)
	}

	// Update primary tree
	updated := it.primary.Set(key, value)

	// Update all indexes
	it.addToIndexes(key, value)

	return updated
}

// Remove removes a value from the primary tree and all indexes
func (it *IndexedTree) Remove(key string) (interface{}, bool) {
	// Get the value first so we can remove it from indexes
	value, exists := it.primary.Get(key)
	if !exists {
		return nil, false
	}

	// Remove from indexes first
	it.removeFromIndexes(key, value)

	// Remove from primary tree
	return it.primary.Remove(key)
}

// GetByIndex retrieves values from a secondary index
func (it *IndexedTree) GetByIndex(indexName string, indexKey string) []interface{} {
	index, exists := it.indexes[indexName]
	if !exists {
		return nil
	}

	var results []interface{}

	// Get all primary keys that match the index key
	value, exists := index.tree.Get(indexKey)
	if exists {
		primaryKeys := value.([]string)
		for _, primaryKey := range primaryKeys {
			if primaryValue, exists := it.primary.Get(primaryKey); exists {
				results = append(results, primaryValue)
			}
		}
	}

	return results
}

// Internal helper methods

func (it *IndexedTree) addToIndexes(primaryKey string, value interface{}) {
	for _, index := range it.indexes {
		indexKey := index.extract(value)

		// Get existing keys or create new slice
		var keys []string
		if existing, exists := index.tree.Get(indexKey); exists {
			keys = existing.([]string)
		}
		keys = append(keys, primaryKey)

		index.tree.Set(indexKey, keys)
	}
}

func (it *IndexedTree) removeFromIndexes(primaryKey string, value interface{}) {
	for _, index := range it.indexes {
		indexKey := index.extract(value)

		// Get existing keys
		if existing, exists := index.tree.Get(indexKey); exists {
			keys := existing.([]string)
			// Remove the primary key from the slice
			newKeys := make([]string, 0)
			for _, k := range keys {
				if k != primaryKey {
					newKeys = append(newKeys, k)
				}
			}
			// If there are still keys, update the index, otherwise remove it
			if len(newKeys) > 0 {
				index.tree.Set(indexKey, newKeys)
			} else {
				index.tree.Remove(indexKey)
			}
		}
	}
}

func (it *IndexedTree) GetIndexTree(name string) avl.TreeInterface {
	if idx, exists := it.indexes[name]; exists {
		return idx.tree
	}
	return nil
}

func (it *IndexedTree) GetPrimary() avl.TreeInterface {
	return it.primary
}

func (it *IndexedTree) Update(key string, oldValue interface{}, newValue interface{}) bool {
	// Remove old value from indexes
	it.removeFromIndexes(key, oldValue)

	// Update primary tree
	updated := it.primary.Set(key, newValue)

	// Add new value to indexes
	it.addToIndexes(key, newValue)

	return updated
}
