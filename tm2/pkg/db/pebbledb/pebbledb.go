package pebbledb

import (
	"bytes"
	goerrors "errors"
	"path/filepath"
	"slices"

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
	return NewPebbleDBWithOpts(name, dir, &pebble.Options{})
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
func (pdb *PebbleDB) Get(key []byte) []byte {
	key = internal.NonNilBytes(key)
	res, closer, err := pdb.db.Get(key)
	if err != nil {
		if goerrors.Is(err, pebble.ErrNotFound) {
			return nil
		}
		panic(err)
	}

	// The caller should not modify the contents of the returned slice,
	// but it is safe to modify the contents of the argument after db.Get returns.
	// The returned slice will remain valid until the returned Closer is closed.
	// On success, the caller MUST call closer.Close() or a memory leak will occur.
	defer closer.Close()
	out := make([]byte, len(res))
	copy(out, res)
	return out
}

// Implements DB.
func (pdb *PebbleDB) Has(key []byte) bool {
	return pdb.Get(key) != nil
}

// Implements DB.
func (pdb *PebbleDB) Set(key []byte, value []byte) {
	key = internal.NonNilBytes(key)
	value = internal.NonNilBytes(value)
	err := pdb.db.Set(key, value, pebble.NoSync)
	if err != nil {
		panic(err)
	}
}

// Implements DB.
func (pdb *PebbleDB) SetSync(key []byte, value []byte) {
	key = internal.NonNilBytes(key)
	value = internal.NonNilBytes(value)
	err := pdb.db.Set(key, value, pebble.Sync)
	if err != nil {
		panic(err)
	}
}

// Implements DB.
func (pdb *PebbleDB) Delete(key []byte) {
	key = internal.NonNilBytes(key)
	err := pdb.db.Delete(key, pebble.NoSync)
	if err != nil {
		panic(err)
	}
}

// Implements DB.
func (pdb *PebbleDB) DeleteSync(key []byte) {
	key = internal.NonNilBytes(key)
	err := pdb.db.Delete(key, pebble.Sync)
	if err != nil {
		panic(err)
	}
}

// Implements DB.
func (pdb *PebbleDB) Close() error {
	return pdb.db.Close()
}

// Implements DB.
func (pdb *PebbleDB) Print() {
}

// Implements DB.
func (pdb *PebbleDB) Stats() map[string]string {
	return nil
}

// ----------------------------------------
// Batch

// Implements DB.
func (pdb *PebbleDB) NewBatch() db.Batch {
	return &pebbleDBBatch{pdb, pdb.db.NewBatch()}
}

type pebbleDBBatch struct {
	db    *PebbleDB
	batch *pebble.Batch
}

// Implements Batch.
func (mBatch *pebbleDBBatch) Set(key, value []byte) {
	if err := mBatch.batch.Set(key, value, pebble.NoSync); err != nil {
		panic(err)
	}
}

// Implements Batch.
func (mBatch *pebbleDBBatch) Delete(key []byte) {
	if err := mBatch.batch.Delete(key, pebble.NoSync); err != nil {
		panic(err)
	}
}

// Implements Batch.
func (mBatch *pebbleDBBatch) Write() {
	if err := mBatch.batch.Commit(pebble.Sync); err != nil {
		panic(err)
	}
}

// Implements Batch.
func (mBatch *pebbleDBBatch) WriteSync() {
	if err := mBatch.batch.Commit(pebble.Sync); err != nil {
		panic(err)
	}
}

// Implements Batch.
func (mBatch *pebbleDBBatch) Close() {
	mBatch.batch.Close()
}

// Implements DB.
func (pdb *PebbleDB) Iterator(start, end []byte) db.Iterator {
	it, err := pdb.db.NewIter(&pebble.IterOptions{
		LowerBound: start,
		UpperBound: end,
	})
	if err != nil {
		panic(err)
	}

	return newPebbleDBIterator(it, start, end, false)
}

// Implements DB.
func (pdb *PebbleDB) ReverseIterator(start, end []byte) db.Iterator {
	it, err := pdb.db.NewIter(nil)
	if err != nil {
		panic(err)
	}

	return newPebbleDBIterator(it, start, end, true)
}

type pebbleDBIterator struct {
	source    *pebble.Iterator
	start     []byte
	end       []byte
	isReverse bool
	isInvalid bool
}

var _ db.Iterator = (*pebbleDBIterator)(nil)

func newPebbleDBIterator(source *pebble.Iterator, start, end []byte, isReverse bool) *pebbleDBIterator {
	if isReverse {
		if end == nil {
			source.Last()
		} else {
			valid := source.SeekGE(end)
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
			source.SeekGE(start)
		}
	}
	return &pebbleDBIterator{
		source:    source,
		start:     start,
		end:       end,
		isReverse: isReverse,
		isInvalid: false,
	}
}

// Implements Iterator.
func (itr *pebbleDBIterator) Domain() ([]byte, []byte) {
	return itr.start, itr.end
}

// Implements Iterator.
func (itr *pebbleDBIterator) Valid() bool {
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
func (itr *pebbleDBIterator) Key() []byte {
	itr.assertNoError()
	itr.assertIsValid()
	return slices.Clone(itr.source.Key())
}

// Implements Iterator.
func (itr *pebbleDBIterator) Value() []byte {
	itr.assertNoError()
	itr.assertIsValid()
	return slices.Clone(itr.source.Value())
}

// Implements Iterator.
func (itr *pebbleDBIterator) Next() {
	itr.assertNoError()
	itr.assertIsValid()
	if itr.isReverse {
		itr.source.Prev()
	} else {
		itr.source.Next()
	}
}

// Implements Iterator.
func (itr *pebbleDBIterator) Close() {
	itr.source.Close()
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
