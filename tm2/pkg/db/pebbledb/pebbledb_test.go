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

// The registry opens every node store through NewPebbleDB, so the tuning in
// DefaultPebbleOptions only reaches production if NewPebbleDB applies it.
func TestNewPebbleDBUsesDefaultOptions(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	opened, err := db.NewDB(name, db.PebbleDBBackend, t.TempDir())
	require.NoError(t, err)
	pdb := opened.(*PebbleDB)
	defer pdb.Close()

	for i := range 1000 {
		require.NoError(t, pdb.Set(fmt.Appendf(nil, "key-%04d", i), []byte("value")))
	}
	require.NoError(t, pdb.db.Flush())

	// A point lookup of an absent key inside the table's range must be
	// answered by the bloom filter, without reading a data block.
	for i := range 100 {
		v, err := pdb.Get(fmt.Appendf(nil, "key-%04d-absent", i))
		require.NoError(t, err)
		require.Nil(t, v)
	}

	m := pdb.db.Metrics()
	assert.Positive(t, m.Filter.Hits, "absent keys should be filtered by the bloom filter")
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
