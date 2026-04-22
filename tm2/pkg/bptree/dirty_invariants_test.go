package bptree

import (
	"bytes"
	"testing"
)

// Tests pinning two related cache-coherence contracts in the leaf-level
// dirty-state machinery:
//
//  1. The slotsDirty bitmap tracks which slotHashes entries are stale
//     and must be recomputed by rebuildMiniMerkleIncremental. Any
//     mutation that shifts slot data (insert / remove / redistribute /
//     merge) must shift slotsDirty in parallel — otherwise dirty bits
//     stay at pre-shift indices, the rebuild trusts a stale slotHashes
//     entry for the now-different slot, and the leaf emits a corrupt
//     mini-merkle root.
//
//  2. InnerNode.Clone must copy miniTreeDirty. A clone that drops the
//     flag would answer Hash() from the pre-mutation mini-merkle heap
//     because ensureMiniMerkleBuilt skips the rebuild on a clean
//     clone — silently returning a stale root.
//
// Both contracts can be violated silently (no panic, no error). The
// reproducers below exercise the smallest path that exposes each.

// TestLeafInsert_HashAgreesAcrossInsertOrders exercises contract (1).
// Two trees built from the same key/value set in different orders must
// produce identical WorkingHash. The insert-then-shift order
// (Set(K20), Set(K30), Set(K10)) leaves the second insert's dirty bit
// at index 1; the third insert's slot shift moves the K30 data to
// index 2 without shifting the dirty bit, and the next Hash() call
// rebuilds with stale slotHashes for the K30 slot.
func TestLeafInsert_HashAgreesAcrossInsertOrders(t *testing.T) {
	mkTree := func() *MutableTree { return NewMutableTreeMem() }

	// Reverse-then-forward triggers the shift path on every insert
	// after the first.
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

	// Sorted order never shifts; per-slot cache stays trivially
	// consistent. Reference hash.
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

// TestInnerNodeClone_PreservesMiniTreeDirty exercises contract (2).
// Direct unit test on the Clone function: construct an InnerNode with
// miniTreeDirty set, verify the clone carries it; symmetric for a
// clean inner.
func TestInnerNodeClone_PreservesMiniTreeDirty(t *testing.T) {
	dirty := &InnerNode{
		miniTree:      NewMiniMerkle(),
		miniTreeDirty: true,
	}
	if c := dirty.Clone(); !c.miniTreeDirty {
		t.Fatalf("Clone() lost miniTreeDirty: orig=true, clone=false")
	}

	clean := &InnerNode{
		miniTree:      NewMiniMerkle(),
		miniTreeDirty: false,
	}
	if c := clean.Clone(); c.miniTreeDirty {
		t.Fatalf("Clone() flipped clean to dirty: orig=false, clone=true")
	}
}
