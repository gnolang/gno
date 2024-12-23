// Package collection provides a generic collection implementation with support for
// multiple indexes, including unique indexes, multi-value indexes, and case-insensitive
// indexes. It is designed to be used with any type and allows efficient lookups using
// different fields or computed values.
//
// Basic usage:
//
//	// Create a new collection
//	c := collection.New()
//
//	// Add various types of indexes
//	c.AddIndex("name", func(v interface{}) string {
//	    return v.(*User).Name
//	}, UniqueIndex)
//
//	c.AddIndex("email", func(v interface{}) string {
//	    return v.(*User).Email
//	}, UniqueIndex | CaseInsensitiveIndex)
//
//	c.AddIndex("age", func(v interface{}) string {
//	    return strconv.Itoa(v.(*User).Age)
//	}, DefaultIndex)
//
//	c.AddIndex("tags", func(v interface{}) []string {
//	    return v.(*User).Tags
//	}, MultiValueIndex)
//
//	// Store an object
//	id := c.Set(&User{
//	    Name:  "Alice",
//	    Email: "alice@example.com",
//	    Age:   30,
//	    Tags:  []string{"admin", "staff"},
//	})
//
//	// Retrieve by any index
//	entry = c.Get("email", "alice@example.com")
//	entries := c.GetAll("tags", "admin")
//
// Index options can be combined using the bitwise OR operator:
//   - UniqueIndex: Ensures values are unique within the index
//   - MultiValueIndex: Allows storing multiple values for a single key
//   - CaseInsensitiveIndex: Makes string comparisons case-insensitive
//   - SparseIndex: Skips indexing empty values
//   - DefaultIndex: Regular index with no special behavior
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

	// SparseIndex only indexes non-empty values
	SparseIndex

	// MultiValueIndex supports multiple key entries for a single object.
	// WARNING: The current implementation uses a simple slice to store multiple values,
	// which is not optimized for large datasets. This should only be used in controlled
	// environments with a small number of values per key.
	// TODO(moul): Switch to avl/list implementation for better scalability with large datasets.
	MultiValueIndex
)

// Index represents an index with its configuration and data.
//
// If an index is MultiValueIndex, then its function is expected to
// return []string (rather than a single string). We store one or more
// object IDs in the backing tree for each key.
type Index struct {
	// fn should return either a single string if !(options & MultiValueIndex),
	// or a []string if (options & MultiValueIndex) != 0.
	fn      interface{}
	options IndexOption
	tree    avl.ITree
}

// AddIndex adds a new index to the collection with the specified options
//
// Parameters:
//   - name: the unique name of the index (e.g., "tags")
//   - indexFn: a function that extracts either a string or []string from an object
//   - options: bit flags for index configuration (e.g., MultiValueIndex)
func (c *Collection) AddIndex(name string, indexFn interface{}, options IndexOption) {
	if name == IDIndex {
		panic("_id is a reserved index name")
	}
	c.indexes[name] = &Index{
		fn:      indexFn,
		options: options,
		tree:    avl.NewTree(),
	}
}

// storeIndex handles how we store an ID in the index tree
func storeIndex(idx *Index, key, idStr string) {
	if idx.options&MultiValueIndex != 0 {
		// Get existing value
		existing, exists := idx.tree.Get(key)

		var ids []string
		if exists {
			switch v := existing.(type) {
			case []string:
				// Copy existing slice, checking for duplicates
				for _, id := range v {
					if id == idStr {
						return // duplicate found, exit early
					}
					ids = append(ids, id)
				}
			case string:
				if v != idStr { // Only add if not duplicate
					ids = append(ids, v)
				}
			}
		}

		// Add new ID
		ids = append(ids, idStr)
		idx.tree.Set(key, ids)
	} else {
		idx.tree.Set(key, idStr)
	}
}

// removeIndex handles how we remove an ID from the index tree
func removeIndex(idx *Index, key, idStr string) {
	if idx.options&MultiValueIndex != 0 {
		existing, exists := idx.tree.Get(key)
		if !exists {
			return
		}
		if ids, ok := existing.([]string); ok {
			// Filter out the ID to remove
			newIds := make([]string, 0, len(ids)-1)
			for _, id := range ids {
				if id != idStr {
					newIds = append(newIds, id)
				}
			}
			if len(newIds) == 0 {
				idx.tree.Remove(key)
			} else {
				idx.tree.Set(key, newIds)
			}
		}
		return
	}

	// Single-value index logic remains the same
	if stored, exists := idx.tree.Get(key); exists {
		if storedID, ok := stored.(string); ok && storedID == idStr {
			idx.tree.Remove(key)
		}
	}
}

// generateKeys extracts one or more keys from an object for a given index.
// If the index is MultiValueIndex, indexFn should return []string;
// otherwise, a single string is cast as []string for uniform handling.
//
// This function recovers from any panic inside the index function.
func generateKeys(idx *Index, obj interface{}) ([]string, bool) {
	if obj == nil {
		return nil, false
	}

	defer func() {
		_ = recover() // recover from any panic
	}()

	switch fnTyped := idx.fn.(type) {
	case func(interface{}) string:
		// Single-value index
		key := fnTyped(obj)
		return []string{key}, true
	case func(interface{}) []string:
		// Multi-value index
		keys := fnTyped(obj)
		return keys, true
	default:
		panic("invalid index function type")
	}
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
		keys, ok := generateKeys(idx, obj)
		if !ok {
			return 0
		}

		for _, key := range keys {
			// Skip empty values for sparse indexes
			if idx.options&SparseIndex != 0 && key == "" {
				continue
			}
			if idx.options&CaseInsensitiveIndex != 0 {
				key = strings.ToLower(key)
			}
			// Only check uniqueness for unique + single-value indexes
			// (MultiValueIndex + UniqueIndex is ambiguous; skipping that scenario)
			if idx.options&UniqueIndex != 0 && idx.options&MultiValueIndex == 0 {
				if existing, exists := idx.tree.Get(key); exists && existing != nil {
					// Uniqueness constraint violated
					return 0
				}
			}
		}
	}

	// Store in _id index first (the actual object)
	c.indexes[IDIndex].tree.Set(idStr, obj)

	// Store in all other indexes
	for name, idx := range c.indexes {
		if name == IDIndex {
			continue
		}
		keys, ok := generateKeys(idx, obj)
		if !ok {
			// Rollback: remove from _id index
			c.indexes[IDIndex].tree.Remove(idStr)
			return 0
		}

		for _, key := range keys {
			if idx.options&SparseIndex != 0 && key == "" {
				continue
			}
			if idx.options&CaseInsensitiveIndex != 0 {
				key = strings.ToLower(key)
			}
			storeIndex(idx, key, idStr)
		}
	}

	return uint64(id)
}

// Get retrieves an object by index and key, returns an Entry or nil if not found.
func (c *Collection) Get(indexName, key string) *Entry {
	idx, exists := c.indexes[indexName]
	if !exists {
		return nil
	}

	if indexName == IDIndex {
		obj, exists := idx.tree.Get(key)
		if !exists {
			return nil
		}
		return &Entry{ID: key, Obj: obj}
	}

	// For other indexes
	if idx.options&CaseInsensitiveIndex != 0 {
		key = strings.ToLower(key)
	}

	idData, exists := idx.tree.Get(key)
	if !exists {
		return nil
	}

	// If multi-value, just return the first object found
	if idx.options&MultiValueIndex != 0 {
		list, ok := idData.([]string)
		if !ok || len(list) == 0 {
			return nil
		}
		// Get the first ID from the slice
		idStr := list[0]
		obj, exists := c.indexes[IDIndex].tree.Get(idStr)
		if !exists {
			return nil
		}
		return &Entry{ID: idStr, Obj: obj}
	}

	// single-value index
	idStr, ok := idData.(string)
	if !ok {
		return nil
	}
	obj, exists := c.indexes[IDIndex].tree.Get(idStr)
	if !exists {
		return nil
	}
	return &Entry{ID: idStr, Obj: obj}
}

// Entry represents a single object with its ID in the collection
type Entry struct {
	ID  string
	Obj interface{}
}

// GetAll retrieves all entries matching the given key in the specified index.
func (c *Collection) GetAll(indexName, key string) []Entry {
	idx, exists := c.indexes[indexName]
	if !exists {
		return nil
	}

	if idx.options&CaseInsensitiveIndex != 0 {
		key = strings.ToLower(key)
	}

	// Special handling for ID index
	if indexName == IDIndex {
		if obj, exists := idx.tree.Get(key); exists {
			return []Entry{{ID: key, Obj: obj}}
		}
		return nil
	}

	idData, exists := idx.tree.Get(key)
	if !exists {
		return nil
	}

	// For multi-value indexes
	if idx.options&MultiValueIndex != 0 {
		if ids, ok := idData.([]string); ok {
			result := make([]Entry, 0, len(ids))
			for _, idStr := range ids {
				if obj, exists := c.indexes[IDIndex].tree.Get(idStr); exists {
					result = append(result, Entry{ID: idStr, Obj: obj})
				}
			}
			return result
		}
		return nil
	}

	// For single-value indexes
	if idStr, ok := idData.(string); ok {
		if obj, exists := c.indexes[IDIndex].tree.Get(idStr); exists {
			return []Entry{{ID: idStr, Obj: obj}}
		}
	}
	return nil
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
		// remove from _id or any other index
		if name == IDIndex {
			idx.tree.Remove(idStr)
			continue
		}
		keys, ok := generateKeys(idx, obj)
		if !ok {
			continue
		}
		for _, key := range keys {
			if idx.options&CaseInsensitiveIndex != 0 {
				key = strings.ToLower(key)
			}
			removeIndex(idx, key, idStr)
		}
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

	// Check uniqueness constraints for single-value unique indexes
	for name, idx := range c.indexes {
		if name == IDIndex {
			continue
		}
		// We only enforce uniqueness if it's a single-value, unique index
		if (idx.options&UniqueIndex != 0) && (idx.options&MultiValueIndex == 0) {
			newKeys, newOk := generateKeys(idx, obj)
			_, oldOk := generateKeys(idx, oldObj)
			if !newOk || !oldOk {
				return 0
			}
			// If any new key changed and already exists, fail uniqueness
			for _, newKey := range newKeys {
				if idx.options&CaseInsensitiveIndex != 0 {
					newKey = strings.ToLower(newKey)
				}
				found, _ := idx.tree.Get(newKey)
				// If this key is found, but not tied to the same ID, uniqueness is violated
				if found != nil {
					switch v := found.(type) {
					case string:
						if v != idStr {
							return 0
						}
					case []string:
						// For a single-value unique index, we don't expect a slice,
						// so treat it as a conflict.
						return 0
					}
				}
			}
		}
	}

	// Store old index entries for potential rollback
	oldEntries := make(map[string][]string)
	for name, idx := range c.indexes {
		if name == IDIndex {
			continue
		}
		oldKeys, ok := generateKeys(idx, oldObj)
		if !ok {
			continue
		}
		var adjusted []string
		for _, okey := range oldKeys {
			if idx.options&CaseInsensitiveIndex != 0 {
				okey = strings.ToLower(okey)
			}
			// Remove the oldObj from the index right away
			removeIndex(idx, okey, idStr)
			adjusted = append(adjusted, okey)
		}
		oldEntries[name] = adjusted
	}

	// Update the object in the _id index
	c.indexes[IDIndex].tree.Set(idStr, obj)

	// Add new index entries
	for name, idx := range c.indexes {
		if name == IDIndex {
			continue
		}
		newKeys, ok := generateKeys(idx, obj)
		if !ok {
			// Rollback: restore old object and old index entries
			c.indexes[IDIndex].tree.Set(idStr, oldObj)
			for idxName, keys := range oldEntries {
				for _, oldKey := range keys {
					storeIndex(c.indexes[idxName], oldKey, idStr)
				}
			}
			return 0
		}
		for _, nkey := range newKeys {
			if idx.options&CaseInsensitiveIndex != 0 {
				nkey = strings.ToLower(nkey)
			}
			storeIndex(idx, nkey, idStr)
		}
	}

	return id
}
