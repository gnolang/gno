package bptree

import (
	"testing"
)

// TestTreeInsert_UpdateDoesNotCopyKey verifies that an update of an
// existing key does not allocate a fresh copy of the key slice. The
// previous treeInsert took a defensive copyKey unconditionally; for
// update-heavy workloads that's one wasted 16+N-byte allocation per
// Set. We only need the copy when the key is actually stored — i.e.,
// on the new-insert and split paths in leafInsert.
//
// The update path still has other allocations (sha256.Sum256 internals
// triggered by later hashing, map/slice growth elsewhere), but the
// specific copyKey allocation is observable and worth guarding against
// a regression: if someone reintroduces the unconditional copy in
// treeInsert, the per-Set allocation count on the update path jumps
// noticeably.
func TestTreeInsert_UpdateDoesNotCopyKey(t *testing.T) {
	tree := newMemTree()
	key := []byte("fixed-key-exercising-update-path")
	// Initial insert (this one DOES copy; we're measuring updates only).
	if _, err := tree.Set(key, []byte("v0")); err != nil {
		t.Fatalf("initial Set: %v", err)
	}

	// Measure allocations during repeated updates of the same key.
	updateVal := []byte("v1")
	n := testing.AllocsPerRun(200, func() {
		_, _ = tree.Set(key, updateVal)
	})

	// Before the fix, treeInsert did copyKey(key) unconditionally, adding ~1
	// allocation per update. The value-staging path still allocates per Set
	// (value copy + batch entry), so we can't assert 0; instead we bound the
	// total to confirm the key copy is elided on the update path. A regression
	// on that elision pushes allocations back up. Small margin for tweaks.
	const maxAllowed = 10
	if n > maxAllowed {
		t.Fatalf("update path allocates %f per Set; want <= %d (regression on key-copy elision)", n, maxAllowed)
	}
}

// TestTreeInsert_NewKeyStillCopies verifies that new-insert paths still
// defensively copy the key, so callers can reuse their input slice.
func TestTreeInsert_NewKeyStillCopies(t *testing.T) {
	tree := newMemTree()

	key := []byte("k")
	if _, err := tree.Set(key, []byte("v")); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Mutate the caller's slice: the tree must have taken its own copy.
	key[0] = 'X'

	v, err := tree.Get([]byte("k"))
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(v) != "v" {
		t.Fatalf("Get after caller mutated key slice = %q, want v (the tree did not take a defensive copy)", v)
	}

	// And the mutated key is NOT in the tree.
	if has, _ := tree.Has([]byte("X")); has {
		t.Fatalf("mutated key was found in tree (should not be)")
	}
}
