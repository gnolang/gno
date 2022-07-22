package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func TestGoLevelDBNewGoLevelDB(t *testing.T) {
	dir := t.TempDir()
	name := fmt.Sprintf("test_%x", randStr(12))

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

func BenchmarkGoLevelDBRandomReadsWrites(b *testing.B) {
	name := fmt.Sprintf("test_%x", randStr(12))
	db, err := NewGoLevelDB(name, b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	benchmarkRandomReadsWrites(b, db)
}
