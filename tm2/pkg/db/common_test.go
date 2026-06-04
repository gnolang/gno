package db_test

import (
	"encoding/binary"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/internal"
)

// ----------------------------------------
// Test Helper functions.

func checkValue(t *testing.T, db db.DB, key []byte, valueWanted []byte) {
	t.Helper()

	valueGot, err := db.Get(key)
	require.NoError(t, err)
	assert.Equal(t, valueWanted, valueGot)
}

func checkValid(t *testing.T, itr db.Iterator, expected bool) {
	t.Helper()

	valid := itr.Valid()
	require.Equal(t, expected, valid)
}

func checkNext(t *testing.T, itr db.Iterator, expected bool) {
	t.Helper()

	itr.Next()
	valid := itr.Valid()
	require.Equal(t, expected, valid)
}

func checkNextPanics(t *testing.T, itr db.Iterator) {
	t.Helper()

	assert.Panics(t, func() { itr.Next() }, "checkNextPanics expected panic but didn't")
}

func checkDomain(t *testing.T, itr db.Iterator, start, end []byte) {
	t.Helper()

	ds, de := itr.Domain()
	assert.Equal(t, start, ds, "checkDomain domain start incorrect")
	assert.Equal(t, end, de, "checkDomain domain end incorrect")
}

func checkItem(t *testing.T, itr db.Iterator, key []byte, value []byte) {
	t.Helper()

	k, v := itr.Key(), itr.Value()
	assert.Exactly(t, key, k)
	assert.Exactly(t, value, v)
}

func checkInvalid(t *testing.T, itr db.Iterator) {
	t.Helper()

	checkValid(t, itr, false)
	checkKeyPanics(t, itr)
	checkValuePanics(t, itr)
	checkNextPanics(t, itr)
}

func checkKeyPanics(t *testing.T, itr db.Iterator) {
	t.Helper()

	assert.Panics(t, func() { itr.Key() }, "checkKeyPanics expected panic but didn't")
}

func checkValuePanics(t *testing.T, itr db.Iterator) {
	t.Helper()

	assert.Panics(t, func() { itr.Value() }, "checkValuePanics expected panic but didn't")
}

// ----------------------------------------
// mockDB

// NOTE: not actually goroutine safe.
// If you want something goroutine safe, maybe you just want a MemDB.
type mockDB struct {
	mtx   sync.Mutex
	calls map[string]int
}

func newMockDB() *mockDB {
	return &mockDB{
		calls: make(map[string]int),
	}
}

func (mdb *mockDB) Mutex() *sync.Mutex {
	return &(mdb.mtx)
}

func (mdb *mockDB) Get([]byte) []byte {
	mdb.calls["Get"]++
	return nil
}

func (mdb *mockDB) Has([]byte) bool {
	mdb.calls["Has"]++
	return false
}

func (mdb *mockDB) Set([]byte, []byte) {
	mdb.calls["Set"]++
}

func (mdb *mockDB) SetSync([]byte, []byte) {
	mdb.calls["SetSync"]++
}

func (mdb *mockDB) SetNoLock([]byte, []byte) {
	mdb.calls["SetNoLock"]++
}

func (mdb *mockDB) SetNoLockSync([]byte, []byte) {
	mdb.calls["SetNoLockSync"]++
}

func (mdb *mockDB) Delete([]byte) {
	mdb.calls["Delete"]++
}

func (mdb *mockDB) DeleteSync([]byte) {
	mdb.calls["DeleteSync"]++
}

func (mdb *mockDB) DeleteNoLock([]byte) {
	mdb.calls["DeleteNoLock"]++
}

func (mdb *mockDB) DeleteNoLockSync([]byte) {
	mdb.calls["DeleteNoLockSync"]++
}

func (mdb *mockDB) Iterator(start, end []byte) db.Iterator {
	mdb.calls["Iterator"]++
	return &internal.MockIterator{}
}

func (mdb *mockDB) ReverseIterator(start, end []byte) db.Iterator {
	mdb.calls["ReverseIterator"]++
	return &internal.MockIterator{}
}

func (mdb *mockDB) Close() error {
	mdb.calls["Close"]++

	return nil
}

func (mdb *mockDB) NewBatch() db.Batch {
	mdb.calls["NewBatch"]++
	return &internal.MemBatch{
		DB:   mdb,
		Ops:  []internal.Operation{},
		Size: 0,
	}
}

func (mdb *mockDB) Print() {
	mdb.calls["Print"]++
	fmt.Printf("mockDB{%v}", mdb.Stats())
}

func (mdb *mockDB) Stats() map[string]string {
	mdb.calls["Stats"]++

	res := make(map[string]string)
	for key, count := range mdb.calls {
		res[key] = fmt.Sprintf("%d", count)
	}
	return res
}

func int642Bytes(i int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func bytes2Int64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

// For testing convenience.
func bz(s string) []byte {
	return []byte(s)
}
