package main

import (
	dbm "github.com/gnolang/gno/tm2/pkg/db"
)

// Wrapper wraps a dbm.DB to implement DB.
type Wrapper struct {
	dbm.DB
}

// newWrapper returns a new Wrapper.
// Wrapper must be implemented against rocksdb.DB and pebbleDB separately
func newWrapper(db dbm.DB) *Wrapper {
	return &Wrapper{DB: db}
}

// Iterator implements DB.
func (db *Wrapper) Iterator(start, end []byte) (dbm.Iterator, error) {
	return db.DB.Iterator(start, end)
}

// ReverseIterator implements DB.
func (db *Wrapper) ReverseIterator(start, end []byte) (dbm.Iterator, error) {
	return db.DB.ReverseIterator(start, end)
}

// NewBatch implements DB.
func (db *Wrapper) NewBatch() dbm.Batch {
	return db.DB.NewBatch()
}

// NewBatchWithSize implements DB.
func (db *Wrapper) NewBatchWithSize(size int) dbm.Batch {
	return db.DB.NewBatchWithSize(size)
}
