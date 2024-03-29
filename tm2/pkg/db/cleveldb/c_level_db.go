//go:build cgo

package cleveldb

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/internal"
	"github.com/jmhodges/levigo"
)

func init() {
	dbCreator := func(name string, dir string) (db.DB, error) {
		return NewCLevelDB(name, dir)
	}
	db.InternalRegisterDBCreator(db.CLevelDBBackend, dbCreator, false)
}

var _ db.DB = (*CLevelDB)(nil)

type CLevelDB struct {
	db     *levigo.DB
	ro     *levigo.ReadOptions
	wo     *levigo.WriteOptions
	woSync *levigo.WriteOptions
}

func NewCLevelDB(name string, dir string) (*CLevelDB, error) {
	dbPath := filepath.Join(dir, name+".db")

	opts := levigo.NewOptions()
	opts.SetCache(levigo.NewLRUCache(1 << 30))
	opts.SetCreateIfMissing(true)
	db, err := levigo.Open(dbPath, opts)
	if err != nil {
		return nil, err
	}
	ro := levigo.NewReadOptions()
	wo := levigo.NewWriteOptions()
	woSync := levigo.NewWriteOptions()
	woSync.SetSync(true)
	database := &CLevelDB{
		db:     db,
		ro:     ro,
		wo:     wo,
		woSync: woSync,
	}
	return database, nil
}

// Implements DB.
func (db *CLevelDB) Get(key []byte) []byte {
	key = internal.NonNilBytes(key)
	res, err := db.db.Get(db.ro, key)
	if err != nil {
		panic(err)
	}
	return res
}

// Implements DB.
func (db *CLevelDB) Has(key []byte) bool {
	return db.Get(key) != nil
}

// Implements DB.
func (db *CLevelDB) Set(key []byte, value []byte) {
	key = internal.NonNilBytes(key)
	value = internal.NonNilBytes(value)
	err := db.db.Put(db.wo, key, value)
	if err != nil {
		panic(err)
	}
}

// Implements DB.
func (db *CLevelDB) SetSync(key []byte, value []byte) {
	key = internal.NonNilBytes(key)
	value = internal.NonNilBytes(value)
	err := db.db.Put(db.woSync, key, value)
	if err != nil {
		panic(err)
	}
}

// Implements DB.
func (db *CLevelDB) Delete(key []byte) {
	key = internal.NonNilBytes(key)
	err := db.db.Delete(db.wo, key)
	if err != nil {
		panic(err)
	}
}

// Implements DB.
func (db *CLevelDB) DeleteSync(key []byte) {
	key = internal.NonNilBytes(key)
	err := db.db.Delete(db.woSync, key)
	if err != nil {
		panic(err)
	}
}

func (db *CLevelDB) DB() *levigo.DB {
	return db.db
}

// Implements DB.
func (db *CLevelDB) Close() {
	db.db.Close()
	db.ro.Close()
	db.wo.Close()
	db.woSync.Close()
}

// Implements DB.
func (db *CLevelDB) Print() {
	itr := db.Iterator(nil, nil)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		key := itr.Key()
		value := itr.Value()
		fmt.Printf("[%X]:\t[%X]\n", key, value)
	}
}

// Implements DB.
func (db *CLevelDB) Stats() map[string]string {
	keys := []string{
		"leveldb.aliveiters",
		"leveldb.alivesnaps",
		"leveldb.blockpool",
		"leveldb.cachedblock",
		"leveldb.num-files-at-level{n}",
		"leveldb.openedtables",
		"leveldb.sstables",
		"leveldb.stats",
	}

	stats := make(map[string]string, len(keys))
	for _, key := range keys {
		str := db.db.PropertyValue(key)
		stats[key] = str
	}
	return stats
}

//----------------------------------------
// Batch

// Implements DB.
func (db *CLevelDB) NewBatch() db.Batch {
	batch := levigo.NewWriteBatch()
	return &cLevelDBBatch{db, batch}
}

type cLevelDBBatch struct {
	db    *CLevelDB
	batch *levigo.WriteBatch
}

// Implements Batch.
func (mBatch *cLevelDBBatch) Set(key, value []byte) {
	mBatch.batch.Put(key, value)
}

// Implements Batch.
func (mBatch *cLevelDBBatch) Delete(key []byte) {
	mBatch.batch.Delete(key)
}

// Implements Batch.
func (mBatch *cLevelDBBatch) Write() {
	err := mBatch.db.db.Write(mBatch.db.wo, mBatch.batch)
	if err != nil {
		panic(err)
	}
}

// Implements Batch.
func (mBatch *cLevelDBBatch) WriteSync() {
	err := mBatch.db.db.Write(mBatch.db.woSync, mBatch.batch)
	if err != nil {
		panic(err)
	}
}

// Implements Batch.
func (mBatch *cLevelDBBatch) Close() {
	mBatch.batch.Close()
}

//----------------------------------------
// Iterator
// NOTE This is almost identical to db/go_level_db.Iterator
// Before creating a third version, refactor.

func (db *CLevelDB) Iterator(start, end []byte) db.Iterator {
	itr := db.db.NewIterator(db.ro)
	return newCLevelDBIterator(itr, start, end, false)
}

func (db *CLevelDB) ReverseIterator(start, end []byte) db.Iterator {
	itr := db.db.NewIterator(db.ro)
	return newCLevelDBIterator(itr, start, end, true)
}

var _ db.Iterator = (*cLevelDBIterator)(nil)

type cLevelDBIterator struct {
	source     *levigo.Iterator
	start, end []byte
	isReverse  bool
	isInvalid  bool
}

func newCLevelDBIterator(source *levigo.Iterator, start, end []byte, isReverse bool) *cLevelDBIterator {
	if isReverse {
		if end == nil {
			source.SeekToLast()
		} else {
			source.Seek(end)
			if source.Valid() {
				eoakey := source.Key() // end or after key
				if bytes.Compare(end, eoakey) <= 0 {
					source.Prev()
				}
			} else {
				source.SeekToLast()
			}
		}
	} else {
		if start == nil {
			source.SeekToFirst()
		} else {
			source.Seek(start)
		}
	}
	return &cLevelDBIterator{
		source:    source,
		start:     start,
		end:       end,
		isReverse: isReverse,
		isInvalid: false,
	}
}

func (itr cLevelDBIterator) Domain() ([]byte, []byte) {
	return itr.start, itr.end
}

func (itr cLevelDBIterator) Valid() bool {
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

	// It's valid.
	return true
}

func (itr cLevelDBIterator) Key() []byte {
	itr.assertNoError()
	itr.assertIsValid()
	return itr.source.Key()
}

func (itr cLevelDBIterator) Value() []byte {
	itr.assertNoError()
	itr.assertIsValid()
	return itr.source.Value()
}

func (itr cLevelDBIterator) Next() {
	itr.assertNoError()
	itr.assertIsValid()
	if itr.isReverse {
		itr.source.Prev()
	} else {
		itr.source.Next()
	}
}

func (itr cLevelDBIterator) Close() {
	itr.source.Close()
}

func (itr cLevelDBIterator) assertNoError() {
	if err := itr.source.GetError(); err != nil {
		panic(err)
	}
}

func (itr cLevelDBIterator) assertIsValid() {
	if !itr.Valid() {
		panic("cLevelDBIterator is invalid")
	}
}
