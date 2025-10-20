package boltdb

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"go.etcd.io/bbolt"

	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/internal"
)

var bucket = []byte("tm")

func init() {
	db.InternalRegisterDBCreator(db.BoltDBBackend, func(name, dir string) (db.DB, error) {
		return New(name, dir)
	}, false)
}

// BoltDB is a wrapper around etcd's fork of bolt
// (https://go.etcd.io/bbolt).
//
// NOTE: All operations (including Set, Delete) are synchronous by default. One
// can globally turn it off by using NoSync config option (not recommended).
//
// A single bucket ([]byte("tm")) is used per a database instance. This could
// lead to performance issues when/if there will be lots of keys.
type BoltDB struct {
	db *bbolt.DB
}

// New returns a BoltDB with default options.
func New(name, dir string) (db.DB, error) {
	return NewWithOptions(name, dir, bbolt.DefaultOptions)
}

// NewWithOptions allows you to supply *bbolt.Options. ReadOnly: true is not
// supported because NewWithOptions creates a global bucket.
func NewWithOptions(name string, dir string, opts *bbolt.Options) (db.DB, error) {
	if opts.ReadOnly {
		return nil, errors.New("ReadOnly: true is not supported")
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("error creating dir: %w", err)
	}

	dbPath := filepath.Join(dir, name+".db")
	db, err := bbolt.Open(dbPath, os.ModePerm, opts)
	if err != nil {
		return nil, err
	}

	// create a global bucket
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucket)
		return err
	})
	if err != nil {
		return nil, err
	}

	return &BoltDB{db: db}, nil
}

func (bdb *BoltDB) Get(key []byte) (value []byte, err error) {
	key = nonEmptyKey(internal.NonNilBytes(key))
	err = bdb.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if v := b.Get(key); v != nil {
			value = slices.Clone(v)
		}
		return nil
	})
	return
}

func (bdb *BoltDB) Has(key []byte) (bool, error) {
	v, err := bdb.Get(key)
	return v != nil, err
}

func (bdb *BoltDB) Set(key, value []byte) error {
	key = nonEmptyKey(internal.NonNilBytes(key))
	value = internal.NonNilBytes(value)
	return bdb.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		return b.Put(key, value)
	})
}

func (bdb *BoltDB) SetSync(key, value []byte) error {
	return bdb.Set(key, value)
}

func (bdb *BoltDB) Delete(key []byte) error {
	key = nonEmptyKey(internal.NonNilBytes(key))
	return bdb.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucket).Delete(key)
	})
}

func (bdb *BoltDB) DeleteSync(key []byte) error {
	return bdb.Delete(key)
}

func (bdb *BoltDB) Close() error {
	return bdb.db.Close()
}

func (bdb *BoltDB) Print() error {
	stats := bdb.db.Stats()
	fmt.Printf("%v\n", stats)

	return bdb.db.View(func(tx *bbolt.Tx) error {
		tx.Bucket(bucket).ForEach(func(k, v []byte) error {
			fmt.Printf("[%X]:\t[%X]\n", k, v)
			return nil
		})
		return nil
	})
}

func (bdb *BoltDB) Stats() map[string]string {
	stats := bdb.db.Stats()
	m := make(map[string]string)

	// Freelist stats
	m["FreePageN"] = fmt.Sprintf("%v", stats.FreePageN)
	m["PendingPageN"] = fmt.Sprintf("%v", stats.PendingPageN)
	m["FreeAlloc"] = fmt.Sprintf("%v", stats.FreeAlloc)
	m["FreelistInuse"] = fmt.Sprintf("%v", stats.FreelistInuse)

	// Transaction stats
	m["TxN"] = fmt.Sprintf("%v", stats.TxN)
	m["OpenTxN"] = fmt.Sprintf("%v", stats.OpenTxN)

	return m
}

// boltDBBatch stores key values in sync.Map and dumps them to the underlying
// DB upon Write call.
type boltDBBatch struct {
	db   *BoltDB
	ops  []internal.Operation
	size int
}

// NewBatch returns a new batch.
func (bdb *BoltDB) NewBatch() db.Batch {
	bdb.db.MaxBatchSize = bbolt.DefaultMaxBatchSize
	return &boltDBBatch{
		ops:  nil,
		db:   bdb,
		size: 0,
	}
}

// NewBatchWithSize returns a new batch.
func (bdb *BoltDB) NewBatchWithSize(size int) db.Batch {
	bdb.db.MaxBatchSize = size
	return &boltDBBatch{
		ops:  nil,
		db:   bdb,
		size: 0,
	}
}

// It is safe to modify the contents of the argument after Set returns but not
// before.
func (bdb *boltDBBatch) Set(key, value []byte) error {
	bdb.size += len(key) + len(value)
	bdb.ops = append(bdb.ops, internal.Operation{OpType: internal.OpTypeSet, Key: key, Value: value})
	return nil
}

// It is safe to modify the contents of the argument after Delete returns but
// not before.
func (bdb *boltDBBatch) Delete(key []byte) error {
	bdb.size += len(key)
	bdb.ops = append(bdb.ops, internal.Operation{OpType: internal.OpTypeDelete, Key: key})
	return nil
}

// NOTE: the operation is synchronous (see BoltDB for reasons)
func (bdb *boltDBBatch) Write() error {
	return bdb.db.db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		for _, op := range bdb.ops {
			key := nonEmptyKey(internal.NonNilBytes(op.Key))
			switch op.OpType {
			case internal.OpTypeSet:
				if putErr := b.Put(key, op.Value); putErr != nil {
					return putErr
				}
			case internal.OpTypeDelete:
				if delErr := b.Delete(key); delErr != nil {
					return delErr
				}
			}
		}
		return nil
	})
}

func (bdb *boltDBBatch) WriteSync() error {
	return bdb.Write()
}

func (bdb *boltDBBatch) Close() error {
	bdb.size = 0
	return nil
}

func (bdb *boltDBBatch) GetByteSize() (int, error) {
	if bdb.ops == nil {
		return 0, errors.New("boltdb: batch has been written or closed")
	}
	return bdb.size, nil
}

// WARNING: Any concurrent writes or reads will block until the iterator is
// closed.
func (bdb *BoltDB) Iterator(start, end []byte) (db.Iterator, error) {
	tx, err := bdb.db.Begin(false)
	if err != nil {
		return nil, err
	}
	return newBoltDBIterator(tx, start, end, false), nil
}

// WARNING: Any concurrent writes or reads will block until the iterator is
// closed.
func (bdb *BoltDB) ReverseIterator(start, end []byte) (db.Iterator, error) {
	tx, err := bdb.db.Begin(false)
	if err != nil {
		return nil, err
	}
	return newBoltDBIterator(tx, start, end, true), nil
}

// boltDBIterator allows you to iterate on range of keys/values given some
// start / end keys (nil & nil will result in doing full scan).
type boltDBIterator struct {
	tx *bbolt.Tx

	itr   *bbolt.Cursor
	start []byte
	end   []byte

	currentKey   []byte
	currentValue []byte

	isInvalid bool
	isReverse bool
}

func newBoltDBIterator(tx *bbolt.Tx, start, end []byte, isReverse bool) *boltDBIterator {
	itr := tx.Bucket(bucket).Cursor()

	var ck, cv []byte
	if isReverse {
		if end == nil {
			ck, cv = itr.Last()
		} else {
			_, _ = itr.Seek(end) // after key
			ck, cv = itr.Prev()  // return to end key
		}
	} else {
		if start == nil {
			ck, cv = itr.First()
		} else {
			ck, cv = itr.Seek(start)
		}
	}

	return &boltDBIterator{
		tx:           tx,
		itr:          itr,
		start:        start,
		end:          end,
		currentKey:   ck,
		currentValue: cv,
		isReverse:    isReverse,
		isInvalid:    false,
	}
}

func (itr *boltDBIterator) Domain() ([]byte, []byte) {
	return itr.start, itr.end
}

func (itr *boltDBIterator) Valid() bool {
	if itr.isInvalid {
		return false
	}

	// iterated to the end of the cursor
	if len(itr.currentKey) == 0 {
		itr.isInvalid = true
		return false
	}

	if itr.isReverse {
		if itr.start != nil && bytes.Compare(itr.currentKey, itr.start) < 0 {
			itr.isInvalid = true
			return false
		}
	} else {
		if itr.end != nil && bytes.Compare(itr.end, itr.currentKey) <= 0 {
			itr.isInvalid = true
			return false
		}
	}

	// Valid
	return true
}

func (itr *boltDBIterator) Next() {
	itr.assertIsValid()
	if itr.isReverse {
		itr.currentKey, itr.currentValue = itr.itr.Prev()
	} else {
		itr.currentKey, itr.currentValue = itr.itr.Next()
	}
}

func (itr *boltDBIterator) Key() []byte {
	itr.assertIsValid()
	return slices.Clone(itr.currentKey)
}

func (itr *boltDBIterator) Value() []byte {
	itr.assertIsValid()
	var value []byte
	if itr.currentValue != nil {
		value = slices.Clone(itr.currentValue)
	}
	return value
}

func (*boltDBIterator) Error() error {
	return nil // famous last words
}

func (itr *boltDBIterator) Close() error {
	return itr.tx.Rollback()
}

func (itr *boltDBIterator) assertIsValid() {
	if !itr.Valid() {
		panic("Boltdb-iterator is invalid")
	}
}

// nonEmptyKey returns a []byte("nil") if key is empty.
// WARNING: this may collude with "nil" user key!
func nonEmptyKey(key []byte) []byte {
	if len(key) == 0 {
		return []byte("nil")
	}
	return key
}
