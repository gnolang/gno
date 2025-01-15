package pebbledb

import (
	goerrors "errors"
	"path/filepath"

	"github.com/cockroachdb/pebble"

	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/internal"
)

func init() {
	dbCreator := func(name string, dir string) (db.DB, error) {
		return NewPebbleDB(name, dir)
	}
	db.InternalRegisterDBCreator(db.PebbleDBBackend, dbCreator, false)
}

var _ db.DB = (*PebbleDB)(nil)

type PebbleDB struct {
	db *pebble.DB
}

func NewPebbleDB(name string, dir string) (*PebbleDB, error) {
	return NewPebbleDBWithOpts(name, dir, nil)
}

func NewPebbleDBWithOpts(name string, dir string, o *pebble.Options) (*PebbleDB, error) {
	dbPath := filepath.Join(dir, name+".db")
	db, err := pebble.Open(dbPath, o)
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
	key = internal.NonNilBytes(key)
	res, closer, err := db.db.Get(key)
	if err != nil {
		if goerrors.Is(err, pebble.ErrNotFound) {
			return nil
		}
		panic(err)
	}

	defer closer.Close()

	return res
}

// Implements DB.
func (db *PebbleDB) Has(key []byte) bool {
	return db.Get(key) != nil
}

// Implements DB.
func (db *PebbleDB) Set(key []byte, value []byte) {
	key = internal.NonNilBytes(key)
	value = internal.NonNilBytes(value)
	err := db.db.Set(key, value, pebble.NoSync)
	if err != nil {
		panic(err)
	}
}

// Implements DB.
func (db *PebbleDB) SetSync(key []byte, value []byte) {
	key = internal.NonNilBytes(key)
	value = internal.NonNilBytes(value)
	err := db.db.Set(key, value, pebble.Sync)
	if err != nil {
		panic(err)
	}
}

// Implements DB.
func (db *PebbleDB) Delete(key []byte) {
	key = internal.NonNilBytes(key)
	err := db.db.Delete(key, pebble.NoSync)
	if err != nil {
		panic(err)
	}
}

// Implements DB.
func (db *PebbleDB) DeleteSync(key []byte) {
	key = internal.NonNilBytes(key)
	err := db.db.Delete(key, pebble.Sync)
	if err != nil {
		panic(err)
	}
}

// Implements DB.
func (db *PebbleDB) Close() {
	db.db.Close()
}

// Implements DB.
func (db *PebbleDB) Print() {

}

// Implements DB.
func (db *PebbleDB) Stats() map[string]string {
	return nil
}

// ----------------------------------------
// Batch

// Implements DB.
func (db *PebbleDB) NewBatch() db.Batch {
	return &goLevelDBBatch{db, db.db.NewBatch()}
}

type goLevelDBBatch struct {
	db    *PebbleDB
	batch *pebble.Batch
}

// Implements Batch.
func (mBatch *goLevelDBBatch) Set(key, value []byte) {
	if err := mBatch.batch.Set(key, value, pebble.NoSync); err != nil {
		panic(err)
	}
}

// Implements Batch.
func (mBatch *goLevelDBBatch) Delete(key []byte) {
	if err := mBatch.batch.Delete(key, pebble.NoSync); err != nil {
		panic(err)
	}
}

// Implements Batch.
func (mBatch *goLevelDBBatch) Write() {
	if err := mBatch.batch.Commit(pebble.NoSync); err != nil {
		panic(err)
	}
}

// Implements Batch.
func (mBatch *goLevelDBBatch) WriteSync() {
	if err := mBatch.batch.Commit(pebble.Sync); err != nil {
		panic(err)
	}
}

// Implements Batch.
// Close is no-op for goLevelDBBatch.
func (mBatch *goLevelDBBatch) Close() {
	mBatch.batch.Close()
}

// Implements DB.
func (db *PebbleDB) Iterator(start, end []byte) db.Iterator {
	it, err := db.db.NewIter(&pebble.IterOptions{
		LowerBound: start,
		UpperBound: end,
	})
	if err != nil {
		panic(err)
	}

	return &pebbleIter{
		pebbleBaseIter: pebbleBaseIter{
			i:     it,
			start: start,
			end:   end,
		},
	}
}

// Implements DB.
func (db *PebbleDB) ReverseIterator(start, end []byte) db.Iterator {
	it, err := db.db.NewIter(&pebble.IterOptions{
		LowerBound: start,
		UpperBound: end,
	})
	if err != nil {
		panic(err)
	}

	return &pebbleReverseIter{
		pebbleBaseIter: pebbleBaseIter{
			i:     it,
			start: start,
			end:   end,
		},
	}
}

type pebbleBaseIter struct {
	i *pebble.Iterator

	init bool

	start, end []byte
}

func (pi *pebbleBaseIter) Domain() (start []byte, end []byte) {
	return pi.start, pi.end
}

// TODO no error handling on the Iterator interface
// func (pi *pebbleBaseIter) Error() error {
// 	return pi.i.Error()
// }

func (pi *pebbleBaseIter) Value() []byte {
	return pi.i.Value()
}

func (pi *pebbleBaseIter) Key() []byte {
	return pi.i.Key()
}

func (pi *pebbleBaseIter) Valid() bool {
	return pi.i.Valid()
}

func (pi *pebbleBaseIter) Close() {
	// TODO no error handling
	pi.i.Close()
}

var _ db.Iterator = &pebbleIter{}

type pebbleIter struct {
	pebbleBaseIter
}

func (pi *pebbleIter) Next() {
	if !pi.init {
		pi.init = true

		pi.i.First()
	}

	pi.i.Next()
}

var _ db.Iterator = &pebbleReverseIter{}

type pebbleReverseIter struct {
	pebbleBaseIter
}

func (pi *pebbleReverseIter) Next() {
	if !pi.init {
		pi.init = true

		pi.i.Last()
	}

	pi.i.Prev()
}
