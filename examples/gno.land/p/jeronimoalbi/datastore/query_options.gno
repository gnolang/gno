package datastore

import (
	"errors"
	"strings"
)

var (
	ErrEmptyQueryIndexName = errors.New("query index name is empty")
	ErrInvalidQueryOffset  = errors.New("minimum allowed query offset is 0")
	ErrInvalidQuerySize    = errors.New("minimum allowed query size is 1")
)

// QueryOption configures queries.
type QueryOption func(*Query) error

// WithOffset assigns the offset or position of the first record that query must return.
// The minimum allowed offset is 0.
func WithOffset(offset int) QueryOption {
	return func(q *Query) error {
		if offset < 0 {
			return ErrInvalidQueryOffset
		}

		q.offset = offset
		return nil
	}
}

// WithSize assigns the maximum number of records that query can return.
// The minimum allowed size is 1.
func WithSize(size int) QueryOption {
	return func(q *Query) error {
		if size < 1 {
			return ErrInvalidQuerySize
		}

		q.size = size
		return nil
	}
}

// UseIndex assigns the index that the query must use to get the records.
// Using an index requires a key value to locate the records within the index.
func UseIndex(name, key string) QueryOption {
	return func(q *Query) error {
		q.indexName = strings.TrimSpace(name)
		if q.indexName == "" {
			return ErrEmptyQueryIndexName
		}

		q.indexKey = key
		return nil
	}
}
