package bptree

import (
	"errors"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestGetNode_ReturnsErrNodeNotFound verifies that nodeDB.GetNode returns
// the ErrNodeNotFound sentinel when the underlying DB has no record for
// the given NodeKey. This is the only error GetNode surfaces to callers;
// every other failure panics. See Finding #5.
func TestGetNode_ReturnsErrNodeNotFound(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 100, NewNopLogger())

	// Fabricate a NodeKey that could not have been saved: version 42,
	// nonce 9999 on a fresh tree.
	nk := &NodeKey{Version: 42, Nonce: 9999}
	nkBytes := nk.GetKey()

	node, err := tree.ndb.GetNode(nkBytes)
	if node != nil {
		t.Fatalf("GetNode on missing node returned non-nil node")
	}
	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("GetNode on missing node: got %v, want ErrNodeNotFound", err)
	}
}

// TestGetNode_CallersDetectNotFound documents the idiomatic callsite
// pattern: use errors.Is to distinguish the missing-node signal from
// any other error chain.
func TestGetNode_CallersDetectNotFound(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 100, NewNopLogger())

	nk := &NodeKey{Version: 1, Nonce: 1}
	_, err := tree.ndb.GetNode(nk.GetKey())
	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("callers expected to detect ErrNodeNotFound via errors.Is; got %v", err)
	}
}

// TestGetImmutable_PropagatesNotFoundFromGetNode verifies that when
// GetImmutable's underlying loadNode receives ErrNodeNotFound it
// surfaces the error (wrapped or otherwise) rather than succeeding
// with an invalid snapshot. This is the end-to-end guarantee of the
// Finding #5 redesign on the GetImmutable path.
func TestGetImmutable_PropagatesNotFoundFromGetNode(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	// Build two versions.
	tree.Set([]byte("k1"), []byte("v1"))
	tree.SaveVersion()
	tree.Set([]byte("k2"), []byte("v2"))
	tree.SaveVersion()

	// Prune v1 legitimately.
	if err := tree.PruneVersionsTo(1); err != nil {
		t.Fatalf("PruneVersionsTo(1): %v", err)
	}
	// v1 is now gone. GetImmutable must return a real error, not panic.
	imm, err := tree.GetImmutable(1)
	if err == nil {
		imm.Close()
		t.Fatalf("GetImmutable(1) after prune should have failed")
	}
}
