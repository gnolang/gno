//go:build pebbledb

package db

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/cockroachdb/pebble"
)

const (
	defaultBytesPerSync = 1024 * 512       // 512KB
	defaultCacheSize    = 1024 * 1024 * 16 // 16MB
)

func init() {
	dbCreator := func(name string, dir string) (DB, error) {
		return NewPebbleDB(name, dir)
	}
	registerDBCreator(PebbleDBBackend, dbCreator, false)
}

var _ DB = (*PebbleDB)(nil)

type PebbleDB struct {
	db *pebble.DB
}

func NewPebbleDB(name string, dir string) (*PebbleDB, error) {
	cache := pebble.NewCache(defaultCacheSize)
	defer cache.Unref()
	opts := &pebble.Options{
		BytesPerSync: defaultBytesPerSync,
		Cache:        cache,
	}
	return NewPebbleDBWithOpts(name, dir, opts)
}

func NewPebbleDBWithOpts(name string, dir string, opts *pebble.Options) (*PebbleDB, error) {
	dbPath := filepath.Join(dir, name+".db")
	db, err := pebble.Open(dbPath, opts)
	if err != nil {
		return nil, err
	}
	database := &PebbleDB{
		db: db,
	}
	return database, nil
}

// Implements DB.
func (db *PebbleDB) Get(key []byte) []byte {
	key = nonNilBytes(key)
	value, closer, err := db.db.Get(key)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return nil
		}
		panic(err)
	}
	defer closer.Close()
	return cp(value)
}

// Implements DB.
func (db *PebbleDB) Has(key []byte) bool {
	return db.Get(key) != nil
}

// Implements DB.
func (db *PebbleDB) Set(key []byte, value []byte) {
	key = nonNilBytes(key)
	value = nonNilBytes(value)
	if err := db.db.Set(key, value, pebble.NoSync); err != nil {
		panic(err)
	}
}

// Implements DB.
func (db *PebbleDB) SetSync(key []byte, value []byte) {
	key = nonNilBytes(key)
	value = nonNilBytes(value)
	if err := db.db.Set(key, value, pebble.Sync); err != nil {
		panic(err)
	}
}

// Implements DB.
func (db *PebbleDB) Delete(key []byte) {
	key = nonNilBytes(key)
	if err := db.db.Delete(key, pebble.NoSync); err != nil {
		panic(err)
	}
}

// Implements DB.
func (db *PebbleDB) DeleteSync(key []byte) {
	key = nonNilBytes(key)
	if err := db.db.Delete(key, pebble.Sync); err != nil {
		panic(err)
	}
}

func (db *PebbleDB) DB() *pebble.DB {
	return db.db
}

// Implements DB.
func (db *PebbleDB) Close() {
	if err := db.db.Close(); err != nil {
		panic(err)
	}
}

// Implements DB.
func (db *PebbleDB) Print() {
	itr := db.db.NewIter(nil)
	for itr.First(); itr.Valid(); itr.Next() {
		fmt.Printf("[%X]:\t[%X]\n", itr.Key(), itr.Value())
	}
	itr.Close()
}

// Implements DB.
func (db *PebbleDB) Stats() map[string]string {
	stats := make(map[string]string)
	// TODO: what stats are there?
	return stats
}

// ----------------------------------------
// Batch

// Implements DB.
func (db *PebbleDB) NewBatch() Batch {
	batch := new(pebble.Batch)
	return &pebbleDBBatch{db, batch}
}

type pebbleDBBatch struct {
	db    *PebbleDB
	batch *pebble.Batch
}

// Implements Batch.
func (mBatch *pebbleDBBatch) Set(key, value []byte) {
	mBatch.batch.Set(key, value, nil)
}

// Implements Batch.
func (mBatch *pebbleDBBatch) Delete(key []byte) {
	mBatch.batch.Delete(key, nil)
}

// Implements Batch.
func (mBatch *pebbleDBBatch) Write() {
	if err := mBatch.db.db.Apply(mBatch.batch, pebble.NoSync); err != nil {
		panic(err)
	}
}

// Implements Batch.
func (mBatch *pebbleDBBatch) WriteSync() {
	if err := mBatch.db.db.Apply(mBatch.batch, pebble.Sync); err != nil {
		panic(err)
	}
}

// Implements Batch.
func (mBatch *pebbleDBBatch) Close() {
	if err := mBatch.batch.Close(); err != nil {
		panic(err)
	}
}

// ----------------------------------------
// Iterator

// Implements DB.
func (db *PebbleDB) Iterator(start, end []byte) Iterator {
	return db.newPebbleDBIterator(start, end, false)
}

// Implements DB.
func (db *PebbleDB) ReverseIterator(start, end []byte) Iterator {
	return db.newPebbleDBIterator(start, end, true)
}

type pebbleDBIterator struct {
	source    *pebble.Iterator
	start     []byte
	end       []byte
	isReverse bool
}

var _ Iterator = (*pebbleDBIterator)(nil)

func (db *PebbleDB) newPebbleDBIterator(start, end []byte, isReverse bool) *pebbleDBIterator {
	source := db.db.NewIter(&pebble.IterOptions{
		LowerBound: start,
		UpperBound: end,
	})
	if isReverse {
		source.Last()
	} else {
		source.First()
	}
	return &pebbleDBIterator{
		source:    source,
		start:     start,
		end:       end,
		isReverse: isReverse,
	}
}

// Implements Iterator.
func (itr *pebbleDBIterator) Domain() ([]byte, []byte) {
	return itr.start, itr.end
}

// Implements Iterator.
func (itr *pebbleDBIterator) Valid() bool {
	itr.assertNoError()
	return itr.source.Valid()
}

// Implements Iterator.
func (itr *pebbleDBIterator) Key() []byte {
	itr.assertIsValid()
	return cp(itr.source.Key())
}

// Implements Iterator.
func (itr *pebbleDBIterator) Value() []byte {
	itr.assertIsValid()
	return cp(itr.source.Value())
}

// Implements Iterator.
func (itr *pebbleDBIterator) Next() {
	itr.assertIsValid()
	if itr.isReverse {
		itr.source.Prev()
	} else {
		itr.source.Next()
	}
}

// Implements Iterator.
func (itr *pebbleDBIterator) Close() {
	if err := itr.source.Close(); err != nil {
		panic(err)
	}
}

func (itr *pebbleDBIterator) assertNoError() {
	if err := itr.source.Error(); err != nil {
		panic(err)
	}
}

func (itr pebbleDBIterator) assertIsValid() {
	if !itr.Valid() {
		panic("pebbleDBIterator is invalid")
	}
}
