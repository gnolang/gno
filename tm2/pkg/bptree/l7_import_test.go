package bptree

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestImport_RollsBackPendingSession verifies that Import discards any
// uncommitted working-session state (pending batch values, orphan list,
// value-nonce counter, working root) before reconstructing — so stale state
// from a prior un-committed session can't leak into the import's SaveVersion
// (L7). Without the Rollback in Import, the assertions below fail (pendingVals
// holds the stale value, nextValueNonce is 1, root holds the stale leaf).
func TestImport_RollsBackPendingSession(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 1000, NewNopLogger())

	// Uncommitted working-session state.
	if _, err := tree.Set([]byte("stale"), []byte("v")); err != nil {
		t.Fatal(err)
	}
	if len(tree.ndb.pendingVals) == 0 || tree.nextValueNonce == 0 || tree.root == nil {
		t.Fatal("setup: expected pending session state after an uncommitted Set")
	}

	// Creating the Importer must roll that state back.
	if _, err := tree.Import(1); err != nil {
		t.Fatal(err)
	}

	if n := len(tree.ndb.pendingVals); n != 0 {
		t.Fatalf("Import did not drop pending values: %d remain", n)
	}
	if tree.nextValueNonce != 0 {
		t.Fatalf("Import did not reset nextValueNonce: got %d", tree.nextValueNonce)
	}
	if n := len(tree.versionOrphans); n != 0 {
		t.Fatalf("Import did not clear versionOrphans: %d remain", n)
	}
	if tree.root != nil {
		t.Fatalf("Import did not revert root to lastSaved (nil for a fresh tree)")
	}
}
