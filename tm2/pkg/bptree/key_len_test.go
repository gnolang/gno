package bptree

import (
	"bytes"
	"errors"
	"testing"
)

// TestSet_RejectsKeyOverMax asserts Set enforces a MaxKeyLen cap. Without
// it, a caller could write a key longer than readBytes's 1 MiB cap on the
// read path; the node would serialize successfully but fail to deserialize
// on the next Load, wedging the version permanently.
func TestSet_RejectsKeyOverMax(t *testing.T) {
	tree := NewMutableTreeMem()

	// Just under the limit is fine.
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
	tree := NewMutableTreeMem()
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
