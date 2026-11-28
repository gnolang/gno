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

// Sizes measured by replaying gnoland-1 blocks; see
// tm2/adr/pr6480_pebble_block_cache_size.md.
const (
	// blockCacheSize replaces pebble's 8 MiB default. pebble reserves
	// 2 x MemTableSize of the cache for memtables (the live one and one kept
	// for reuse), so the default 4 MiB memtable took all 8 MiB and no block
	// stayed cached. 64 MiB leaves 32 MiB for blocks next to 16 MiB
	// memtables. The cache fills only as blocks are read.
	blockCacheSize = 64 << 20
	// memTableSize replaces pebble's 4 MiB default, which flushed and
	// compacted several times per busy block.
	memTableSize = 16 << 20
)

// DefaultPebbleOptions returns the pebble.Options every node database opens
// with. The caller holds a reference on the returned Cache: drop it with
// Unref once pebble.Open has taken its own.
func DefaultPebbleOptions() *pebble.Options {
	return &pebble.Options{
		Cache:        pebble.NewCache(blockCacheSize),
		MemTableSize: memTableSize,
	}
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

// NewPebbleDB opens a database with DefaultPebbleOptions. It is the creator
// the db registry uses, so every node database opens through it.
func NewPebbleDB(name string, dir string) (*PebbleDB, error) {
	opts := DefaultPebbleOptions()
	defer opts.Cache.Unref() // pebble.Open takes its own reference
	return NewPebbleDBWithOpts(name, dir, opts)
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
	it, err := pdb.db.NewIter(&pebble.IterOptions{
		LowerBound: start,
		UpperBound: end,
	})
	if err != nil {
		return nil, err
	}

	return newPebbleDBIterator(it, start, end, false), nil
}

// Implements DB.
func (pdb *PebbleDB) ReverseIterator(start, end []byte) (db.Iterator, error) {
	it, err := pdb.db.NewIter(nil)
	if err != nil {
		return nil, err
	}

	return newPebbleDBIterator(it, start, end, true), nil
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
	it, err := s.snapshot.NewIter(&pebble.IterOptions{
		LowerBound: start,
		UpperBound: end,
	})
	if err != nil {
		return nil, err
	}
	return newPebbleDBIterator(it, start, end, false), nil
}

// Implements Snapshot.
func (s *pebbleSnapshot) ReverseIterator(start, end []byte) (db.Iterator, error) {
	it, err := s.snapshot.NewIter(nil)
	if err != nil {
		return nil, err
	}
	return newPebbleDBIterator(it, start, end, true), nil
}
