//go:build pebbledb

package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPebbleDB(t *testing.T) {
	name := fmt.Sprintf("test_%x", randStr(12))
	db, err := NewPebbleDB(name, t.TempDir())
	assert.Nil(t, err)
	db.Close()
}

func BenchmarkPebbleRandomReadsWrites(b *testing.B) {
	name := fmt.Sprintf("test_%x", randStr(12))
	db, err := NewPebbleDB(name, b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	benchmarkRandomReadsWrites(b, db)
}
