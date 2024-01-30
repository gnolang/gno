package memdb

import (
	"fmt"
	"sort"
	"sync"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/internal"
	"github.com/gnolang/gno/tm2/pkg/strings"
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
func (db *MemDB) Get(key []byte) []byte {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	key = internal.NonNilBytes(key)

	value := db.db[string(key)]
	return value
}

// Implements DB.
func (db *MemDB) Has(key []byte) bool {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	key = internal.NonNilBytes(key)

	_, ok := db.db[string(key)]
	return ok
}

// Implements DB.
func (db *MemDB) Set(key []byte, value []byte) {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	db.SetNoLock(key, value)
}

// Implements DB.
func (db *MemDB) SetSync(key []byte, value []byte) {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	db.SetNoLock(key, value)
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
func (db *MemDB) Delete(key []byte) {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	db.DeleteNoLock(key)
}

// Implements DB.
func (db *MemDB) DeleteSync(key []byte) {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	db.DeleteNoLock(key)
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
func (db *MemDB) Close() {
	// Close is a noop since for an in-memory
	// database, we don't have a destination
	// to flush contents to nor do we want
	// any data loss on invoking Close()
	// See the discussion in https://github.com/tendermint/classic/libs/pull/56
}

// Implements DB.
func (db *MemDB) Print() {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	for key, value := range db.db {
		var keystr, valstr string
		if strings.IsASCIIText(key) {
			keystr = key
		} else {
			keystr = fmt.Sprintf("0x%X", []byte(key))
		}
		if strings.IsASCIIText(string(value)) {
			valstr = string(value)
		} else {
			valstr = fmt.Sprintf("0x%X", value)
		}
		fmt.Printf("%s:\t%s\n", keystr, valstr)
	}
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

	return &internal.MemBatch{db, nil}
}

// ----------------------------------------
// Iterator

// Implements DB.
func (db *MemDB) Iterator(start, end []byte) dbm.Iterator {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	keys := db.getSortedKeys(start, end, false)
	return internal.NewMemIterator(db, keys, start, end)
}

// Implements DB.
func (db *MemDB) ReverseIterator(start, end []byte) dbm.Iterator {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	keys := db.getSortedKeys(start, end, true)
	return internal.NewMemIterator(db, keys, start, end)
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
		for i := 0; i < nkeys/2; i++ {
			temp := keys[i]
			keys[i] = keys[nkeys-i-1]
			keys[nkeys-i-1] = temp
		}
	}
	return keys
}
