package datastore

import (
	"gno.land/p/moul/collection"
)

var defaultQuery = Query{indexName: collection.IDIndex}

// Query contains arguments for querying a storage.
type Query struct {
	offset    int
	size      int
	indexName string
	indexKey  string
}

// Offset returns the position of the first record to return.
// The minimum offset value is 0.
func (q Query) Offset() int {
	return q.offset
}

// Size returns the maximum number of records a query returns.
func (q Query) Size() int {
	return q.size
}

// IndexName returns the name of the storage index being used for the query.
func (q Query) IndexName() string {
	return q.indexName
}

// IndexKey return the index key value to locate the records.
// An empty string is returned when all indexed records match the query.
func (q Query) IndexKey() string {
	return q.indexKey
}

// IsEmpty checks if the query is empty.
// Empty queries return no records.
func (q Query) IsEmpty() bool {
	return q.indexName == ""
}
