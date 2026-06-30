package bptree

import (
	"fmt"
	"testing"
)

// TestPrune_BelowFloorDoesNotRewindFirstVersion verifies that pruning to a
// version below the current floor is a no-op and does NOT rewind firstVersion
// (L1). Without the guard, PruneVersionsTo(1) after the floor advanced to 3
// would set firstVersion=2, making AvailableVersions scan versions that were
// already pruned.
func TestPrune_BelowFloorDoesNotRewindFirstVersion(t *testing.T) {
	tree := newPruneTree(t)
	for v := 1; v <= 3; v++ {
		for i := 0; i < 10; i++ {
			tree.Set(fmt.Appendf(nil, "fl%03d", i), fmt.Appendf(nil, "v%d", v))
		}
		if _, _, err := tree.SaveVersion(); err != nil {
			t.Fatal(err)
		}
	}

	if err := tree.DeleteVersionsTo(2); err != nil {
		t.Fatalf("prune to 2: %v", err)
	}
	if first := tree.ndb.getFirstVersion(); first != 3 {
		t.Fatalf("after prune to 2: firstVersion = %d, want 3", first)
	}

	// Re-pruning at or below the floor must be a no-op, not a floor rewind.
	for _, toV := range []int64{2, 1, 0, -5} {
		if err := tree.DeleteVersionsTo(toV); err != nil {
			t.Fatalf("prune to %d (below floor): %v", toV, err)
		}
		if first := tree.ndb.getFirstVersion(); first != 3 {
			t.Fatalf("prune to %d rewound firstVersion to %d, want 3", toV, first)
		}
	}

	if avail := tree.AvailableVersions(); len(avail) != 1 || avail[0] != 3 {
		t.Fatalf("AvailableVersions = %v, want [3]", avail)
	}
}
