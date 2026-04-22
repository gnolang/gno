package bptree

// removeResult is returned by recursive remove functions.
type removeResult struct {
	found      bool
	oldPayload slotPayload // the displaced slot (inline bytes or valueKey)
	underflow  bool        // node now has fewer than minimum entries
}

// treeRemove removes a key from the tree rooted at root. Returns the
// (possibly new) root, the displaced slot payload, and whether the key
// was found. Callers orphan oldPayload.valueKey only when external.
//
// The caller is responsible for COW-cloning the root before calling
// this function (MutableTree.Remove does this via cowRoot()). See
// Finding #17.
func treeRemove(root Node, key []byte) (Node, slotPayload, bool) {
	if root == nil {
		return nil, slotPayload{}, false
	}
	res := nodeRemove(root, key)
	if !res.found {
		return root, slotPayload{}, false
	}

	// Check for root collapse
	if inner, ok := root.(*InnerNode); ok && inner.numKeys == 0 {
		return inner.getChild(0), res.oldPayload, true
	}
	if leaf, ok := root.(*LeafNode); ok && leaf.numKeys == 0 {
		// Empty tree
		return nil, res.oldPayload, true
	}
	return root, res.oldPayload, true
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

	// Capture the displaced slot payload (either inline bytes or an
	// external valueKey) before shifting.
	old := captureSlotPayload(leaf, pos)

	n := int(leaf.numKeys)
	for i := pos; i < n-1; i++ {
		leaf.keys[i] = leaf.keys[i+1]
		leaf.valueHashes[i] = leaf.valueHashes[i+1]
		leaf.valueKeys[i] = leaf.valueKeys[i+1]
		leaf.inlineValues[i] = leaf.inlineValues[i+1]
		leaf.slotHashes[i] = leaf.slotHashes[i+1]
	}
	leaf.keys[n-1] = nil
	leaf.valueHashes[n-1] = Hash{}
	leaf.valueKeys[n-1] = nil
	leaf.inlineValues[n-1] = nil
	leaf.slotHashes[n-1] = Hash{}
	// Shift inlineMask AND slotsDirty bits [pos+1, n) down by one in
	// parallel with the slot-data shift above; the vacated top bit
	// falls off naturally. See shiftSlotsDirtyDown for the corruption
	// hazard the missing parallel shift causes.
	shiftInlineMaskDown(leaf, pos)
	shiftSlotsDirtyDown(leaf, pos)
	leaf.numKeys--
	leaf.miniTreeDirty = true

	return removeResult{found: true, oldPayload: old, underflow: leaf.numKeys < MinKeys}
}

// shiftInlineMaskDown shifts inlineMask bits [pos+1, 32) down by one
// to close the gap left by removing slot `pos`.
func shiftInlineMaskDown(leaf *LeafNode, pos int) {
	low := leaf.inlineMask & ((uint32(1) << uint(pos)) - 1)
	high := leaf.inlineMask >> uint(pos+1) << uint(pos)
	leaf.inlineMask = low | high
}

// shiftSlotsDirtyDown shifts slotsDirty bits [pos+1, 32) down by one
// to follow the parallel shift of slot data on remove. The vacated bit
// at the old top of the occupied range falls off naturally (cleared by
// the right-shift). See shiftSlotsDirtyUp comment for the corruption
// hazard the missing parallel shift causes.
func shiftSlotsDirtyDown(leaf *LeafNode, pos int) {
	low := leaf.slotsDirty & ((uint32(1) << uint(pos)) - 1)
	high := leaf.slotsDirty >> uint(pos+1) << uint(pos)
	leaf.slotsDirty = low | high
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
		// Single-slot update (one child hash changed). Prefer the
		// incremental SetSlot (5 hashes up the mini-merkle) over a
		// full 31-hash rebuild.
		inner.ensureMiniMerkleBuilt()
		inner.miniTree.SetSlot(childIdx, inner.childHashes[childIdx])
		return removeResult{found: true, oldPayload: res.oldPayload}
	}

	// Fix underflow
	merged := fixUnderflow(inner, childIdx)
	inner.miniTreeDirty = true

	// Inner node underflows if it has fewer than MinKeys-1 separators
	// (MinKeys-1 because inner minimum is ceil(B/2)-1 = 15 separators)
	return removeResult{
		found:      true,
		oldPayload: res.oldPayload,
		underflow:  merged && inner.numKeys < int16(MinKeys-1),
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
//
// Precondition: both children must be COW-cloned (nodeKey == nil). A
// shared child still bearing its persisted nodeKey would be mutated in
// place here, corrupting the retained version that also references it.
// fixUnderflow's call sites clone both sides; the assertion catches a
// future caller that forgets. See Finding #24.
func redistributeRight(parent *InnerNode, idx int) {
	left := parent.getChild(idx)
	right := parent.getChild(idx + 1)
	assertCloned(left, "redistributeRight: left child not cloned")
	assertCloned(right, "redistributeRight: right child not cloned")

	switch l := left.(type) {
	case *LeafNode:
		r := right.(*LeafNode)
		lastIdx := int(l.numKeys) - 1
		// Shift right's entries (plus slotHashes cache + inlineValues)
		// to make room.
		rn := int(r.numKeys)
		for i := rn; i > 0; i-- {
			r.keys[i] = r.keys[i-1]
			r.valueHashes[i] = r.valueHashes[i-1]
			r.valueKeys[i] = r.valueKeys[i-1]
			r.inlineValues[i] = r.inlineValues[i-1]
			r.slotHashes[i] = r.slotHashes[i-1]
		}
		// Shift r.inlineMask AND r.slotsDirty up by 1 bit (inserting a
		// vacant bit at 0) — parallel with the slot-data shift above.
		// See shiftSlotsDirtyUp comment (insert.go) for the corruption
		// hazard the missing parallel shift causes.
		r.inlineMask <<= 1
		r.slotsDirty <<= 1
		// Move the last slot from l to position 0 of r, preserving
		// inline/external status.
		r.keys[0] = l.keys[lastIdx]
		r.valueHashes[0] = l.valueHashes[lastIdx]
		r.valueKeys[0] = l.valueKeys[lastIdx]
		r.inlineValues[0] = l.inlineValues[lastIdx]
		r.slotHashes[0] = l.slotHashes[lastIdx]
		if l.inlineMask&(uint32(1)<<uint(lastIdx)) != 0 {
			r.inlineMask |= 1
		}
		// Mirror the borrowed slot's dirty bit so r.slotHashes[0]'s
		// validity tracks l.slotHashes[lastIdx]'s pre-move state.
		if l.slotsDirty&(uint32(1)<<uint(lastIdx)) != 0 {
			r.slotsDirty |= 1
		}
		r.numKeys++
		l.keys[lastIdx] = nil
		l.valueHashes[lastIdx] = Hash{}
		l.valueKeys[lastIdx] = nil
		l.inlineValues[lastIdx] = nil
		l.slotHashes[lastIdx] = Hash{}
		l.inlineMask &^= uint32(1) << uint(lastIdx)
		l.slotsDirty &^= uint32(1) << uint(lastIdx)
		l.numKeys--
		// Update separator and parent childSizes
		parent.keys[idx] = copyKey(r.keys[0])
		parent.childSizes[idx]--
		parent.childSizes[idx+1]++
		l.miniTreeDirty = true
		r.miniTreeDirty = true
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
		// Demote separator, promote left's last key. Both are copied so
		// the receiving node owns its key bytes (Finding #20).
		r.keys[0] = copyKey(parent.keys[idx])
		r.childNodes[0] = movedChild
		r.children[0] = l.children[lastChildIdx]
		r.childHashes[0] = l.childHashes[lastChildIdx]
		r.childSizes[0] = movedSize
		r.numKeys++
		parent.keys[idx] = copyKey(l.keys[lastKeyIdx])
		l.keys[lastKeyIdx] = nil
		l.childNodes[lastChildIdx] = nil
		l.children[lastChildIdx] = nil
		l.childHashes[lastChildIdx] = Hash{}
		l.childSizes[lastChildIdx] = 0
		l.numKeys--
		// Update parent's view of these children's sizes
		parent.childSizes[idx] -= movedSize
		parent.childSizes[idx+1] += movedSize
		l.rebuildChildLoaded()
		r.rebuildChildLoaded()
		l.miniTreeDirty = true
		r.miniTreeDirty = true
		parent.childHashes[idx] = l.Hash()
		parent.childHashes[idx+1] = r.Hash()
	}
}

// redistributeLeft moves the first entry from parent.child[idx+1] to
// parent.child[idx], updating the separator at parent.keys[idx].
//
// Precondition: both children must be COW-cloned. See redistributeRight
// and Finding #24.
func redistributeLeft(parent *InnerNode, idx int) {
	left := parent.getChild(idx)
	right := parent.getChild(idx + 1)
	assertCloned(left, "redistributeLeft: left child not cloned")
	assertCloned(right, "redistributeLeft: right child not cloned")

	switch r := right.(type) {
	case *LeafNode:
		l := left.(*LeafNode)
		// Append right's first entry (plus its slot hash and inline
		// status) to left.
		dst := int(l.numKeys)
		l.keys[dst] = r.keys[0]
		l.valueHashes[dst] = r.valueHashes[0]
		l.valueKeys[dst] = r.valueKeys[0]
		l.inlineValues[dst] = r.inlineValues[0]
		l.slotHashes[dst] = r.slotHashes[0]
		if r.inlineMask&1 != 0 {
			l.inlineMask |= uint32(1) << uint(dst)
		}
		// Mirror the borrowed slot's dirty bit so l.slotHashes[dst]'s
		// validity tracks r.slotHashes[0]'s pre-move state
		//.
		if r.slotsDirty&1 != 0 {
			l.slotsDirty |= uint32(1) << uint(dst)
		}
		l.numKeys++
		// Shift right left (with slotHashes cache + inline values).
		rn := int(r.numKeys)
		for i := 0; i < rn-1; i++ {
			r.keys[i] = r.keys[i+1]
			r.valueHashes[i] = r.valueHashes[i+1]
			r.valueKeys[i] = r.valueKeys[i+1]
			r.inlineValues[i] = r.inlineValues[i+1]
			r.slotHashes[i] = r.slotHashes[i+1]
		}
		// Shift inlineMask AND slotsDirty down by one — parallel with
		// the slot-data shift above.
		r.inlineMask >>= 1
		r.slotsDirty >>= 1
		r.keys[rn-1] = nil
		r.valueHashes[rn-1] = Hash{}
		r.valueKeys[rn-1] = nil
		r.inlineValues[rn-1] = nil
		r.slotHashes[rn-1] = Hash{}
		r.numKeys--
		parent.keys[idx] = copyKey(r.keys[0])
		parent.childSizes[idx]++
		parent.childSizes[idx+1]--
		l.miniTreeDirty = true
		r.miniTreeDirty = true
		parent.childHashes[idx] = l.Hash()
		parent.childHashes[idx+1] = r.Hash()

	case *InnerNode:
		l := left.(*InnerNode)
		movedSize := r.childSizes[0]
		movedChild := r.childNodes[0]
		// Demote separator to end of left, promote right's first key.
		// Both are copied so the receiving node owns its key bytes
		// (Finding #20).
		l.keys[l.numKeys] = copyKey(parent.keys[idx])
		lnc := l.NumChildren()
		l.childNodes[lnc] = movedChild
		l.children[lnc] = r.children[0]
		l.childHashes[lnc] = r.childHashes[0]
		l.childSizes[lnc] = movedSize
		l.numKeys++
		parent.keys[idx] = copyKey(r.keys[0])
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
		l.rebuildChildLoaded()
		r.rebuildChildLoaded()
		l.miniTreeDirty = true
		r.miniTreeDirty = true
		parent.childHashes[idx] = l.Hash()
		parent.childHashes[idx+1] = r.Hash()
	}
}

// merge merges parent.child[idx+1] into parent.child[idx], removing
// the separator at parent.keys[idx].
//
// Precondition: the left (destination) child must be COW-cloned.
// fixUnderflow always clones left before calling merge; the assertion
// guards against a future caller that forgets. See Finding #24.
func merge(parent *InnerNode, idx int) {
	left := parent.getChild(idx)
	right := parent.getChild(idx + 1)
	assertCloned(left, "merge: destination (left) child not cloned")

	switch l := left.(type) {
	case *LeafNode:
		r := right.(*LeafNode)
		lBase := int(l.numKeys)
		for i := 0; i < int(r.numKeys); i++ {
			dst := lBase + i
			// copyKey so the merged-left node owns its key bytes
			// independently of the (now dead) right node, in line with
			// Finding #20's uniform discipline (matches the InnerNode
			// case below). valueKeys are content-addressed identifiers
			// (12-byte NodeKey form, allocated immutable per session);
			// inlineValues/slotHashes are similarly transferable as
			// slice headers. See BUG-5 in PR-5571 / PR-5750 history.
			l.keys[dst] = copyKey(r.keys[i])
			l.valueHashes[dst] = r.valueHashes[i]
			l.valueKeys[dst] = r.valueKeys[i]
			l.inlineValues[dst] = r.inlineValues[i]
			l.slotHashes[dst] = r.slotHashes[i]
			if r.inlineMask&(uint32(1)<<uint(i)) != 0 {
				l.inlineMask |= uint32(1) << uint(dst)
			}
			// Mirror the absorbed slot's dirty bit so l.slotHashes[dst]'s
			// validity tracks r.slotHashes[i]'s pre-merge state. Without
			// this, dirty bits from r are dropped — the next incremental
			// rebuild on l would trust uninitialised slotHashes for the
			// absorbed range.
			if r.slotsDirty&(uint32(1)<<uint(i)) != 0 {
				l.slotsDirty |= uint32(1) << uint(dst)
			}
		}
		l.numKeys += r.numKeys
		l.miniTreeDirty = true

	case *InnerNode:
		r := right.(*InnerNode)
		// Demote separator. Copied so the merged left node owns its key
		// bytes (Finding #20).
		l.keys[l.numKeys] = copyKey(parent.keys[idx])
		l.numKeys++
		// Append right's keys and children. copyKey for parity with
		// Finding #20 — see the LeafNode case above. (BUG-5.)
		for i := 0; i < int(r.numKeys); i++ {
			l.keys[l.numKeys] = copyKey(r.keys[i])
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
		l.rebuildChildLoaded()
		l.miniTreeDirty = true
	}

	// Remove separator and right child from parent
	pn := int(parent.numKeys)
	// Update parent's childSizes for the merged child
	parent.childSizes[idx] += parent.childSizes[idx+1]
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
	parent.rebuildChildLoaded()
	parent.childHashes[idx] = left.Hash()
}

func copyKey(k []byte) []byte {
	c := make([]byte, len(k))
	copy(c, k)
	return c
}

// assertCloned panics if node still carries a persisted nodeKey. A cloned
// node always has nodeKey == nil (see cloneNode); the helper codifies the
// contract that structural mutations must operate only on COW clones. See
// Finding #24.
func assertCloned(node Node, msg string) {
	if node == nil {
		return
	}
	if node.GetNodeKey() != nil {
		panic(msg)
	}
}
