package pebbledb

import (
	"bytes"
	goerrors "errors"
	"path/filepath"
	"slices"

	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/bloom"

	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/internal"
)

// DefaultPebbleOptions returns pebble.Options tuned for gno.land production.
func DefaultPebbleOptions() *pebble.Options {
	cache := pebble.NewCache(500 << 20) // 500 MB
	opts := &pebble.Options{
		Cache:                    cache,
		MemTableSize:             64 << 20, // 64 MB
		MaxConcurrentCompactions: func() int { return 3 },
		Levels: []pebble.LevelOptions{
			{FilterPolicy: bloom.FilterPolicy(10)},
		},
	}
	return opts
}

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
func (pdb *PebbleDB) Get(key []byte) ([]byte, error) {
	key = internal.NonNilBytes(key)
	res, closer, err := pdb.db.Get(key)
	if err != nil {
		if goerrors.Is(err, pebble.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// The caller should not modify the contents of the returned slice,
	// but it is safe to modify the contents of the argument after db.Get returns.
	// The returned slice will remain valid until the returned Closer is closed.
	// On success, the caller MUST call closer.Close() or a memory leak will occur.
	defer closer.Close()
	out := make([]byte, len(res))
	copy(out, res)
	return out, nil
}

// Implements DB.
func (pdb *PebbleDB) Has(key []byte) (bool, error) {
	v, err := pdb.Get(key)
	if err != nil {
		return false, err
	}
	return v != nil, nil
}

// Implements DB.
func (pdb *PebbleDB) Set(key []byte, value []byte) error {
	key = internal.NonNilBytes(key)
	value = internal.NonNilBytes(value)
	return pdb.db.Set(key, value, pebble.NoSync)
}

// Implements DB.
func (pdb *PebbleDB) SetSync(key []byte, value []byte) error {
	key = internal.NonNilBytes(key)
	value = internal.NonNilBytes(value)
	return pdb.db.Set(key, value, pebble.Sync)
}

// Implements DB.
func (pdb *PebbleDB) Delete(key []byte) error {
	key = internal.NonNilBytes(key)
	return pdb.db.Delete(key, pebble.NoSync)
}

// Implements DB.
func (pdb *PebbleDB) DeleteSync(key []byte) error {
	key = internal.NonNilBytes(key)
	return pdb.db.Delete(key, pebble.Sync)
}

// Implements DB.
func (pdb *PebbleDB) Close() error {
	return pdb.db.Close()
}

// Implements DB.
func (pdb *PebbleDB) Print() error {
	return nil
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

// Implements DB.
func (pdb *PebbleDB) NewBatchWithSize(s int) db.Batch {
	return &pebbleDBBatch{pdb, pdb.db.NewBatchWithSize(s)}
}

type pebbleDBBatch struct {
	db    *PebbleDB
	batch *pebble.Batch
}

// Implements Batch.
func (mBatch *pebbleDBBatch) Set(key, value []byte) error {
	return mBatch.batch.Set(internal.NonNilBytes(key), internal.NonNilBytes(value), pebble.NoSync)
}

// Implements Batch.
func (mBatch *pebbleDBBatch) Delete(key []byte) error {
	return mBatch.batch.Delete(key, pebble.NoSync)
}

// Implements Batch.
func (mBatch *pebbleDBBatch) Write() error {
	return mBatch.batch.Commit(pebble.Sync)
}

// Implements Batch.
func (mBatch *pebbleDBBatch) WriteSync() error {
	return mBatch.batch.Commit(pebble.Sync)
}

// Implements Batch.
func (mBatch *pebbleDBBatch) Close() error {
	return mBatch.batch.Close()
}

// Implements Batch.
func (mBatch *pebbleDBBatch) GetByteSize() (int, error) {
	return mBatch.batch.Len(), nil
}

// Implements DB.
func (pdb *PebbleDB) Iterator(start, end []byte) (db.Iterator, error) {
	return newRangeIterator(pdb.db, start, end, false)
}

// Implements DB.
func (pdb *PebbleDB) ReverseIterator(start, end []byte) (db.Iterator, error) {
	return newRangeIterator(pdb.db, start, end, true)
}

func newRangeIterator(r pebble.Reader, start, end []byte, reverse bool) (db.Iterator, error) {
	source, err := r.NewIter(&pebble.IterOptions{
		LowerBound: start,
		UpperBound: end,
	})
	if err != nil {
		return nil, err
	}
	if reverse {
		source.Last()
	} else {
		source.First()
	}
	return &pebbleDBIterator{
		source:    source,
		start:     start,
		end:       end,
		isReverse: reverse,
	}, nil
}

type pebbleDBIterator struct {
	source    *pebble.Iterator
	start     []byte
	end       []byte
	isReverse bool
	isInvalid bool
}

var _ db.Iterator = (*pebbleDBIterator)(nil)

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

	// pebble can ignore an empty UpperBound, which then admits every key.
	if itr.end != nil && bytes.Compare(itr.end, itr.source.Key()) <= 0 {
		itr.isInvalid = true
		return false
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
func (itr *pebbleDBIterator) Close() error {
	return itr.source.Close()
}

// Implements Iterator.
func (itr *pebbleDBIterator) Error() error {
	return itr.source.Error()
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

func (pdb *PebbleDB) NewSnapshot() (db.Snapshot, error) {
	snap := pdb.db.NewSnapshot()
	return &pebbleSnapshot{snapshot: snap}, nil
}

type pebbleSnapshot struct {
	snapshot *pebble.Snapshot
}

var _ db.Snapshot = (*pebbleSnapshot)(nil)

func (s *pebbleSnapshot) Close() error {
	return s.snapshot.Close()
}

// Implements Snapshot.
func (s *pebbleSnapshot) Get(key []byte) ([]byte, error) {
	key = internal.NonNilBytes(key)
	res, closer, err := s.snapshot.Get(key)
	if err != nil {
		if goerrors.Is(err, pebble.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// The caller should not modify the contents of the returned slice,
	// but it is safe to modify the contents of the argument after db.Get returns.
	// The returned slice will remain valid until the returned Closer is closed.
	// On success, the caller MUST call closer.Close() or a memory leak will occur.
	defer closer.Close()
	out := make([]byte, len(res))
	copy(out, res)
	return out, nil
}

// Implements Snapshot.
func (s *pebbleSnapshot) Has(key []byte) (bool, error) {
	v, err := s.Get(key)
	if err != nil {
		return false, err
	}
	return v != nil, nil
}

// Implements Snapshot.
func (s *pebbleSnapshot) Iterator(start, end []byte) (db.Iterator, error) {
	return newRangeIterator(s.snapshot, start, end, false)
}

// Implements Snapshot.
func (s *pebbleSnapshot) ReverseIterator(start, end []byte) (db.Iterator, error) {
	return newRangeIterator(s.snapshot, start, end, true)
}
