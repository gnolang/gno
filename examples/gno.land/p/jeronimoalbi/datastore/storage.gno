package datastore

import (
	"strings"

	"gno.land/p/moul/collection"
)

// NewStorage creates a new records storage.
func NewStorage(name string, options ...StorageOption) Storage {
	s := Storage{
		name:         name,
		currentIndex: collection.IDIndex,
		collection:   collection.New(),
		schema:       NewSchema(strings.Title(name)),
	}

	for _, apply := range options {
		apply(&s)
	}
	return s
}

// Storage stores a collection of records.
//
// By default it searches records by record ID but it allows
// using custom user defined indexes for other record fields.
//
// When a storage is created it defines a default schema that
// keeps track of record fields. Storage can be optionally
// created with a user defined schema in cases where the number
// of fields has to be pre-defined or when new records must have
// one or more fields initialized to default values.
type Storage struct {
	name         string
	currentIndex string
	collection   *collection.Collection
	schema       *Schema
}

// Name returns storage's name.
func (s Storage) Name() string {
	return s.name
}

// CurrentIndex returns the name of the index that is used
// by default for search and iteration.
func (s Storage) CurrentIndex() string {
	return s.currentIndex
}

// Collection returns the undelying collection used by the
// storage to store all records.
func (s Storage) Collection() *collection.Collection {
	return s.collection
}

// Schema returns the schema being used to track record fields.
func (s Storage) Schema() *Schema {
	return s.schema
}

// Size returns the number of records that the storage have.
func (s Storage) Size() int {
	return s.collection.GetIndex(collection.IDIndex).Size()
}

// NewRecord creates a new storage record.
//
// If a custom schema with default field values is assigned to
// storage it's used to assign initial default values when new
// records are created.
//
// Creating a new record doesn't assign an ID to it, a new ID
// is generated and assigned to the record when it's saved for
// the first time.
func (s Storage) NewRecord() Record {
	r := &record{
		schema:     s.schema,
		collection: s.collection,
	}

	// Assign default record values if the schema defines them
	for i, name := range s.schema.Fields() {
		if v, found := s.schema.GetDefaultByIndex(i); found {
			r.Set(name, v)
		}
	}
	return r
}

// WithIndex returns a copy of the storage with a different default index.
//
// Example usage:
//
//	// Create a storage that index users by ID (default) and email
//	storage := NewStorage("users", WithIndex(emailIdx))
//
//	// The "email" index has to be used instead of the default "ID"
//	// index to search a user by email.
//	user := storage.WithIndex("email").GetFirst("foo@bar.org")
func (s Storage) WithIndex(name string) Storage {
	s.currentIndex = name
	return s
}

// ForEach iterates each storage record using the current index.
func (s Storage) ForEach(fn func(Record) bool) bool {
	idx := s.collection.GetIndex(s.currentIndex)
	if idx == nil {
		return false
	}

	return idx.Iterate("", "", func(_ string, v interface{}) bool {
		return fn(v.(Record))
	})
}

// Get returns all records found for a key searched using current index.
func (s Storage) Get(key string) []Record {
	var (
		records []Record
		iter    = s.collection.Get(s.currentIndex, key)
	)
	for iter.Next() {
		records = append(records, iter.Value().Obj.(Record))
	}
	return records
}

// GetFirst returns the first record found for a key searched within current index.
// It returns nil when no records are found for the given key.
func (s Storage) GetFirst(key string) Record {
	var (
		record Record
		iter   = s.collection.Get(s.currentIndex, key)
	)
	if iter.Next() {
		record = iter.Value().Obj.(Record)
	}
	return record
}

// Delete deletes a record from the storage.
func (s Storage) Delete(id uint64) bool {
	return s.collection.Delete(id)
}
