//go:build grocksdb
// +build grocksdb

package db

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGRocksDBBackend(t *testing.T) {
	name := fmt.Sprintf("test_%x", randStr(12))
	dir := os.TempDir()
	db := NewDB(name, GRocksDBBackend, dir)
	defer cleanupDBDir(dir, name)

	_, ok := db.(*RocksDB)
	assert.True(t, ok)
}

func TestGRocksDBStats(t *testing.T) {
	name := fmt.Sprintf("test_%x", randStr(12))
	dir := os.TempDir()
	db := NewDB(name, GRocksDBBackend, dir)
	defer cleanupDBDir(dir, name)

	assert.NotEmpty(t, db.Stats())
}
