package pebbledb

import (
	"fmt"
	"runtime"
	"slices"
	"testing"

	"github.com/cockroachdb/pebble"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/internal"
)

func TestPebbleDBNewGoLevelDB(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	name := fmt.Sprintf("test_%x", internal.RandStr(12))

	// Test we can't open the db twice for writing
	wr1, err := NewPebbleDB(name, dir)
	require.Nil(t, err)
	_, err = NewPebbleDB(name, dir)
	require.NotNil(t, err)
	wr1.Close() // Close the db to release the lock
}

func TestPebbleDBBackend(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	db, err := db.NewDB(name, db.PebbleDBBackend, t.TempDir())
	require.NoError(t, err)

	_, ok := db.(*PebbleDB)
	assert.True(t, ok)
}

// newPrunedPebbleDB opens a PebbleDB and writes a PrunedRange to it, leaving the
// deleted keys as tombstones in an SST.
func newPrunedPebbleDB(tb testing.TB, deleted int) (*PebbleDB, internal.PrunedRange) {
	tb.Helper()

	pdb, err := NewPebbleDB("pruned", tb.TempDir())
	require.NoError(tb, err)
	tb.Cleanup(func() { require.NoError(tb, pdb.Close()) })
	r := internal.WritePrunedRange(tb, pdb, deleted)
	require.NoError(tb, pdb.db.Flush())
	return pdb, r
}

func TestPebbleIteratorsStayInRange(t *testing.T) {
	t.Parallel()

	// An iteration over [start, end) reads its few in-range keys and nothing
	// of the deleted run after end.
	const maxPoints = 100
	pdb, r := newPrunedPebbleDB(t, 10_000)
	snap, err := pdb.NewSnapshot()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, snap.Close()) })

	reversed := slices.Clone(r.Live)
	slices.Reverse(reversed)
	cases := map[string]struct {
		open func(start, end []byte) (db.Iterator, error)
		want [][]byte
	}{
		"db/forward":       {open: pdb.Iterator, want: r.Live},
		"db/reverse":       {open: pdb.ReverseIterator, want: reversed},
		"snapshot/forward": {open: snap.Iterator, want: r.Live},
		"snapshot/reverse": {open: snap.ReverseIterator, want: reversed},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			itr, err := tc.open(r.Start, r.End)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, itr.Close()) })
			var got [][]byte
			for ; itr.Valid(); itr.Next() {
				got = append(got, itr.Key())
			}
			require.NoError(t, itr.Error())
			require.Equal(t, tc.want, got)

			points := itr.(*pebbleDBIterator).source.Stats().InternalStats.PointCount
			require.LessOrEqual(t, points, uint64(maxPoints), "iterator stepped over keys outside [start, end)")
		})
	}
}

func TestPebbleEmptyEndYieldsNothing(t *testing.T) {
	// Not parallel: pebble reads an empty UpperBound as none while its pooled
	// iterator has no bounds buffer, so each subtest empties the pool and no other
	// iterator may refill it.
	pdb, err := NewPebbleDB("emptyend", t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pdb.Close()) })
	require.NoError(t, pdb.Set([]byte("a"), []byte("a")))

	openers := map[string]func(start, end []byte) (db.Iterator, error){
		"forward": pdb.Iterator,
		"reverse": pdb.ReverseIterator,
	}
	for name, open := range openers {
		t.Run(name, func(t *testing.T) {
			runtime.GC()
			runtime.GC()
			itr, err := open(nil, []byte{})
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, itr.Close()) })
			require.False(t, itr.Valid(), "no key sorts before an empty end")
			points := itr.(*pebbleDBIterator).source.Stats().InternalStats.PointCount
			require.Zero(t, points, "an iterator with an empty end read keys")
		})
	}
}

func TestPebbleInvertedRangeIsEmpty(t *testing.T) {
	t.Parallel()

	// Small files spread the keys over several SSTs of one level, and pebble
	// selects those SSTs from the iterator bounds.
	opts := &pebble.Options{Levels: []pebble.LevelOptions{{TargetFileSize: 1 << 10}}}
	pdb, err := NewPebbleDBWithOpts("inverted", t.TempDir(), opts)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pdb.Close()) })
	for i := range 1_000 {
		k := fmt.Appendf(nil, "k%04d", i)
		require.NoError(t, pdb.Set(k, k))
	}
	require.NoError(t, pdb.db.Compact([]byte("k"), []byte("l"), false))
	var maxFiles int64
	for _, level := range pdb.db.Metrics().Levels {
		maxFiles = max(maxFiles, level.NumFiles)
	}
	require.Greater(t, maxFiles, int64(1), "keys must span several files of one level")

	snap, err := pdb.NewSnapshot()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, snap.Close()) })

	start, end := []byte("k0900"), []byte("k0100")
	openers := map[string]func(start, end []byte) (db.Iterator, error){
		"db/forward":       pdb.Iterator,
		"db/reverse":       pdb.ReverseIterator,
		"snapshot/forward": snap.Iterator,
		"snapshot/reverse": snap.ReverseIterator,
	}
	for name, open := range openers {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			itr, err := open(start, end)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, itr.Close()) })
			require.False(t, itr.Valid())
			require.NoError(t, itr.Error())
		})
	}
}

func BenchmarkPebbleDBPrunedRange(b *testing.B) {
	internal.BenchmarkPrunedRange(b, func(tb testing.TB, deleted int) (db.DB, internal.PrunedRange) {
		tb.Helper()
		return newPrunedPebbleDB(tb, deleted)
	})
}

func BenchmarkPebbleDBRandomReadsWrites(b *testing.B) {
	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	db, err := NewPebbleDB(name, b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	internal.BenchmarkRandomReadsWrites(b, db)
}

func BenchmarkPebbleDBBatchWrites(b *testing.B) {
	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	db, err := NewPebbleDB(name, b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	internal.BenchmarkBatchWrites(b, db)
}

func BenchmarkPebbleDBIterator(b *testing.B) {
	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	db, err := NewPebbleDB(name, b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	internal.BenchmarkIterator(b, db)
}
