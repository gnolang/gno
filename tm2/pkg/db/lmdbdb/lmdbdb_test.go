//go:build cgo

package lmdbdb

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/internal"
)

func newTestDB(t *testing.T) *LMDB {
	t.Helper()
	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	db, err := NewLMDB(name, t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

// ----------------------------------------
// Construction and lifecycle

func TestLMDBNew(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	db, err := NewLMDB(name, t.TempDir())
	require.NoError(t, err)
	require.NoError(t, db.Close())
}

func TestLMDBNewWithOptions(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	db, err := NewLMDBWithOptions(name, t.TempDir(), 10<<20, 0)
	require.NoError(t, err)
	require.NoError(t, db.Set([]byte("k"), []byte("v")))
	val, err := db.Get([]byte("k"))
	require.NoError(t, err)
	require.Equal(t, []byte("v"), val)
	require.NoError(t, db.Close())
}

func TestLMDBOpenExisting(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	name := "reopen"

	db1, err := NewLMDB(name, dir)
	require.NoError(t, err)
	require.NoError(t, db1.Set([]byte("persist"), []byte("yes")))
	require.NoError(t, db1.Close())

	db2, err := NewLMDB(name, dir)
	require.NoError(t, err)
	defer db2.Close()
	val, err := db2.Get([]byte("persist"))
	require.NoError(t, err)
	require.Equal(t, []byte("yes"), val)
}

func TestLMDBOpenBadDir(t *testing.T) {
	t.Parallel()
	_, err := NewLMDB("test", "/dev/null/impossible")
	require.Error(t, err)
}

func TestLMDBRegistry(t *testing.T) {
	t.Parallel()
	d, err := db.NewDB("regtest", LMDBBackend, t.TempDir())
	require.NoError(t, err)
	require.NoError(t, d.Set([]byte("k"), []byte("v")))
	val, err := d.Get([]byte("k"))
	require.NoError(t, err)
	require.Equal(t, []byte("v"), val)
	require.NoError(t, d.Close())
}

func TestLMDBDoubleClose(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	db, err := NewLMDB(name, t.TempDir())
	require.NoError(t, err)
	require.NoError(t, db.Close())
	require.Error(t, db.Close())
}

func TestLMDBUseAfterClose(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	db, err := NewLMDB(name, t.TempDir())
	require.NoError(t, err)
	require.NoError(t, db.Close())

	_, err = db.Get([]byte("k"))
	require.ErrorIs(t, err, errClosed)

	err = db.Set([]byte("k"), []byte("v"))
	require.ErrorIs(t, err, errClosed)

	err = db.Delete([]byte("k"))
	require.ErrorIs(t, err, errClosed)

	_, err = db.Has([]byte("k"))
	require.ErrorIs(t, err, errClosed)

	_, err = db.Iterator(nil, nil)
	require.ErrorIs(t, err, errClosed)

	_, err = db.ReverseIterator(nil, nil)
	require.ErrorIs(t, err, errClosed)

	err = db.SetSync([]byte("k"), []byte("v"))
	require.ErrorIs(t, err, errClosed)

	err = db.DeleteSync([]byte("k"))
	require.ErrorIs(t, err, errClosed)

	require.Nil(t, db.Stats())
}

func TestLMDBBatchWriteOnClosedDB(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	db, err := NewLMDB(name, t.TempDir())
	require.NoError(t, err)

	batch := db.NewBatch()
	require.NoError(t, batch.Set([]byte("k"), []byte("v")))

	require.NoError(t, db.Close())

	err = batch.Write()
	require.Error(t, err)
}

// ----------------------------------------
// Get / Set / Delete basics

func TestLMDBGetSetDelete(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	val, err := db.Get([]byte("missing"))
	require.NoError(t, err)
	require.Nil(t, val)

	require.NoError(t, db.Set([]byte("key1"), []byte("value1")))
	val, err = db.Get([]byte("key1"))
	require.NoError(t, err)
	require.Equal(t, []byte("value1"), val)

	require.NoError(t, db.Set([]byte("key1"), []byte("value2")))
	val, err = db.Get([]byte("key1"))
	require.NoError(t, err)
	require.Equal(t, []byte("value2"), val)

	has, err := db.Has([]byte("key1"))
	require.NoError(t, err)
	require.True(t, has)
	has, err = db.Has([]byte("nope"))
	require.NoError(t, err)
	require.False(t, has)

	require.NoError(t, db.Delete([]byte("key1")))
	val, err = db.Get([]byte("key1"))
	require.NoError(t, err)
	require.Nil(t, val)

	require.NoError(t, db.Delete([]byte("key1")))
	require.NoError(t, db.Delete([]byte("never-existed")))
}

func TestLMDBSetSync(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	require.NoError(t, db.SetSync([]byte("k"), []byte("v")))
	val, err := db.Get([]byte("k"))
	require.NoError(t, err)
	require.Equal(t, []byte("v"), val)
}

func TestLMDBDeleteSync(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	require.NoError(t, db.Set([]byte("k"), []byte("v")))
	require.NoError(t, db.DeleteSync([]byte("k")))
	val, err := db.Get([]byte("k"))
	require.NoError(t, err)
	require.Nil(t, val)
}

// ----------------------------------------
// Nil and empty key/value edge cases

func TestLMDBNilAndEmptyKeys(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	require.NoError(t, db.Set(nil, []byte("nilkey")))
	val, err := db.Get(nil)
	require.NoError(t, err)
	require.Equal(t, []byte("nilkey"), val)

	require.NoError(t, db.Set([]byte{}, []byte("emptykey")))
	val, err = db.Get([]byte{})
	require.NoError(t, err)
	require.Equal(t, []byte("emptykey"), val)

	val2, err := db.Get(nil)
	require.NoError(t, err)
	require.Equal(t, val, val2)

	has, err := db.Has(nil)
	require.NoError(t, err)
	require.True(t, has)

	require.NoError(t, db.Delete(nil))
	val, err = db.Get([]byte{})
	require.NoError(t, err)
	require.Nil(t, val)
}

func TestLMDBNilAndEmptyValues(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	require.NoError(t, db.Set([]byte("k1"), nil))
	val, err := db.Get([]byte("k1"))
	require.NoError(t, err)
	require.Equal(t, []byte{}, val)

	require.NoError(t, db.Set([]byte("k2"), []byte{}))
	val, err = db.Get([]byte("k2"))
	require.NoError(t, err)
	require.Equal(t, []byte{}, val)

	has, err := db.Has([]byte("k1"))
	require.NoError(t, err)
	require.True(t, has)
}

// ----------------------------------------
// Copy safety

func TestLMDBGetReturnsCopy(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	require.NoError(t, db.Set([]byte("key"), []byte("original")))
	val1, err := db.Get([]byte("key"))
	require.NoError(t, err)
	val1[0] = 'X'

	val2, err := db.Get([]byte("key"))
	require.NoError(t, err)
	require.Equal(t, []byte("original"), val2)
}

func TestLMDBSetDoesNotRetainInput(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	key := []byte("key")
	val := []byte("value")
	require.NoError(t, db.Set(key, val))
	key[0] = 'X'
	val[0] = 'X'

	got, err := db.Get([]byte("key"))
	require.NoError(t, err)
	require.Equal(t, []byte("value"), got)
}

func TestLMDBMultipleGetsReturnIndependentCopies(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	require.NoError(t, db.Set([]byte("k"), []byte("hello")))
	v1, _ := db.Get([]byte("k"))
	v2, _ := db.Get([]byte("k"))
	v1[0] = 'X'
	require.Equal(t, []byte("hello"), v2)
}

// ----------------------------------------
// Batch operations

func TestLMDBBatch(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	batch := db.NewBatch()
	require.NoError(t, batch.Set([]byte("k1"), []byte("v1")))
	require.NoError(t, batch.Set([]byte("k2"), []byte("v2")))
	require.NoError(t, batch.Delete([]byte("nonexistent")))

	val, err := db.Get([]byte("k1"))
	require.NoError(t, err)
	require.Nil(t, val)

	require.NoError(t, batch.Write())

	val, err = db.Get([]byte("k1"))
	require.NoError(t, err)
	require.Equal(t, []byte("v1"), val)
	val, err = db.Get([]byte("k2"))
	require.NoError(t, err)
	require.Equal(t, []byte("v2"), val)

	require.NoError(t, batch.Close())
}

func TestLMDBBatchAtomic(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	require.NoError(t, db.Set([]byte("survive"), []byte("yes")))

	batch := db.NewBatch()
	require.NoError(t, batch.Set([]byte("new"), []byte("val")))
	require.NoError(t, batch.Delete([]byte("survive")))
	require.NoError(t, batch.Write())
	require.NoError(t, batch.Close())

	val, err := db.Get([]byte("new"))
	require.NoError(t, err)
	require.Equal(t, []byte("val"), val)
	val, err = db.Get([]byte("survive"))
	require.NoError(t, err)
	require.Nil(t, val)
}

func TestLMDBBatchOverwriteInBatch(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	batch := db.NewBatch()
	require.NoError(t, batch.Set([]byte("k"), []byte("first")))
	require.NoError(t, batch.Set([]byte("k"), []byte("second")))
	require.NoError(t, batch.Set([]byte("k"), []byte("third")))
	require.NoError(t, batch.Write())
	require.NoError(t, batch.Close())

	val, err := db.Get([]byte("k"))
	require.NoError(t, err)
	require.Equal(t, []byte("third"), val)
}

func TestLMDBBatchSetThenDeleteInBatch(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	batch := db.NewBatch()
	require.NoError(t, batch.Set([]byte("k"), []byte("v")))
	require.NoError(t, batch.Delete([]byte("k")))
	require.NoError(t, batch.Write())
	require.NoError(t, batch.Close())

	val, err := db.Get([]byte("k"))
	require.NoError(t, err)
	require.Nil(t, val)
}

func TestLMDBBatchDeleteThenSetInBatch(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	require.NoError(t, db.Set([]byte("k"), []byte("original")))

	batch := db.NewBatch()
	require.NoError(t, batch.Delete([]byte("k")))
	require.NoError(t, batch.Set([]byte("k"), []byte("revived")))
	require.NoError(t, batch.Write())
	require.NoError(t, batch.Close())

	val, err := db.Get([]byte("k"))
	require.NoError(t, err)
	require.Equal(t, []byte("revived"), val)
}

func TestLMDBBatchEmptyWrite(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	batch := db.NewBatch()
	require.NoError(t, batch.Write())
	require.NoError(t, batch.Close())
}

func TestLMDBBatchGetByteSize(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	batch := db.NewBatch()
	require.NoError(t, batch.Set([]byte("abc"), []byte("defgh")))
	size, err := batch.GetByteSize()
	require.NoError(t, err)
	require.Equal(t, 8, size)

	require.NoError(t, batch.Delete([]byte("xy")))
	size, err = batch.GetByteSize()
	require.NoError(t, err)
	require.Equal(t, 10, size)

	require.NoError(t, batch.Close())
	_, err = batch.GetByteSize()
	require.Error(t, err)
}

func TestLMDBBatchWriteSync(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	batch := db.NewBatch()
	require.NoError(t, batch.Set([]byte("k1"), []byte("v1")))
	require.NoError(t, batch.WriteSync())
	require.NoError(t, batch.Close())

	val, err := db.Get([]byte("k1"))
	require.NoError(t, err)
	require.Equal(t, []byte("v1"), val)
}

func TestLMDBBatchWithSize(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	batch := db.NewBatchWithSize(100)
	require.NoError(t, batch.Set([]byte("k"), []byte("v")))
	require.NoError(t, batch.Write())
	require.NoError(t, batch.Close())

	val, err := db.Get([]byte("k"))
	require.NoError(t, err)
	require.Equal(t, []byte("v"), val)
}

func TestLMDBBatchNilKey(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	batch := db.NewBatch()
	require.NoError(t, batch.Set(nil, []byte("nilval")))
	require.NoError(t, batch.Write())
	require.NoError(t, batch.Close())

	val, err := db.Get(nil)
	require.NoError(t, err)
	require.Equal(t, []byte("nilval"), val)
}

func TestLMDBBatchUseAfterWrite(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	batch := db.NewBatch()
	require.NoError(t, batch.Set([]byte("k"), []byte("v")))
	require.NoError(t, batch.Write())

	// After Write, further operations should error.
	require.Error(t, batch.Set([]byte("k2"), []byte("v2")))
	require.Error(t, batch.Delete([]byte("k")))
	require.Error(t, batch.Write())
	require.Error(t, batch.WriteSync())
	_, err := batch.GetByteSize()
	require.Error(t, err)

	// Close is still safe.
	require.NoError(t, batch.Close())
}

func TestLMDBBatchUseAfterClose(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	batch := db.NewBatch()
	require.NoError(t, batch.Set([]byte("k"), []byte("v")))
	require.NoError(t, batch.Close())

	require.Error(t, batch.Set([]byte("k2"), []byte("v2")))
	require.Error(t, batch.Delete([]byte("k")))
	require.Error(t, batch.Write())
	_, err := batch.GetByteSize()
	require.Error(t, err)
}

func TestLMDBBatchDoubleClose(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	batch := db.NewBatch()
	require.NoError(t, batch.Set([]byte("k"), []byte("v")))
	require.NoError(t, batch.Close())
	require.NoError(t, batch.Close()) // idempotent
}

func TestLMDBBatchCloseDiscardsWrites(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	batch := db.NewBatch()
	require.NoError(t, batch.Set([]byte("discard"), []byte("me")))
	require.NoError(t, batch.Close()) // without Write()

	// Data should NOT be in the DB.
	val, err := db.Get([]byte("discard"))
	require.NoError(t, err)
	require.Nil(t, val)
}

func TestLMDBCloseBlocksUntilIteratorClosed(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	require.NoError(t, db.Set([]byte("a"), []byte("1")))

	itr, err := db.Iterator(nil, nil)
	require.NoError(t, err)

	// Close() should block while iterator is alive.
	// Verify by closing iterator from another goroutine after a short delay.
	done := make(chan error, 1)
	go func() {
		done <- db.Close()
	}()

	// Close hasn't returned yet (iterator holds RLock).
	select {
	case <-done:
		t.Fatal("Close() should block while iterator is open")
	default:
	}

	// Now close the iterator, which releases the RLock.
	require.NoError(t, itr.Close())

	// Close() should now complete.
	err = <-done
	require.NoError(t, err)
}

// ----------------------------------------
// Forward iterator

func TestLMDBIteratorForward(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	require.NoError(t, db.Set([]byte("a"), []byte("1")))
	require.NoError(t, db.Set([]byte("b"), []byte("2")))
	require.NoError(t, db.Set([]byte("c"), []byte("3")))
	require.NoError(t, db.Set([]byte("d"), []byte("4")))
	require.NoError(t, db.Set([]byte("e"), []byte("5")))

	itr, err := db.Iterator(nil, nil)
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b", "c", "d", "e"}, collectKeys(itr))
	require.NoError(t, itr.Close())

	itr, err = db.Iterator([]byte("b"), []byte("d"))
	require.NoError(t, err)
	require.Equal(t, []string{"b", "c"}, collectKeys(itr))
	require.NoError(t, itr.Close())

	itr, err = db.Iterator([]byte("c"), nil)
	require.NoError(t, err)
	require.Equal(t, []string{"c", "d", "e"}, collectKeys(itr))
	require.NoError(t, itr.Close())

	itr, err = db.Iterator(nil, []byte("c"))
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b"}, collectKeys(itr))
	require.NoError(t, itr.Close())

	itr, err = db.Iterator([]byte("x"), []byte("z"))
	require.NoError(t, err)
	require.Empty(t, collectKeys(itr))
	require.NoError(t, itr.Close())

	itr, err = db.Iterator([]byte("b"), []byte("c"))
	require.NoError(t, err)
	require.Equal(t, []string{"b"}, collectKeys(itr))
	require.NoError(t, itr.Close())

	itr, err = db.Iterator([]byte("bb"), []byte("d"))
	require.NoError(t, err)
	require.Equal(t, []string{"c"}, collectKeys(itr))
	require.NoError(t, itr.Close())

	itr, err = db.Iterator([]byte("a"), []byte("cc"))
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b", "c"}, collectKeys(itr))
	require.NoError(t, itr.Close())
}

// ----------------------------------------
// Reverse iterator

func TestLMDBIteratorReverse(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	require.NoError(t, db.Set([]byte("a"), []byte("1")))
	require.NoError(t, db.Set([]byte("b"), []byte("2")))
	require.NoError(t, db.Set([]byte("c"), []byte("3")))
	require.NoError(t, db.Set([]byte("d"), []byte("4")))
	require.NoError(t, db.Set([]byte("e"), []byte("5")))

	itr, err := db.ReverseIterator(nil, nil)
	require.NoError(t, err)
	require.Equal(t, []string{"e", "d", "c", "b", "a"}, collectKeys(itr))
	require.NoError(t, itr.Close())

	itr, err = db.ReverseIterator([]byte("b"), []byte("d"))
	require.NoError(t, err)
	require.Equal(t, []string{"c", "b"}, collectKeys(itr))
	require.NoError(t, itr.Close())

	itr, err = db.ReverseIterator([]byte("b"), nil)
	require.NoError(t, err)
	require.Equal(t, []string{"e", "d", "c", "b"}, collectKeys(itr))
	require.NoError(t, itr.Close())

	itr, err = db.ReverseIterator(nil, []byte("d"))
	require.NoError(t, err)
	require.Equal(t, []string{"c", "b", "a"}, collectKeys(itr))
	require.NoError(t, itr.Close())

	itr, err = db.ReverseIterator([]byte("bb"), []byte("e"))
	require.NoError(t, err)
	require.Equal(t, []string{"d", "c"}, collectKeys(itr))
	require.NoError(t, itr.Close())

	itr, err = db.ReverseIterator([]byte("a"), []byte("cc"))
	require.NoError(t, err)
	require.Equal(t, []string{"c", "b", "a"}, collectKeys(itr))
	require.NoError(t, itr.Close())

	itr, err = db.ReverseIterator([]byte("x"), []byte("z"))
	require.NoError(t, err)
	require.Empty(t, collectKeys(itr))
	require.NoError(t, itr.Close())
}

// ----------------------------------------
// Iterator edge cases

func TestLMDBIteratorEmptyDB(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	itr, err := db.Iterator(nil, nil)
	require.NoError(t, err)
	require.False(t, itr.Valid())
	require.NoError(t, itr.Close())

	itr, err = db.ReverseIterator(nil, nil)
	require.NoError(t, err)
	require.False(t, itr.Valid())
	require.NoError(t, itr.Close())
}

func TestLMDBIteratorDomain(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	require.NoError(t, db.Set([]byte("a"), []byte("1")))

	itr, err := db.Iterator([]byte("start"), []byte("end"))
	require.NoError(t, err)
	s, e := itr.Domain()
	require.Equal(t, []byte("start"), s)
	require.Equal(t, []byte("end"), e)
	require.NoError(t, itr.Close())
}

func TestLMDBIteratorReturnsCopies(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	require.NoError(t, db.Set([]byte("key"), []byte("value")))

	itr, err := db.Iterator(nil, nil)
	require.NoError(t, err)
	require.True(t, itr.Valid())

	k := itr.Key()
	v := itr.Value()
	k[0] = 'X'
	v[0] = 'X'

	require.Equal(t, []byte("key"), itr.Key())
	require.Equal(t, []byte("value"), itr.Value())
	require.NoError(t, itr.Close())
}

func TestLMDBIteratorValues(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	require.NoError(t, db.Set([]byte("a"), []byte("1")))
	require.NoError(t, db.Set([]byte("b"), []byte("2")))
	require.NoError(t, db.Set([]byte("c"), []byte("3")))

	itr, err := db.Iterator(nil, nil)
	require.NoError(t, err)
	var vals []string
	for ; itr.Valid(); itr.Next() {
		vals = append(vals, string(itr.Value()))
	}
	require.NoError(t, itr.Close())
	require.Equal(t, []string{"1", "2", "3"}, vals)
}

func TestLMDBIteratorCloseReleasesResources(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	require.NoError(t, db.Set([]byte("a"), []byte("1")))

	for i := 0; i < 200; i++ {
		itr, err := db.Iterator(nil, nil)
		require.NoError(t, err)
		require.NoError(t, itr.Close())
	}

	val, err := db.Get([]byte("a"))
	require.NoError(t, err)
	require.Equal(t, []byte("1"), val)
}

func TestLMDBIteratorSingleKey(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	require.NoError(t, db.Set([]byte("only"), []byte("one")))

	itr, err := db.Iterator(nil, nil)
	require.NoError(t, err)
	require.True(t, itr.Valid())
	require.Equal(t, []byte("only"), itr.Key())
	require.Equal(t, []byte("one"), itr.Value())
	itr.Next()
	require.False(t, itr.Valid())
	require.NoError(t, itr.Close())

	itr, err = db.ReverseIterator(nil, nil)
	require.NoError(t, err)
	require.True(t, itr.Valid())
	require.Equal(t, []byte("only"), itr.Key())
	itr.Next()
	require.False(t, itr.Valid())
	require.NoError(t, itr.Close())
}

func TestLMDBIteratorError(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	require.NoError(t, db.Set([]byte("a"), []byte("1")))

	itr, err := db.Iterator(nil, nil)
	require.NoError(t, err)
	require.Nil(t, itr.Error())
	itr.Next()
	require.Nil(t, itr.Error())
	require.NoError(t, itr.Close())
}

func TestLMDBIteratorDoubleClose(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	require.NoError(t, db.Set([]byte("a"), []byte("1")))
	itr, err := db.Iterator(nil, nil)
	require.NoError(t, err)
	require.NoError(t, itr.Close())
	require.NoError(t, itr.Close()) // should not panic
}

func TestLMDBIteratorNextPanicsOnInvalid(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	require.NoError(t, db.Set([]byte("a"), []byte("1")))
	itr, err := db.Iterator(nil, nil)
	require.NoError(t, err)
	itr.Next() // exhaust
	require.False(t, itr.Valid())

	require.Panics(t, func() { itr.Next() })
	require.NoError(t, itr.Close())
}

func TestLMDBIteratorKeyPanicsOnInvalid(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	itr, err := db.Iterator(nil, nil) // empty DB
	require.NoError(t, err)
	require.False(t, itr.Valid())

	require.Panics(t, func() { itr.Key() })
	require.Panics(t, func() { itr.Value() })
	require.NoError(t, itr.Close())
}

func TestLMDBIteratorUseAfterClose(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	require.NoError(t, db.Set([]byte("a"), []byte("1")))
	itr, err := db.Iterator(nil, nil)
	require.NoError(t, err)
	require.True(t, itr.Valid())
	require.NoError(t, itr.Close())

	// After close, Valid() returns false, Key()/Value()/Next() panic.
	require.False(t, itr.Valid())
	require.Panics(t, func() { itr.Key() })
	require.Panics(t, func() { itr.Value() })
	require.Panics(t, func() { itr.Next() })
}

// ----------------------------------------
// Sentinel key and iterator ordering

func TestLMDBSentinelKeyNoCollision(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	require.NoError(t, db.Set([]byte{0x00}, []byte("sentinel")))

	val, err := db.Get(nil)
	require.NoError(t, err)
	require.Equal(t, []byte("sentinel"), val)
	val, err = db.Get([]byte{})
	require.NoError(t, err)
	require.Equal(t, []byte("sentinel"), val)
	val, err = db.Get([]byte{0x00})
	require.NoError(t, err)
	require.Equal(t, []byte("sentinel"), val)
}

func TestLMDBSentinelKeyIteratorOrder(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	// Empty key (→ sentinel 0x00) should sort before 0x01.
	require.NoError(t, db.Set([]byte{}, []byte("empty")))
	require.NoError(t, db.Set([]byte{0x01}, []byte("one")))
	require.NoError(t, db.Set([]byte{0x02}, []byte("two")))

	itr, err := db.Iterator(nil, nil)
	require.NoError(t, err)
	var keys [][]byte
	for ; itr.Valid(); itr.Next() {
		keys = append(keys, itr.Key())
	}
	require.NoError(t, itr.Close())

	require.Len(t, keys, 3)
	require.Equal(t, []byte{0x00}, keys[0]) // sentinel
	require.Equal(t, []byte{0x01}, keys[1])
	require.Equal(t, []byte{0x02}, keys[2])
}

func TestLMDBSentinelKeyDistinctFromMultiByte(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	// {0x00} (sentinel for empty key) must be distinct from {0x00, 0x01}.
	require.NoError(t, db.Set([]byte{}, []byte("empty")))
	require.NoError(t, db.Set([]byte{0x00, 0x01}, []byte("multi")))

	val, err := db.Get(nil)
	require.NoError(t, err)
	require.Equal(t, []byte("empty"), val)

	val, err = db.Get([]byte{0x00, 0x01})
	require.NoError(t, err)
	require.Equal(t, []byte("multi"), val)
}

// ----------------------------------------
// Concurrent access

func TestLMDBConcurrentReads(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	for i := 0; i < 100; i++ {
		k := fmt.Sprintf("key%03d", i)
		v := fmt.Sprintf("val%03d", i)
		require.NoError(t, db.Set([]byte(k), []byte(v)))
	}

	var wg sync.WaitGroup
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				k := fmt.Sprintf("key%03d", i)
				v := fmt.Sprintf("val%03d", i)
				got, err := db.Get([]byte(k))
				if err != nil {
					t.Errorf("goroutine %d: Get(%s) error: %v", g, k, err)
					return
				}
				if string(got) != v {
					t.Errorf("goroutine %d: Get(%s) = %s, want %s", g, k, got, v)
					return
				}
			}
		}(g)
	}
	wg.Wait()
}

func TestLMDBConcurrentIterators(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	for i := 0; i < 50; i++ {
		require.NoError(t, db.Set([]byte(fmt.Sprintf("k%03d", i)), []byte(fmt.Sprintf("v%03d", i))))
	}

	var wg sync.WaitGroup
	for g := 0; g < 5; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			itr, err := db.Iterator(nil, nil)
			if err != nil {
				t.Errorf("Iterator error: %v", err)
				return
			}
			count := 0
			for ; itr.Valid(); itr.Next() {
				_ = itr.Key()
				_ = itr.Value()
				count++
			}
			itr.Close()
			if count != 50 {
				t.Errorf("expected 50 keys, got %d", count)
			}
		}()
	}
	wg.Wait()
}

func TestLMDBConcurrentWrites(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	var wg sync.WaitGroup
	for g := 0; g < 5; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				k := fmt.Sprintf("g%d_k%03d", g, i)
				v := fmt.Sprintf("g%d_v%03d", g, i)
				if err := db.Set([]byte(k), []byte(v)); err != nil {
					t.Errorf("goroutine %d: Set error: %v", g, err)
					return
				}
			}
		}(g)
	}
	wg.Wait()

	// Verify all writes landed.
	for g := 0; g < 5; g++ {
		for i := 0; i < 100; i++ {
			k := fmt.Sprintf("g%d_k%03d", g, i)
			v := fmt.Sprintf("g%d_v%03d", g, i)
			got, err := db.Get([]byte(k))
			require.NoError(t, err)
			require.Equal(t, []byte(v), got, "key %s", k)
		}
	}
}

// ----------------------------------------
// Large and many keys

func TestLMDBLargeValues(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	bigVal := make([]byte, 1<<20)
	for i := range bigVal {
		bigVal[i] = byte(i)
	}
	require.NoError(t, db.Set([]byte("big"), bigVal))
	got, err := db.Get([]byte("big"))
	require.NoError(t, err)
	require.Equal(t, bigVal, got)
}

func TestLMDBLargeKey(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	bigKey := make([]byte, 500)
	for i := range bigKey {
		bigKey[i] = byte(i)
	}
	require.NoError(t, db.Set(bigKey, []byte("val")))
	got, err := db.Get(bigKey)
	require.NoError(t, err)
	require.Equal(t, []byte("val"), got)
}

func TestLMDBManyKeys(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	const n = 10000
	for i := 0; i < n; i++ {
		k := fmt.Sprintf("key%06d", i)
		v := fmt.Sprintf("val%06d", i)
		require.NoError(t, db.Set([]byte(k), []byte(v)))
	}

	val, err := db.Get([]byte("key005000"))
	require.NoError(t, err)
	require.Equal(t, []byte("val005000"), val)

	itr, err := db.Iterator(nil, nil)
	require.NoError(t, err)
	count := 0
	for ; itr.Valid(); itr.Next() {
		count++
	}
	require.NoError(t, itr.Close())
	require.Equal(t, n, count)
}

func TestLMDBManyKeysBatch(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	const n = 5000
	batch := db.NewBatch()
	for i := 0; i < n; i++ {
		k := fmt.Sprintf("bk%06d", i)
		v := fmt.Sprintf("bv%06d", i)
		require.NoError(t, batch.Set([]byte(k), []byte(v)))
	}
	require.NoError(t, batch.Write())
	require.NoError(t, batch.Close())

	for i := 0; i < n; i++ {
		k := fmt.Sprintf("bk%06d", i)
		v := fmt.Sprintf("bv%06d", i)
		got, err := db.Get([]byte(k))
		require.NoError(t, err)
		require.Equal(t, []byte(v), got)
	}
}

// ----------------------------------------
// MapFull error

func TestLMDBMapFull(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	// Use page-size-aware map size: small enough to fill quickly.
	pageSize := os.Getpagesize()
	mapSize := int64(pageSize) * 8 // 8 pages
	db, err := NewLMDBWithOptions(name, t.TempDir(), mapSize, 0)
	require.NoError(t, err)
	defer db.Close()

	// Write enough data to exceed the map.
	bigVal := make([]byte, pageSize)
	var writeErr error
	for i := 0; i < 1000; i++ {
		k := fmt.Sprintf("key%04d", i)
		writeErr = db.Set([]byte(k), bigVal)
		if writeErr != nil {
			break
		}
	}
	require.Error(t, writeErr, "expected MDB_MAP_FULL error")
}

// ----------------------------------------
// Stats and Print

func TestLMDBStats(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	require.NoError(t, db.Set([]byte("k"), []byte("v")))

	stats := db.Stats()
	require.NotNil(t, stats)
	require.Contains(t, stats, "Entries")
	require.Contains(t, stats, "Depth")
	require.Contains(t, stats, "PageSize")
	require.Contains(t, stats, "BranchPages")
	require.Contains(t, stats, "LeafPages")
	require.Contains(t, stats, "OverflowPages")
}

func TestLMDBPrint(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	require.NoError(t, db.Print())
}

// ----------------------------------------
// Shared benchmarks from internal

func BenchmarkLMDBRandomReadsWrites(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping testing in short mode")
	}
	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	db, err := NewLMDB(name, b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()
	internal.BenchmarkRandomReadsWrites(b, db)
}

func BenchmarkLMDBIterator(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping testing in short mode")
	}
	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	db, err := NewLMDB(name, b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()
	internal.BenchmarkIterator(b, db)
}

func BenchmarkLMDBBatchWrites(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping testing in short mode")
	}
	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	db, err := NewLMDB(name, b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()
	internal.BenchmarkBatchWrites(b, db)
}

// ----------------------------------------
// Helpers

func collectKeys(itr interface {
	Valid() bool
	Next()
	Key() []byte
}) []string {
	var keys []string
	for ; itr.Valid(); itr.Next() {
		keys = append(keys, string(itr.Key()))
	}
	return keys
}
