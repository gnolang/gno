package bptree

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestSaveVersion_IdempotentLegacyEmptyBlob verifies that an empty-tree
// re-save at a version whose existing root-blob was written in the
// legacy zero-length format (nk=nil, hash=nil) is recognised as
// equivalent instead of failing with a false "hash mismatch" error.
// See Finding #26.
func TestSaveVersion_IdempotentLegacyEmptyBlob(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	// Save two empty versions (modern format includes emptyHash).
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("V1 SaveVersion: %v", err)
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("V2 SaveVersion: %v", err)
	}

	// Corrupt V2's root-blob into legacy format (zero-length, pre-hash).
	// GetRoot will return (nil, nil, nil) for this blob, which is the
	// edge case Finding #26 targets.
	if err := db.Set(rootDBKey(2), []byte{}); err != nil {
		t.Fatalf("poking legacy blob: %v", err)
	}

	// Load V1 and re-save at V2 (WorkingVersion = 2). With the fix,
	// the idempotent check recognises the legacy blob as empty and
	// succeeds; pre-fix, it would error "already exists with a
	// different hash".
	if _, err := tree.LoadVersion(1); err != nil {
		t.Fatalf("LoadVersion(1): %v", err)
	}
	_, ver, err := tree.SaveVersion()
	if err != nil {
		t.Fatalf("SaveVersion over legacy empty V2: %v", err)
	}
	if ver != 2 {
		t.Fatalf("saved version = %d, want 2", ver)
	}
}

// TestSaveVersion_IdempotentEmptyMatchesModernBlob is a regression
// test for the normal path — ensure that re-saving an empty tree
// against a modern-format empty-V blob still works (no collateral
// damage from the Finding #26 fix).
func TestSaveVersion_IdempotentEmptyMatchesModernBlob(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("V1: %v", err)
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("V2: %v", err)
	}

	if _, err := tree.LoadVersion(1); err != nil {
		t.Fatalf("LoadVersion(1): %v", err)
	}
	_, ver, err := tree.SaveVersion()
	if err != nil {
		t.Fatalf("idempotent empty re-save: %v", err)
	}
	if ver != 2 {
		t.Fatalf("version = %d, want 2", ver)
	}
}

// TestSaveVersion_MismatchStillDetected ensures the Finding #26 fix
// does not weaken the real-mismatch detection: saving a non-empty
// tree against an empty legacy blob should still fail.
func TestSaveVersion_MismatchStillDetected(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("V1: %v", err)
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("V2: %v", err)
	}

	// Corrupt V2 to legacy format.
	if err := db.Set(rootDBKey(2), []byte{}); err != nil {
		t.Fatalf("poking legacy blob: %v", err)
	}

	if _, err := tree.LoadVersion(1); err != nil {
		t.Fatalf("LoadVersion(1): %v", err)
	}
	// Now add a key — the tree is no longer empty, must mismatch.
	if _, err := tree.Set([]byte("k"), []byte("v")); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if _, _, err := tree.SaveVersion(); err == nil {
		t.Fatal("expected mismatch error; got nil")
	}
}
