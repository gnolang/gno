//go:build grocksdb
// +build grocksdb

package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGRocksDBBackend(t *testing.T) {
	name := fmt.Sprintf("test_%x", randStr(12))
	db := NewDB(name, GRocksDBBackend, t.TempDir())

	_, ok := db.(*RocksDB)
	assert.True(t, ok)
}

func TestGRocksDBStats(t *testing.T) {
	name := fmt.Sprintf("test_%x", randStr(12))
	db := NewDB(name, GRocksDBBackend, t.TempDir())

	assert.NotEmpty(t, db.Stats())
}
