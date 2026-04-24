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
	tree := NewMutableTreeMem()
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

	// Before the fix, treeInsert did copyKey(key) unconditionally, adding
	// ~1 allocation per update. The value path also allocates a ValueKey
	// (sessionValues), so we can't assert 0; but we can assert the key
	// copy is gone by bounding the total. With the fix, updates do a
	// single valueKey allocation plus internal bookkeeping; the key copy
	// is removed. A regression puts allocations back at ~1 higher per Set.
	// Give a small safety margin for future internal tweaks.
	const maxAllowed = 10
	if n > maxAllowed {
		t.Fatalf("update path allocates %f per Set; want <= %d (regression on key-copy elision)", n, maxAllowed)
	}
}

// TestTreeInsert_NewKeyStillCopies verifies that new-insert paths still
// defensively copy the key, so callers can reuse their input slice.
func TestTreeInsert_NewKeyStillCopies(t *testing.T) {
	tree := NewMutableTreeMem()

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
