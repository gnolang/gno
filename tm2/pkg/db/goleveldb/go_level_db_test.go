package goleveldb

import (
	"fmt"
	"slices"
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

// newPrunedGoLevelDB opens a GoLevelDB with opts and writes a PrunedRange to it,
// leaving the deleted keys as tombstones in a table.
func newPrunedGoLevelDB(tb testing.TB, opts *opt.Options, deleted int) (*GoLevelDB, internal.PrunedRange) {
	tb.Helper()

	dir := tb.TempDir()
	ldb, err := NewGoLevelDBWithOpts("pruned", dir, opts)
	require.NoError(tb, err)
	r := internal.WritePrunedRange(tb, ldb, deleted)
	// Reopening flushes the memdb into a table.
	require.NoError(tb, ldb.Close())
	ldb, err = NewGoLevelDBWithOpts("pruned", dir, opts)
	require.NoError(tb, err)
	tb.Cleanup(func() { require.NoError(tb, ldb.Close()) })
	return ldb, r
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
	// Not parallel: each subtest reads its delta of one shared counter.

	// Iterating [start, end) compares keys near its few in-range keys and
	// none of the deleted run after end.
	const maxCompares = 1_000
	counter := &countingComparer{Comparer: comparer.DefaultComparer}
	ldb, r := newPrunedGoLevelDB(t, &opt.Options{Comparer: counter}, 10_000)

	reversed := slices.Clone(r.Live)
	slices.Reverse(reversed)
	cases := map[string]struct {
		open func(start, end []byte) (db.Iterator, error)
		want [][]byte
	}{
		"forward": {open: ldb.Iterator, want: r.Live},
		"reverse": {open: ldb.ReverseIterator, want: reversed},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			before := counter.calls.Load()
			itr, err := tc.open(r.Start, r.End)
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

	// Small tables spread the keys over several files of one level, and goleveldb
	// selects those files from the range bounds.
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
			t.Parallel()

			itr, err := open(start, end)
			require.NoError(t, err)
			require.False(t, itr.Valid())
			require.NoError(t, itr.Close())
		})
	}
}

func BenchmarkGoLevelDBPrunedRange(b *testing.B) {
	internal.BenchmarkPrunedRange(b, func(tb testing.TB, deleted int) (db.DB, internal.PrunedRange) {
		tb.Helper()
		return newPrunedGoLevelDB(tb, nil, deleted)
	})
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
