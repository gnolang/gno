package datastore

import (
	"errors"
	"strings"

	"gno.land/p/moul/collection"
	"gno.land/p/nt/seqid"
)

// NewStorage creates a new records storage.
func NewStorage(name string, options ...StorageOption) Storage {
	s := Storage{
		name:       name,
		collection: collection.New(),
		schema:     NewSchema(strings.Title(name)),
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
	name       string
	collection *collection.Collection
	schema     *Schema
}

// Name returns storage's name.
func (s Storage) Name() string {
	return s.name
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

// Query returns a recordset that matches the query parameters.
// By default query selects records using the ID index.
//
// Example usage:
//
//	// Get 50 records starting from the one at position 100
//	rs, _ := storage.Query(
//		WithOffset(100),
//		WithSize(50),
//	)
//
//	// Iterate records to create a new slice
//	var records []Record
//	rs.Iterate(func (r Record) bool {
//		records = append(records, r)
//		return false
//	})
func (s Storage) Query(options ...QueryOption) (Recordset, error) {
	// Initialize the recordset for the query
	rs := recordset{
		query:   defaultQuery,
		records: s.collection.GetIndex(collection.IDIndex),
	}

	for _, apply := range options {
		if err := apply(&rs.query); err != nil {
			return nil, err
		}
	}

	indexName := rs.query.IndexName()
	if indexName != collection.IDIndex {
		// When using a custom index get the keys to get records from the ID index
		keys, err := s.getIndexRecordsKeys(indexName, rs.query.IndexKey())
		if err != nil {
			return nil, err
		}

		// Adjust the number of keys to match available query options
		if offset := rs.query.Offset(); offset > 0 {
			if offset > len(keys) {
				keys = nil
			} else {
				keys = keys[offset:]
			}
		}

		if size := rs.query.Size(); size > 0 && size < len(keys) {
			keys = keys[:size]
		}

		rs.keys = keys
		rs.size = len(keys)
	} else {
		// When using the default ID index init size with the total number of records
		rs.size = rs.records.Size()

		// Adjust recordset size to match available query options
		if offset := rs.query.Offset(); offset > 0 {
			if offset > rs.size {
				rs.size = 0
			} else {
				rs.size -= offset
			}
		}

		if size := rs.query.Size(); size > 0 && size < rs.size {
			rs.size = size
		}
	}

	return rs, nil
}

// MustQuery returns a recordset that matches the query parameters or panics on error.
// By default query selects records using the ID index.
//
// Example usage:
//
//	// Get 50 records starting from the one at position 100
//	var records []Record
//	storage.MustQuery(
//		WithOffset(100),
//		WithSize(50),
//	).Iterate(func (r Record) bool {
//		records = append(records, r)
//		return false
//	})
func (s Storage) MustQuery(options ...QueryOption) Recordset {
	rs, err := s.Query(options...)
	if err != nil {
		panic(err)
	}
	return rs
}

// Get returns the first record found for a key within a storage index.
//
// This is a convenience method to get a single record. A multi index will
// always return the first record value for the specified key in this case.
// To get multiple records create a query using a custom index and key value
// or use the underlying storage collection.
func (s Storage) Get(indexName, indexKey string) (_ Record, found bool) {
	iter := s.collection.Get(indexName, indexKey)
	if iter.Next() {
		return iter.Value().Obj.(Record), true
	}
	return nil, false
}

// GetByID returns a record whose ID matches the specified ID.
func (s Storage) GetByID(id uint64) (_ Record, found bool) {
	iter := s.collection.Get(collection.IDIndex, seqid.ID(id).String())
	if iter.Next() {
		return iter.Value().Obj.(Record), true
	}
	return nil, false
}

// Delete deletes a record from the storage.
func (s Storage) Delete(id uint64) bool {
	return s.collection.Delete(id)
}

func (s Storage) getIndexRecordsKeys(indexName, indexKey string) ([]string, error) {
	idx := s.collection.GetIndex(indexName)
	if idx == nil {
		return nil, errors.New("storage index for query not found: " + indexName)
	}

	var keys []string
	if v, found := idx.Get(indexKey); found {
		keys = castIfaceToRecordKeys(v)
		if keys == nil {
			return nil, errors.New("unexpected storage index key format")
		}
	}
	return keys, nil
}

func castIfaceToRecordKeys(v interface{}) []string {
	switch k := v.(type) {
	case []string:
		return k
	case string:
		return []string{k}
	}
	return nil
}
