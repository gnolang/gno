package goleveldb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb/opt"

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
