package datasource

import "gno.land/p/nt/avl"

// DefaultQueryRecords defines the default number of records returned by queries.
const DefaultQueryRecords = 50

var defaultQuery = Query{Count: DefaultQueryRecords}

type (
	// QueryOption configures datasource queries.
	QueryOption func(*Query)

	// Query contains datasource query options.
	Query struct {
		// Offset of the first record to return during iteration.
		Offset int

		// Count contains the number to records that query should return.
		Count int

		// Tag contains a tag to use as filter for the records.
		Tag string

		// Filters contains optional query filters by field value.
		Filters avl.Tree
	}
)

// WithOffset configures query to return records starting from an offset.
func WithOffset(offset int) QueryOption {
	return func(q *Query) {
		q.Offset = offset
	}
}

// WithCount configures the number of records that query returns.
func WithCount(count int) QueryOption {
	return func(q *Query) {
		if count < 1 {
			count = DefaultQueryRecords
		}
		q.Count = count
	}
}

// ByTag configures query to filter by tag.
func ByTag(tag string) QueryOption {
	return func(q *Query) {
		q.Tag = tag
	}
}

// WithFilter assigns a new filter argument to a query.
// This option can be used multiple times if more than one
// filter has to be given to the query.
func WithFilter(field string, value any) QueryOption {
	return func(q *Query) {
		q.Filters.Set(field, value)
	}
}

// NewQuery creates a new datasource query.
func NewQuery(options ...QueryOption) Query {
	q := defaultQuery
	for _, apply := range options {
		apply(&q)
	}
	return q
}
