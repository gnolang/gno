package bptree

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestFastNode_GetHit verifies that after a Get, a second Get for the
// same key returns the same value without touching the tree. We prove
// this by Set → Get (populates cache) → nil'ing the root and calling
// Get again; the second Get must still succeed from cache.
func TestFastNode_GetHit(t *testing.T) {
	tree := NewMutableTreeMem()
	if _, err := tree.Set([]byte("k"), []byte("v")); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if _, err := tree.Get([]byte("k")); err != nil {
		t.Fatalf("Get: %v", err)
	}
	// Simulate a stripped root to prove the second Get is cache-only.
	savedRoot := tree.root
	tree.root = nil
	defer func() { tree.root = savedRoot }()

	v, err := tree.Get([]byte("k"))
	if err != nil {
		t.Fatalf("Get after warm cache: %v", err)
	}
	if !bytes.Equal(v, []byte("v")) {
		t.Fatalf("Get returned %q, want %q (cache miss despite prior Get)", v, "v")
	}
}

// TestFastNode_SetRefreshesCache verifies that an update to an existing
// key refreshes the cache with the new value, not the stale one.
func TestFastNode_SetRefreshesCache(t *testing.T) {
	tree := NewMutableTreeMem()
	tree.Set([]byte("k"), []byte("v1"))
	tree.Get([]byte("k")) // populate cache with v1
	tree.Set([]byte("k"), []byte("v2"))

	v, _ := tree.Get([]byte("k"))
	if !bytes.Equal(v, []byte("v2")) {
		t.Fatalf("Get after update = %q, want v2 (stale cache)", v)
	}
}

// TestFastNode_RemoveInvalidates verifies that Remove drops the entry
// so a subsequent Get falls through to the tree (which must return the
// missing-key nil).
func TestFastNode_RemoveInvalidates(t *testing.T) {
	tree := NewMutableTreeMem()
	tree.Set([]byte("k"), []byte("v"))
	tree.Get([]byte("k")) // populate
	tree.Remove([]byte("k"))

	v, err := tree.Get([]byte("k"))
	if err != nil {
		t.Fatalf("Get after remove: %v", err)
	}
	if v != nil {
		t.Fatalf("Get after remove returned %q, want nil", v)
	}
}

// TestFastNode_HasHit verifies Has short-circuits on a cached key.
func TestFastNode_HasHit(t *testing.T) {
	tree := NewMutableTreeMem()
	tree.Set([]byte("k"), []byte("v"))
	tree.Get([]byte("k")) // populate

	// Strip the root; Has must still answer true from cache.
	savedRoot := tree.root
	tree.root = nil
	defer func() { tree.root = savedRoot }()

	ok, _ := tree.Has([]byte("k"))
	if !ok {
		t.Fatalf("Has returned false on cached key")
	}
}

// TestFastNode_RollbackPurges verifies that Rollback drops the entire
// cache, so cached working-session writes cannot leak into queries
// issued after the rollback.
func TestFastNode_RollbackPurges(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())
	tree.Set([]byte("saved"), []byte("svalue"))
	tree.SaveVersion()
	tree.Get([]byte("saved")) // populate cache from saved version

	tree.Set([]byte("unsaved"), []byte("uvalue")) // populates cache via Set path
	tree.Rollback()

	if _, ok := tree.fastNodes.Get("unsaved"); ok {
		t.Fatalf("rollback left an unsaved key in the cache")
	}
	if _, ok := tree.fastNodes.Get("saved"); ok {
		t.Fatalf("rollback did not purge the saved key from the cache (expected wholesale purge)")
	}
}

// TestFastNode_LoadVersionPurges verifies that LoadVersion clears the
// cache, since the new root's value mappings may differ from the
// previously-loaded root.
func TestFastNode_LoadVersionPurges(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())
	tree.Set([]byte("k"), []byte("v1"))
	tree.SaveVersion()
	tree.Set([]byte("k"), []byte("v2"))
	tree.SaveVersion()

	tree.Get([]byte("k")) // populate cache with v2
	if _, err := tree.LoadVersion(1); err != nil {
		t.Fatalf("LoadVersion: %v", err)
	}
	if _, ok := tree.fastNodes.Get("k"); ok {
		t.Fatalf("LoadVersion did not purge the cache")
	}
	v, _ := tree.Get([]byte("k"))
	if !bytes.Equal(v, []byte("v1")) {
		t.Fatalf("Get after LoadVersion(1) = %q, want v1", v)
	}
}

// TestFastNode_DisabledByOption verifies that a negative
// FastNodeCacheSize disables the cache.
func TestFastNode_DisabledByOption(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger(), FastNodeCacheSizeOption(-1))
	if tree.fastNodes != nil {
		t.Fatalf("cache should be disabled with FastNodeCacheSizeOption(-1)")
	}
}

// TestFastNode_ExplicitSize verifies a positive FastNodeCacheSize is
// honoured.
func TestFastNode_ExplicitSize(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger(), FastNodeCacheSizeOption(5))
	for i := 0; i < 10; i++ {
		tree.Set(fmt.Appendf(nil, "k%02d", i), []byte("v"))
		tree.Get(fmt.Appendf(nil, "k%02d", i))
	}
	if n := tree.fastNodes.Len(); n > 5 {
		t.Fatalf("cache length = %d, want ≤ 5 (LRU bound)", n)
	}
}
