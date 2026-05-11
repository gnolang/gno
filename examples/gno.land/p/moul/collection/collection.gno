// Package collection provides a generic collection implementation with support for
// multiple indexes, including unique indexes and case-insensitive indexes.
// It is designed to be used with any type and allows efficient lookups using
// different fields or computed values.
//
// Example usage:
//
//	// Define a data type
//	type User struct {
//	    Name     string
//	    Email    string
//	    Age      int
//	    Username string
//	    Tags     []string
//	}
//
//	// Create a new collection
//	c := collection.New()
//
//	// Add indexes with different options
//	c.AddIndex("name", func(v any) string {
//	    return v.(*User).Name
//	}, UniqueIndex)
//
//	c.AddIndex("email", func(v any) string {
//	    return v.(*User).Email
//	}, UniqueIndex|CaseInsensitiveIndex)
//
//	c.AddIndex("age", func(v any) string {
//	    return strconv.Itoa(v.(*User).Age)
//	}, DefaultIndex)  // Non-unique index
//
//	c.AddIndex("username", func(v any) string {
//	    return v.(*User).Username
//	}, UniqueIndex|SparseIndex)  // Allow empty usernames
//
//	// For tags, we index all tags for the user
//	c.AddIndex("tag", func(v any) []string {
//	    return v.(*User).Tags
//	}, DefaultIndex)  // Non-unique to allow multiple users with same tag
//
//	// Store an object
//	id := c.Set(&User{
//	    Name:  "Alice",
//	    Email: "alice@example.com",
//	    Age:   30,
//	    Tags:  []string{"admin", "moderator"},  // User can have multiple tags
//	})
//
//	// Retrieve by any index
//	entry := c.GetFirst("email", "alice@example.com")
//	adminUsers := c.GetAll("tag", "admin")      // Find all users with admin tag
//	modUsers := c.GetAll("tag", "moderator")    // Find all users with moderator tag
//
// Index options can be combined using the bitwise OR operator.
// Available options:
//   - DefaultIndex: Regular index with no special behavior
//   - UniqueIndex: Ensures values are unique within the index
//   - CaseInsensitiveIndex: Makes string comparisons case-insensitive
//   - SparseIndex: Skips indexing empty values (nil or empty string)
//
// Example: UniqueIndex|CaseInsensitiveIndex for a case-insensitive unique index
package collection

import (
	"errors"
	"strings"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/seqid/v0"
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
)

// Index represents an index with its configuration and data.
// The index function can return either:
//   - string: for single-value indexes
//   - []string: for multi-value indexes where one object can be indexed under multiple keys
//
// The backing tree stores either a single ID or []string for multiple IDs per key.
type Index struct {
	fn      any
	options IndexOption
	tree    avl.ITree
}

// AddIndex adds a new index to the collection with the specified options
//
// Parameters:
//   - name: the unique name of the index (e.g., "tags")
//   - indexFn: a function that extracts either a string or []string from an object
//   - options: bit flags for index configuration (e.g., UniqueIndex)
func (c *Collection) AddIndex(name string, indexFn any, options IndexOption) {
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
func (idx *Index) store(key string, idStr string) {
	stored, exists := idx.tree.Get(key)
	if !exists {
		// First entry for this key
		idx.tree.Set(key, idStr)
		return
	}

	// Handle existing entries
	switch existing := stored.(type) {
	case string:
		if existing == idStr {
			return // Already stored
		}
		// Convert to array
		idx.tree.Set(key, []string{existing, idStr})
	case []string:
		// Check if ID already exists
		for _, id := range existing {
			if id == idStr {
				return
			}
		}
		// Append new ID
		idx.tree.Set(key, append(existing, idStr))
	}
}

// removeIndex handles how we remove an ID from the index tree
func (idx *Index) remove(key string, idStr string) {
	stored, exists := idx.tree.Get(key)
	if !exists {
		return
	}

	switch existing := stored.(type) {
	case string:
		if existing == idStr {
			idx.tree.Remove(key)
		}
	case []string:
		newIds := make([]string, 0, len(existing))
		for _, id := range existing {
			if id != idStr {
				newIds = append(newIds, id)
			}
		}
		if len(newIds) == 0 {
			idx.tree.Remove(key)
		} else if len(newIds) == 1 {
			idx.tree.Set(key, newIds[0])
		} else {
			idx.tree.Set(key, newIds)
		}
	}
}

// generateKeys extracts one or more keys from an object for a given index.
func generateKeys(idx *Index, obj any) ([]string, bool) {
	if obj == nil {
		return nil, false
	}

	switch fnTyped := idx.fn.(type) {
	case func(any) string:
		// Single-value index
		key := fnTyped(obj)
		return []string{key}, true
	case func(any) []string:
		// Multi-value index
		keys := fnTyped(obj)
		return keys, true
	default:
		panic("invalid index function type")
	}
}

// Set adds or updates an object in the collection.
// Returns a positive ID if successful.
// Returns 0 if:
//   - The object is nil
//   - A uniqueness constraint would be violated
//   - Index generation fails for any index
func (c *Collection) Set(obj any) uint64 {
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
			// (UniqueIndex is ambiguous; skipping that scenario)
			if idx.options&UniqueIndex != 0 {
				if existing, exists := idx.tree.Get(key); exists && existing != nil {
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
			idx.store(key, idStr)
		}
	}

	return uint64(id)
}

// Get retrieves entries matching the given key in the specified index.
// Returns an iterator over the matching entries.
func (c *Collection) Get(indexName string, key string) EntryIterator {
	idx, exists := c.indexes[indexName]
	if !exists {
		return EntryIterator{err: errors.New("index not found: " + indexName)}
	}

	if idx.options&CaseInsensitiveIndex != 0 {
		key = strings.ToLower(key)
	}

	if indexName == IDIndex {
		// For ID index, validate the ID format first
		_, err := seqid.FromString(key)
		if err != nil {
			return EntryIterator{err: err}
		}
	}

	return EntryIterator{
		collection: c,
		indexName:  indexName,
		key:        key,
	}
}

// GetFirst returns the first matching entry or nil if none found
func (c *Collection) GetFirst(indexName, key string) *Entry {
	iter := c.Get(indexName, key)
	if iter.Next() {
		return iter.Value()
	}
	return nil
}

// Delete removes an object by its ID and returns true if something was deleted
func (c *Collection) Delete(id uint64) bool {
	idStr := seqid.ID(id).String()

	// Get the object first to clean up other indexes
	obj, exists := c.indexes[IDIndex].tree.Get(idStr)
	if !exists {
		return false
	}

	// Remove from all indexes
	for name, idx := range c.indexes {
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
			idx.remove(key, idStr)
		}
	}
	return true
}

// Update updates an existing object and returns true if successful
// Returns true if the update was successful.
// Returns false if:
//   - The object is nil
//   - The ID doesn't exist
//   - A uniqueness constraint would be violated
//   - Index generation fails for any index
//
// If the update fails, the collection remains unchanged.
func (c *Collection) Update(id uint64, obj any) bool {
	if obj == nil {
		return false
	}
	idStr := seqid.ID(id).String()
	oldObj, exists := c.indexes[IDIndex].tree.Get(idStr)
	if !exists {
		return false
	}

	// Check unique constraints
	for name, idx := range c.indexes {
		if name == IDIndex {
			continue
		}

		if idx.options&UniqueIndex != 0 {
			newKeys, newOk := generateKeys(idx, obj)
			_, oldOk := generateKeys(idx, oldObj)
			if !newOk || !oldOk {
				return false
			}

			for _, newKey := range newKeys {
				if idx.options&CaseInsensitiveIndex != 0 {
					newKey = strings.ToLower(newKey)
				}

				found, _ := idx.tree.Get(newKey)
				if found != nil {
					if storedID, ok := found.(string); !ok || storedID != idStr {
						return false
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
			idx.remove(okey, idStr)
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
					c.indexes[idxName].store(oldKey, idStr)
				}
			}
			return false
		}
		for _, nkey := range newKeys {
			if idx.options&CaseInsensitiveIndex != 0 {
				nkey = strings.ToLower(nkey)
			}
			idx.store(nkey, idStr)
		}
	}

	return true
}

// GetAll retrieves all entries matching the given key in the specified index.
func (c *Collection) GetAll(indexName string, key string) []Entry {
	idx, exists := c.indexes[indexName]
	if !exists {
		return nil
	}

	if idx.options&CaseInsensitiveIndex != 0 {
		key = strings.ToLower(key)
	}

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

	// Handle both single and multi-value cases based on the actual data type
	switch stored := idData.(type) {
	case []string:
		result := make([]Entry, 0, len(stored))
		for _, idStr := range stored {
			if obj, exists := c.indexes[IDIndex].tree.Get(idStr); exists {
				result = append(result, Entry{ID: idStr, Obj: obj})
			}
		}
		return result
	case string:
		if obj, exists := c.indexes[IDIndex].tree.Get(stored); exists {
			return []Entry{{ID: stored, Obj: obj}}
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
