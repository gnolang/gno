package bptree

// splitResult is returned when a node split occurs during insertion.
// The caller (parent) must insert the separator key and the new right child.
type splitResult struct {
	separator []byte     // first key of the right node (copy for parent)
	right     Node       // the new right sibling
}

// splitLeaf splits a leaf that has B+1 keys (overflow after insert at pos).
// insertPos is the position where the new key was inserted in the overflow array.
//
// 90/10 split: if insertPos == B (key appended at the end), left gets B-1
// keys, right gets 2 keys. Otherwise 50/50: left gets ceil((B+1)/2) = 17.
//
// Each slot carries either an inline payload (when bit i is set in
// inlineMask — inlineValues[i] holds the bytes) or an external
// valueKey (inlineMask bit i cleared — valueKeys[i] is the 12-byte
// reference). The split partitions inlineMask along with the other
// arrays.
func splitLeaf(keys [][]byte, valueHashes []Hash, valueKeys [][]byte, inlineValues [][]byte, inlineMask uint64, insertPos int) (*LeafNode, splitResult) {
	total := len(keys) // B+1
	var splitPoint int

	if insertPos == B {
		splitPoint = total - 2 // B-1
	} else {
		splitPoint = (total + 1) / 2
	}

	left := &LeafNode{}
	left.numKeys = int16(splitPoint)
	copy(left.keys[:], keys[:splitPoint])
	copy(left.valueHashes[:], valueHashes[:splitPoint])
	copy(left.valueKeys[:], valueKeys[:splitPoint])
	copy(left.inlineValues[:], inlineValues[:splitPoint])
	left.inlineMask = uint32(inlineMask & ((uint64(1) << uint(splitPoint)) - 1))

	rightCount := total - splitPoint
	right := &LeafNode{}
	right.numKeys = int16(rightCount)
	copy(right.keys[:], keys[splitPoint:])
	copy(right.valueHashes[:], valueHashes[splitPoint:])
	copy(right.valueKeys[:], valueKeys[splitPoint:])
	copy(right.inlineValues[:], inlineValues[splitPoint:])
	// Right-half's inlineMask: shift the high bits down by splitPoint
	// so they align with right's slot indices.
	right.inlineMask = uint32(inlineMask >> uint(splitPoint))

	// Mark every occupied slot dirty on both halves so the next
	// ensureMiniMerkleBuilt rebuilds via rebuildMiniMerkleIncremental
	// against fresh slotHashes. The slotHashes [B]Hash arrays on these
	// fresh nodes are zero-initialised; without flagging them dirty,
	// rebuildMiniMerkleIncremental would trust the all-zero cache
	// (slotsDirty == 0 means "all hashes are valid") and emit a
	// corrupt root hash.
	left.markLeafSlotsDirtyRange(0, int(left.numKeys))
	right.markLeafSlotsDirtyRange(0, int(right.numKeys))

	sep := make([]byte, len(right.keys[0]))
	copy(sep, right.keys[0])

	return left, splitResult{
		separator: sep,
		right:     right,
	}
}

// splitInner splits an inner node that has B children (overflow: B keys, B+1 children).
// The overflow entry is passed as sorted slices.
//
// For inner nodes, the separator is CONSUMED (promoted to the parent),
// not duplicated. Left gets splitPoint keys and splitPoint+1 children.
// Right gets the remaining keys and children.
//
// The promoted separator is defensively copied so that the parent owns
// its own byte storage, matching the invariant established by splitLeaf
// and the redistribute paths (see Finding #20). Callers never mutate
// key bytes in place today, but the defensive copy makes the
// node-level key ownership invariant unconditional.
func splitInner(keys [][]byte, children [][]byte, childHashes []Hash, height int16, sizes []int64) (*InnerNode, splitResult) {
	totalKeys := len(keys) // B (one more than max B-1)
	splitPoint := totalKeys / 2 // B/2 = 16

	left := &InnerNode{height: height}
	left.numKeys = int16(splitPoint)
	copy(left.keys[:], keys[:splitPoint])
	copy(left.children[:], children[:splitPoint+1])
	copy(left.childHashes[:], childHashes[:splitPoint+1])
	copy(left.childSizes[:], sizes[:splitPoint+1])

	// The separator is keys[splitPoint] — consumed, promoted to the
	// parent as a freshly-owned byte slice.
	sep := copyKey(keys[splitPoint])

	rightKeys := keys[splitPoint+1:]
	rightChildren := children[splitPoint+1:]
	rightChildHashes := childHashes[splitPoint+1:]
	rightSizes := sizes[splitPoint+1:]

	right := &InnerNode{height: height}
	right.numKeys = int16(len(rightKeys))
	copy(right.keys[:], rightKeys)
	copy(right.children[:], rightChildren)
	copy(right.childHashes[:], rightChildHashes)
	copy(right.childSizes[:], rightSizes)

	return left, splitResult{
		separator: sep,
		right:     right,
	}
}
