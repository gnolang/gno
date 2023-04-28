package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testBackendGetSetDelete(t *testing.T, backend BackendType) {
	t.Helper()

	// Default
	db, err := NewDB("testdb", backend, t.TempDir())
	require.NoError(t, err)

	// A nonexistent key should return nil, even if the key is empty
	require.Nil(t, db.Get([]byte("")))

	// A nonexistent key should return nil, even if the key is nil
	require.Nil(t, db.Get(nil))

	// A nonexistent key should return nil.
	key := []byte("abc")
	require.Nil(t, db.Get(key))

	// Set empty value.
	db.SetSync(key, []byte(""))
	require.NotNil(t, db.Get(key))
	require.Empty(t, db.Get(key))

	// Set nil value.
	db.SetSync(key, nil)
	require.NotNil(t, db.Get(key))
	require.Empty(t, db.Get(key))

	// Delete.
	db.DeleteSync(key)
	require.Nil(t, db.Get(key))
}

func TestBackendsGetSetDelete(t *testing.T) {
	for dbType := range backends {
		t.Run(string(dbType), func(t *testing.T) {
			testBackendGetSetDelete(t, dbType)
		})
	}
}

func withDB(t *testing.T, creator dbCreator, fn func(DB)) {
	t.Helper()

	name := fmt.Sprintf("test_%x", randStr(12))
	db, err := creator(name, t.TempDir())
	require.Nil(t, err)
	fn(db)
	db.Close()
}

func TestBackendsNilKeys(t *testing.T) {
	// Test all backends.
	for dbType, creator := range backends {
		withDB(t, creator, func(db DB) {
			t.Run(fmt.Sprintf("Testing %s", dbType), func(t *testing.T) {
				// Nil keys are treated as the empty key for most operations.
				expect := func(key, value []byte) {
					if len(key) == 0 { // nil or empty
						assert.Equal(t, db.Get(nil), db.Get([]byte("")))
						assert.Equal(t, db.Has(nil), db.Has([]byte("")))
					}
					assert.Equal(t, db.Get(key), value)
					assert.Equal(t, db.Has(key), value != nil)
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

func TestGoLevelDBBackend(t *testing.T) {
	name := fmt.Sprintf("test_%x", randStr(12))
	db, err := NewDB(name, GoLevelDBBackend, t.TempDir())
	require.NoError(t, err)

	_, ok := db.(*GoLevelDB)
	assert.True(t, ok)
}

func TestDBIterator(t *testing.T) {
	for dbType := range backends {
		t.Run(fmt.Sprintf("%v", dbType), func(t *testing.T) {
			t.Helper()

			testDBIterator(t, dbType)
		})
	}
}

func testDBIterator(t *testing.T, backend BackendType) {
	t.Helper()

	name := fmt.Sprintf("test_%x", randStr(12))
	db, err := NewDB(name, backend, t.TempDir())
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		if i != 6 { // but skip 6.
			db.Set(int642Bytes(int64(i)), nil)
		}
	}

	verifyIterator(t, db.Iterator(nil, nil), []int64{0, 1, 2, 3, 4, 5, 7, 8, 9}, "forward iterator")
	verifyIterator(t, db.ReverseIterator(nil, nil), []int64{9, 8, 7, 5, 4, 3, 2, 1, 0}, "reverse iterator")

	verifyIterator(t, db.Iterator(nil, int642Bytes(0)), []int64(nil), "forward iterator to 0")
	verifyIterator(t, db.ReverseIterator(int642Bytes(10), nil), []int64(nil), "reverse iterator from 10 (ex)")

	verifyIterator(t, db.Iterator(int642Bytes(0), nil), []int64{0, 1, 2, 3, 4, 5, 7, 8, 9}, "forward iterator from 0")
	verifyIterator(t, db.Iterator(int642Bytes(1), nil), []int64{1, 2, 3, 4, 5, 7, 8, 9}, "forward iterator from 1")
	verifyIterator(t, db.ReverseIterator(nil, int642Bytes(10)),
		[]int64{9, 8, 7, 5, 4, 3, 2, 1, 0}, "reverse iterator from 10 (ex)")
	verifyIterator(t, db.ReverseIterator(nil, int642Bytes(9)),
		[]int64{8, 7, 5, 4, 3, 2, 1, 0}, "reverse iterator from 9 (ex)")
	verifyIterator(t, db.ReverseIterator(nil, int642Bytes(8)),
		[]int64{7, 5, 4, 3, 2, 1, 0}, "reverse iterator from 8 (ex)")

	verifyIterator(t, db.Iterator(int642Bytes(5), int642Bytes(6)), []int64{5}, "forward iterator from 5 to 6")
	verifyIterator(t, db.Iterator(int642Bytes(5), int642Bytes(7)), []int64{5}, "forward iterator from 5 to 7")
	verifyIterator(t, db.Iterator(int642Bytes(5), int642Bytes(8)), []int64{5, 7}, "forward iterator from 5 to 8")
	verifyIterator(t, db.Iterator(int642Bytes(6), int642Bytes(7)), []int64(nil), "forward iterator from 6 to 7")
	verifyIterator(t, db.Iterator(int642Bytes(6), int642Bytes(8)), []int64{7}, "forward iterator from 6 to 8")
	verifyIterator(t, db.Iterator(int642Bytes(7), int642Bytes(8)), []int64{7}, "forward iterator from 7 to 8")

	verifyIterator(t, db.ReverseIterator(int642Bytes(4), int642Bytes(5)), []int64{4}, "reverse iterator from 5 (ex) to 4")
	verifyIterator(t, db.ReverseIterator(int642Bytes(4), int642Bytes(6)),
		[]int64{5, 4}, "reverse iterator from 6 (ex) to 4")
	verifyIterator(t, db.ReverseIterator(int642Bytes(4), int642Bytes(7)),
		[]int64{5, 4}, "reverse iterator from 7 (ex) to 4")
	verifyIterator(t, db.ReverseIterator(int642Bytes(5), int642Bytes(6)), []int64{5}, "reverse iterator from 6 (ex) to 5")
	verifyIterator(t, db.ReverseIterator(int642Bytes(5), int642Bytes(7)), []int64{5}, "reverse iterator from 7 (ex) to 5")
	verifyIterator(t, db.ReverseIterator(int642Bytes(6), int642Bytes(7)),
		[]int64(nil), "reverse iterator from 7 (ex) to 6")

	verifyIterator(t, db.Iterator(int642Bytes(0), int642Bytes(1)), []int64{0}, "forward iterator from 0 to 1")
	verifyIterator(t, db.ReverseIterator(int642Bytes(8), int642Bytes(9)), []int64{8}, "reverse iterator from 9 (ex) to 8")

	verifyIterator(t, db.Iterator(int642Bytes(2), int642Bytes(4)), []int64{2, 3}, "forward iterator from 2 to 4")
	verifyIterator(t, db.Iterator(int642Bytes(4), int642Bytes(2)), []int64(nil), "forward iterator from 4 to 2")
	verifyIterator(t, db.ReverseIterator(int642Bytes(2), int642Bytes(4)),
		[]int64{3, 2}, "reverse iterator from 4 (ex) to 2")
	verifyIterator(t, db.ReverseIterator(int642Bytes(4), int642Bytes(2)),
		[]int64(nil), "reverse iterator from 2 (ex) to 4")
}

func verifyIterator(t *testing.T, itr Iterator, expected []int64, msg string) {
	t.Helper()

	var list []int64
	for itr.Valid() {
		list = append(list, bytes2Int64(itr.Key()))
		itr.Next()
	}
	assert.Equal(t, expected, list, msg)
}
