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

func TestPebbleReverseIteratorStaysInRange(t *testing.T) {
	t.Parallel()

	const (
		deleted = 10_000
		// A reverse iteration over [start, end) reads its few in-range keys and
		// nothing of the deleted run after end.
		maxPoints = 100
	)
	pdb, err := NewPebbleDB("reverse", t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pdb.Close()) })

	start, end := []byte("b"), []byte("c")
	inRange := [][]byte{[]byte("b1"), []byte("b2"), []byte("b3")}
	for _, k := range inRange {
		require.NoError(t, pdb.Set(k, k))
	}
	batch := pdb.NewBatch()
	for i := range deleted {
		k := fmt.Appendf(nil, "c%05d", i)
		require.NoError(t, batch.Set(k, k))
		require.NoError(t, batch.Delete(k))
	}
	require.NoError(t, batch.Set([]byte("d"), []byte("d")))
	require.NoError(t, batch.Write())
	require.NoError(t, batch.Close())
	// Tombstones in an SST, as pruned keys sit until compaction drops them.
	require.NoError(t, pdb.db.Flush())

	snap, err := pdb.NewSnapshot()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, snap.Close()) })

	openers := map[string]func(start, end []byte) (db.Iterator, error){
		"db":       pdb.ReverseIterator,
		"snapshot": snap.ReverseIterator,
	}
	for name, open := range openers {
		t.Run(name, func(t *testing.T) {
			itr, err := open(start, end)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, itr.Close()) })
			var got [][]byte
			for ; itr.Valid(); itr.Next() {
				got = append(got, itr.Key())
			}
			require.NoError(t, itr.Error())
			require.Equal(t, [][]byte{inRange[2], inRange[1], inRange[0]}, got)

			points := itr.(*pebbleDBIterator).source.Stats().InternalStats.PointCount
			require.LessOrEqual(t, points, uint64(maxPoints), "reverse iterator stepped over keys outside [start, end)")
		})
	}
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
