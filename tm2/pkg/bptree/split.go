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
func splitLeaf(keys [][]byte, valueHashes []Hash, insertPos int) (*LeafNode, splitResult) {
	total := len(keys) // B+1
	var splitPoint int

	if insertPos == B {
		// Append pattern: new key was inserted at the end.
		// 90/10: left gets B-1, right gets 2.
		splitPoint = total - 2 // B-1
	} else {
		// 50/50: left gets ceil((B+1)/2) = 17 for B=32
		splitPoint = (total + 1) / 2
	}

	left := &LeafNode{}
	left.numKeys = int16(splitPoint)
	copy(left.keys[:], keys[:splitPoint])
	copy(left.valueHashes[:], valueHashes[:splitPoint])

	rightCount := total - splitPoint
	right := &LeafNode{}
	right.numKeys = int16(rightCount)
	copy(right.keys[:], keys[splitPoint:])
	copy(right.valueHashes[:], valueHashes[splitPoint:])

	// Separator is a copy of the first key of the right leaf
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
func splitInner(keys [][]byte, children [][]byte, childHashes []Hash, height int16, sizes []int64) (*InnerNode, splitResult) {
	totalKeys := len(keys) // B (one more than max B-1)
	splitPoint := totalKeys / 2 // B/2 = 16

	left := &InnerNode{height: height}
	left.numKeys = int16(splitPoint)
	copy(left.keys[:], keys[:splitPoint])
	copy(left.children[:], children[:splitPoint+1])
	copy(left.childHashes[:], childHashes[:splitPoint+1])
	copy(left.childSizes[:], sizes[:splitPoint+1])

	// The separator is keys[splitPoint] — consumed, not in either node
	sep := keys[splitPoint]

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

func sumSizes(sizes []int64) int64 {
	var s int64
	for _, v := range sizes {
		s += v
	}
	return s
}

