package db

import (
	goerrors "errors"
	"fmt"
	"path/filepath"

	"github.com/gnolang/goleveldb/leveldb"
	"github.com/gnolang/goleveldb/leveldb/errors"
	"github.com/gnolang/goleveldb/leveldb/iterator"
	"github.com/gnolang/goleveldb/leveldb/opt"
	"github.com/gnolang/goleveldb/leveldb/util"
)

func init() {
	dbCreator := func(name string, dir string) (DB, error) {
		return NewGoLevelDB(name, dir)
	}
	registerDBCreator(GoLevelDBBackend, dbCreator, false)
}

var _ DB = (*GoLevelDB)(nil)

type GoLevelDB struct {
	db *leveldb.DB
}

func NewGoLevelDB(name string, dir string) (*GoLevelDB, error) {
	return NewGoLevelDBWithOpts(name, dir, nil)
}

func NewGoLevelDBWithOpts(name string, dir string, o *opt.Options) (*GoLevelDB, error) {
	dbPath := filepath.Join(dir, name+".db")
	db, err := leveldb.OpenFile(dbPath, o)
	if err != nil {
		return nil, err
	}
	database := &GoLevelDB{
		db: db,
	}
	return database, nil
}

// Implements DB.
func (db *GoLevelDB) Get(key []byte) []byte {
	key = nonNilBytes(key)
	res, err := db.db.Get(key, nil)
	if err != nil {
		if goerrors.Is(err, errors.ErrNotFound) {
			return nil
		}
		panic(err)
	}
	return res
}

// Implements DB.
func (db *GoLevelDB) Has(key []byte) bool {
	return db.Get(key) != nil
}

// Implements DB.
func (db *GoLevelDB) Set(key []byte, value []byte) {
	key = nonNilBytes(key)
	value = nonNilBytes(value)
	err := db.db.Put(key, value, nil)
	if err != nil {
		panic(err)
	}
}

// Implements DB.
func (db *GoLevelDB) SetSync(key []byte, value []byte) {
	key = nonNilBytes(key)
	value = nonNilBytes(value)
	err := db.db.Put(key, value, &opt.WriteOptions{Sync: true})
	if err != nil {
		panic(err)
	}
}

// Implements DB.
func (db *GoLevelDB) Delete(key []byte) {
	key = nonNilBytes(key)
	err := db.db.Delete(key, nil)
	if err != nil {
		panic(err)
	}
}

// Implements DB.
func (db *GoLevelDB) DeleteSync(key []byte) {
	key = nonNilBytes(key)
	err := db.db.Delete(key, &opt.WriteOptions{Sync: true})
	if err != nil {
		panic(err)
	}
}

func (db *GoLevelDB) DB() *leveldb.DB {
	return db.db
}

// Implements DB.
func (db *GoLevelDB) Close() {
	db.db.Close()
}

// Implements DB.
func (db *GoLevelDB) Print() {
	str, _ := db.db.GetProperty("leveldb.stats")
	fmt.Printf("%v\n", str)

	itr := db.db.NewIterator(nil, &opt.ReadOptions{DontFillCache: true})
	for itr.Next() {
		key := itr.Key()
		value := itr.Value()
		fmt.Printf("[%X]:\t[%X]\n", key, value)
	}
	itr.Release()
}

// Implements DB.
func (db *GoLevelDB) Stats() map[string]string {
	keys := []string{
		"leveldb.num-files-at-level{n}",
		"leveldb.stats",
		"leveldb.iostats",
		"leveldb.writedelay",
		"leveldb.sstables",
		"leveldb.blockpool",
		"leveldb.cachedblock",
		"leveldb.openedtables",
		"leveldb.alivesnaps",
		"leveldb.aliveiters",
	}

	stats := make(map[string]string)
	for _, key := range keys {
		str, err := db.db.GetProperty(key)
		if err == nil {
			stats[key] = str
		}
	}
	return stats
}

// ----------------------------------------
// Batch

// Implements DB.
func (db *GoLevelDB) NewBatch() Batch {
	batch := new(leveldb.Batch)
	return &goLevelDBBatch{db, batch}
}

type goLevelDBBatch struct {
	db    *GoLevelDB
	batch *leveldb.Batch
}

// Implements Batch.
func (mBatch *goLevelDBBatch) Set(key, value []byte) {
	mBatch.batch.Put(nonNilBytes(key), nonNilBytes(value))
}

// Implements Batch.
func (mBatch *goLevelDBBatch) Delete(key []byte) {
	mBatch.batch.Delete(nonNilBytes(key))
}

// Implements Batch.
func (mBatch *goLevelDBBatch) Write() {
	err := mBatch.db.db.Write(mBatch.batch, &opt.WriteOptions{Sync: false})
	if err != nil {
		panic(err)
	}
}

// Implements Batch.
func (mBatch *goLevelDBBatch) WriteSync() {
	err := mBatch.db.db.Write(mBatch.batch, &opt.WriteOptions{Sync: true})
	if err != nil {
		panic(err)
	}
}

// Implements Batch.
// Close is no-op for goLevelDBBatch.
func (mBatch *goLevelDBBatch) Close() {}

// ----------------------------------------
// Iterator
// NOTE This is almost identical to db/c_level_db.Iterator
// Before creating a third version, refactor.

// Implements DB.
func (db *GoLevelDB) Iterator(start, end []byte) Iterator {
	return db.newGoLevelDBIterator(start, end, false)
}

// Implements DB.
func (db *GoLevelDB) ReverseIterator(start, end []byte) Iterator {
	return db.newGoLevelDBIterator(start, end, true)
}

type goLevelDBIterator struct {
	source  iterator.Iterator
	start   []byte
	end     []byte
	reverse bool
}

var _ Iterator = (*goLevelDBIterator)(nil)

func (db *GoLevelDB) newGoLevelDBIterator(start, end []byte, reverse bool) *goLevelDBIterator {
	source := db.db.NewIterator(&util.Range{Start: start, Limit: end}, nil)
	if reverse {
		source.Last()
	} else {
		source.First()
	}
	return &goLevelDBIterator{
		source:  source,
		start:   start,
		end:     end,
		reverse: reverse,
	}
}

// Implements Iterator.
func (itr *goLevelDBIterator) Domain() ([]byte, []byte) {
	return itr.start, itr.end
}

// Implements Iterator.
func (itr *goLevelDBIterator) Valid() bool {
	// Panic on DB error. No way to recover.
	itr.assertNoError()

	// If source is invalid, invalid.
	if !itr.source.Valid() {
		return false
	}
	return true
}

// Implements Iterator.
func (itr *goLevelDBIterator) Key() []byte {
	// Key returns a copy of the current key.
	// See https://github.com/gnolang/goleveldb/blob/52c212e6c196a1404ea59592d3f1c227c9f034b2/leveldb/iterator/iter.go#L88
	itr.assertIsValid()
	return cp(itr.source.Key())
}

// Implements Iterator.
func (itr *goLevelDBIterator) Value() []byte {
	// Value returns a copy of the current value.
	// See https://github.com/gnolang/goleveldb/blob/52c212e6c196a1404ea59592d3f1c227c9f034b2/leveldb/iterator/iter.go#L88
	itr.assertIsValid()
	return cp(itr.source.Value())
}

// Implements Iterator.
func (itr *goLevelDBIterator) Next() {
	itr.assertIsValid()
	if itr.reverse {
		itr.source.Prev()
	} else {
		itr.source.Next()
	}
}

// Implements Iterator.
func (itr *goLevelDBIterator) Close() {
	itr.source.Release()
}

func (itr *goLevelDBIterator) assertNoError() {
	if err := itr.source.Error(); err != nil {
		panic(err)
	}
}

func (itr goLevelDBIterator) assertIsValid() {
	if !itr.Valid() {
		panic("goLevelDBIterator is invalid")
	}
}
