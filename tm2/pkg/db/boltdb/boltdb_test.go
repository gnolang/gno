package boltdb

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/internal"
	"github.com/stretchr/testify/require"
)

func TestBoltDBNew(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("test_%x", internal.RandStr(12))

	db, err := New(name, t.TempDir())
	require.NoError(t, err)
	db.Close()
}

func BenchmarkBoltDBRandomReadsWrites(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping testing in short mode")
	}

	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	db, err := New(name, b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	internal.BenchmarkRandomReadsWrites(b, db)
}
