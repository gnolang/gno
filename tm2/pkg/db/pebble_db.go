//go:build pebbledb

package db

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/cockroachdb/pebble"
)

const (
	defaultBytesPerSync = 0 // use the distribution default, 512KB
	defaultCacheSize    = 0 // use the distribution default, 8MB
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
	opts := &pebble.Options{}
	if defaultBytesPerSync > 0 {
		opts.BytesPerSync = defaultBytesPerSync
	}
	if defaultCacheSize > 0 {
		cache := pebble.NewCache(defaultCacheSize)
		defer cache.Unref()
		opts.Cache = cache
	}
	return NewPebbleDBWithOpts(name, dir, opts)
}

func NewPebbleDBWithOpts(name string, dir string, opts *pebble.Options) (*PebbleDB, error) {
	db, err := pebble.Open(filepath.Join(dir, name+".db"), opts)
	if err != nil {
		return nil, err
	}
	return &PebbleDB{
		db: db,
	}, nil
}

// Implements DB.
func (db *PebbleDB) Get(key []byte) []byte {
	value, closer, err := db.db.Get(nonNilBytes(key))
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
	_, closer, err := db.db.Get(nonNilBytes(key))
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return false
		}
		panic(err)
	}
	closer.Close()
	return true
}

// Implements DB.
func (db *PebbleDB) Set(key []byte, value []byte) {
	if err := db.db.Set(nonNilBytes(key), nonNilBytes(value), pebble.NoSync); err != nil {
		panic(err)
	}
}

// Implements DB.
func (db *PebbleDB) SetSync(key []byte, value []byte) {
	if err := db.db.Set(nonNilBytes(key), nonNilBytes(value), pebble.Sync); err != nil {
		panic(err)
	}
}

// Implements DB.
func (db *PebbleDB) Delete(key []byte) {
	if err := db.db.Delete(nonNilBytes(key), pebble.NoSync); err != nil {
		panic(err)
	}
}

// Implements DB.
func (db *PebbleDB) DeleteSync(key []byte) {
	if err := db.db.Delete(nonNilBytes(key), pebble.Sync); err != nil {
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
	it, err := db.db.NewIter(nil)
	if err != nil {
		panic(err)
	}
	for it.First(); it.Valid(); it.Next() {
		fmt.Printf("[%X]:\t[%X]\n", it.Key(), it.Value())
	}
	it.Close()
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
func (ba *pebbleDBBatch) Set(key, value []byte) {
	if err := ba.batch.Set(nonNilBytes(key), nonNilBytes(value), nil); err != nil {
		panic(err)
	}
}

// Implements Batch.
func (ba *pebbleDBBatch) Delete(key []byte) {
	if err := ba.batch.Delete(nonNilBytes(key), nil); err != nil {
		panic(err)
	}
}

// Implements Batch.
func (ba *pebbleDBBatch) Write() {
	if err := ba.db.db.Apply(ba.batch, pebble.NoSync); err != nil {
		panic(err)
	}
}

// Implements Batch.
func (ba *pebbleDBBatch) WriteSync() {
	if err := ba.db.db.Apply(ba.batch, pebble.Sync); err != nil {
		panic(err)
	}
}

// Implements Batch.
func (ba *pebbleDBBatch) Close() {
	if err := ba.batch.Close(); err != nil {
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
	iter    *pebble.Iterator
	start   []byte
	end     []byte
	reverse bool
}

var _ Iterator = (*pebbleDBIterator)(nil)

func (db *PebbleDB) newPebbleDBIterator(start, end []byte, reverse bool) *pebbleDBIterator {
	iter, err := db.db.NewIter(&pebble.IterOptions{
		LowerBound: start,
		UpperBound: end,
	})
	if err != nil {
		panic(err)
	}
	if reverse {
		iter.Last()
	} else {
		iter.First()
	}
	return &pebbleDBIterator{
		iter:    iter,
		start:   start,
		end:     end,
		reverse: reverse,
	}
}

// Implements Iterator.
func (it *pebbleDBIterator) Domain() ([]byte, []byte) {
	return it.start, it.end
}

// Implements Iterator.
func (it *pebbleDBIterator) Valid() bool {
	if err := it.iter.Error(); err != nil {
		panic(err)
	}
	return it.iter.Valid()
}

// Implements Iterator.
func (it *pebbleDBIterator) Key() []byte {
	it.assertValid()
	return cp(it.iter.Key())
}

// Implements Iterator.
func (it *pebbleDBIterator) Value() []byte {
	it.assertValid()
	return cp(it.iter.Value())
}

// Implements Iterator.
func (it *pebbleDBIterator) Next() {
	it.assertValid()
	if it.reverse {
		it.iter.Prev()
	} else {
		it.iter.Next()
	}
}

// Implements Iterator.
func (it *pebbleDBIterator) Close() {
	if err := it.iter.Close(); err != nil {
		panic(err)
	}
}

func (it pebbleDBIterator) assertValid() {
	if !it.Valid() {
		panic("pebbleDBIterator is invalid")
	}
}
