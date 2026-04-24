package bptree

import (
	"errors"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestGetValue_MissingValueReturnsError verifies that if the tree references
// a ValueKey that is not present in the DB (corruption or a pruning bug),
// GetValue returns a wrapped ErrKeyDoesNotExist rather than silently
// returning (nil, nil). The silent form caused corruption to look like a
// legitimate "empty value" to the caller.
func TestGetValue_MissingValueReturnsError(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	if _, err := tree.Set([]byte("k"), []byte("v")); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Capture the vk via tree lookup, then delete the value directly from DB
	// to simulate corruption.
	_, _, vk, found := treeLookup(tree.root, []byte("k"))
	if !found {
		t.Fatalf("setup: key not found")
	}
	if err := db.Delete(valueDBKey(vk)); err != nil {
		t.Fatalf("delete value: %v", err)
	}

	_, err := tree.ndb.GetValue(vk)
	if err == nil {
		t.Fatalf("GetValue on missing vk should return error")
	}
	if !errors.Is(err, ErrKeyDoesNotExist) {
		t.Fatalf("error should wrap ErrKeyDoesNotExist, got %v", err)
	}
}

// TestGetValue_EmptyValueReturnsEmptySlice verifies that a stored empty
// value is returned as a non-nil []byte{} — distinguishable from missing.
// Previously, depending on the backend, empty values could round-trip as
// nil, making them indistinguishable from "not found".
func TestGetValue_EmptyValueReturnsEmptySlice(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	if _, err := tree.Set([]byte("k"), []byte{}); err != nil {
		t.Fatalf("Set: %v", err)
	}

	val, err := tree.Get([]byte("k"))
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val == nil {
		t.Fatalf("Get on stored empty value should return non-nil, got nil")
	}
	if len(val) != 0 {
		t.Fatalf("Get on stored empty value = %q, want empty", val)
	}

	// After SaveVersion: same contract.
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("SaveVersion: %v", err)
	}
	val, err = tree.Get([]byte("k"))
	if err != nil || val == nil || len(val) != 0 {
		t.Fatalf("Get after save: val=%q err=%v", val, err)
	}
}
