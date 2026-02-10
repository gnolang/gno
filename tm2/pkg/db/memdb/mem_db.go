package memdb

import (
	"fmt"
	"sort"
	"sync"

	"github.com/gnolang/gno/tm2/pkg/colors"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/internal"
)

func init() {
	dbm.InternalRegisterDBCreator(dbm.MemDBBackend, func(name, dir string) (dbm.DB, error) {
		return NewMemDB(), nil
	}, false)
}

var _ dbm.DB = (*MemDB)(nil)

type MemDB struct {
	mtx sync.Mutex
	db  map[string][]byte
}

func NewMemDB() *MemDB {
	database := &MemDB{
		db: make(map[string][]byte),
	}
	return database
}

// Implements internal.AtomicSetDeleter.
func (db *MemDB) Mutex() *sync.Mutex {
	return &(db.mtx)
}

// Implements DB.
func (db *MemDB) Get(key []byte) ([]byte, error) {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	key = internal.NonNilBytes(key)

	value := db.db[string(key)]
	return value, nil
}

// Implements DB.
func (db *MemDB) Has(key []byte) (bool, error) {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	key = internal.NonNilBytes(key)

	_, ok := db.db[string(key)]
	return ok, nil
}

// Implements DB.
func (db *MemDB) Set(key []byte, value []byte) error {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	db.SetNoLock(key, value)
	return nil
}

// Implements DB.
func (db *MemDB) SetSync(key []byte, value []byte) error {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	db.SetNoLock(key, value)
	return nil
}

// Implements internal.AtomicSetDeleter.
func (db *MemDB) SetNoLock(key []byte, value []byte) {
	db.SetNoLockSync(key, value)
}

// Implements internal.AtomicSetDeleter.
func (db *MemDB) SetNoLockSync(key []byte, value []byte) {
	key = internal.NonNilBytes(key)
	value = internal.NonNilBytes(value)

	db.db[string(key)] = value
}

// Implements DB.
func (db *MemDB) Delete(key []byte) error {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	db.DeleteNoLock(key)
	return nil
}

// Implements DB.
func (db *MemDB) DeleteSync(key []byte) error {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	db.DeleteNoLock(key)
	return nil
}

// Implements internal.AtomicSetDeleter.
func (db *MemDB) DeleteNoLock(key []byte) {
	db.DeleteNoLockSync(key)
}

// Implements internal.AtomicSetDeleter.
func (db *MemDB) DeleteNoLockSync(key []byte) {
	key = internal.NonNilBytes(key)

	delete(db.db, string(key))
}

// Implements DB.
func (db *MemDB) Close() error {
	// Close is a noop since for an in-memory
	// database, we don't have a destination
	// to flush contents to nor do we want
	// any data loss on invoking Close()
	// See the discussion in https://github.com/tendermint/tmlibs/pull/56
	return nil
}

// Implements DB.
func (db *MemDB) Print() error {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	for key, value := range db.db {
		var keystr, valstr string
		keystr = colors.DefaultColoredBytesN([]byte(key), 50)
		valstr = colors.DefaultColoredBytesN(value, 100)
		fmt.Printf("%s: %s\n", keystr, valstr)
	}
	return nil
}

// Implements DB.
func (db *MemDB) Stats() map[string]string {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	stats := make(map[string]string)
	stats["database.type"] = "memDB"
	stats["database.size"] = fmt.Sprintf("%d", len(db.db))
	return stats
}

// Implements DB.
func (db *MemDB) NewBatch() dbm.Batch {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	return &internal.MemBatch{
		DB:   db,
		Ops:  []internal.Operation{},
		Size: 0,
	}
}

// Implements DB.
// It does the same thing as NewBatch because we can't pre-allocate MemDB.
func (db *MemDB) NewBatchWithSize(_ int) dbm.Batch {
	return db.NewBatch()
}

// ----------------------------------------
// Iterator

// Implements DB.
func (db *MemDB) Iterator(start, end []byte) (dbm.Iterator, error) {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	keys := db.getSortedKeys(start, end, false)
	return internal.NewMemIterator(db, keys, start, end), nil
}

// Implements DB.
func (db *MemDB) ReverseIterator(start, end []byte) (dbm.Iterator, error) {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	keys := db.getSortedKeys(start, end, true)
	return internal.NewMemIterator(db, keys, start, end), nil
}

// ----------------------------------------
// Misc.

func (db *MemDB) getSortedKeys(start, end []byte, reverse bool) []string {
	keys := []string{}
	for key := range db.db {
		inDomain := dbm.IsKeyInDomain([]byte(key), start, end)
		if inDomain {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	if reverse {
		nkeys := len(keys)
		for i := range nkeys / 2 {
			temp := keys[i]
			keys[i] = keys[nkeys-i-1]
			keys[nkeys-i-1] = temp
		}
	}
	return keys
}
