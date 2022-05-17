//go:build gorocksdb
// +build gorocksdb

package db

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoRocksDBBackend(t *testing.T) {
	name := fmt.Sprintf("test_%x", randStr(12))
	dir := os.TempDir()
	db := NewDB(name, GoRocksDBBackend, dir)
	defer cleanupDBDir(dir, name)

	_, ok := db.(*RocksDB)
	assert.True(t, ok)
}

func TestGoRocksDBStats(t *testing.T) {
	name := fmt.Sprintf("test_%x", randStr(12))
	dir := os.TempDir()
	db := NewDB(name, GoRocksDBBackend, dir)
	defer cleanupDBDir(dir, name)

	assert.NotEmpty(t, db.Stats())
}
