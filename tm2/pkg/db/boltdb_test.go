//go:build boltdb

package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBoltDBNewBoltDB(t *testing.T) {
	name := fmt.Sprintf("test_%x", randStr(12))

	db, err := NewBoltDB(name, t.TempDir())
	require.NoError(t, err)
	db.Close()
}

func BenchmarkBoltDBRandomReadsWrites(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping testing in short mode")
	}

	name := fmt.Sprintf("test_%x", randStr(12))
	db, err := NewBoltDB(name, b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	benchmarkRandomReadsWrites(b, db)
}
