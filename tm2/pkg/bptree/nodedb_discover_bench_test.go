package bptree

import (
	"fmt"
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
// automatic compactions off so tombstones outlive the run.
func benchPebbleOptions() *pebble.Options {
	return &pebble.Options{Logger: quietPebbleLogger{}, DisableAutomaticCompactions: true}
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
	require.NoError(b, db.Close())

	raw, err := pebble.Open(filepath.Join(dir, name+".db"), benchPebbleOptions())
	require.NoError(b, err)
	if pruned > 0 {
		del := raw.NewBatch()
		for v := int64(1); v <= pruned; v++ {
			require.NoError(b, del.Delete(rootDBKey(v), nil))
		}
		require.NoError(b, del.Commit(pebble.Sync))
		require.NoError(b, del.Close())
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
			db := buildVersionRootsDB(b, c.n, c.pruned)
			for b.Loop() {
				snap, err := db.NewSnapshot()
				require.NoError(b, err)
				tree := NewMutableTreeWithDB(dbm.NewSnapshotDB(snap), 10000, NewNopLogger(), FastIndexOption(true))
				version, err := tree.LoadReadonly()
				require.NoError(b, err)
				require.Equal(b, c.n, version)
				require.NoError(b, snap.Close())
			}
		})
	}
}
