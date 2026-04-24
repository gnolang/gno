package bptree

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestOrphans_FirstVersionEmpty locks in the invariant that orphans[first]
// is always empty (not written) for a freshly saved first version. The
// pruning logic depends on this: pruneVersion only processes orphans[nextV]
// values, not orphans[v] values — relying on orphans[first] being empty
// because the first version has no prior state to displace.
func TestOrphans_FirstVersionEmpty(t *testing.T) {
	cases := []struct {
		name           string
		initialVersion uint64
		setup          func(tree *MutableTree)
	}{
		{
			name: "default initialVersion, no sets",
			setup: func(_ *MutableTree) {},
		},
		{
			name: "default initialVersion, with sets",
			setup: func(tree *MutableTree) {
				tree.Set([]byte("a"), []byte("1"))
				tree.Set([]byte("b"), []byte("2"))
				tree.Set([]byte("c"), []byte("3"))
			},
		},
		{
			name: "default initialVersion, set+remove+set",
			setup: func(tree *MutableTree) {
				tree.Set([]byte("a"), []byte("1"))
				tree.Set([]byte("a"), []byte("2"))
				tree.Remove([]byte("a"))
				tree.Set([]byte("a"), []byte("3"))
			},
		},
		{
			name:           "initialVersion=100, with sets",
			initialVersion: 100,
			setup: func(tree *MutableTree) {
				tree.Set([]byte("a"), []byte("1"))
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db := memdb.NewMemDB()
			var tree *MutableTree
			if c.initialVersion > 0 {
				tree = NewMutableTreeWithDB(db, 100, NewNopLogger(),
					InitialVersionOption(c.initialVersion))
			} else {
				tree = NewMutableTreeWithDB(db, 100, NewNopLogger())
			}
			c.setup(tree)
			_, version, err := tree.SaveVersion()
			if err != nil {
				t.Fatalf("SaveVersion: %v", err)
			}

			orphans, err := tree.ndb.LoadOrphans(version)
			if err != nil {
				t.Fatalf("LoadOrphans(%d): %v", version, err)
			}
			if len(orphans) != 0 {
				t.Fatalf("orphans[%d] has %d entries, want 0. Pruning assumes "+
					"orphans[first] is empty since no prune consumes it; "+
					"breaking this invariant leaks values.",
					version, len(orphans))
			}
		})
	}
}

// TestPrune_ConsumesOrphansOfFirstVersion defends the pruning change that
// processes orphans[v] in addition to orphans[nextV]. We seed a non-empty
// orphans[first] record directly (simulating a future regression where the
// first-version-is-empty invariant breaks) and verify that PruneVersionsTo
// deletes those values rather than leaking them.
func TestPrune_ConsumesOrphansOfFirstVersion(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	// Create two real versions so we can prune v=1.
	tree.Set([]byte("k1"), []byte("v1"))
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("SaveVersion(1): %v", err)
	}
	tree.Set([]byte("k2"), []byte("v2"))
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("SaveVersion(2): %v", err)
	}

	// Plant a synthetic value and a synthetic orphans[1] record that points
	// at it. This represents the "some external initialization wrote
	// orphans[first]" scenario the defensive code guards against.
	planted := (&NodeKey{Version: 0, Nonce: 42}).GetKey()
	if err := tree.ndb.SaveValue([]byte("PLANTED"), planted); err != nil {
		t.Fatalf("SaveValue: %v", err)
	}
	if err := tree.ndb.SaveOrphans(1, [][]byte{planted}); err != nil {
		t.Fatalf("SaveOrphans: %v", err)
	}
	if err := tree.ndb.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	// Sanity: the planted value is present.
	if v, _ := tree.ndb.GetValue(planted); string(v) != "PLANTED" {
		t.Fatalf("setup: planted value missing, got %q", v)
	}

	// Prune v=1. The fix must consume orphans[1] (not just orphans[2]).
	if err := tree.DeleteVersionsTo(1); err != nil {
		t.Fatalf("DeleteVersionsTo(1): %v", err)
	}

	// The planted value must be gone.
	v, err := tree.ndb.GetValue(planted)
	if err == nil && v != nil {
		t.Fatalf("planted value not cleaned up by prune: still present as %q", v)
	}
}
