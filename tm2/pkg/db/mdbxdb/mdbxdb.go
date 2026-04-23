//go:build cgo

package mdbxdb

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/erigontech/mdbx-go/mdbx"

	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/internal"
)

const (
	MDBXBackend db.BackendType = "mdbxdb"
)

var errClosed = errors.New("mdbxdb: database is closed")
var errBatchDone = errors.New("mdbxdb: batch already written or closed")

func init() {
	db.InternalRegisterDBCreator(MDBXBackend, func(name, dir string) (db.DB, error) {
		return NewMDBX(name, dir)
	}, false)
}

var _ db.DB = (*MDBX)(nil)

// MDBX wraps an mdbx.Env + DBI handle.
type MDBX struct {
	mu     sync.RWMutex
	env    *mdbx.Env
	dbi    mdbx.DBI
	closed bool
}

// DefaultMapSize is the default maximum database size (1 TB).
const DefaultMapSize int = 1 << 40

// NewMDBX opens an MDBX database with production defaults.
func NewMDBX(name, dir string) (*MDBX, error) {
	return NewMDBXWithOptions(name, dir, DefaultMapSize, 0)
}

// NewMDBXWithOptions opens an MDBX database with custom map size and flags.
// Extra flags are OR'd with the base flags (NoMetaSync | WriteMap | NoReadahead).
func NewMDBXWithOptions(name, dir string, mapSize int, extraFlags uint) (*MDBX, error) {
	dbDir := filepath.Join(dir, name+".db")
	if err := os.MkdirAll(dbDir, 0o700); err != nil {
		return nil, fmt.Errorf("error creating dir: %w", err)
	}

	env, err := mdbx.NewEnv("")
	if err != nil {
		return nil, err
	}
	// MDBX uses SetGeometry instead of SetMapSize.
	// -1 means "use default" for parameters we don't care about.
	if err := env.SetGeometry(-1, -1, mapSize, -1, -1, -1); err != nil {
		env.Close()
		return nil, err
	}
	if err := env.SetOption(mdbx.OptMaxDB, 1); err != nil {
		env.Close()
		return nil, err
	}

	flags := uint(mdbx.NoMetaSync | mdbx.WriteMap | mdbx.NoReadahead)
	flags |= extraFlags
	if err := env.Open(dbDir, flags, 0o644); err != nil {
		env.Close()
		return nil, err
	}

	// Open the default (unnamed) database.
	var dbi mdbx.DBI
	err = env.Update(func(txn *mdbx.Txn) error {
		var err error
		dbi, err = txn.OpenRoot(mdbx.Create)
		return err
	})
	if err != nil {
		env.Close()
		return nil, err
	}

	return &MDBX{env: env, dbi: dbi}, nil
}

// nonEmptyKey maps empty keys to a sentinel since MDBX does not support them.
func nonEmptyKey(key []byte) []byte {
	if len(key) == 0 {
		return []byte{0x00}
	}
	return key
}

func (m *MDBX) checkClosed() error {
	if m.closed {
		return errClosed
	}
	return nil
}

func (m *MDBX) Get(key []byte) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if err := m.checkClosed(); err != nil {
		return nil, err
	}
	key = nonEmptyKey(internal.NonNilBytes(key))
	var val []byte
	err := m.env.View(func(txn *mdbx.Txn) error {
		v, err := txn.Get(m.dbi, key)
		if mdbx.IsNotFound(err) {
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

func (m *MDBX) Has(key []byte) (bool, error) {
	v, err := m.Get(key)
	return v != nil, err
}

func (m *MDBX) Set(key, value []byte) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if err := m.checkClosed(); err != nil {
		return err
	}
	key = nonEmptyKey(internal.NonNilBytes(key))
	value = internal.NonNilBytes(value)
	return m.env.Update(func(txn *mdbx.Txn) error {
		return txn.Put(m.dbi, key, value, 0)
	})
}

func (m *MDBX) SetSync(key, value []byte) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if err := m.checkClosed(); err != nil {
		return err
	}
	key = nonEmptyKey(internal.NonNilBytes(key))
	value = internal.NonNilBytes(value)
	err := m.env.Update(func(txn *mdbx.Txn) error {
		return txn.Put(m.dbi, key, value, 0)
	})
	if err != nil {
		return err
	}
	return m.env.Sync(true, false)
}

func (m *MDBX) Delete(key []byte) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if err := m.checkClosed(); err != nil {
		return err
	}
	key = nonEmptyKey(internal.NonNilBytes(key))
	return m.env.Update(func(txn *mdbx.Txn) error {
		err := txn.Del(m.dbi, key, nil)
		if mdbx.IsNotFound(err) {
			return nil
		}
		return err
	})
}

func (m *MDBX) DeleteSync(key []byte) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if err := m.checkClosed(); err != nil {
		return err
	}
	key = nonEmptyKey(internal.NonNilBytes(key))
	err := m.env.Update(func(txn *mdbx.Txn) error {
		err := txn.Del(m.dbi, key, nil)
		if mdbx.IsNotFound(err) {
			return nil
		}
		return err
	})
	if err != nil {
		return err
	}
	return m.env.Sync(true, false)
}

func (m *MDBX) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return errClosed
	}
	m.closed = true
	m.env.Close() // MDBX Close() returns nothing
	return nil
}

func (m *MDBX) Print() error {
	return nil
}

func (m *MDBX) Stats() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.closed {
		return nil
	}
	stat, err := m.env.Stat()
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

func (m *MDBX) NewBatch() db.Batch {
	return &mdbxBatch{db: m}
}

func (m *MDBX) NewBatchWithSize(_ int) db.Batch {
	return &mdbxBatch{db: m}
}

// Iterator implements db.DB.
// The caller MUST call Close() on the returned iterator to release resources.
func (m *MDBX) Iterator(start, end []byte) (db.Iterator, error) {
	m.mu.RLock()
	if err := m.checkClosed(); err != nil {
		m.mu.RUnlock()
		return nil, err
	}
	itr, err := newMDBXIterator(m, start, end, false)
	if err != nil {
		m.mu.RUnlock()
		return nil, err
	}
	return itr, nil // RLock held until itr.Close()
}

// ReverseIterator implements db.DB.
// The caller MUST call Close() on the returned iterator to release resources.
func (m *MDBX) ReverseIterator(start, end []byte) (db.Iterator, error) {
	m.mu.RLock()
	if err := m.checkClosed(); err != nil {
		m.mu.RUnlock()
		return nil, err
	}
	itr, err := newMDBXIterator(m, start, end, true)
	if err != nil {
		m.mu.RUnlock()
		return nil, err
	}
	return itr, nil // RLock held until itr.Close()
}

// ----------------------------------------
// Batch

type mdbxBatch struct {
	db   *MDBX
	ops  []internal.Operation
	size int
	done bool
}

func (b *mdbxBatch) Set(key, value []byte) error {
	if b.done {
		return errBatchDone
	}
	b.size += len(key) + len(value)
	b.ops = append(b.ops, internal.Operation{OpType: internal.OpTypeSet, Key: key, Value: value})
	return nil
}

func (b *mdbxBatch) Delete(key []byte) error {
	if b.done {
		return errBatchDone
	}
	b.size += len(key)
	b.ops = append(b.ops, internal.Operation{OpType: internal.OpTypeDelete, Key: key})
	return nil
}

func (b *mdbxBatch) write(sync bool) error {
	if b.done {
		return errBatchDone
	}
	b.db.mu.RLock()
	defer b.db.mu.RUnlock()
	if err := b.db.checkClosed(); err != nil {
		return err
	}
	err := b.db.env.Update(func(txn *mdbx.Txn) error {
		for _, op := range b.ops {
			key := nonEmptyKey(internal.NonNilBytes(op.Key))
			switch op.OpType {
			case internal.OpTypeSet:
				if err := txn.Put(b.db.dbi, key, internal.NonNilBytes(op.Value), 0); err != nil {
					return err
				}
			case internal.OpTypeDelete:
				err := txn.Del(b.db.dbi, key, nil)
				if err != nil && !mdbx.IsNotFound(err) {
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
		return b.db.env.Sync(true, false)
	}
	return nil
}

func (b *mdbxBatch) Write() error {
	return b.write(false)
}

func (b *mdbxBatch) WriteSync() error {
	return b.write(true)
}

func (b *mdbxBatch) Close() error {
	b.done = true
	b.ops = nil
	b.size = 0
	return nil
}

func (b *mdbxBatch) GetByteSize() (int, error) {
	if b.done {
		return 0, errBatchDone
	}
	return b.size, nil
}

// ----------------------------------------
// Iterator

type mdbxIterator struct {
	db  *MDBX // for releasing RLock on Close
	txn *mdbx.Txn
	cur *mdbx.Cursor

	start, end   []byte
	currentKey   []byte
	currentValue []byte
	isReverse    bool
	isInvalid    bool
}

func newMDBXIterator(m *MDBX, start, end []byte, isReverse bool) (*mdbxIterator, error) {
	txn, err := m.env.BeginTxn(nil, mdbx.Readonly)
	if err != nil {
		return nil, err
	}
	cur, err := txn.OpenCursor(m.dbi)
	if err != nil {
		txn.Abort()
		return nil, err
	}

	itr := &mdbxIterator{
		db:        m,
		txn:       txn,
		cur:       cur,
		start:     start,
		end:       end,
		isReverse: isReverse,
	}

	// Position cursor.
	if isReverse {
		if end == nil {
			itr.currentKey, itr.currentValue, err = cur.Get(nil, nil, mdbx.Last)
		} else {
			itr.currentKey, itr.currentValue, err = cur.Get(end, nil, mdbx.SetRange)
			if mdbx.IsNotFound(err) {
				// end is past all keys, go to last.
				itr.currentKey, itr.currentValue, err = cur.Get(nil, nil, mdbx.Last)
			} else if err == nil {
				// SetRange lands on >= end, we want < end.
				if bytes.Compare(itr.currentKey, end) >= 0 {
					itr.currentKey, itr.currentValue, err = cur.Get(nil, nil, mdbx.Prev)
				}
			}
		}
	} else {
		if start == nil {
			itr.currentKey, itr.currentValue, err = cur.Get(nil, nil, mdbx.First)
		} else {
			itr.currentKey, itr.currentValue, err = cur.Get(start, nil, mdbx.SetRange)
		}
	}

	if mdbx.IsNotFound(err) {
		itr.isInvalid = true
		err = nil
	} else if err != nil {
		cur.Close()
		txn.Abort()
		return nil, err
	}

	return itr, nil
}

func (itr *mdbxIterator) Domain() ([]byte, []byte) {
	return itr.start, itr.end
}

func (itr *mdbxIterator) Valid() bool {
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

func (itr *mdbxIterator) Next() {
	if !itr.Valid() {
		panic("mdbxIterator is invalid")
	}
	var err error
	if itr.isReverse {
		itr.currentKey, itr.currentValue, err = itr.cur.Get(nil, nil, mdbx.Prev)
	} else {
		itr.currentKey, itr.currentValue, err = itr.cur.Get(nil, nil, mdbx.Next)
	}
	if mdbx.IsNotFound(err) {
		itr.isInvalid = true
	} else if err != nil {
		panic(err)
	}
}

func (itr *mdbxIterator) Key() []byte {
	if !itr.Valid() {
		panic("mdbxIterator is invalid")
	}
	return slices.Clone(itr.currentKey)
}

func (itr *mdbxIterator) Value() []byte {
	if !itr.Valid() {
		panic("mdbxIterator is invalid")
	}
	return slices.Clone(itr.currentValue)
}

func (itr *mdbxIterator) Error() error {
	return nil
}

func (itr *mdbxIterator) Close() error {
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
