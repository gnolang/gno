package db

import "fmt"

type ImmutableDB struct {
	db DB
}

// NewImmutableDB wraps a db to make it immutable.
// ImmutableDB panics on mutation operations.
func NewImmutableDB(db DB) *ImmutableDB {
	return &ImmutableDB{
		db: db,
	}
}

// Implements DB.
func (idb *ImmutableDB) Get(key []byte) ([]byte, error) {
	return idb.db.Get(key)
}

// Implements DB.
func (idb *ImmutableDB) Has(key []byte) (bool, error) {
	return idb.db.Has(key)
}

// Implements DB.
func (idb *ImmutableDB) Set(key []byte, value []byte) error {
	panic("Cannot mutate *ImmutableDB by calling .Set()")
}

// Implements DB.
func (idb *ImmutableDB) SetSync(key []byte, value []byte) error {
	panic("Cannot mutate *ImmutableDB by calling .SetSync()")
}

// Implements DB.
func (idb *ImmutableDB) Delete(key []byte) error {
	panic("Cannot mutate *ImmutableDB by calling .Delete()")
}

// Implements DB.
func (idb *ImmutableDB) DeleteSync(key []byte) error {
	panic("Cannot mutate *ImmutableDB by calling .DeleteSync()")
}

// Implements DB.
func (idb *ImmutableDB) Iterator(start, end []byte) (Iterator, error) {
	return idb.db.Iterator(start, end)
}

// Implements DB.
func (idb *ImmutableDB) ReverseIterator(start, end []byte) (Iterator, error) {
	return idb.db.ReverseIterator(start, end)
}

// Implements DB.
func (idb *ImmutableDB) NewBatch() Batch {
	return nil // XXX
}

// Implements DB.
func (idb *ImmutableDB) NewBatchWithSize(_ int) Batch {
	return nil // XXX
}

// Implements DB.
func (idb *ImmutableDB) Close() error {
	return idb.db.Close()
}

// Implements DB.
func (idb *ImmutableDB) Print() error {
	fmt.Print("(immutable) ")
	return idb.db.Print()
}

// Implements DB.
func (idb *ImmutableDB) Stats() map[string]string {
	return idb.db.Stats()
}
