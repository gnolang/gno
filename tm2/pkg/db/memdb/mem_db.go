package memdb

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gnolang/gno/tm2/pkg/colors"
	"github.com/gnolang/gno/tm2/pkg/db"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/internal"
	"github.com/tidwall/btree"
)

func init() {
	dbm.InternalRegisterDBCreator(dbm.MemDBBackend, func(name, dir string) (dbm.DB, error) {
		return NewMemDB(), nil
	}, false)
}

var _ dbm.DB = (*MemDB)(nil)

type MemDB struct {
	mtx sync.Mutex
	db  *btree.Map[string, []byte]
}

func NewMemDB() *MemDB {
	database := &MemDB{
		db: btree.NewMap[string, []byte](0),
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

	value, _ := db.db.Get(string(key))
	return value
}

// Implements DB.
func (db *MemDB) Has(key []byte) bool {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	key = internal.NonNilBytes(key)

	_, ok := db.db.Get(string(key))
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
	value = internal.NonNilBytes(value)

	db.db.Set(string(key), internal.NonNilBytes(value))
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

	db.db.Delete(string(key))
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

	db.db.Scan(func(key string, value []byte) bool {
		keystr := colors.DefaultColoredBytesN([]byte(key), 50)
		valstr := colors.DefaultColoredBytesN(value, 100)
		fmt.Printf("%s: %s\n", keystr, valstr)
		return true
	})
}

// Implements DB.
func (db *MemDB) Stats() map[string]string {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	stats := make(map[string]string)
	stats["database.type"] = "memDB"
	stats["database.size"] = fmt.Sprintf("%d", db.db.Len())
	return stats
}

// Implements DB.
func (db *MemDB) NewBatch() dbm.Batch {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	return &internal.MemBatch{DB: db}
}

// ----------------------------------------
// Iterator

// Implements DB.
func (db *MemDB) Iterator(start, end []byte) dbm.Iterator {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	res := &iterator{
		it:    db.db.Iter(),
		start: string(start),
		end:   string(end),
	}
	if start == nil {
		if !res.it.First() {
			res.invalid = true
		}
	} else {
		if !res.it.Seek(res.start) {
			res.invalid = true
		}
	}
	return res
}

// Implements DB.
func (db *MemDB) ReverseIterator(start, end []byte) dbm.Iterator {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	res := &iterator{
		it:      db.db.Iter(),
		start:   string(start),
		end:     string(end),
		reverse: true,
	}
	if end == nil {
		if !res.it.Last() {
			res.invalid = true
		}
	} else {
		valid := res.it.Seek(res.end)
		if valid {
			eoakey := res.it.Key() // end or after key
			if strings.Compare(res.end, eoakey) <= 0 {
				if !res.it.Prev() {
					res.invalid = true
				}
			}
		} else {
			if !res.it.Last() {
				res.invalid = true
			}
		}
	}
	return res
}

type iterator struct {
	it         btree.MapIter[string, []byte]
	start, end string
	invalid    bool
	reverse    bool
}

var _ db.Iterator = (*iterator)(nil)

func (i *iterator) Domain() (start []byte, end []byte) {
	return []byte(i.start), []byte(i.end)
}

// Valid returns whether the current position is valid.
// Once invalid, an Iterator is forever invalid.
func (i *iterator) Valid() bool {
	// Once invalid, forever invalid.
	if i.invalid {
		return false
	}

	// If key is end or past it, invalid.
	key := i.it.Key()

	if i.reverse {
		if i.start != "" && key < i.start {
			i.invalid = true
			return false
		}
	} else {
		if i.end != "" && key >= i.end {
			i.invalid = true
			return false
		}
	}

	// Valid
	return true
}

func (i *iterator) assertIsValid() {
	if !i.Valid() {
		panic("memdb iterator is invalid")
	}
}

// Next moves the iterator to the next sequential key in the database, as
// defined by order of iteration.
//
// If Valid returns false, this method will panic.
func (i *iterator) Next() {
	i.assertIsValid()

	if i.reverse {
		if !i.it.Prev() {
			i.invalid = true
		}
	} else {
		if !i.it.Next() {
			i.invalid = true
		}
	}
}

// Key returns the key of the cursor.
// If Valid returns false, this method will panic.
// CONTRACT: key readonly []byte
func (i *iterator) Key() []byte {
	i.assertIsValid()
	return []byte(i.it.Key())
}

// Value returns the value of the cursor.
// If Valid returns false, this method will panic.
// CONTRACT: value readonly []byte
func (i *iterator) Value() []byte {
	i.assertIsValid()
	return i.it.Value()
}

// Close releases the Iterator.
func (i *iterator) Close() {}
