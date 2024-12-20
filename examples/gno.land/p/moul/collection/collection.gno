package collection

import (
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
)

// New creates a new Collection instance with an initialized ID index.
// The ID index is a special unique index that is always present and
// serves as the primary key for all objects in the collection.
func New() *Collection {
	c := &Collection{
		indexes: make(map[string]*Index),
		idGen:   seqid.ID(0),
	}
	// Initialize _id index
	c.indexes[IDIndex] = &Index{
		options: UniqueIndex,
		tree:    avl.NewTree(),
	}
	return c
}

// Collection represents a collection of objects with multiple indexes
type Collection struct {
	indexes map[string]*Index
	idGen   seqid.ID
}

const (
	// IDIndex is the reserved name for the primary key index
	IDIndex = "_id"
)

// IndexOption represents configuration options for an index using bit flags
type IndexOption uint64

const (
	// DefaultIndex is a basic index with no special options
	DefaultIndex IndexOption = 0

	// UniqueIndex ensures no duplicate values are allowed
	UniqueIndex IndexOption = 1 << iota

	// CaseInsensitiveIndex automatically converts string values to lowercase
	CaseInsensitiveIndex

	// SparseIndex only indexes non-null/non-empty values
	SparseIndex

	// TODO: Add support for MultiValueIndex
	// TODO: Add support for ReverseSorting
)

// Index represents an index with its configuration and data
type Index struct {
	fn      func(interface{}) string
	options IndexOption
	tree    avl.ITree
}

// AddIndex adds a new index to the collection with the specified options
//
// Parameters:
//   - name: the unique name of the index (e.g., "age", "email", "username")
//   - indexFn: a function that extracts the index key from an object
//   - options: bit flags for index configuration
//
// Example usage:
//
//	// Create a unique, case-insensitive index for email
//	c.AddIndex("email", func(v interface{}) string {
//	    return v.(*User).Email
//	}, UniqueIndex|CaseInsensitiveIndex)
//
//	// Create a basic index for age
//	c.AddIndex("age", func(v interface{}) string {
//	    return strconv.Itoa(v.(*User).Age)
//	}, DefaultIndex)
func (c *Collection) AddIndex(name string, indexFn func(interface{}) string, options IndexOption) {
	if name == IDIndex {
		panic("_id is a reserved index name")
	}
	c.indexes[name] = &Index{
		fn:      indexFn,
		options: options,
		tree:    avl.NewTree(),
	}
}

// safeGenerateKey safely generates an index key from an object
func safeGenerateKey(fn func(interface{}) string, obj interface{}) (string, bool) {
	if obj == nil {
		return "", false
	}

	defer func() {
		recover() // recover from any panic
	}()

	return fn(obj), true
}

// Set adds or updates an object in the collection
func (c *Collection) Set(obj interface{}) uint64 {
	if obj == nil {
		return 0
	}

	// Generate new ID
	id := c.idGen.Next()
	idStr := id.String()

	// Check uniqueness constraints first
	for name, idx := range c.indexes {
		if name == IDIndex {
			continue
		}
		key, ok := safeGenerateKey(idx.fn, obj)
		if !ok {
			return 0
		}

		// Skip empty values for sparse indexes
		if idx.options&SparseIndex != 0 && key == "" {
			continue
		}

		if idx.options&CaseInsensitiveIndex != 0 {
			key = strings.ToLower(key)
		}

		// Only check uniqueness for unique indexes
		if idx.options&UniqueIndex != 0 {
			if existing, exists := idx.tree.Get(key); exists && existing != nil {
				return 0 // Uniqueness constraint violated
			}
		}
	}

	// Store in _id index first
	c.indexes[IDIndex].tree.Set(idStr, obj)

	// Store in all other indexes
	for name, idx := range c.indexes {
		if name == IDIndex {
			continue
		}
		key, ok := safeGenerateKey(idx.fn, obj)
		if !ok {
			// Rollback: remove from _id index
			c.indexes[IDIndex].tree.Remove(idStr)
			return 0
		}

		// Skip empty values for sparse indexes
		if idx.options&SparseIndex != 0 && key == "" {
			continue
		}

		if idx.options&CaseInsensitiveIndex != 0 {
			key = strings.ToLower(key)
		}

		// For non-unique indexes, we store the ID as the value
		idx.tree.Set(key, idStr)
	}

	return uint64(id)
}

// Get retrieves an object by index and key, returns (object, id)
func (c *Collection) Get(indexName, key string) (interface{}, uint64) {
	idx, exists := c.indexes[indexName]
	if !exists {
		return nil, 0
	}

	if indexName == IDIndex {
		obj, exists := idx.tree.Get(key)
		if !exists {
			return nil, 0
		}
		id, err := seqid.FromString(key)
		if err != nil {
			return nil, 0
		}
		return obj, uint64(id)
	}

	// For other indexes
	if idx.options&CaseInsensitiveIndex != 0 {
		key = strings.ToLower(key)
	}

	idStr, exists := idx.tree.Get(key)
	if !exists {
		return nil, 0
	}

	// Get the actual object from _id index
	obj, exists := c.indexes[IDIndex].tree.Get(idStr.(string))
	if !exists {
		return nil, 0
	}

	id, err := seqid.FromString(idStr.(string))
	if err != nil {
		return nil, 0
	}
	return obj, uint64(id)
}

// GetIndex returns the underlying tree for an index
func (c *Collection) GetIndex(name string) avl.ITree {
	idx, exists := c.indexes[name]
	if !exists {
		return nil
	}
	return idx.tree
}

// Delete removes an object by its ID
func (c *Collection) Delete(id uint64) {
	idStr := seqid.ID(id).String()

	// Get the object first to clean up other indexes
	obj, exists := c.indexes[IDIndex].tree.Get(idStr)
	if !exists {
		return
	}

	// Remove from all indexes
	for name, idx := range c.indexes {
		if name == IDIndex {
			idx.tree.Remove(idStr)
			continue
		}
		key, ok := safeGenerateKey(idx.fn, obj)
		if !ok {
			continue
		}
		// Handle case-insensitive indexes
		if idx.options&CaseInsensitiveIndex != 0 {
			key = strings.ToLower(key)
		}
		idx.tree.Remove(key)
	}
}

// Update updates an existing object and returns its ID (0 if not found)
func (c *Collection) Update(id uint64, obj interface{}) uint64 {
	if obj == nil {
		return 0
	}

	idStr := seqid.ID(id).String()

	// Check if object exists
	oldObj, exists := c.indexes[IDIndex].tree.Get(idStr)
	if !exists {
		return 0
	}

	// Check uniqueness constraints
	for name, idx := range c.indexes {
		if name == IDIndex {
			continue
		}
		if idx.options&UniqueIndex != 0 {
			newKey, ok := safeGenerateKey(idx.fn, obj)
			if !ok {
				return 0
			}
			oldKey, ok := safeGenerateKey(idx.fn, oldObj)
			if !ok {
				return 0
			}
			// If the key changed and new key already exists
			if newKey != oldKey {
				if existing, _ := idx.tree.Get(newKey); existing != nil {
					return 0 // Uniqueness constraint violated
				}
			}
		}
	}

	// Store old index entries for potential rollback
	oldEntries := make(map[string]string)
	for name, idx := range c.indexes {
		if name == IDIndex {
			continue
		}
		oldKey, ok := safeGenerateKey(idx.fn, oldObj)
		if !ok {
			continue
		}
		if idx.options&CaseInsensitiveIndex != 0 {
			oldKey = strings.ToLower(oldKey)
		}
		oldEntries[name] = oldKey
		idx.tree.Remove(oldKey)
	}

	// Add new index entries
	for name, idx := range c.indexes {
		if name == IDIndex {
			idx.tree.Set(idStr, obj)
			continue
		}
		newKey, ok := safeGenerateKey(idx.fn, obj)
		if !ok {
			// Rollback: restore old object and all index entries
			c.indexes[IDIndex].tree.Set(idStr, oldObj)
			for idxName, oldKey := range oldEntries {
				c.indexes[idxName].tree.Set(oldKey, idStr)
			}
			return 0
		}

		if idx.options&CaseInsensitiveIndex != 0 {
			newKey = strings.ToLower(newKey)
		}
		idx.tree.Set(newKey, idStr)
	}

	return id
}

// TODO: Add support for GetAll to retrieve multiple objects matching a given index key
// This will be particularly useful for non-unique indexes like "age" or "status"
