package bptree

// removeResult is returned by recursive remove functions.
type removeResult struct {
	found     bool
	oldValue  Hash
	underflow bool // node now has fewer than minimum entries
}

// treeRemove removes a key from the tree rooted at root.
// Returns the (possibly new) root, the old value hash, and whether found.
func treeRemove(root Node, key []byte) (Node, Hash, bool) {
	if root == nil {
		return nil, Hash{}, false
	}
	root = cloneNode(root)
	res := nodeRemove(root, key)
	if !res.found {
		return root, Hash{}, false
	}

	// Check for root collapse
	if inner, ok := root.(*InnerNode); ok && inner.numKeys == 0 {
		// Root has single child — collapse
		return inner.getChild(0), res.oldValue, true
	}
	if leaf, ok := root.(*LeafNode); ok && leaf.numKeys == 0 {
		// Empty tree
		return nil, res.oldValue, true
	}
	return root, res.oldValue, true
}

func nodeRemove(node Node, key []byte) removeResult {
	switch n := node.(type) {
	case *LeafNode:
		return leafRemove(n, key)
	case *InnerNode:
		return innerRemove(n, key)
	default:
		panic("unknown node type")
	}
}

func leafRemove(leaf *LeafNode, key []byte) removeResult {
	pos, found := searchLeaf(leaf, key)
	if !found {
		return removeResult{}
	}

	oldVH := leaf.valueHashes[pos]
	n := int(leaf.numKeys)
	for i := pos; i < n-1; i++ {
		leaf.keys[i] = leaf.keys[i+1]
		leaf.valueHashes[i] = leaf.valueHashes[i+1]
	}
	leaf.keys[n-1] = nil
	leaf.valueHashes[n-1] = Hash{}
	leaf.numKeys--
	leaf.RebuildMiniMerkle()

	return removeResult{found: true, oldValue: oldVH, underflow: leaf.numKeys < MinKeys}
}

func innerRemove(inner *InnerNode, key []byte) removeResult {
	childIdx := searchInner(inner, key)
	child := inner.getChild(childIdx)
	if child == nil {
		panic("nil child in innerRemove")
	}

	child = cloneNode(child)
	inner.setChild(childIdx, child)
	res := nodeRemove(child, key)
	if !res.found {
		return res
	}

	inner.childSizes[childIdx]--
	inner.childHashes[childIdx] = child.Hash()

	if !res.underflow {
		inner.miniTree.SetSlot(childIdx, inner.childHashes[childIdx])
		return removeResult{found: true, oldValue: res.oldValue}
	}

	// Fix underflow
	merged := fixUnderflow(inner, childIdx)
	inner.RebuildMiniMerkle()

	// Inner node underflows if it has fewer than MinKeys-1 separators
	// (MinKeys-1 because inner minimum is ceil(B/2)-1 = 15 separators)
	return removeResult{
		found:     true,
		oldValue:  res.oldValue,
		underflow: merged && inner.numKeys < int16(MinKeys-1),
	}
}

// fixUnderflow fixes an underflowing child at childIdx by redistributing
// or merging. Returns true if a merge occurred.
func fixUnderflow(parent *InnerNode, childIdx int) bool {
	// Try redistribute from left sibling
	if childIdx > 0 {
		left := parent.getChild(childIdx - 1)
		if canSpare(left) {
			leftClone := cloneNode(left)
			parent.setChild(childIdx-1, leftClone)
			redistributeRight(parent, childIdx-1) // move from left to child
			return false
		}
	}

	// Try redistribute from right sibling
	if childIdx < int(parent.numKeys) {
		right := parent.getChild(childIdx + 1)
		if canSpare(right) {
			rightClone := cloneNode(right)
			parent.setChild(childIdx+1, rightClone)
			redistributeLeft(parent, childIdx) // move from right to child
			return false
		}
	}

	// Must merge — always clone the sibling before merging into it.
	// We unconditionally clone because the sibling may be shared with
	// lastSaved (for rollback) or other versions. The previous heuristic
	// of checking GetNodeKey() != nil was incorrect for in-memory trees
	// where nodeKeys are always nil.
	if childIdx > 0 {
		leftClone := cloneNode(parent.getChild(childIdx - 1))
		parent.setChild(childIdx-1, leftClone)
		merge(parent, childIdx-1)
	} else {
		rightClone := cloneNode(parent.getChild(childIdx + 1))
		parent.setChild(childIdx+1, rightClone)
		merge(parent, childIdx)
	}
	return true
}

func canSpare(n Node) bool {
	switch n := n.(type) {
	case *LeafNode:
		return n.numKeys > int16(MinKeys)
	case *InnerNode:
		return n.numKeys > int16(MinKeys-1)
	default:
		return false
	}
}

// redistributeRight moves the last entry from parent.child[idx] to
// parent.child[idx+1], updating the separator at parent.keys[idx].
func redistributeRight(parent *InnerNode, idx int) {
	left := parent.getChild(idx)
	right := parent.getChild(idx + 1)

	switch l := left.(type) {
	case *LeafNode:
		r := right.(*LeafNode)
		lastIdx := int(l.numKeys) - 1
		// Shift right's entries to make room at position 0
		rn := int(r.numKeys)
		for i := rn; i > 0; i-- {
			r.keys[i] = r.keys[i-1]
			r.valueHashes[i] = r.valueHashes[i-1]
		}
		r.keys[0] = l.keys[lastIdx]
		r.valueHashes[0] = l.valueHashes[lastIdx]
		r.numKeys++
		l.keys[lastIdx] = nil
		l.valueHashes[lastIdx] = Hash{}
		l.numKeys--
		// Update separator and parent childSizes
		parent.keys[idx] = copyKey(r.keys[0])
		parent.childSizes[idx]--
		parent.childSizes[idx+1]++
		l.RebuildMiniMerkle()
		r.RebuildMiniMerkle()
		parent.childHashes[idx] = l.Hash()
		parent.childHashes[idx+1] = r.Hash()

	case *InnerNode:
		r := right.(*InnerNode)
		lastKeyIdx := int(l.numKeys) - 1
		lastChildIdx := int(l.numKeys)
		movedSize := l.childSizes[lastChildIdx]
		movedChild := l.childNodes[lastChildIdx]
		// Shift right's entries (including childSizes)
		rn := int(r.numKeys)
		for i := rn; i > 0; i-- {
			r.keys[i] = r.keys[i-1]
		}
		for i := rn + 1; i > 0; i-- {
			r.childNodes[i] = r.childNodes[i-1]
			r.children[i] = r.children[i-1]
			r.childHashes[i] = r.childHashes[i-1]
			r.childSizes[i] = r.childSizes[i-1]
		}
		// Demote separator, promote left's last key
		r.keys[0] = parent.keys[idx]
		r.childNodes[0] = movedChild
		r.children[0] = l.children[lastChildIdx]
		r.childHashes[0] = l.childHashes[lastChildIdx]
		r.childSizes[0] = movedSize
		r.numKeys++
		parent.keys[idx] = l.keys[lastKeyIdx]
		l.keys[lastKeyIdx] = nil
		l.childNodes[lastChildIdx] = nil
		l.children[lastChildIdx] = nil
		l.childHashes[lastChildIdx] = Hash{}
		l.childSizes[lastChildIdx] = 0
		l.numKeys--
		// Update parent's view of these children's sizes
		parent.childSizes[idx] -= movedSize
		parent.childSizes[idx+1] += movedSize
		l.RebuildMiniMerkle()
		r.RebuildMiniMerkle()
		parent.childHashes[idx] = l.Hash()
		parent.childHashes[idx+1] = r.Hash()
	}
}

// redistributeLeft moves the first entry from parent.child[idx+1] to
// parent.child[idx], updating the separator at parent.keys[idx].
func redistributeLeft(parent *InnerNode, idx int) {
	left := parent.getChild(idx)
	right := parent.getChild(idx + 1)

	switch r := right.(type) {
	case *LeafNode:
		l := left.(*LeafNode)
		// Append right's first entry to left
		l.keys[l.numKeys] = r.keys[0]
		l.valueHashes[l.numKeys] = r.valueHashes[0]
		l.numKeys++
		// Shift right left
		rn := int(r.numKeys)
		for i := 0; i < rn-1; i++ {
			r.keys[i] = r.keys[i+1]
			r.valueHashes[i] = r.valueHashes[i+1]
		}
		r.keys[rn-1] = nil
		r.valueHashes[rn-1] = Hash{}
		r.numKeys--
		parent.keys[idx] = copyKey(r.keys[0])
		parent.childSizes[idx]++
		parent.childSizes[idx+1]--
		l.RebuildMiniMerkle()
		r.RebuildMiniMerkle()
		parent.childHashes[idx] = l.Hash()
		parent.childHashes[idx+1] = r.Hash()

	case *InnerNode:
		l := left.(*InnerNode)
		movedSize := r.childSizes[0]
		movedChild := r.childNodes[0]
		// Demote separator to end of left, promote right's first key
		l.keys[l.numKeys] = parent.keys[idx]
		lnc := l.NumChildren()
		l.childNodes[lnc] = movedChild
		l.children[lnc] = r.children[0]
		l.childHashes[lnc] = r.childHashes[0]
		l.childSizes[lnc] = movedSize
		l.numKeys++
		parent.keys[idx] = r.keys[0]
		// Shift right left (including childSizes)
		rn := int(r.numKeys)
		for i := 0; i < rn-1; i++ {
			r.keys[i] = r.keys[i+1]
		}
		for i := 0; i < rn; i++ {
			r.childNodes[i] = r.childNodes[i+1]
			r.children[i] = r.children[i+1]
			r.childHashes[i] = r.childHashes[i+1]
			r.childSizes[i] = r.childSizes[i+1]
		}
		r.childNodes[rn] = nil
		r.children[rn] = nil
		r.childHashes[rn] = Hash{}
		r.childSizes[rn] = 0
		r.keys[rn-1] = nil
		r.numKeys--
		// Update parent's view of these children's sizes
		parent.childSizes[idx] += movedSize
		parent.childSizes[idx+1] -= movedSize
		l.RebuildMiniMerkle()
		r.RebuildMiniMerkle()
		parent.childHashes[idx] = l.Hash()
		parent.childHashes[idx+1] = r.Hash()
	}
}

// merge merges parent.child[idx+1] into parent.child[idx], removing
// the separator at parent.keys[idx].
func merge(parent *InnerNode, idx int) {
	left := parent.getChild(idx)
	right := parent.getChild(idx + 1)

	switch l := left.(type) {
	case *LeafNode:
		r := right.(*LeafNode)
		for i := 0; i < int(r.numKeys); i++ {
			l.keys[l.numKeys] = r.keys[i]
			l.valueHashes[l.numKeys] = r.valueHashes[i]
			l.numKeys++
		}
		l.RebuildMiniMerkle()

	case *InnerNode:
		r := right.(*InnerNode)
		// Demote separator
		l.keys[l.numKeys] = parent.keys[idx]
		l.numKeys++
		// Append right's keys and children
		for i := 0; i < int(r.numKeys); i++ {
			l.keys[l.numKeys] = r.keys[i]
			l.numKeys++
		}
		// Children: left already has some, append right's (including childSizes)
		leftChildBase := int(l.numKeys) - int(r.numKeys) // position after demoted separator
		for i := 0; i < r.NumChildren(); i++ {
			l.childNodes[leftChildBase+i] = r.childNodes[i]
			l.children[leftChildBase+i] = r.children[i]
			l.childHashes[leftChildBase+i] = r.childHashes[i]
			l.childSizes[leftChildBase+i] = r.childSizes[i]
		}
		l.RebuildMiniMerkle()
	}

	// Remove separator and right child from parent
	pn := int(parent.numKeys)
	// Update parent's childSizes for the merged child
	parent.childSizes[idx] = parent.childSizes[idx] + parent.childSizes[idx+1]
	for i := idx; i < pn-1; i++ {
		parent.keys[i] = parent.keys[i+1]
	}
	parent.keys[pn-1] = nil
	for i := idx + 1; i < pn; i++ {
		parent.childNodes[i] = parent.childNodes[i+1]
		parent.children[i] = parent.children[i+1]
		parent.childHashes[i] = parent.childHashes[i+1]
		parent.childSizes[i] = parent.childSizes[i+1]
	}
	parent.childNodes[pn] = nil
	parent.children[pn] = nil
	parent.childHashes[pn] = Hash{}
	parent.childSizes[pn] = 0
	parent.numKeys--
	parent.childHashes[idx] = left.Hash()
}

func copyKey(k []byte) []byte {
	c := make([]byte, len(k))
	copy(c, k)
	return c
}
