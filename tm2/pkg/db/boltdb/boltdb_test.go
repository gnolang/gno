package boltdb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/db/internal"
)

func TestBoltDBNew(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("test_%x", internal.RandStr(12))

	db, err := New(name, t.TempDir())
	require.NoError(t, err)
	require.NoError(t, db.Print())
	require.NoError(t, db.Close())
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
