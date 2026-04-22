//go:build cgo

package lmdbdb

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/bmatsuo/lmdb-go/lmdb"

	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/internal"
)

const (
	LMDBBackend db.BackendType = "lmdbdb"
)

var errClosed = errors.New("lmdbdb: database is closed")
var errBatchDone = errors.New("lmdbdb: batch already written or closed")

func init() {
	db.InternalRegisterDBCreator(LMDBBackend, func(name, dir string) (db.DB, error) {
		return NewLMDB(name, dir)
	}, false)
}

var _ db.DB = (*LMDB)(nil)

// LMDB wraps an lmdb.Env + DBI handle.
type LMDB struct {
	mu     sync.RWMutex
	env    *lmdb.Env
	dbi    lmdb.DBI
	closed bool
}

// DefaultMapSize is the default maximum database size (1 TB).
const DefaultMapSize int64 = 1 << 40

// NewLMDB opens an LMDB database with production defaults.
func NewLMDB(name, dir string) (*LMDB, error) {
	return NewLMDBWithOptions(name, dir, DefaultMapSize, 0)
}

// NewLMDBWithOptions opens an LMDB database with custom map size and flags.
// Extra flags are OR'd with the base flags (NoMetaSync | WriteMap | NoReadahead).
func NewLMDBWithOptions(name, dir string, mapSize int64, extraFlags uint) (*LMDB, error) {
	dbDir := filepath.Join(dir, name+".db")
	if err := os.MkdirAll(dbDir, 0o700); err != nil {
		return nil, fmt.Errorf("error creating dir: %w", err)
	}

	env, err := lmdb.NewEnv()
	if err != nil {
		return nil, err
	}
	if err := env.SetMapSize(mapSize); err != nil {
		env.Close()
		return nil, err
	}
	if err := env.SetMaxDBs(1); err != nil {
		env.Close()
		return nil, err
	}

	flags := uint(lmdb.NoMetaSync | lmdb.WriteMap | lmdb.NoReadahead)
	flags |= extraFlags
	if err := env.Open(dbDir, flags, 0o644); err != nil {
		env.Close()
		return nil, err
	}

	// Open the default (unnamed) database.
	var dbi lmdb.DBI
	err = env.Update(func(txn *lmdb.Txn) error {
		var err error
		dbi, err = txn.OpenRoot(lmdb.Create)
		return err
	})
	if err != nil {
		env.Close()
		return nil, err
	}

	return &LMDB{env: env, dbi: dbi}, nil
}

// nonEmptyKey maps empty keys to a sentinel since LMDB does not support them.
func nonEmptyKey(key []byte) []byte {
	if len(key) == 0 {
		return []byte{0x00}
	}
	return key
}

func (l *LMDB) checkClosed() error {
	if l.closed {
		return errClosed
	}
	return nil
}

func (l *LMDB) Get(key []byte) ([]byte, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if err := l.checkClosed(); err != nil {
		return nil, err
	}
	key = nonEmptyKey(internal.NonNilBytes(key))
	var val []byte
	err := l.env.View(func(txn *lmdb.Txn) error {
		txn.RawRead = true
		v, err := txn.Get(l.dbi, key)
		if lmdb.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}
		val = slices.Clone(v)
		return nil
	})
	return val, err
}

func (l *LMDB) Has(key []byte) (bool, error) {
	v, err := l.Get(key)
	return v != nil, err
}

func (l *LMDB) Set(key, value []byte) error {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if err := l.checkClosed(); err != nil {
		return err
	}
	key = nonEmptyKey(internal.NonNilBytes(key))
	value = internal.NonNilBytes(value)
	return l.env.Update(func(txn *lmdb.Txn) error {
		return txn.Put(l.dbi, key, value, 0)
	})
}

func (l *LMDB) SetSync(key, value []byte) error {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if err := l.checkClosed(); err != nil {
		return err
	}
	key = nonEmptyKey(internal.NonNilBytes(key))
	value = internal.NonNilBytes(value)
	err := l.env.Update(func(txn *lmdb.Txn) error {
		return txn.Put(l.dbi, key, value, 0)
	})
	if err != nil {
		return err
	}
	return l.env.Sync(true)
}

func (l *LMDB) Delete(key []byte) error {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if err := l.checkClosed(); err != nil {
		return err
	}
	key = nonEmptyKey(internal.NonNilBytes(key))
	return l.env.Update(func(txn *lmdb.Txn) error {
		err := txn.Del(l.dbi, key, nil)
		if lmdb.IsNotFound(err) {
			return nil
		}
		return err
	})
}

func (l *LMDB) DeleteSync(key []byte) error {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if err := l.checkClosed(); err != nil {
		return err
	}
	key = nonEmptyKey(internal.NonNilBytes(key))
	err := l.env.Update(func(txn *lmdb.Txn) error {
		err := txn.Del(l.dbi, key, nil)
		if lmdb.IsNotFound(err) {
			return nil
		}
		return err
	})
	if err != nil {
		return err
	}
	return l.env.Sync(true)
}

func (l *LMDB) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return errClosed
	}
	l.closed = true
	return l.env.Close()
}

func (l *LMDB) Print() error {
	return nil
}

func (l *LMDB) Stats() map[string]string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.closed {
		return nil
	}
	stat, err := l.env.Stat()
	if err != nil {
		return nil
	}
	return map[string]string{
		"PageSize":      fmt.Sprintf("%d", stat.PSize),
		"Depth":         fmt.Sprintf("%d", stat.Depth),
		"BranchPages":   fmt.Sprintf("%d", stat.BranchPages),
		"LeafPages":     fmt.Sprintf("%d", stat.LeafPages),
		"OverflowPages": fmt.Sprintf("%d", stat.OverflowPages),
		"Entries":       fmt.Sprintf("%d", stat.Entries),
	}
}

func (l *LMDB) NewBatch() db.Batch {
	return &lmdbBatch{db: l}
}

func (l *LMDB) NewBatchWithSize(_ int) db.Batch {
	return &lmdbBatch{db: l}
}

// Iterator implements db.DB.
// The caller MUST call Close() on the returned iterator to release resources.
// Close() on the LMDB will block until all iterators are closed.
func (l *LMDB) Iterator(start, end []byte) (db.Iterator, error) {
	l.mu.RLock()
	if err := l.checkClosed(); err != nil {
		l.mu.RUnlock()
		return nil, err
	}
	itr, err := newLMDBIterator(l, start, end, false)
	if err != nil {
		l.mu.RUnlock()
		return nil, err
	}
	return itr, nil // RLock held until itr.Close()
}

// ReverseIterator implements db.DB.
// The caller MUST call Close() on the returned iterator to release resources.
// Close() on the LMDB will block until all iterators are closed.
func (l *LMDB) ReverseIterator(start, end []byte) (db.Iterator, error) {
	l.mu.RLock()
	if err := l.checkClosed(); err != nil {
		l.mu.RUnlock()
		return nil, err
	}
	itr, err := newLMDBIterator(l, start, end, true)
	if err != nil {
		l.mu.RUnlock()
		return nil, err
	}
	return itr, nil // RLock held until itr.Close()
}

// ----------------------------------------
// Batch

type lmdbBatch struct {
	db   *LMDB
	ops  []internal.Operation
	size int
	done bool
}

func (b *lmdbBatch) Set(key, value []byte) error {
	if b.done {
		return errBatchDone
	}
	b.size += len(key) + len(value)
	b.ops = append(b.ops, internal.Operation{OpType: internal.OpTypeSet, Key: key, Value: value})
	return nil
}

func (b *lmdbBatch) Delete(key []byte) error {
	if b.done {
		return errBatchDone
	}
	b.size += len(key)
	b.ops = append(b.ops, internal.Operation{OpType: internal.OpTypeDelete, Key: key})
	return nil
}

func (b *lmdbBatch) write(sync bool) error {
	if b.done {
		return errBatchDone
	}
	b.db.mu.RLock()
	defer b.db.mu.RUnlock()
	if err := b.db.checkClosed(); err != nil {
		return err
	}
	err := b.db.env.Update(func(txn *lmdb.Txn) error {
		for _, op := range b.ops {
			key := nonEmptyKey(internal.NonNilBytes(op.Key))
			switch op.OpType {
			case internal.OpTypeSet:
				if err := txn.Put(b.db.dbi, key, internal.NonNilBytes(op.Value), 0); err != nil {
					return err
				}
			case internal.OpTypeDelete:
				err := txn.Del(b.db.dbi, key, nil)
				if err != nil && !lmdb.IsNotFound(err) {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	b.done = true
	b.ops = nil
	if sync {
		return b.db.env.Sync(true)
	}
	return nil
}

func (b *lmdbBatch) Write() error {
	return b.write(false)
}

func (b *lmdbBatch) WriteSync() error {
	return b.write(true)
}

func (b *lmdbBatch) Close() error {
	b.done = true
	b.ops = nil
	b.size = 0
	return nil
}

func (b *lmdbBatch) GetByteSize() (int, error) {
	if b.done {
		return 0, errBatchDone
	}
	return b.size, nil
}

// ----------------------------------------
// Iterator

type lmdbIterator struct {
	db  *LMDB // for releasing RLock on Close
	txn *lmdb.Txn
	cur *lmdb.Cursor

	start, end   []byte
	currentKey   []byte
	currentValue []byte
	isReverse    bool
	isInvalid    bool
}

func newLMDBIterator(l *LMDB, start, end []byte, isReverse bool) (*lmdbIterator, error) {
	txn, err := l.env.BeginTxn(nil, lmdb.Readonly)
	if err != nil {
		return nil, err
	}
	txn.RawRead = true

	cur, err := txn.OpenCursor(l.dbi)
	if err != nil {
		txn.Abort()
		return nil, err
	}

	itr := &lmdbIterator{
		db:        l,
		txn:       txn,
		cur:       cur,
		start:     start,
		end:       end,
		isReverse: isReverse,
	}

	// Position cursor.
	if isReverse {
		if end == nil {
			itr.currentKey, itr.currentValue, err = cur.Get(nil, nil, lmdb.Last)
		} else {
			itr.currentKey, itr.currentValue, err = cur.Get(end, nil, lmdb.SetRange)
			if lmdb.IsNotFound(err) {
				// end is past all keys, go to last.
				itr.currentKey, itr.currentValue, err = cur.Get(nil, nil, lmdb.Last)
			} else if err == nil {
				// SetRange lands on >= end, we want < end.
				if bytes.Compare(itr.currentKey, end) >= 0 {
					itr.currentKey, itr.currentValue, err = cur.Get(nil, nil, lmdb.Prev)
				}
			}
		}
	} else {
		if start == nil {
			itr.currentKey, itr.currentValue, err = cur.Get(nil, nil, lmdb.First)
		} else {
			itr.currentKey, itr.currentValue, err = cur.Get(start, nil, lmdb.SetRange)
		}
	}

	if lmdb.IsNotFound(err) {
		itr.isInvalid = true
		err = nil
	} else if err != nil {
		cur.Close()
		txn.Abort()
		return nil, err
	}

	return itr, nil
}

func (itr *lmdbIterator) Domain() ([]byte, []byte) {
	return itr.start, itr.end
}

func (itr *lmdbIterator) Valid() bool {
	if itr.isInvalid {
		return false
	}
	if itr.currentKey == nil {
		itr.isInvalid = true
		return false
	}
	if itr.isReverse {
		if itr.start != nil && bytes.Compare(itr.currentKey, itr.start) < 0 {
			itr.isInvalid = true
			return false
		}
	} else {
		if itr.end != nil && bytes.Compare(itr.currentKey, itr.end) >= 0 {
			itr.isInvalid = true
			return false
		}
	}
	return true
}

func (itr *lmdbIterator) Next() {
	if !itr.Valid() {
		panic("lmdbIterator is invalid")
	}
	var err error
	if itr.isReverse {
		itr.currentKey, itr.currentValue, err = itr.cur.Get(nil, nil, lmdb.Prev)
	} else {
		itr.currentKey, itr.currentValue, err = itr.cur.Get(nil, nil, lmdb.Next)
	}
	if lmdb.IsNotFound(err) {
		itr.isInvalid = true
	} else if err != nil {
		panic(err)
	}
}

func (itr *lmdbIterator) Key() []byte {
	if !itr.Valid() {
		panic("lmdbIterator is invalid")
	}
	return slices.Clone(itr.currentKey)
}

func (itr *lmdbIterator) Value() []byte {
	if !itr.Valid() {
		panic("lmdbIterator is invalid")
	}
	return slices.Clone(itr.currentValue)
}

func (itr *lmdbIterator) Error() error {
	return nil
}

func (itr *lmdbIterator) Close() error {
	if itr.cur != nil {
		itr.cur.Close()
		itr.cur = nil
	}
	if itr.txn != nil {
		itr.txn.Abort()
		itr.txn = nil
	}
	if itr.db != nil {
		itr.db.mu.RUnlock()
		itr.db = nil
	}
	itr.isInvalid = true
	itr.currentKey = nil
	itr.currentValue = nil
	return nil
}
