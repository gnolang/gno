package prefix

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/iavl"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storeiavl "github.com/gnolang/gno/tm2/pkg/store/iavl"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// copied from iavl/store_test.go
var (
	cacheSize        = 100
	numRecent  int64 = 5
	storeEvery int64 = 3
)

func bz(s string) []byte { return []byte(s) }

type kvpair struct {
	key   []byte
	value []byte
}

func genRandomKVPairs() []kvpair {
	kvps := make([]kvpair, 20)

	for i := range 20 {
		kvps[i].key = make([]byte, 32)
		rand.Read(kvps[i].key)
		kvps[i].value = make([]byte, 32)
		rand.Read(kvps[i].value)
	}

	return kvps
}

func setRandomKVPairs(store types.Store) []kvpair {
	kvps := genRandomKVPairs()
	for _, kvp := range kvps {
		store.Set(nil, kvp.key, kvp.value)
	}
	return kvps
}

func testPrefixStore(t *testing.T, baseStore types.Store, prefix []byte) {
	t.Helper()

	prefixStore := New(baseStore, prefix)
	prefixPrefixStore := New(prefixStore, []byte("prefix"))

	require.Panics(t, func() { prefixStore.Get(nil, nil) })
	require.Panics(t, func() { prefixStore.Set(nil, nil, []byte{}) })

	kvps := setRandomKVPairs(prefixPrefixStore)

	for i := range 20 {
		key := kvps[i].key
		value := kvps[i].value
		require.True(t, prefixPrefixStore.Has(nil, key))
		require.Equal(t, value, prefixPrefixStore.Get(nil, key))

		key = append([]byte("prefix"), key...)
		require.True(t, prefixStore.Has(nil, key))
		require.Equal(t, value, prefixStore.Get(nil, key))
		key = append(prefix, key...)
		require.True(t, baseStore.Has(nil, key))
		require.Equal(t, value, baseStore.Get(nil, key))

		key = kvps[i].key
		prefixPrefixStore.Delete(nil, key)
		require.False(t, prefixPrefixStore.Has(nil, key))
		require.Nil(t, prefixPrefixStore.Get(nil, key))
		key = append([]byte("prefix"), key...)
		require.False(t, prefixStore.Has(nil, key))
		require.Nil(t, prefixStore.Get(nil, key))
		key = append(prefix, key...)
		require.False(t, baseStore.Has(nil, key))
		require.Nil(t, baseStore.Get(nil, key))
	}
}

func TestIAVLStorePrefix(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	tree := iavl.NewMutableTree(db, cacheSize, false, iavl.NewNopLogger())
	iavlStore := storeiavl.UnsafeNewStore(tree, types.StoreOptions{
		PruningOptions: types.PruningOptions{
			KeepRecent: numRecent,
			KeepEvery:  storeEvery,
		},
	})

	testPrefixStore(t, iavlStore, []byte("test"))
}

func TestPrefixStoreNoNilSet(t *testing.T) {
	t.Parallel()

	mem := dbadapter.Store{DB: memdb.NewMemDB()}
	cs := mem.CacheWrap()
	require.Panics(t, func() { cs.Set(nil, []byte("key"), nil) }, "setting a nil value should panic")
}

func TestPrefixStoreIterate(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	baseStore := dbadapter.Store{DB: db}
	prefix := []byte("test")
	prefixStore := New(baseStore, prefix)

	setRandomKVPairs(prefixStore)

	bIter := types.PrefixIterator(nil, baseStore, prefix)
	pIter := types.PrefixIterator(nil, prefixStore, nil)

	for bIter.Valid() && pIter.Valid() {
		require.Equal(t, bIter.Key(), append(prefix, pIter.Key()...))
		require.Equal(t, bIter.Value(), pIter.Value())

		bIter.Next()
		pIter.Next()
	}

	bIter.Close()
	pIter.Close()
}

func incFirstByte(bz []byte) {
	bz[0]++
}

func TestCloneAppend(t *testing.T) {
	t.Parallel()

	kvps := genRandomKVPairs()
	for _, kvp := range kvps {
		bz := cloneAppend(kvp.key, kvp.value)
		require.Equal(t, bz, append(kvp.key, kvp.value...))

		incFirstByte(bz)
		require.NotEqual(t, bz, append(kvp.key, kvp.value...))

		bz = cloneAppend(kvp.key, kvp.value)
		incFirstByte(kvp.key)
		require.NotEqual(t, bz, append(kvp.key, kvp.value...))

		bz = cloneAppend(kvp.key, kvp.value)
		incFirstByte(kvp.value)
		require.NotEqual(t, bz, append(kvp.key, kvp.value...))
	}
}

func TestPrefixStoreIteratorEdgeCase(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	baseStore := dbadapter.Store{DB: db}

	// overflow in cpIncr
	prefix := []byte{0xAA, 0xFF, 0xFF}
	prefixStore := New(baseStore, prefix)

	// ascending order
	baseStore.Set(nil, []byte{0xAA, 0xFF, 0xFE}, []byte{})
	baseStore.Set(nil, []byte{0xAA, 0xFF, 0xFE, 0x00}, []byte{})
	baseStore.Set(nil, []byte{0xAA, 0xFF, 0xFF}, []byte{})
	baseStore.Set(nil, []byte{0xAA, 0xFF, 0xFF, 0x00}, []byte{})
	baseStore.Set(nil, []byte{0xAB}, []byte{})
	baseStore.Set(nil, []byte{0xAB, 0x00}, []byte{})
	baseStore.Set(nil, []byte{0xAB, 0x00, 0x00}, []byte{})

	iter := prefixStore.Iterator(nil, nil, nil)

	checkDomain(t, iter, nil, nil)
	checkItem(t, iter, []byte{}, bz(""))
	checkNext(t, iter, true)
	checkItem(t, iter, []byte{0x00}, bz(""))
	checkNext(t, iter, false)

	checkInvalid(t, iter)

	iter.Close()
}

func TestPrefixStoreReverseIteratorEdgeCase(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	baseStore := dbadapter.Store{DB: db}

	// overflow in cpIncr
	prefix := []byte{0xAA, 0xFF, 0xFF}
	prefixStore := New(baseStore, prefix)

	// descending order
	baseStore.Set(nil, []byte{0xAB, 0x00, 0x00}, []byte{})
	baseStore.Set(nil, []byte{0xAB, 0x00}, []byte{})
	baseStore.Set(nil, []byte{0xAB}, []byte{})
	baseStore.Set(nil, []byte{0xAA, 0xFF, 0xFF, 0x00}, []byte{})
	baseStore.Set(nil, []byte{0xAA, 0xFF, 0xFF}, []byte{})
	baseStore.Set(nil, []byte{0xAA, 0xFF, 0xFE, 0x00}, []byte{})
	baseStore.Set(nil, []byte{0xAA, 0xFF, 0xFE}, []byte{})

	iter := prefixStore.ReverseIterator(nil, nil, nil)

	checkDomain(t, iter, nil, nil)
	checkItem(t, iter, []byte{0x00}, bz(""))
	checkNext(t, iter, true)
	checkItem(t, iter, []byte{}, bz(""))
	checkNext(t, iter, false)

	checkInvalid(t, iter)

	iter.Close()

	db = memdb.NewMemDB()
	baseStore = dbadapter.Store{DB: db}

	// underflow in cpDecr
	prefix = []byte{0xAA, 0x00, 0x00}
	prefixStore = New(baseStore, prefix)

	baseStore.Set(nil, []byte{0xAB, 0x00, 0x01, 0x00, 0x00}, []byte{})
	baseStore.Set(nil, []byte{0xAB, 0x00, 0x01, 0x00}, []byte{})
	baseStore.Set(nil, []byte{0xAB, 0x00, 0x01}, []byte{})
	baseStore.Set(nil, []byte{0xAA, 0x00, 0x00, 0x00}, []byte{})
	baseStore.Set(nil, []byte{0xAA, 0x00, 0x00}, []byte{})
	baseStore.Set(nil, []byte{0xA9, 0xFF, 0xFF, 0x00}, []byte{})
	baseStore.Set(nil, []byte{0xA9, 0xFF, 0xFF}, []byte{})

	iter = prefixStore.ReverseIterator(nil, nil, nil)

	checkDomain(t, iter, nil, nil)
	checkItem(t, iter, []byte{0x00}, bz(""))
	checkNext(t, iter, true)
	checkItem(t, iter, []byte{}, bz(""))
	checkNext(t, iter, false)

	checkInvalid(t, iter)

	iter.Close()
}

// Tests below are ported from https://github.com/tendermint/classic/blob/master/libs/db/prefix_db_test.go

func mockStoreWithStuff() types.Store {
	db := memdb.NewMemDB()
	store := dbadapter.Store{DB: db}
	// Under "key" prefix
	store.Set(nil, bz("key"), bz("value"))
	store.Set(nil, bz("key1"), bz("value1"))
	store.Set(nil, bz("key2"), bz("value2"))
	store.Set(nil, bz("key3"), bz("value3"))
	store.Set(nil, bz("something"), bz("else"))
	store.Set(nil, bz(""), bz(""))
	store.Set(nil, bz("k"), bz("g"))
	store.Set(nil, bz("ke"), bz("valu"))
	store.Set(nil, bz("kee"), bz("valuu"))
	return store
}

func checkValue(t *testing.T, store types.Store, key []byte, expected []byte) {
	t.Helper()

	bz := store.Get(nil, key)
	require.Equal(t, expected, bz)
}

func checkValid(t *testing.T, itr types.Iterator, expected bool) {
	t.Helper()

	valid := itr.Valid()
	require.Equal(t, expected, valid)
}

func checkNext(t *testing.T, itr types.Iterator, expected bool) {
	t.Helper()

	itr.Next()
	valid := itr.Valid()
	require.Equal(t, expected, valid)
}

func checkDomain(t *testing.T, itr types.Iterator, start, end []byte) {
	t.Helper()

	ds, de := itr.Domain()
	require.Equal(t, start, ds)
	require.Equal(t, end, de)
}

func checkItem(t *testing.T, itr types.Iterator, key, value []byte) {
	t.Helper()

	require.Exactly(t, key, itr.Key())
	require.Exactly(t, value, itr.Value())
}

func checkInvalid(t *testing.T, itr types.Iterator) {
	t.Helper()

	checkValid(t, itr, false)
	checkKeyPanics(t, itr)
	checkValuePanics(t, itr)
	checkNextPanics(t, itr)
}

func checkKeyPanics(t *testing.T, itr types.Iterator) {
	t.Helper()

	require.Panics(t, func() { itr.Key() })
}

func checkValuePanics(t *testing.T, itr types.Iterator) {
	t.Helper()

	require.Panics(t, func() { itr.Value() })
}

func checkNextPanics(t *testing.T, itr types.Iterator) {
	t.Helper()

	require.Panics(t, func() { itr.Next() })
}

func TestPrefixDBSimple(t *testing.T) {
	t.Parallel()

	store := mockStoreWithStuff()
	pstore := New(store, bz("key"))

	checkValue(t, pstore, bz("key"), nil)
	checkValue(t, pstore, bz(""), bz("value"))
	checkValue(t, pstore, bz("key1"), nil)
	checkValue(t, pstore, bz("1"), bz("value1"))
	checkValue(t, pstore, bz("key2"), nil)
	checkValue(t, pstore, bz("2"), bz("value2"))
	checkValue(t, pstore, bz("key3"), nil)
	checkValue(t, pstore, bz("3"), bz("value3"))
	checkValue(t, pstore, bz("something"), nil)
	checkValue(t, pstore, bz("k"), nil)
	checkValue(t, pstore, bz("ke"), nil)
	checkValue(t, pstore, bz("kee"), nil)
}

func TestPrefixDBIterator1(t *testing.T) {
	t.Parallel()

	store := mockStoreWithStuff()
	pstore := New(store, bz("key"))

	itr := pstore.Iterator(nil, nil, nil)
	checkDomain(t, itr, nil, nil)
	checkItem(t, itr, bz(""), bz("value"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("1"), bz("value1"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("2"), bz("value2"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("3"), bz("value3"))
	checkNext(t, itr, false)
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBIterator2(t *testing.T) {
	t.Parallel()

	store := mockStoreWithStuff()
	pstore := New(store, bz("key"))

	itr := pstore.Iterator(nil, nil, bz(""))
	checkDomain(t, itr, nil, bz(""))
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBIterator3(t *testing.T) {
	t.Parallel()

	store := mockStoreWithStuff()
	pstore := New(store, bz("key"))

	itr := pstore.Iterator(nil, bz(""), nil)
	checkDomain(t, itr, bz(""), nil)
	checkItem(t, itr, bz(""), bz("value"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("1"), bz("value1"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("2"), bz("value2"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("3"), bz("value3"))
	checkNext(t, itr, false)
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBIterator4(t *testing.T) {
	t.Parallel()

	store := mockStoreWithStuff()
	pstore := New(store, bz("key"))

	itr := pstore.Iterator(nil, bz(""), bz(""))
	checkDomain(t, itr, bz(""), bz(""))
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBReverseIterator1(t *testing.T) {
	t.Parallel()

	store := mockStoreWithStuff()
	pstore := New(store, bz("key"))

	itr := pstore.ReverseIterator(nil, nil, nil)
	checkDomain(t, itr, nil, nil)
	checkItem(t, itr, bz("3"), bz("value3"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("2"), bz("value2"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("1"), bz("value1"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz(""), bz("value"))
	checkNext(t, itr, false)
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBReverseIterator2(t *testing.T) {
	t.Parallel()

	store := mockStoreWithStuff()
	pstore := New(store, bz("key"))

	itr := pstore.ReverseIterator(nil, bz(""), nil)
	checkDomain(t, itr, bz(""), nil)
	checkItem(t, itr, bz("3"), bz("value3"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("2"), bz("value2"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz("1"), bz("value1"))
	checkNext(t, itr, true)
	checkItem(t, itr, bz(""), bz("value"))
	checkNext(t, itr, false)
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBReverseIterator3(t *testing.T) {
	t.Parallel()

	store := mockStoreWithStuff()
	pstore := New(store, bz("key"))

	itr := pstore.ReverseIterator(nil, nil, bz(""))
	checkDomain(t, itr, nil, bz(""))
	checkInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBReverseIterator4(t *testing.T) {
	t.Parallel()

	store := mockStoreWithStuff()
	pstore := New(store, bz("key"))

	itr := pstore.ReverseIterator(nil, bz(""), bz(""))
	checkInvalid(t, itr)
	itr.Close()
}
