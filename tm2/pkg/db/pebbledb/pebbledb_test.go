package pebbledb

import (
	"fmt"
	"testing"

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

// pebble reserves 2 x MemTableSize of the block cache for memtables. With its
// defaults (8 MiB cache, 4 MiB memtables) no block stayed cached once the
// memtables had grown, and every read missed the cache.
func TestNewPebbleDBCacheHoldsBlocksAfterWrites(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	opened, err := db.NewDB(name, db.PebbleDBBackend, t.TempDir())
	require.NoError(t, err)
	pdb := opened.(*PebbleDB)
	defer pdb.Close()

	// 64 MiB of writes: memtables grow one doubling per rotation, so they
	// reach full size well before the end.
	val := make([]byte, 1024)
	for i := range 64 << 10 {
		require.NoError(t, pdb.Set(fmt.Appendf(nil, "key-%06d", i), val))
	}
	// Read a flushed working set twice: the second pass should hit.
	for range 2 {
		for i := range 1000 {
			_, err := pdb.Get(fmt.Appendf(nil, "key-%06d", i))
			require.NoError(t, err)
		}
	}

	m := pdb.db.Metrics()
	hitRate := float64(m.BlockCache.Hits) / float64(m.BlockCache.Hits+m.BlockCache.Misses)
	assert.Greater(t, hitRate, 0.4, "block cache hits %d, misses %d", m.BlockCache.Hits, m.BlockCache.Misses)
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
