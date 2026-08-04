package bptree

// Helpers for ported IAVL tests.

import (
	"encoding/binary"
	"math/rand"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// getTestTree creates a DB-backed tree with `size` random keys pre-inserted.
func getTestTree(size int) *MutableTree {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	for i := range size {
		tree.Set(i2b(i), []byte{})
	}
	return tree
}

// newMemTree creates an empty in-memory (memdb-backed) tree for tests.
func newMemTree() *MutableTree {
	return NewMutableTreeWithDB(memdb.NewMemDB(), 1000, NewNopLogger())
}

// i2b converts an int to a big-endian 4-byte key.
func i2b(i int) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(i))
	return b
}

// randstr generates a random alphanumeric string of length n.
func randstr(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// traverser collects keys seen during IterateRange.
type traverser struct {
	first, last string
	count       int
}

func (t *traverser) view(key, value []byte) bool {
	if t.first == "" {
		t.first = string(key)
	}
	t.last = string(key)
	t.count++
	return false
}

func expectTraverse(t *testing.T, trav traverser, first, last string, count int) {
	t.Helper()
	if trav.count != count {
		t.Errorf("Expected %d items, got %d", count, trav.count)
	}
	if first != "" && trav.first != first {
		t.Errorf("Expected first=%q, got %q", first, trav.first)
	}
	if last != "" && trav.last != last {
		t.Errorf("Expected last=%q, got %q", last, trav.last)
	}
}
