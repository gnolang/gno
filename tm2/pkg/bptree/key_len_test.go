package bptree

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestSet_RejectsKeyOverMax asserts Set enforces a MaxKeyLen cap. Without
// it, a caller could write a key longer than readBytes's 1 MiB cap on the
// read path; the node would serialize successfully but fail to deserialize
// on the next Load, wedging the version permanently.
func TestSet_RejectsKeyOverMax(t *testing.T) {
	tree := newMemTree()

	// Exactly at the limit is fine.
	ok := bytes.Repeat([]byte{'a'}, MaxKeyLen)
	if _, err := tree.Set(ok, []byte("v")); err != nil {
		t.Fatalf("Set at MaxKeyLen: %v", err)
	}

	// Over the limit is rejected.
	tooBig := bytes.Repeat([]byte{'b'}, MaxKeyLen+1)
	_, err := tree.Set(tooBig, []byte("v"))
	if err == nil {
		t.Fatalf("Set over MaxKeyLen should return ErrKeyTooLong")
	}
	if !errors.Is(err, ErrKeyTooLong) {
		t.Fatalf("error should wrap ErrKeyTooLong, got %v", err)
	}

	// Tree state is unchanged by the rejected Set.
	if has, _ := tree.Has(tooBig); has {
		t.Fatalf("rejected Set should not be reflected in the tree")
	}
}

// TestImport_RejectsKeyOverMax asserts the Importer also enforces
// MaxKeyLen; an untrusted export stream must not poison a fresh tree.
func TestImport_RejectsKeyOverMax(t *testing.T) {
	tree := newMemTree()
	imp, err := tree.Import(1)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}

	err = imp.Add(&ExportNode{
		Height: 0,
		Key:    bytes.Repeat([]byte{'x'}, MaxKeyLen+1),
		Value:  []byte("v"),
	})
	if err == nil {
		t.Fatalf("Importer.Add should reject key over MaxKeyLen")
	}
}

// TestImport_RejectsSeparatorOverMax asserts the Importer enforces MaxKeyLen
// on inner separator keys too, not just leaf keys.
func TestImport_RejectsSeparatorOverMax(t *testing.T) {
	tree := newMemTree()
	imp, err := tree.Import(1)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}

	// Two single-key leaves so the inner marker's child-count check passes.
	for _, k := range []string{"a", "b"} {
		if err := imp.Add(&ExportNode{Height: 0, Key: []byte(k), Value: []byte("v")}); err != nil {
			t.Fatalf("Add leaf entry %q: %v", k, err)
		}
		if err := imp.Add(&ExportNode{Height: -1, NumKeys: 1}); err != nil {
			t.Fatalf("Add leaf marker %q: %v", k, err)
		}
	}

	err = imp.Add(&ExportNode{
		Height:        1,
		NumKeys:       1,
		SeparatorKeys: [][]byte{bytes.Repeat([]byte{'x'}, MaxKeyLen+1)},
	})
	if err == nil {
		t.Fatalf("Importer.Add should reject separator key over MaxKeyLen")
	}
	if !strings.Contains(err.Error(), "separator") {
		t.Fatalf("rejection should come from the separator check, got: %v", err)
	}
}

// TestSet_MaxKeyLenRoundTripsFromDisk pins the MaxKeyLen == maxReadBytesLen
// boundary coupling: a key written at exactly the write-side cap must decode
// through readBytes's read-side cap on a cold reload. Lowering one constant
// without the other breaks this (the version would wedge on reload).
func TestSet_MaxKeyLenRoundTripsFromDisk(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	key := bytes.Repeat([]byte{'a'}, MaxKeyLen)
	if _, err := tree.Set(key, []byte("v")); err != nil {
		t.Fatalf("Set at MaxKeyLen: %v", err)
	}
	if _, version, err := tree.SaveVersion(); err != nil {
		t.Fatalf("SaveVersion: %v", err)
	} else if version != 1 {
		t.Fatalf("SaveVersion: version = %d, want 1", version)
	}

	// Fresh handle: empty node cache, so Get must deserialize from disk.
	fresh := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	if _, err := fresh.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	got, err := fresh.Get(key)
	if err != nil {
		t.Fatalf("Get after reload: %v", err)
	}
	if !bytes.Equal(got, []byte("v")) {
		t.Fatalf("Get after reload: got %q, want %q", got, "v")
	}
}
