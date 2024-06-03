//go:build cgo

package rocksdb

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGRocksDBBackend(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	db, err := db.NewDB(name, db.RocksDBBackend, t.TempDir())
	require.NoError(t, err)

	_, ok := db.(*RocksDB)
	assert.True(t, ok)
}

func TestGRocksDBStats(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("test_%x", internal.RandStr(12))
	db, err := db.NewDB(name, db.RocksDBBackend, t.TempDir())
	require.NoError(t, err)

	assert.NotEmpty(t, db.Stats())
}
