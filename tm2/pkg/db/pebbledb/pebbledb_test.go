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

func TestPebbleSnapshotHas(t *testing.T) {
	t.Parallel()

	database, err := NewPebbleDB("has", t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, database.Close()) })

	require.NoError(t, database.Set(nil, nil))
	snapshot, err := database.NewSnapshot()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, snapshot.Close()) })
	require.NoError(t, database.Delete(nil))

	found, err := database.Has(nil)
	require.NoError(t, err)
	require.False(t, found)
	found, err = snapshot.Has(nil)
	require.NoError(t, err)
	require.True(t, found)
	found, err = snapshot.Has([]byte("absent"))
	require.NoError(t, err)
	require.False(t, found)
}

func BenchmarkPebbleHas(b *testing.B) {
	database, err := NewPebbleDB("has", b.TempDir())
	require.NoError(b, err)
	b.Cleanup(func() { require.NoError(b, database.Close()) })

	key := []byte("present")
	require.NoError(b, database.Set(key, make([]byte, 1024)))
	require.NoError(b, database.db.Flush())
	snapshot, err := database.NewSnapshot()
	require.NoError(b, err)
	b.Cleanup(func() { require.NoError(b, snapshot.Close()) })

	for _, tc := range []struct {
		name string
		has  func([]byte) (bool, error)
	}{
		{"db", database.Has},
		{"snapshot", snapshot.Has},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				found, err := tc.has(key)
				if err != nil || !found {
					b.Fatalf("Has(%q) = %v, %v", key, found, err)
				}
			}
		})
	}
}
