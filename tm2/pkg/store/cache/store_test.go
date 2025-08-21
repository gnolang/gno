package cache_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/random"
	"github.com/gnolang/gno/tm2/pkg/store/cache"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

func newCacheStore() types.Store {
	mem := dbadapter.Store{DB: memdb.NewMemDB()}
	return cache.New(mem)
}

func keyFmt(i int) []byte { return bz(fmt.Sprintf("key%0.8d", i)) }
func valFmt(i int) []byte { return bz(fmt.Sprintf("value%0.8d", i)) }

func TestCacheStore(t *testing.T) {
	t.Parallel()

	mem := dbadapter.Store{DB: memdb.NewMemDB()}
	st := cache.New(mem)

	require.Empty(t, st.Get(keyFmt(1)), "Expected `key1` to be empty")

	// put something in mem and in cache
	mem.Set(keyFmt(1), valFmt(1))
	st.Set(keyFmt(1), valFmt(1))
	require.Equal(t, valFmt(1), st.Get(keyFmt(1)))

	// update it in cache, shouldn't change mem
	st.Set(keyFmt(1), valFmt(2))
	require.Equal(t, valFmt(2), st.Get(keyFmt(1)))
	require.Equal(t, valFmt(1), mem.Get(keyFmt(1)))

	// write it. should change mem
	st.Write()
	require.Equal(t, valFmt(2), mem.Get(keyFmt(1)))
	require.Equal(t, valFmt(2), st.Get(keyFmt(1)))

	// more writes and checks
	st.Write()
	st.Write()
	require.Equal(t, valFmt(2), mem.Get(keyFmt(1)))
	require.Equal(t, valFmt(2), st.Get(keyFmt(1)))

	// make a new one, check it
	st = cache.New(mem)
	require.Equal(t, valFmt(2), st.Get(keyFmt(1)))

	// make a new one and delete - should not be removed from mem
	st = cache.New(mem)
	st.Delete(keyFmt(1))
	require.Empty(t, st.Get(keyFmt(1)))
	require.Equal(t, mem.Get(keyFmt(1)), valFmt(2))

	// Write. should now be removed from both
	st.Write()
	require.Empty(t, st.Get(keyFmt(1)), "Expected `key1` to be empty")
	require.Empty(t, mem.Get(keyFmt(1)), "Expected `key1` to be empty")
}

func TestCacheStoreNoNilSet(t *testing.T) {
	t.Parallel()

	mem := dbadapter.Store{DB: memdb.NewMemDB()}
	st := cache.New(mem)
	require.Panics(t, func() { st.Set([]byte("key"), nil) }, "setting a nil value should panic")
}

func TestCacheStoreNested(t *testing.T) {
	t.Parallel()

	mem := dbadapter.Store{DB: memdb.NewMemDB()}
	st := cache.New(mem)

	// set. check its there on st and not on mem.
	st.Set(keyFmt(1), valFmt(1))
	require.Empty(t, mem.Get(keyFmt(1)))
	require.Equal(t, valFmt(1), st.Get(keyFmt(1)))

	// make a new from st and check
	st2 := cache.New(st)
	require.Equal(t, valFmt(1), st2.Get(keyFmt(1)))

	// update the value on st2, check it only effects st2
	st2.Set(keyFmt(1), valFmt(3))
	require.Equal(t, []byte(nil), mem.Get(keyFmt(1)))
	require.Equal(t, valFmt(1), st.Get(keyFmt(1)))
	require.Equal(t, valFmt(3), st2.Get(keyFmt(1)))

	// st2 writes to its parent, st. doesn't effect mem
	st2.Write()
	require.Equal(t, []byte(nil), mem.Get(keyFmt(1)))
	require.Equal(t, valFmt(3), st.Get(keyFmt(1)))

	// updates mem
	st.Write()
	require.Equal(t, valFmt(3), mem.Get(keyFmt(1)))
}

func TestCacheKVIteratorBounds(t *testing.T) {
	t.Parallel()

	st := newCacheStore()

	// set some items
	nItems := 5
	for i := range nItems {
		st.Set(keyFmt(i), valFmt(i))
	}

	// iterate over all of them
	itr := st.Iterator(nil, nil)
	i := 0
	for ; itr.Valid(); itr.Next() {
		k, v := itr.Key(), itr.Value()
		require.Equal(t, keyFmt(i), k)
		require.Equal(t, valFmt(i), v)
		i++
	}
	require.Equal(t, nItems, i)
	require.NoError(t, itr.Error())
	require.NoError(t, itr.Close())

	// iterate over none
	itr = st.Iterator(bz("money"), nil)
	i = 0
	for ; itr.Valid(); itr.Next() {
		i++
	}
	require.Equal(t, 0, i)
	require.NoError(t, itr.Error())
	require.NoError(t, itr.Close())

	// iterate over lower
	itr = st.Iterator(keyFmt(0), keyFmt(3))
	i = 0
	for ; itr.Valid(); itr.Next() {
		k, v := itr.Key(), itr.Value()
		require.Equal(t, keyFmt(i), k)
		require.Equal(t, valFmt(i), v)
		i++
	}
	require.Equal(t, 3, i)
	require.NoError(t, itr.Error())
	require.NoError(t, itr.Close())

	// iterate over upper
	itr = st.Iterator(keyFmt(2), keyFmt(4))
	i = 2
	for ; itr.Valid(); itr.Next() {
		k, v := itr.Key(), itr.Value()
		require.Equal(t, keyFmt(i), k)
		require.Equal(t, valFmt(i), v)
		i++
	}
	require.Equal(t, 4, i)
	require.NoError(t, itr.Error())
	require.NoError(t, itr.Close())
}

func TestCacheKVReverseIteratorBounds(t *testing.T) {
	t.Parallel()

	st := newCacheStore()

	// set some items
	nItems := 5
	for i := range nItems {
		st.Set(keyFmt(i), valFmt(i))
	}

	// iterate over all of them in reverse
	i := nItems - 1
	for itr := st.ReverseIterator(nil, nil); itr.Valid(); itr.Next() {
		require.Equal(t, keyFmt(i), itr.Key())
		require.Equal(t, valFmt(i), itr.Value())
		i--
	}
	require.Equal(t, -1, i)

	// iterate over none
	i = 0
	for itr := st.ReverseIterator(bz("money"), nil); itr.Valid(); itr.Next() {
		i++
	}
	require.Equal(t, 0, i)

	// iterate over lower
	i = 2
	for itr := st.ReverseIterator(keyFmt(0), keyFmt(3)); itr.Valid(); itr.Next() {
		require.Equal(t, keyFmt(i), itr.Key())
		require.Equal(t, valFmt(i), itr.Value())
		i--
	}
	require.Equal(t, -1, i)

	// iterate over upper
	i = 3
	for itr := st.ReverseIterator(keyFmt(2), keyFmt(4)); itr.Valid(); itr.Next() {
		require.Equal(t, keyFmt(i), itr.Key())
		require.Equal(t, valFmt(i), itr.Value())
		i--
	}
	require.Equal(t, 1, i)
}

func TestCacheKVMergeIteratorBasics(t *testing.T) {
	t.Parallel()

	st := newCacheStore()

	// set and delete an item in the cache, iterator should be empty
	k, v := keyFmt(0), valFmt(0)
	st.Set(k, v)
	st.Delete(k)
	assertIterateDomain(t, st, 0)

	// now set it and assert its there
	st.Set(k, v)
	assertIterateDomain(t, st, 1)

	// write it and assert its there
	st.Write()
	assertIterateDomain(t, st, 1)

	// remove it in cache and assert its not
	st.Delete(k)
	assertIterateDomain(t, st, 0)

	// write the delete and assert its not there
	st.Write()
	assertIterateDomain(t, st, 0)

	// add two keys and assert they're there
	k1, v1 := keyFmt(1), valFmt(1)
	st.Set(k, v)
	st.Set(k1, v1)
	assertIterateDomain(t, st, 2)

	// write it and assert they're there
	st.Write()
	assertIterateDomain(t, st, 2)

	// remove one in cache and assert its not
	st.Delete(k1)
	assertIterateDomain(t, st, 1)

	// write the delete and assert its not there
	st.Write()
	assertIterateDomain(t, st, 1)

	// delete the other key in cache and asserts its empty
	st.Delete(k)
	assertIterateDomain(t, st, 0)
}

func TestCacheKVMergeIteratorDeleteLast(t *testing.T) {
	t.Parallel()

	st := newCacheStore()

	// set some items and write them
	nItems := 5
	for i := range nItems {
		st.Set(keyFmt(i), valFmt(i))
	}
	st.Write()

	// set some more items and leave dirty
	for i := nItems; i < nItems*2; i++ {
		st.Set(keyFmt(i), valFmt(i))
	}

	// iterate over all of them
	assertIterateDomain(t, st, nItems*2)

	// delete them all
	for i := range nItems * 2 {
		last := nItems*2 - 1 - i
		st.Delete(keyFmt(last))
		assertIterateDomain(t, st, last)
	}
}

func TestCacheKVMergeIteratorDeletes(t *testing.T) {
	t.Parallel()

	st := newCacheStore()
	truth := memdb.NewMemDB()

	// set some items and write them
	nItems := 10
	for i := range nItems {
		doOp(st, truth, opSet, i)
	}
	st.Write()

	// delete every other item, starting from 0
	for i := 0; i < nItems; i += 2 {
		doOp(st, truth, opDel, i)
		assertIterateDomainCompare(t, st, truth)
	}

	// reset
	st = newCacheStore()
	truth = memdb.NewMemDB()

	// set some items and write them
	for i := range nItems {
		doOp(st, truth, opSet, i)
	}
	st.Write()

	// delete every other item, starting from 1
	for i := 1; i < nItems; i += 2 {
		doOp(st, truth, opDel, i)
		assertIterateDomainCompare(t, st, truth)
	}
}

func TestCacheKVMergeIteratorChunks(t *testing.T) {
	t.Parallel()

	st := newCacheStore()

	// Use the truth to check values on the merge iterator
	truth := memdb.NewMemDB()

	// sets to the parent
	setRange(st, truth, 0, 20)
	setRange(st, truth, 40, 60)
	st.Write()

	// sets to the cache
	setRange(st, truth, 20, 40)
	setRange(st, truth, 60, 80)
	assertIterateDomainCheck(t, st, truth, []keyRange{{0, 80}})

	// remove some parents and some cache
	deleteRange(st, truth, 15, 25)
	assertIterateDomainCheck(t, st, truth, []keyRange{{0, 15}, {25, 80}})

	// remove some parents and some cache
	deleteRange(st, truth, 35, 45)
	assertIterateDomainCheck(t, st, truth, []keyRange{{0, 15}, {25, 35}, {45, 80}})

	// write, add more to the cache, and delete some cache
	st.Write()
	setRange(st, truth, 38, 42)
	deleteRange(st, truth, 40, 43)
	assertIterateDomainCheck(t, st, truth, []keyRange{{0, 15}, {25, 35}, {38, 40}, {45, 80}})
}

func TestCacheKVMergeIteratorRandom(t *testing.T) {
	t.Parallel()

	st := newCacheStore()
	truth := memdb.NewMemDB()

	start, end := 25, 975
	maxVal := 1000
	setRange(st, truth, start, end)

	// do an op, test the iterator
	for range 2000 {
		doRandomOp(st, truth, maxVal)
		assertIterateDomainCompare(t, st, truth)
	}
}

// -------------------------------------------------------------------------------------------
// do some random ops

const (
	opSet      = 0
	opSetRange = 1
	opDel      = 2
	opDelRange = 3
	opWrite    = 4

	totalOps = 5 // number of possible operations
)

func randInt(n int) int {
	return random.RandInt() % n
}

// useful for replaying a error case if we find one
func doOp(st types.Store, truth dbm.DB, op int, args ...int) {
	switch op {
	case opSet:
		k := args[0]
		st.Set(keyFmt(k), valFmt(k))
		truth.Set(keyFmt(k), valFmt(k))
	case opSetRange:
		start := args[0]
		end := args[1]
		setRange(st, truth, start, end)
	case opDel:
		k := args[0]
		st.Delete(keyFmt(k))
		truth.Delete(keyFmt(k))
	case opDelRange:
		start := args[0]
		end := args[1]
		deleteRange(st, truth, start, end)
	case opWrite:
		st.Write()
	}
}

func doRandomOp(st types.Store, truth dbm.DB, maxKey int) {
	r := randInt(totalOps)
	switch r {
	case opSet:
		k := randInt(maxKey)
		st.Set(keyFmt(k), valFmt(k))
		truth.Set(keyFmt(k), valFmt(k))
	case opSetRange:
		start := randInt(maxKey - 2)
		end := randInt(maxKey-start) + start
		setRange(st, truth, start, end)
	case opDel:
		k := randInt(maxKey)
		st.Delete(keyFmt(k))
		truth.Delete(keyFmt(k))
	case opDelRange:
		start := randInt(maxKey - 2)
		end := randInt(maxKey-start) + start
		deleteRange(st, truth, start, end)
	case opWrite:
		st.Write()
	}
}

// -------------------------------------------------------------------------------------------

// iterate over whole domain
func assertIterateDomain(t *testing.T, st types.Store, expectedN int) {
	t.Helper()

	itr := st.Iterator(nil, nil)
	i := 0
	for ; itr.Valid(); itr.Next() {
		k, v := itr.Key(), itr.Value()
		require.Equal(t, keyFmt(i), k)
		require.Equal(t, valFmt(i), v)
		i++
	}
	require.Equal(t, expectedN, i)
	require.NoError(t, itr.Error())
	require.NoError(t, itr.Close())
}

func assertIterateDomainCheck(t *testing.T, st types.Store, mem dbm.DB, r []keyRange) {
	t.Helper()

	// iterate over each and check they match the other
	itr := st.Iterator(nil, nil)
	itr2, err := mem.Iterator(nil, nil) // ground truth
	require.NoError(t, err)

	krc := newKeyRangeCounter(r)
	i := 0

	for ; krc.valid(); krc.next() {
		require.True(t, itr.Valid())
		require.True(t, itr2.Valid())

		// check the key/val matches the ground truth
		k, v := itr.Key(), itr.Value()
		k2, v2 := itr2.Key(), itr2.Value()
		require.Equal(t, k, k2)
		require.Equal(t, v, v2)

		// check they match the counter
		require.Equal(t, k, keyFmt(krc.key()))

		itr.Next()
		itr2.Next()
		i++
	}

	require.False(t, itr.Valid())
	require.False(t, itr2.Valid())
}

func assertIterateDomainCompare(t *testing.T, st types.Store, mem dbm.DB) {
	t.Helper()

	// iterate over each and check they match the other
	itr := st.Iterator(nil, nil)
	itr2, err := mem.Iterator(nil, nil) // ground truth
	require.NoError(t, err)
	checkIterators(t, itr, itr2)
	checkIterators(t, itr2, itr)
}

func checkIterators(t *testing.T, itr, itr2 types.Iterator) {
	t.Helper()

	for ; itr.Valid(); itr.Next() {
		require.True(t, itr2.Valid())
		k, v := itr.Key(), itr.Value()
		k2, v2 := itr2.Key(), itr2.Value()
		require.Equal(t, k, k2)
		require.Equal(t, v, v2)
		itr2.Next()
	}
	require.False(t, itr.Valid())
	require.False(t, itr2.Valid())
}

// --------------------------------------------------------

func setRange(st types.Store, mem dbm.DB, start, end int) {
	for i := start; i < end; i++ {
		st.Set(keyFmt(i), valFmt(i))
		mem.Set(keyFmt(i), valFmt(i))
	}
}

func deleteRange(st types.Store, mem dbm.DB, start, end int) {
	for i := start; i < end; i++ {
		st.Delete(keyFmt(i))
		mem.Delete(keyFmt(i))
	}
}

// --------------------------------------------------------

type keyRange struct {
	start int
	end   int
}

func (kr keyRange) len() int {
	return kr.end - kr.start
}

func newKeyRangeCounter(kr []keyRange) *keyRangeCounter {
	return &keyRangeCounter{keyRanges: kr}
}

// we can iterate over this and make sure our real iterators have all the right keys
type keyRangeCounter struct {
	rangeIdx  int
	idx       int
	keyRanges []keyRange
}

func (krc *keyRangeCounter) valid() bool {
	maxRangeIdx := len(krc.keyRanges) - 1
	maxRange := krc.keyRanges[maxRangeIdx]

	// if we're not in the max range, we're valid
	if krc.rangeIdx <= maxRangeIdx &&
		krc.idx < maxRange.len() {
		return true
	}

	return false
}

func (krc *keyRangeCounter) next() {
	thisKeyRange := krc.keyRanges[krc.rangeIdx]
	if krc.idx == thisKeyRange.len()-1 {
		krc.rangeIdx++
		krc.idx = 0
	} else {
		krc.idx++
	}
}

func (krc *keyRangeCounter) key() int {
	thisKeyRange := krc.keyRanges[krc.rangeIdx]
	return thisKeyRange.start + krc.idx
}

// --------------------------------------------------------

func bz(s string) []byte { return []byte(s) }

func BenchmarkCacheStoreGetNoKeyFound(b *testing.B) {
	st := newCacheStore()
	b.ResetTimer()
	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		st.Get([]byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)})
	}
}

func BenchmarkCacheStoreGetKeyFound(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping testing in short mode")
	}

	st := newCacheStore()
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		st.Set(arr, arr)
	}
	b.ResetTimer()
	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		st.Get([]byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)})
	}
}
