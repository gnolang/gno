package goleveldb

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/internal"
)

func TestGoLevelDBNewGoLevelDB(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	name := fmt.Sprintf("test_%x", internal.RandStr(12))

	// Test we can't open the db twice for writing
	wr1, err := NewGoLevelDB(name, dir)
	require.Nil(t, err)
	_, err = NewGoLevelDB(name, dir)
	require.NotNil(t, err)
	wr1.Close() // Close the db to release the lock

	// Test we can open the db twice for reading only
	ro1, err := NewGoLevelDBWithOpts(name, dir, &opt.Options{ReadOnly: true})
	require.Nil(t, err)
	defer ro1.Close()
	ro2, err := NewGoLevelDBWithOpts(name, dir, &opt.Options{ReadOnly: true})
	require.Nil(t, err)
	defer ro2.Close()
}

func TestGoLevelDBBackend(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	db, err := db.NewDB(name, db.GoLevelDBBackend, t.TempDir())
	require.NoError(t, err)

	_, ok := db.(*GoLevelDB)
	assert.True(t, ok)
}

// countingComparer orders keys as the comparer it wraps and counts the
// comparisons goleveldb makes.
type countingComparer struct {
	comparer.Comparer
	calls atomic.Int64
}

func (c *countingComparer) Compare(a, b []byte) int {
	c.calls.Add(1)
	return c.Comparer.Compare(a, b)
}

func TestGoLevelDBIteratorsStayInRange(t *testing.T) {
	t.Parallel()

	const (
		deleted = 10_000
		// Iterating [start, end) compares keys near its few in-range keys and
		// none of the deleted run after end.
		maxCompares = 1_000
	)
	dir := t.TempDir()
	counter := &countingComparer{Comparer: comparer.DefaultComparer}
	opts := &opt.Options{Comparer: counter}
	ldb, err := NewGoLevelDBWithOpts("range", dir, opts)
	require.NoError(t, err)

	start, end := []byte("b"), []byte("c")
	inRange := [][]byte{[]byte("b1"), []byte("b2"), []byte("b3")}
	for _, k := range inRange {
		require.NoError(t, ldb.Set(k, k))
	}
	batch := ldb.NewBatch()
	for i := range deleted {
		k := fmt.Appendf(nil, "c%05d", i)
		require.NoError(t, batch.Set(k, k))
		require.NoError(t, batch.Delete(k))
	}
	require.NoError(t, batch.Set([]byte("d"), []byte("d")))
	require.NoError(t, batch.Write())
	require.NoError(t, batch.Close())
	// Reopening flushes the memdb into a table, where pruned keys sit as
	// tombstones until compaction drops them.
	require.NoError(t, ldb.Close())
	ldb, err = NewGoLevelDBWithOpts("range", dir, opts)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, ldb.Close()) })

	cases := map[string]struct {
		open func(start, end []byte) (db.Iterator, error)
		want [][]byte
	}{
		"forward": {open: ldb.Iterator, want: inRange},
		"reverse": {open: ldb.ReverseIterator, want: [][]byte{inRange[2], inRange[1], inRange[0]}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			before := counter.calls.Load()
			itr, err := tc.open(start, end)
			require.NoError(t, err)
			var got [][]byte
			for ; itr.Valid(); itr.Next() {
				got = append(got, itr.Key())
			}
			compares := counter.calls.Load() - before
			require.NoError(t, itr.Error())
			require.NoError(t, itr.Close())
			require.Equal(t, tc.want, got)
			require.LessOrEqual(t, compares, int64(maxCompares), "iterator stepped over keys outside [start, end)")
		})
	}
}

func TestGoLevelDBInvertedRangeIsEmpty(t *testing.T) {
	t.Parallel()

	// Small tables spread the keys over several files of one level, which
	// goleveldb selects from the range bounds.
	ldb, err := NewGoLevelDBWithOpts("inverted", t.TempDir(), &opt.Options{CompactionTableSize: 1 << 10})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, ldb.Close()) })
	for i := range 1_000 {
		k := fmt.Appendf(nil, "k%04d", i)
		require.NoError(t, ldb.Set(k, k))
	}
	require.NoError(t, ldb.DB().CompactRange(util.Range{}))
	files, err := ldb.DB().GetProperty("leveldb.num-files-at-level1")
	require.NoError(t, err)
	n, err := strconv.Atoi(files)
	require.NoError(t, err)
	require.Greater(t, n, 1, "keys must span several files of level 1")

	start, end := []byte("k0900"), []byte("k0100")
	openers := map[string]func(start, end []byte) (db.Iterator, error){
		"forward": ldb.Iterator,
		"reverse": ldb.ReverseIterator,
	}
	for name, open := range openers {
		t.Run(name, func(t *testing.T) {
			itr, err := open(start, end)
			require.NoError(t, err)
			require.False(t, itr.Valid())
			require.NoError(t, itr.Close())
		})
	}
}

func BenchmarkGoLevelDBRandomReadsWrites(b *testing.B) {
	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	db, err := NewGoLevelDB(name, b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	internal.BenchmarkRandomReadsWrites(b, db)
}

func BenchmarkGoLevelDBBatchWrites(b *testing.B) {
	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	db, err := NewGoLevelDB(name, b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	internal.BenchmarkBatchWrites(b, db)
}

func BenchmarkGoLevelDBIterator(b *testing.B) {
	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	db, err := NewGoLevelDB(name, b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	internal.BenchmarkIterator(b, db)
}
