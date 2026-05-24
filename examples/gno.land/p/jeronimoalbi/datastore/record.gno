package datastore

import (
	"errors"

	"gno.land/p/moul/collection"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/seqid/v0"
)

// ErrUndefinedField indicates that a field in not defined in a record's schema.
var ErrUndefinedField = errors.New("undefined field")

type (
	// Record stores values for one or more fields.
	Record interface {
		ReadOnlyRecord

		// Set assings a value to a record field.
		// If the field doesn't exist it's created if the underlying schema allows it.
		// Storage schema can optionally be strict in which case no new fields other than
		// the ones that were previously defined are allowed.
		Set(field string, value interface{}) error

		// Save assigns an ID to newly created records and update storage indexes.
		Save() bool
	}

	// ReadOnlyRecord defines an interface for read-only records.
	ReadOnlyRecord interface {
		// ID returns record's ID
		ID() uint64

		// Key returns a string representation of the record's ID.
		// It's used to be able to search records within the ID index.
		Key() string

		// Type returns the record's type.
		Type() string

		// Fields returns the list of the record's field names.
		Fields() []string

		// IsEmpty checks if the record has no values.
		IsEmpty() bool

		// HasField checks if the record has a specific field.
		HasField(name string) bool

		// Get returns the value of a record's field.
		Get(field string) (value interface{}, found bool)

		// MustGet returns the value of a record's field or panics when the field is not found.
		MustGet(field string) interface{}
	}

	// RecordIterFn defines a type for record iteration functions.
	RecordIterFn func(Record) (stop bool)

	// Recordset defines an interface that allows iterating multiple records.
	Recordset interface {
		// Iterate iterates records in order.
		Iterate(fn RecordIterFn) (stopped bool)

		// ReverseIterate iterates records in reverse order.
		ReverseIterate(fn RecordIterFn) (stopped bool)

		// Size returns the number of records in the recordset.
		Size() int
	}
)

type record struct {
	id         uint64
	schema     *Schema
	collection *collection.Collection
	values     avl.Tree // string(field index) -> interface{}
}

// ID returns record's ID
func (r record) ID() uint64 {
	return r.id
}

// Key returns a string representation of the record's ID.
// It's used to be able to search records within the ID index.
func (r record) Key() string {
	return seqid.ID(r.id).String()
}

// Type returns the record's type.
func (r record) Type() string {
	return r.schema.Name()
}

// Fields returns the list of the record's field names.
func (r record) Fields() []string {
	return r.schema.Fields()
}

// IsEmpty checks if the record has no values.
func (r record) IsEmpty() bool {
	return r.values.Size() == 0
}

// HasField checks if the record has a specific field.
func (r record) HasField(name string) bool {
	return r.schema.HasField(name)
}

// Set assings a value to a record field.
// If the field doesn't exist it's created if the underlying schema allows it.
// Storage schema can optionally be strict in which case no new fields other than
// the ones that were previously defined are allowed.
func (r *record) Set(field string, value interface{}) error {
	i := r.schema.GetFieldIndex(field)
	if i == -1 {
		if r.schema.IsStrict() {
			return ErrUndefinedField
		}

		i, _ = r.schema.AddField(field, nil)
	}

	key := castIntToKey(i)
	r.values.Set(key, value)
	return nil
}

// Get returns the value of a record's field.
func (r record) Get(field string) (value interface{}, found bool) {
	i := r.schema.GetFieldIndex(field)
	if i == -1 {
		return nil, false
	}

	key := castIntToKey(i)
	return r.values.Get(key)
}

// MustGet returns the value of a record's field or panics when the field is not found.
func (r record) MustGet(field string) interface{} {
	v, found := r.Get(field)
	if !found {
		panic("field not found: " + field)
	}
	return v
}

// Save assigns an ID to newly created records and update storage indexes.
func (r *record) Save() bool {
	if r.id == 0 {
		r.id = r.collection.Set(r)
		return r.id != 0
	}
	return r.collection.Update(r.id, r)
}

type recordset struct {
	query   Query
	records avl.ITree
	keys    []string
	size    int
}

// Iterate iterates records in order.
func (rs recordset) Iterate(fn RecordIterFn) (stopped bool) {
	if rs.isUsingCustomIndex() {
		for _, k := range rs.keys {
			v, found := rs.records.Get(k)
			if !found {
				continue
			}

			if fn(v.(Record)) {
				return true
			}
		}

		return false
	}

	offset := rs.query.Offset()
	count := rs.query.Size()
	if count == 0 {
		count = rs.records.Size()
	}

	return rs.records.IterateByOffset(offset, count, func(_ string, v interface{}) bool {
		return fn(v.(Record))
	})
}

// ReverseIterate iterates records in reverse order.
func (rs recordset) ReverseIterate(fn RecordIterFn) (stopped bool) {
	if rs.isUsingCustomIndex() {
		for i := len(rs.keys) - 1; i >= 0; i-- {
			v, found := rs.records.Get(rs.keys[i])
			if !found {
				continue
			}

			if fn(v.(Record)) {
				return true
			}
		}

		return false
	}

	offset := rs.query.Offset()
	count := rs.query.Size()
	if count == 0 {
		count = rs.records.Size()
	}

	return rs.records.ReverseIterateByOffset(offset, count, func(_ string, v interface{}) bool {
		return fn(v.(Record))
	})
}

// Size returns the number of records in the recordset.
func (rs recordset) Size() int {
	return rs.size
}

func (rs recordset) isUsingCustomIndex() bool {
	return rs.query.IndexName() != collection.IDIndex
}
