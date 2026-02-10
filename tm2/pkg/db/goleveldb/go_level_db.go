package goleveldb

import (
	"bytes"
	goerrors "errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/gnolang/gno/tm2/pkg/colors"
	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/internal"
)

func init() {
	dbCreator := func(name string, dir string) (db.DB, error) {
		return NewGoLevelDB(name, dir)
	}
	db.InternalRegisterDBCreator(db.GoLevelDBBackend, dbCreator, false)
}

var _ db.DB = (*GoLevelDB)(nil)

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
func (db *GoLevelDB) Get(key []byte) ([]byte, error) {
	key = internal.NonNilBytes(key)
	res, err := db.db.Get(key, nil)
	if err != nil {
		if goerrors.Is(err, errors.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return res, nil
}

// Implements DB.
func (db *GoLevelDB) Has(key []byte) (bool, error) {
	v, err := db.Get(key)
	return v != nil, err
}

// Implements DB.
func (db *GoLevelDB) Set(key []byte, value []byte) error {
	key = internal.NonNilBytes(key)
	value = internal.NonNilBytes(value)
	return db.db.Put(key, value, nil)
}

// Implements DB.
func (db *GoLevelDB) SetSync(key []byte, value []byte) error {
	key = internal.NonNilBytes(key)
	value = internal.NonNilBytes(value)
	return db.db.Put(key, value, &opt.WriteOptions{Sync: true})
}

// Implements DB.
func (db *GoLevelDB) Delete(key []byte) error {
	key = internal.NonNilBytes(key)
	return db.db.Delete(key, nil)
}

// Implements DB.
func (db *GoLevelDB) DeleteSync(key []byte) error {
	key = internal.NonNilBytes(key)
	return db.db.Delete(key, &opt.WriteOptions{Sync: true})
}

func (db *GoLevelDB) DB() *leveldb.DB {
	return db.db
}

// Implements DB.
func (db *GoLevelDB) Close() error {
	return db.db.Close()
}

// Implements DB.
func (db *GoLevelDB) Print() error {
	str, _ := db.db.GetProperty("leveldb.stats")
	fmt.Printf("%v\n", str)

	itr := db.db.NewIterator(nil, nil)
	for itr.Next() {
		key := colors.DefaultColoredBytesN(itr.Key(), 50)
		value := colors.DefaultColoredBytesN(itr.Value(), 100)
		fmt.Printf("%v: %v\n", key, value)
	}
	return nil
}

// Implements DB.
func (db *GoLevelDB) Stats() map[string]string {
	keys := []string{
		"leveldb.num-files-at-level{n}",
		"leveldb.stats",
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
func (db *GoLevelDB) NewBatch() db.Batch {
	batch := new(leveldb.Batch)
	return &goLevelDBBatch{db, batch}
}

// Implements DB.
func (db *GoLevelDB) NewBatchWithSize(size int) db.Batch {
	batch := leveldb.MakeBatch(size)
	return &goLevelDBBatch{db, batch}
}

type goLevelDBBatch struct {
	db    *GoLevelDB
	batch *leveldb.Batch
}

// Implements Batch.
func (mBatch *goLevelDBBatch) Set(key, value []byte) error {
	mBatch.batch.Put(key, value)
	return nil
}

// Implements Batch.
func (mBatch *goLevelDBBatch) Delete(key []byte) error {
	mBatch.batch.Delete(key)
	return nil
}

// Implements Batch.
func (mBatch *goLevelDBBatch) Write() error {
	return mBatch.db.db.Write(mBatch.batch, &opt.WriteOptions{Sync: false})
}

// Implements Batch.
func (mBatch *goLevelDBBatch) WriteSync() error {
	return mBatch.db.db.Write(mBatch.batch, &opt.WriteOptions{Sync: true})
}

// Implements Batch.
// Close is no-op for goLevelDBBatch.
func (mBatch *goLevelDBBatch) Close() error { return nil }

// Implements Batch
func (mBatch *goLevelDBBatch) GetByteSize() (int, error) {
	if mBatch.batch == nil {
		return 0, errors.New("goleveldb: batch has been written or closed")
	}
	return len(mBatch.batch.Dump()), nil
}

// ----------------------------------------
// Iterator
// NOTE This is almost identical to db/c_level_db.Iterator
// Before creating a third version, refactor.

// Implements DB.
func (db *GoLevelDB) Iterator(start, end []byte) (db.Iterator, error) {
	itr := db.db.NewIterator(nil, nil)
	return newGoLevelDBIterator(itr, start, end, false), nil
}

// Implements DB.
func (db *GoLevelDB) ReverseIterator(start, end []byte) (db.Iterator, error) {
	itr := db.db.NewIterator(nil, nil)
	return newGoLevelDBIterator(itr, start, end, true), nil
}

type goLevelDBIterator struct {
	source    iterator.Iterator
	start     []byte
	end       []byte
	isReverse bool
	isInvalid bool
}

var _ db.Iterator = (*goLevelDBIterator)(nil)

func newGoLevelDBIterator(source iterator.Iterator, start, end []byte, isReverse bool) *goLevelDBIterator {
	if isReverse {
		if end == nil {
			source.Last()
		} else {
			valid := source.Seek(end)
			if valid {
				eoakey := source.Key() // end or after key
				if bytes.Compare(end, eoakey) <= 0 {
					source.Prev()
				}
			} else {
				source.Last()
			}
		}
	} else {
		if start == nil {
			source.First()
		} else {
			source.Seek(start)
		}
	}
	return &goLevelDBIterator{
		source:    source,
		start:     start,
		end:       end,
		isReverse: isReverse,
		isInvalid: false,
	}
}

// Implements Iterator.
func (itr *goLevelDBIterator) Domain() ([]byte, []byte) {
	return itr.start, itr.end
}

// Implements Iterator.
func (itr *goLevelDBIterator) Valid() bool {
	// Once invalid, forever invalid.
	if itr.isInvalid {
		return false
	}

	// Panic on DB error.  No way to recover.
	itr.assertNoError()

	// If source is invalid, invalid.
	if !itr.source.Valid() {
		itr.isInvalid = true
		return false
	}

	// If key is end or past it, invalid.
	start := itr.start
	end := itr.end
	key := itr.source.Key()

	if itr.isReverse {
		if start != nil && bytes.Compare(key, start) < 0 {
			itr.isInvalid = true
			return false
		}
	} else {
		if end != nil && bytes.Compare(end, key) <= 0 {
			itr.isInvalid = true
			return false
		}
	}

	// Valid
	return true
}

// Implements Iterator.
func (itr *goLevelDBIterator) Key() []byte {
	// Key returns a copy of the current key.
	// See https://github.com/syndtr/goleveldb/blob/52c212e6c196a1404ea59592d3f1c227c9f034b2/leveldb/iterator/iter.go#L88
	itr.assertNoError()
	itr.assertIsValid()
	return slices.Clone(itr.source.Key())
}

// Implements Iterator.
func (itr *goLevelDBIterator) Value() []byte {
	// Value returns a copy of the current value.
	// See https://github.com/syndtr/goleveldb/blob/52c212e6c196a1404ea59592d3f1c227c9f034b2/leveldb/iterator/iter.go#L88
	itr.assertNoError()
	itr.assertIsValid()
	return slices.Clone(itr.source.Value())
}

// Implements Iterator.
func (itr *goLevelDBIterator) Next() {
	itr.assertNoError()
	itr.assertIsValid()
	if itr.isReverse {
		itr.source.Prev()
	} else {
		itr.source.Next()
	}
}

func (itr *goLevelDBIterator) Error() error {
	return itr.source.Error()
}

// Implements Iterator.
func (itr *goLevelDBIterator) Close() error {
	itr.source.Release()
	return nil
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
