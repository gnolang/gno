// Package datasource defines generic interfaces for datasources.
//
// Datasources contain a set of records which can optionally be
// taggable. Tags can optionally be used to filter records by taxonomy.
//
// Datasources can help in cases where the data sent during
// communication between different realms needs to be generic
// to avoid direct dependencies.
package datasource

import "errors"

// ErrInvalidRecord indicates that a datasource contains invalid records.
var ErrInvalidRecord = errors.New("datasource records is not valid")

type (
	// Fields defines an interface for read-only fields.
	Fields interface {
		// Has checks whether a field exists.
		Has(name string) bool

		// Get retrieves the value associated with the given field.
		Get(name string) (value any, found bool)
	}

	// Record defines a datasource record.
	Record interface {
		// ID returns the unique record's identifier.
		ID() string

		// String returns a string representation of the record.
		String() string

		// Fields returns record fields and values.
		Fields() (Fields, error)
	}

	// TaggableRecord defines a datasource record that supports tags.
	// Tags can be used to build a taxonomy to filter records by category.
	TaggableRecord interface {
		// Tags returns a list of tags for the record.
		Tags() []string
	}

	// ContentRecord defines a datasource record that can return content.
	ContentRecord interface {
		// Content returns the record content.
		Content() (string, error)
	}

	// Iterator defines an iterator of datasource records.
	Iterator interface {
		// Next returns true when a new record is available.
		Next() bool

		// Err returns any error raised when reading records.
		Err() error

		// Record returns the current record.
		Record() Record
	}

	// Datasource defines a generic datasource.
	Datasource interface {
		// Records returns a new datasource records iterator.
		Records(Query) Iterator

		// Size returns the total number of records in the datasource.
		// When -1 is returned it means datasource doesn't support size.
		Size() int

		// Record returns a single datasource record.
		Record(id string) (Record, error)
	}
)

// NewIterator returns a new record iterator for a datasource query.
func NewIterator(ds Datasource, options ...QueryOption) Iterator {
	return ds.Records(NewQuery(options...))
}

// QueryRecords return a slice of records for a datasource query.
func QueryRecords(ds Datasource, options ...QueryOption) ([]Record, error) {
	var (
		records []Record
		query   = NewQuery(options...)
		iter    = ds.Records(query)
	)

	for i := 0; i < query.Count && iter.Next(); i++ {
		r := iter.Record()
		if r == nil {
			return nil, ErrInvalidRecord
		}

		records = append(records, r)
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}
	return records, nil
}
