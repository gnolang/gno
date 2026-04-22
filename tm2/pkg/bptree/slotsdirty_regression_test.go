package bptree

import (
	"bytes"
	"testing"
)

// Regression tests for two related cache-coherence bugs surfaced during
// the PR #5571 review pass:
//
//   1. The leaf-shift loops in insert/remove/redistribute/merge moved
//      keys / valueHashes / valueKeys / inlineValues / slotHashes but
//      not the slotsDirty bitmap, so dirty bits stayed at pre-shift
//      indices and rebuildMiniMerkleIncremental trusted stale slot
//      hashes for the now-different slot.
//
//   2. InnerNode.Clone() omitted the new miniTreeDirty field, so a
//      clone of a dirty inner answered Hash() from the pre-mutation
//      mini-merkle heap (clone got the Go zero value `false`,
//      ensureMiniMerkleBuilt skipped the rebuild).
//
// Both bugs are silent — they corrupt the merkle root without
// returning an error or panicking. Each has a deterministic
// reproducer below. Together they pin the dirty-state contract: any
// future change to leaf-shift discipline or to InnerNode.Clone must
// keep these passing.

// TestRegression_SlotsDirtyShiftedOnInsert covers bug (1). Empty tree
// → Set(K20), Set(K30), Set(K10) leaves the second insert's dirty bit
// at index 1; the third insert's shift moves the K30 data to index 2
// without shifting the dirty bit, and the next Hash() call's
// incremental rebuild then trusts an uninitialised slotHashes[2] for
// the now-K30 slot. The fix shifts slotsDirty in parallel with the
// slot data (insert.go shiftSlotsDirtyUp).
//
// The test asserts that WorkingHash() agrees with the same three keys
// committed in sorted order — the only way for the per-slot cache and
// the rebuild path to agree is if the shift discipline holds.
func TestRegression_SlotsDirtyShiftedOnInsert(t *testing.T) {
	mkTree := func() *MutableTree { return NewMutableTreeMem() }

	// Reverse-then-forward: triggers the shift bug in the reproducer order.
	a := mkTree()
	if _, err := a.Set([]byte("K20"), []byte("v20")); err != nil {
		t.Fatalf("Set K20: %v", err)
	}
	if _, err := a.Set([]byte("K30"), []byte("v30")); err != nil {
		t.Fatalf("Set K30: %v", err)
	}
	if _, err := a.Set([]byte("K10"), []byte("v10")); err != nil {
		t.Fatalf("Set K10: %v", err)
	}
	hashShifted := a.WorkingHash()

	// Sorted order: never triggers a shift, so per-slot cache stays
	// trivially consistent — this is the reference hash.
	b := mkTree()
	if _, err := b.Set([]byte("K10"), []byte("v10")); err != nil {
		t.Fatalf("Set K10: %v", err)
	}
	if _, err := b.Set([]byte("K20"), []byte("v20")); err != nil {
		t.Fatalf("Set K20: %v", err)
	}
	if _, err := b.Set([]byte("K30"), []byte("v30")); err != nil {
		t.Fatalf("Set K30: %v", err)
	}
	hashSorted := b.WorkingHash()

	if !bytes.Equal(hashShifted, hashSorted) {
		t.Fatalf("hash mismatch: insert order matters\n  shifted: %x\n  sorted:  %x", hashShifted, hashSorted)
	}
}

// TestRegression_InnerCloneCopiesMiniTreeDirty covers bug (2). The
// reproducer constructs a tree large enough to have an inner node,
// mutates it (marking it dirty), clones, and asserts the clone reports
// the same Hash() as the original. The fix adds
// `miniTreeDirty: n.miniTreeDirty` to the explicit struct literal in
// InnerNode.Clone (node.go).
func TestRegression_InnerCloneCopiesMiniTreeDirty(t *testing.T) {
	tree := NewMutableTreeMem()
	// Force at least one inner-level split: B = 32 keys per leaf, so
	// inserting 64 keys guarantees an inner node above two leaves.
	for i := range 64 {
		k := []byte{0x00, byte(i)}
		v := []byte{byte(i)}
		if _, err := tree.Set(k, v); err != nil {
			t.Fatalf("Set %d: %v", i, err)
		}
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("SaveVersion: %v", err)
	}

	// Mutate to make the root inner node dirty without rebuilding.
	if _, err := tree.Set([]byte{0x00, 0x40}, []byte{0xff}); err != nil {
		t.Fatalf("Set marker: %v", err)
	}

	// Find the root inner node and clone it. The clone must carry the
	// dirty flag; otherwise its Hash() reads the stale miniTree heap
	// and disagrees with the original's Hash() (which rebuilds via
	// ensureMiniMerkleBuilt).
	rootInner, ok := tree.root.(*InnerNode)
	if !ok {
		t.Fatalf("expected inner root after 64 inserts, got %T", tree.root)
	}
	if !rootInner.miniTreeDirty {
		t.Skip("root not dirty after Set — test setup assumption broken; investigate")
	}
	clone := rootInner.Clone()

	origHash := rootInner.Hash()
	cloneHash := clone.Hash()

	if !bytes.Equal(origHash[:], cloneHash[:]) {
		t.Fatalf("clone Hash != original Hash\n  orig:  %x\n  clone: %x", origHash, cloneHash)
	}
}
