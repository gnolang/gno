package db_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/db"
	_ "github.com/gnolang/gno/tm2/pkg/db/_all"
	"github.com/gnolang/gno/tm2/pkg/db/internal"
)

func testBackendGetSetDelete(t *testing.T, backend db.BackendType) {
	t.Helper()

	// Default
	db, err := db.NewDB("testdb", backend, t.TempDir())
	require.NoError(t, err)

	must := func(bz []byte, err error) []byte {
		require.NoError(t, err)
		return bz
	}

	// A nonexistent key should return nil, even if the key is empty
	require.Nil(t, must(db.Get([]byte(""))))

	// A nonexistent key should return nil, even if the key is nil
	require.Nil(t, must(db.Get(nil)))

	// A nonexistent key should return nil.
	key := []byte("abc")
	require.Nil(t, must(db.Get(key)))

	// Set empty value.
	db.SetSync(key, []byte(""))
	require.NotNil(t, must(db.Get(key)))
	require.Empty(t, must(db.Get(key)))

	// Set nil value.
	db.SetSync(key, nil)
	require.NotNil(t, must(db.Get(key)))
	require.Empty(t, must(db.Get(key)))

	// Delete.
	db.DeleteSync(key)
	require.Nil(t, must(db.Get(key)))
}

func TestBackendsGetSetDelete(t *testing.T) {
	t.Parallel()

	for _, dbType := range db.BackendList() {
		t.Run(string(dbType), func(t *testing.T) {
			t.Parallel()

			testBackendGetSetDelete(t, dbType)
		})
	}
}

func withDB(t *testing.T, dbType db.BackendType, fn func(db.DB)) {
	t.Helper()

	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	db, err := db.NewDB(name, dbType, t.TempDir())
	require.Nil(t, err)
	fn(db)

	require.NoError(t, db.Close())
}

func TestBackendsNilKeys(t *testing.T) {
	t.Parallel()

	// Test all backends.
	for _, dbType := range db.BackendList() {
		withDB(t, dbType, func(db db.DB) {
			t.Run(fmt.Sprintf("Testing %s", dbType), func(t *testing.T) {
				// Nil keys are treated as the empty key for most operations.
				expect := func(key, value []byte) {
					if len(key) == 0 { // nil or empty
						exp, err := db.Get(nil)
						require.NoError(t, err)
						got, err := db.Get([]byte(""))
						require.NoError(t, err)
						assert.Equal(t, exp, got)
						exp2, err := db.Has(nil)
						require.NoError(t, err)
						got2, err := db.Has([]byte(""))
						require.NoError(t, err)
						assert.Equal(t, exp2, got2)
					}
					v, err := db.Get(key)
					require.NoError(t, err)
					assert.Equal(t, v, value)
					h, err := db.Has(key)
					require.NoError(t, err)
					assert.Equal(t, h, value != nil)
				}

				// Not set
				expect(nil, nil)

				// Set nil value
				db.SetSync(nil, nil)
				expect(nil, []byte(""))

				// Set empty value
				db.SetSync(nil, []byte(""))
				expect(nil, []byte(""))

				// Set nil, Delete nil
				db.SetSync(nil, []byte("abc"))
				expect(nil, []byte("abc"))
				db.DeleteSync(nil)
				expect(nil, nil)

				// Set nil, Delete empty
				db.SetSync(nil, []byte("abc"))
				expect(nil, []byte("abc"))
				db.DeleteSync([]byte(""))
				expect(nil, nil)

				// Set empty, Delete nil
				db.SetSync([]byte(""), []byte("abc"))
				expect(nil, []byte("abc"))
				db.DeleteSync(nil)
				expect(nil, nil)

				// Set empty, Delete empty
				db.SetSync([]byte(""), []byte("abc"))
				expect(nil, []byte("abc"))
				db.DeleteSync([]byte(""))
				expect(nil, nil)

				// Set nil, Delete nil
				db.SetSync(nil, []byte("abc"))
				expect(nil, []byte("abc"))
				db.DeleteSync(nil)
				expect(nil, nil)

				// Set nil, Delete empty
				db.SetSync(nil, []byte("abc"))
				expect(nil, []byte("abc"))
				db.DeleteSync([]byte(""))
				expect(nil, nil)

				// Set empty, Delete nil
				db.SetSync([]byte(""), []byte("abc"))
				expect(nil, []byte("abc"))
				db.DeleteSync(nil)
				expect(nil, nil)

				// Set empty, Delete empty
				db.SetSync([]byte(""), []byte("abc"))
				expect(nil, []byte("abc"))
				db.DeleteSync([]byte(""))
				expect(nil, nil)
			})
		})
	}
}

func TestDBIterator(t *testing.T) {
	t.Parallel()

	for _, dbType := range db.BackendList() {
		t.Run(fmt.Sprintf("%v", dbType), func(t *testing.T) {
			t.Parallel()

			testDBIterator(t, dbType)
		})
	}
}

func testDBIterator(t *testing.T, backend db.BackendType) {
	t.Helper()

	mustIterator := func(it db.Iterator, err error) db.Iterator {
		require.NoError(t, err)
		return it
	}

	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	db, err := db.NewDB(name, backend, t.TempDir())
	require.NoError(t, err)

	for i := range 10 {
		if i != 6 { // but skip 6.
			db.Set(int642Bytes(int64(i)), nil)
		}
	}

	verifyIterator(t, mustIterator(db.Iterator(nil, nil)), []int64{0, 1, 2, 3, 4, 5, 7, 8, 9}, "forward iterator")
	verifyIterator(t, mustIterator(db.ReverseIterator(nil, nil)), []int64{9, 8, 7, 5, 4, 3, 2, 1, 0}, "reverse iterator")

	verifyIterator(t, mustIterator(db.Iterator(nil, int642Bytes(0))), []int64(nil), "forward iterator to 0")
	verifyIterator(t, mustIterator(db.ReverseIterator(int642Bytes(10), nil)), []int64(nil), "reverse iterator from 10 (ex)")

	verifyIterator(t, mustIterator(db.Iterator(int642Bytes(0), nil)), []int64{0, 1, 2, 3, 4, 5, 7, 8, 9}, "forward iterator from 0")
	verifyIterator(t, mustIterator(db.Iterator(int642Bytes(1), nil)), []int64{1, 2, 3, 4, 5, 7, 8, 9}, "forward iterator from 1")
	verifyIterator(t, mustIterator(db.ReverseIterator(nil, int642Bytes(10))),
		[]int64{9, 8, 7, 5, 4, 3, 2, 1, 0}, "reverse iterator from 10 (ex)")
	verifyIterator(t, mustIterator(db.ReverseIterator(nil, int642Bytes(9))),
		[]int64{8, 7, 5, 4, 3, 2, 1, 0}, "reverse iterator from 9 (ex)")
	verifyIterator(t, mustIterator(db.ReverseIterator(nil, int642Bytes(8))),
		[]int64{7, 5, 4, 3, 2, 1, 0}, "reverse iterator from 8 (ex)")

	verifyIterator(t, mustIterator(db.Iterator(int642Bytes(5), int642Bytes(6))), []int64{5}, "forward iterator from 5 to 6")
	verifyIterator(t, mustIterator(db.Iterator(int642Bytes(5), int642Bytes(7))), []int64{5}, "forward iterator from 5 to 7")
	verifyIterator(t, mustIterator(db.Iterator(int642Bytes(5), int642Bytes(8))), []int64{5, 7}, "forward iterator from 5 to 8")
	verifyIterator(t, mustIterator(db.Iterator(int642Bytes(6), int642Bytes(7))), []int64(nil), "forward iterator from 6 to 7")
	verifyIterator(t, mustIterator(db.Iterator(int642Bytes(6), int642Bytes(8))), []int64{7}, "forward iterator from 6 to 8")
	verifyIterator(t, mustIterator(db.Iterator(int642Bytes(7), int642Bytes(8))), []int64{7}, "forward iterator from 7 to 8")

	verifyIterator(t, mustIterator(db.ReverseIterator(int642Bytes(4), int642Bytes(5))), []int64{4}, "reverse iterator from 5 (ex) to 4")
	verifyIterator(t, mustIterator(db.ReverseIterator(int642Bytes(4), int642Bytes(6))),
		[]int64{5, 4}, "reverse iterator from 6 (ex) to 4")
	verifyIterator(t, mustIterator(db.ReverseIterator(int642Bytes(4), int642Bytes(7))),
		[]int64{5, 4}, "reverse iterator from 7 (ex) to 4")
	verifyIterator(t, mustIterator(db.ReverseIterator(int642Bytes(5), int642Bytes(6))), []int64{5}, "reverse iterator from 6 (ex) to 5")
	verifyIterator(t, mustIterator(db.ReverseIterator(int642Bytes(5), int642Bytes(7))), []int64{5}, "reverse iterator from 7 (ex) to 5")
	verifyIterator(t, mustIterator(db.ReverseIterator(int642Bytes(6), int642Bytes(7))),
		[]int64(nil), "reverse iterator from 7 (ex) to 6")

	verifyIterator(t, mustIterator(db.Iterator(int642Bytes(0), int642Bytes(1))), []int64{0}, "forward iterator from 0 to 1")
	verifyIterator(t, mustIterator(db.ReverseIterator(int642Bytes(8), int642Bytes(9))), []int64{8}, "reverse iterator from 9 (ex) to 8")

	verifyIterator(t, mustIterator(db.Iterator(int642Bytes(2), int642Bytes(4))), []int64{2, 3}, "forward iterator from 2 to 4")
	verifyIterator(t, mustIterator(db.Iterator(int642Bytes(4), int642Bytes(2))), []int64(nil), "forward iterator from 4 to 2")
	verifyIterator(t, mustIterator(db.ReverseIterator(int642Bytes(2), int642Bytes(4))),
		[]int64{3, 2}, "reverse iterator from 4 (ex) to 2")
	verifyIterator(t, mustIterator(db.ReverseIterator(int642Bytes(4), int642Bytes(2))),
		[]int64(nil), "reverse iterator from 2 (ex) to 4")
}

func verifyIterator(t *testing.T, itr db.Iterator, expected []int64, msg string) {
	t.Helper()

	var list []int64
	for itr.Valid() {
		list = append(list, bytes2Int64(itr.Key()))
		itr.Next()
	}
	assert.Equal(t, expected, list, msg)
}

func TestDBIteratorSingleKey(t *testing.T) {
	t.Parallel()

	for _, backend := range db.BackendList() {
		t.Run(fmt.Sprintf("Backend %s", backend), func(t *testing.T) {
			t.Parallel()

			db := newTempDB(t, backend)

			db.SetSync(bz("1"), bz("value_1"))
			itr, err := db.Iterator(nil, nil)
			require.NoError(t, err)

			checkValid(t, itr, true)
			checkNext(t, itr, false)
			checkValid(t, itr, false)
			checkNextPanics(t, itr)

			// Once invalid...
			checkInvalid(t, itr)
		})
	}
}

func TestDBIteratorTwoKeys(t *testing.T) {
	t.Parallel()

	for _, backend := range db.BackendList() {
		t.Run(fmt.Sprintf("Backend %s", backend), func(t *testing.T) {
			t.Parallel()

			db := newTempDB(t, backend)

			db.SetSync(bz("1"), bz("value_1"))
			db.SetSync(bz("2"), bz("value_1"))

			{ // Fail by calling Next too much
				itr, err := db.Iterator(nil, nil)
				require.NoError(t, err)
				checkValid(t, itr, true)

				checkNext(t, itr, true)
				checkValid(t, itr, true)

				checkNext(t, itr, false)
				checkValid(t, itr, false)

				checkNextPanics(t, itr)

				// Once invalid...
				checkInvalid(t, itr)
			}
		})
	}
}

func TestDBIteratorMany(t *testing.T) {
	t.Parallel()

	for _, backend := range db.BackendList() {
		t.Run(fmt.Sprintf("Backend %s", backend), func(t *testing.T) {
			t.Parallel()

			db := newTempDB(t, backend)

			keys := make([][]byte, 100)
			for i := range 100 {
				keys[i] = []byte{byte(i)}
			}

			value := []byte{5}
			for _, k := range keys {
				db.Set(k, value)
			}

			itr, err := db.Iterator(nil, nil)
			require.NoError(t, err)
			defer itr.Close()
			for ; itr.Valid(); itr.Next() {
				v, err := db.Get(itr.Key())
				require.NoError(t, err)
				assert.Equal(t, v, itr.Value())
			}
		})
	}
}

func TestDBIteratorEmpty(t *testing.T) {
	t.Parallel()

	for _, backend := range db.BackendList() {
		t.Run(fmt.Sprintf("Backend %s", backend), func(t *testing.T) {
			t.Parallel()

			db := newTempDB(t, backend)

			itr, err := db.Iterator(nil, nil)
			require.NoError(t, err)

			checkInvalid(t, itr)
		})
	}
}

func TestDBIteratorEmptyBeginAfter(t *testing.T) {
	t.Parallel()

	for _, backend := range db.BackendList() {
		t.Run(fmt.Sprintf("Backend %s", backend), func(t *testing.T) {
			t.Parallel()

			db := newTempDB(t, backend)

			itr, err := db.Iterator(bz("1"), nil)
			require.NoError(t, err)

			checkInvalid(t, itr)
		})
	}
}

func TestDBIteratorNonemptyBeginAfter(t *testing.T) {
	t.Parallel()

	for _, backend := range db.BackendList() {
		t.Run(fmt.Sprintf("Backend %s", backend), func(t *testing.T) {
			t.Parallel()

			db := newTempDB(t, backend)

			db.SetSync(bz("1"), bz("value_1"))
			itr, err := db.Iterator(bz("2"), nil)
			require.NoError(t, err)

			checkInvalid(t, itr)
		})
	}
}

func TestDBBatchWrite(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		modify func(batch db.Batch)
		calls  map[string]int
	}{
		0: {
			func(batch db.Batch) {
				batch.Set(bz("1"), bz("1"))
				batch.Set(bz("2"), bz("2"))
				batch.Delete(bz("3"))
				batch.Set(bz("4"), bz("4"))
				batch.Write()
			},
			map[string]int{
				"Set": 0, "SetSync": 0, "SetNoLock": 3, "SetNoLockSync": 0,
				"Delete": 0, "DeleteSync": 0, "DeleteNoLock": 1, "DeleteNoLockSync": 0,
			},
		},
		1: {
			func(batch db.Batch) {
				batch.Set(bz("1"), bz("1"))
				batch.Set(bz("2"), bz("2"))
				batch.Set(bz("4"), bz("4"))
				batch.Delete(bz("3"))
				batch.Write()
			},
			map[string]int{
				"Set": 0, "SetSync": 0, "SetNoLock": 3, "SetNoLockSync": 0,
				"Delete": 0, "DeleteSync": 0, "DeleteNoLock": 1, "DeleteNoLockSync": 0,
			},
		},
		2: {
			func(batch db.Batch) {
				batch.Set(bz("1"), bz("1"))
				batch.Set(bz("2"), bz("2"))
				batch.Delete(bz("3"))
				batch.Set(bz("4"), bz("4"))
				batch.WriteSync()
			},
			map[string]int{
				"Set": 0, "SetSync": 0, "SetNoLock": 2, "SetNoLockSync": 1,
				"Delete": 0, "DeleteSync": 0, "DeleteNoLock": 1, "DeleteNoLockSync": 0,
			},
		},
		3: {
			func(batch db.Batch) {
				batch.Set(bz("1"), bz("1"))
				batch.Set(bz("2"), bz("2"))
				batch.Set(bz("4"), bz("4"))
				batch.Delete(bz("3"))
				batch.WriteSync()
			},
			map[string]int{
				"Set": 0, "SetSync": 0, "SetNoLock": 3, "SetNoLockSync": 0,
				"Delete": 0, "DeleteSync": 0, "DeleteNoLock": 0, "DeleteNoLockSync": 1,
			},
		},
	}

	for i, tc := range testCases {
		mdb := newMockDB()
		batch := mdb.NewBatch()

		tc.modify(batch)

		for call, exp := range tc.calls {
			got := mdb.calls[call]
			assert.Equal(t, exp, got, "#%v - key: %s", i, call)
		}
	}
}

func newTempDB(t *testing.T, backend db.BackendType) db.DB {
	t.Helper()

	tmpdb, err := db.NewDB("testdb", backend, t.TempDir())
	require.NoError(t, err)

	return tmpdb
}
