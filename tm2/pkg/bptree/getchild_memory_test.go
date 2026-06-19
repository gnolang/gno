package bptree

import (
	"math/rand"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// countPinned counts the in-memory nodes reachable from n via childNodes
// pointers only (it does NOT load from the DB). After SaveVersion the working
// tree should collapse to just the root; with the old read-memoization, reads
// re-grew this graph without bound (the OOM mechanism).
func countPinned(n Node) int {
	inner, ok := n.(*InnerNode)
	if !ok {
		return 1
	}
	total := 1
	for i := 0; i < inner.NumChildren(); i++ {
		if c := inner.childNodes[i]; c != nil {
			total += countPinned(c)
		}
	}
	return total
}

// assertReloadable walks the whole tree via getChild (loading children from the
// cache/DB) and asserts every inner node carries its ndb and every child loads.
// Without the SaveNode ndb assignment, an in-memory-built saved node (e.g. the
// root) has nil ndb and cannot reload its children after clear-on-save.
func assertReloadable(t *testing.T, n Node) {
	t.Helper()
	inner, ok := n.(*InnerNode)
	if !ok {
		return
	}
	if inner.ndb == nil {
		t.Fatalf("reachable InnerNode (height=%d) has nil ndb; getChild cannot lazy-load its children", inner.height)
	}
	for i := 0; i < inner.NumChildren(); i++ {
		child, err := inner.getChild(i)
		if err != nil {
			t.Fatalf("getChild(%d) on InnerNode height=%d: %v", i, inner.height, err)
		}
		if child == nil {
			t.Fatalf("getChild(%d) returned nil on InnerNode height=%d", i, inner.height)
		}
		assertReloadable(t, child)
	}
}

// TestGetChild_WorkingTreeBoundedAfterSave verifies clear-on-save: the working
// tree drops to a single in-memory node after SaveVersion, and read traffic on
// a snapshot does not re-grow the in-memory graph (reads no longer memoize).
func TestGetChild_WorkingTreeBoundedAfterSave(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 200_000, NewNopLogger())
	const n = 20_000
	for i := 0; i < n; i++ {
		if _, err := tree.Set(i2b(i), i2b(i)); err != nil {
			t.Fatal(err)
		}
	}
	_, version, err := tree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}
	if c := countPinned(tree.root); c != 1 {
		t.Fatalf("after SaveVersion the working tree pins %d in-memory nodes, want 1 (root only)", c)
	}

	// Reads on a fresh snapshot must not re-pin nodes on the snapshot's root.
	imm, err := tree.GetImmutable(version)
	if err != nil {
		t.Fatal(err)
	}
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < n; i++ {
		if _, err := imm.Has(i2b(rng.Intn(n))); err != nil {
			t.Fatal(err)
		}
	}
	if c := countPinned(imm.root); c != 1 {
		t.Fatalf("after %d reads the snapshot pins %d in-memory nodes, want 1 (reads must not memoize)", n, c)
	}
}

// TestGetChild_NoReloadSetSave_NdbInvariant exercises change #3 (SaveNode sets
// ndb): build a tree whose root is built in-memory via splits, save it WITHOUT
// reloading, and confirm every saved node can still lazy-load its children and
// that further mutation + reads work. Without change #3 the second Set panics
// with "inner node has nil child".
func TestGetChild_NoReloadSetSave_NdbInvariant(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 100, NewNopLogger())
	const n = 5_000
	for i := 0; i < n; i++ {
		if _, err := tree.Set(i2b(i), i2b(i)); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}
	// Root is in-memory-built then saved (no reload): every reachable inner must
	// carry ndb so getChild can reload after clear-on-save.
	assertReloadable(t, tree.root)

	// Mutate the same (un-reloaded) tree and save again — the path that panics
	// without change #3.
	for i := n; i < n+2_000; i++ {
		if _, err := tree.Set(i2b(i), i2b(i)); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < n+2_000; i++ {
		got, err := tree.Get(i2b(i))
		if err != nil {
			t.Fatal(err)
		}
		if got == nil {
			t.Fatalf("key %d missing after no-reload mutate+save", i)
		}
	}
}

// TestGetChild_NoReloadRemoveMerge exercises the remove/merge path on a saved,
// un-reloaded tree (in-memory-built merge nodes also need ndb).
func TestGetChild_NoReloadRemoveMerge(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 100, NewNopLogger())
	const n = 5_000
	for i := 0; i < n; i++ {
		if _, err := tree.Set(i2b(i), i2b(i)); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}
	// Remove a large contiguous range to force merges, without reloading.
	for i := 1_000; i < 4_000; i++ {
		if _, _, err := tree.Remove(i2b(i)); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < n; i++ {
		got, err := tree.Get(i2b(i))
		if err != nil {
			t.Fatal(err)
		}
		removed := i >= 1_000 && i < 4_000
		switch {
		case removed && got != nil:
			t.Fatalf("key %d should be removed", i)
		case !removed && got == nil:
			t.Fatalf("key %d should survive", i)
		}
	}
}
