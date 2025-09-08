package memdb

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/internal"
)

func BenchmarkMemDBRandomReadsWrites(b *testing.B) {
	db := NewMemDB()
	defer db.Close()

	internal.BenchmarkRandomReadsWrites(b, db)
}

func BenchmarkMemlDBBatchWrites(b *testing.B) {
	db := NewMemDB()
	defer db.Close()

	internal.BenchmarkBatchWrites(b, db)
}

func BenchmarkMemlDBIterator(b *testing.B) {
	db := NewMemDB()
	defer db.Close()

	internal.BenchmarkIterator(b, db)
}
