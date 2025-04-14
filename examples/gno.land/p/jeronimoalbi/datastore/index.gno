package datastore

import (
	"gno.land/p/moul/collection"
)

// DefaultIndexOptions defines the default options for new indexes.
const DefaultIndexOptions = collection.DefaultIndex | collection.SparseIndex

type (
	// IndexFn defines a type for single value indexing functions.
	// This type of function extracts a single string value from
	// a record that is then used to index it.
	IndexFn func(Record) string

	// IndexMultiValueFn defines a type for multi value indexing functions.
	// This type of function extracts multiple string values from a
	// record that are then used to index it.
	IndexMultiValueFn func(Record) []string

	// Index defines a type for custom user defined storage indexes.
	// Storages are by default indexed by the auto geneated record ID
	// but can additionally be indexed by other custom record fields.
	Index struct {
		name    string
		options collection.IndexOption
		fn      interface{}
	}
)

// NewIndex creates a new single value index.
//
// Usage example:
//
//	// Index a User record by email
//	idx := NewIndex("email", func(r Record) string {
//	  return r.MustGet("email").(string)
//	})
func NewIndex(name string, fn IndexFn) Index {
	return Index{
		name:    name,
		options: DefaultIndexOptions,
		fn: func(v interface{}) string {
			return fn(v.(Record))
		},
	}
}

// NewMultiValueIndex creates a new multi value index.
//
// Usage example:
//
//	// Index a Post record by tag
//	idx := NewMultiValueIndex("tag", func(r Record) []string {
//	  return r.MustGet("tags").([]string)
//	})
func NewMultiValueIndex(name string, fn IndexMultiValueFn) Index {
	return Index{
		name:    name,
		options: DefaultIndexOptions,
		fn: func(v interface{}) []string {
			return fn(v.(Record))
		},
	}
}

// Name returns index's name.
func (idx Index) Name() string {
	return idx.name
}

// Options returns current index options.
// These options define the index behavior regarding case sensitivity and uniquenes.
func (idx Index) Options() collection.IndexOption {
	return idx.options
}

// Func returns the function that storage collections apply
// to each record to get the value to use for indexing it.
func (idx Index) Func() interface{} {
	return idx.fn
}

// Unique returns a copy of the index that indexes record values as unique values.
// Returned index contains previous options plus the unique one.
func (idx Index) Unique() Index {
	if idx.options&collection.UniqueIndex == 0 {
		idx.options |= collection.UniqueIndex
	}
	return idx
}

// CaseInsensitive returns a copy of the index that indexes record values ignoring casing.
// Returned index contains previous options plus the case insensitivity one.
func (idx Index) CaseInsensitive() Index {
	if idx.options&collection.CaseInsensitiveIndex == 0 {
		idx.options |= collection.CaseInsensitiveIndex
	}
	return idx
}
