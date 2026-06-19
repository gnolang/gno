package bptree

import (
	"fmt"
	"sync/atomic"
	"testing"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// countingDB wraps a DB and counts Get calls keyed by the first byte of the
// requested key (the prefix: 'B' for nodes, 'V' for values, 'R' for roots,
// 'O' for orphans, 'M' for meta).
type countingDB struct {
	dbm.DB
	nodeGets uint64
}

func (c *countingDB) Get(key []byte) ([]byte, error) {
	if len(key) > 0 && key[0] == PrefixNode {
		atomic.AddUint64(&c.nodeGets, 1)
	}
	return c.DB.Get(key)
}

func (c *countingDB) nodeGetCount() uint64 { return atomic.LoadUint64(&c.nodeGets) }

// TestSaveVersion_DoesNotForceLoadSiblings verifies that a Set on a tree with
// many unloaded siblings does not cause SaveVersion to load those siblings
// from the DB.
//
// Without the fix, saveNode called inner.getChild(i) for every i in every COW'd
// inner node, which eagerly loaded every sibling. With ~30 unloaded siblings per
// COW'd inner and a path length of ~2, that's ~60 extra DB reads per Set.
// The fix iterates only in-memory childNodes, so no sibling loads occur.
func TestSaveVersion_DoesNotForceLoadSiblings(t *testing.T) {
	cdb := &countingDB{DB: memdb.NewMemDB()}
	// Small cache so sibling loads actually hit the DB, not just the LRU.
	tree := NewMutableTreeWithDB(cdb, 2, NewNopLogger())

	// Fill enough keys to get a height>=1 tree with many leaves.
	for i := 0; i < 500; i++ {
		if _, err := tree.Set(fmt.Appendf(nil, "k%06d", i), []byte("v")); err != nil {
			t.Fatalf("Set: %v", err)
		}
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("SaveVersion v1: %v", err)
	}

	// Reload into a fresh tree so all children are unloaded.
	tree2 := NewMutableTreeWithDB(cdb, 2, NewNopLogger())
	if _, err := tree2.LoadVersion(1); err != nil {
		t.Fatalf("LoadVersion: %v", err)
	}
	if tree2.Height() < 1 {
		t.Fatalf("tree too shallow for sibling test: height=%d", tree2.Height())
	}

	// Do a single Set that loads ONE path (root→leaf). The COW path ends up
	// in-memory; siblings along it are still serialized-only.
	beforeGet := cdb.nodeGetCount()
	if _, err := tree2.Set([]byte("k000100"), []byte("updated")); err != nil {
		t.Fatalf("Set: %v", err)
	}
	afterSet := cdb.nodeGetCount()
	setReads := afterSet - beforeGet

	// Now SaveVersion. The fix's claim: no more sibling loads happen here.
	if _, _, err := tree2.SaveVersion(); err != nil {
		t.Fatalf("SaveVersion v2: %v", err)
	}
	saveReads := cdb.nodeGetCount() - afterSet

	// Sanity: the Set itself loaded the path from root to leaf (1 leaf + inner
	// nodes already root). Should be small.
	if setReads > uint64(tree2.Height()+2) {
		t.Fatalf("Set loaded %d nodes, expected <= %d (path length)", setReads, tree2.Height()+2)
	}

	// The fix's actual assertion: SaveVersion must NOT trigger sibling loads.
	// Without the fix, saveNode would getChild(i) for every i in every COW'd
	// inner, loading ~(B-1) = 31 siblings per level.
	if saveReads != 0 {
		t.Fatalf("SaveVersion loaded %d sibling nodes, want 0", saveReads)
	}
}
