package bptree

import (
	"fmt"
	"math"
	"path/filepath"
	"testing"

	"github.com/cockroachdb/pebble"
	"github.com/stretchr/testify/require"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/pebbledb"
)

// quietPebbleLogger drops pebble's info lines, which would otherwise split
// benchmark result lines, and panics on a fatal one.
type quietPebbleLogger struct{}

func (quietPebbleLogger) Infof(string, ...any) {}

func (quietPebbleLogger) Fatalf(format string, args ...any) {
	panic(fmt.Sprintf(format, args...))
}

// benchPebbleOptions are the node's empty options plus a quiet logger, and
// automatic compactions off so tombstones outlive the run. With compactions
// off, flushed L0 files only accumulate, so the L0 write stop is lifted; at
// its default, a build that flushes often stalls forever.
func benchPebbleOptions() *pebble.Options {
	return &pebble.Options{
		Logger:                      quietPebbleLogger{},
		DisableAutomaticCompactions: true,
		L0StopWritesThreshold:       math.MaxInt,
	}
}

// buildVersionRootsDB returns a flushed pebble DB holding one saved tree
// version plus root records for versions 2..n, with the roots of versions
// 1..pruned deleted.
func buildVersionRootsDB(b *testing.B, n, pruned int64) *pebbledb.PebbleDB {
	b.Helper()
	const (
		name      = "discover"
		batchSize = 50_000
	)
	dir := b.TempDir()

	db, err := pebbledb.NewPebbleDBWithOpts(name, dir, benchPebbleOptions())
	require.NoError(b, err)
	tree := NewMutableTreeWithDB(db, 10000, NewNopLogger(), FastIndexOption(true))
	_, err = tree.Set([]byte("k"), []byte("v"))
	require.NoError(b, err)
	_, _, err = tree.SaveVersion()
	require.NoError(b, err)
	root, err := db.Get(rootDBKey(1))
	require.NoError(b, err)
	require.NotEmpty(b, root)

	for from := int64(2); from <= n; from += batchSize {
		batch := db.NewBatch()
		for v := from; v < from+batchSize && v <= n; v++ {
			require.NoError(b, batch.Set(rootDBKey(v), root))
		}
		require.NoError(b, batch.Write())
		require.NoError(b, batch.Close())
	}

	return reopenFlushed(b, db, dir, name, func(raw *pebble.DB) {
		if pruned > 0 {
			del := raw.NewBatch()
			for v := int64(1); v <= pruned; v++ {
				require.NoError(b, del.Delete(rootDBKey(v), nil))
			}
			require.NoError(b, del.Commit(pebble.Sync))
			require.NoError(b, del.Close())
		}
	})
}

// buildPrunedValuesDB returns a flushed pebble DB holding the given number of
// tree versions, each overwriting the same keys, with every version but the
// latest pruned, so the value tombstones pruning writes sit in an SST.
func buildPrunedValuesDB(b *testing.B, versions, keys int) *pebbledb.PebbleDB {
	b.Helper()
	const name = "discover"
	dir := b.TempDir()

	db, err := pebbledb.NewPebbleDBWithOpts(name, dir, benchPebbleOptions())
	require.NoError(b, err)
	tree := NewMutableTreeWithDB(db, 10000, NewNopLogger(), FastIndexOption(true))
	for v := 1; v <= versions; v++ {
		for k := range keys {
			_, err := tree.Set(fmt.Appendf(nil, "key%05d", k), fmt.Appendf(nil, "v%d", v))
			require.NoError(b, err)
		}
		_, _, err := tree.SaveVersion()
		require.NoError(b, err)
	}
	require.NoError(b, tree.PruneVersionsTo(int64(versions-1)))
	return reopenFlushed(b, db, dir, name, nil)
}

// reopenFlushed closes db, runs edit on the raw pebble DB under it when edit
// is non-nil, flushes the memtable to an SST and reopens db, closing it at
// cleanup.
func reopenFlushed(b *testing.B, db *pebbledb.PebbleDB, dir, name string, edit func(raw *pebble.DB)) *pebbledb.PebbleDB {
	b.Helper()
	require.NoError(b, db.Close())

	raw, err := pebble.Open(filepath.Join(dir, name+".db"), benchPebbleOptions())
	require.NoError(b, err)
	if edit != nil {
		edit(raw)
	}
	require.NoError(b, raw.Flush())
	require.NoError(b, raw.Close())

	db, err = pebbledb.NewPebbleDBWithOpts(name, dir, benchPebbleOptions())
	require.NoError(b, err)
	b.Cleanup(func() { require.NoError(b, db.Close()) })
	return db
}

func BenchmarkLoadReadonlyVersionRoots(b *testing.B) {
	cases := []struct{ n, pruned int64 }{
		{1, 0},
		{10_000, 0},
		{100_000, 0},
		{500_000, 0},
		{500_000, 400_000},
	}
	for _, c := range cases {
		name := fmt.Sprintf("roots=%d", c.n)
		if c.pruned > 0 {
			name += fmt.Sprintf("/pruned=%d", c.pruned)
		}
		b.Run(name, func(b *testing.B) {
			benchLoadReadonly(b, buildVersionRootsDB(b, c.n, c.pruned), c.n)
		})
	}
	b.Run("values=1000/versions=1000/pruned=999", func(b *testing.B) {
		benchLoadReadonly(b, buildPrunedValuesDB(b, 1000, 1000), 1000)
	})
}

// benchLoadReadonly times a query view's load of db, as a query builds it.
func benchLoadReadonly(b *testing.B, db *pebbledb.PebbleDB, want int64) {
	b.Helper()
	for b.Loop() {
		snap, err := db.NewSnapshot()
		require.NoError(b, err)
		tree := NewMutableTreeWithDB(dbm.NewSnapshotDB(snap), 10000, NewNopLogger(), FastIndexOption(true))
		version, err := tree.LoadReadonly()
		require.NoError(b, err)
		require.Equal(b, want, version)
		require.NoError(b, snap.Close())
	}
}
