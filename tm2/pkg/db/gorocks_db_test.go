//go:build gorocksdb

package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoRocksDBBackend(t *testing.T) {
	name := fmt.Sprintf("test_%x", randStr(12))
	db := NewDB(name, GoRocksDBBackend, t.TempDir())

	_, ok := db.(*RocksDB)
	assert.True(t, ok)
}

func TestGoRocksDBStats(t *testing.T) {
	name := fmt.Sprintf("test_%x", randStr(12))
	db := NewDB(name, GoRocksDBBackend, t.TempDir())

	assert.NotEmpty(t, db.Stats())
}
