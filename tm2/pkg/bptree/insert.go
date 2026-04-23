package bptree

// slotPayload is the per-slot value information passed from Set down
// through the insert chain. Exactly one of `inline` or `valueKey` is
// non-nil: inline slots carry the raw bytes (stored on the leaf);
// external slots carry a ValueKey reference (resolved via the value
// store). valueHash is always populated so slot hashing stays uniform.
type slotPayload struct {
	valueHash Hash
	inline    []byte // non-nil for inline storage; LeafNode takes ownership
	valueKey  []byte // non-nil for external storage
}

// insertResult is returned by recursive insert functions.
type insertResult struct {
	updated    bool         // true if key already existed (value replaced)
	split      *splitResult // non-nil if the node split
	oldPayload slotPayload  // displaced slot (only valueKey is consulted, for orphan tracking)
}

// treeInsert inserts a key with a pre-computed value hash and payload.
// Returns the (possibly new) root, whether it was an update, and the old
// slot payload if an existing key was overwritten (zero struct otherwise).
//
// The caller is responsible for ensuring the root is COW-cloned if it
// is shared with a snapshot (lastSaved) or came from the node cache.
// MutableTree.Set does this via cowRoot() at most once per working
// version. Re-cloning here on every Set in a single working version is
// a ~4.3 KB struct copy per call that we can avoid. See Finding #17.
func treeInsert(root Node, key []byte, payload slotPayload) (Node, bool, slotPayload) {
	key = copyKey(key) // defensive copy — caller may reuse the slice
	res := nodeInsert(root, key, payload)

	if res.split != nil {
		// Root split — create a new inner root
		sr := res.split
		newRoot := &InnerNode{
			numKeys: 1,
			height:  nodeHeight(root) + 1,
		}
		newRoot.keys[0] = sr.separator
		newRoot.childNodes[0] = root
		newRoot.childNodes[1] = sr.right
		newRoot.childHashes[0] = root.Hash()
		newRoot.childHashes[1] = sr.right.Hash()
		newRoot.childSizes[0] = nodeSize(root)
		newRoot.childSizes[1] = nodeSize(sr.right)
		newRoot.rebuildChildLoaded()
		newRoot.miniTreeDirty = true // defer rebuild; next Hash() materialises it
		return newRoot, res.updated, res.oldPayload
	}
	return root, res.updated, res.oldPayload
}

// nodeInsert recursively inserts into the subtree rooted at node.
// The node must already be COW-cloned by the caller.
func nodeInsert(node Node, key []byte, payload slotPayload) insertResult {
	switch n := node.(type) {
	case *LeafNode:
		return leafInsert(n, key, payload)
	case *InnerNode:
		return innerInsert(n, key, payload)
	default:
		panic("unknown node type")
	}
}

// applyPayloadToSlot writes payload into slot `pos` of the leaf,
// toggling the inline bit appropriately. Used by both the update and
// insert paths to avoid duplicating the inline/external branching.
func applyPayloadToSlot(leaf *LeafNode, pos int, payload slotPayload) {
	leaf.valueHashes[pos] = payload.valueHash
	bit := uint32(1) << uint(pos)
	if payload.inline != nil {
		leaf.inlineValues[pos] = payload.inline
		leaf.valueKeys[pos] = nil
		leaf.inlineMask |= bit
	} else {
		leaf.inlineValues[pos] = nil
		leaf.valueKeys[pos] = payload.valueKey
		leaf.inlineMask &^= bit
	}
}

// captureSlotPayload extracts the current payload at slot pos (for
// orphan tracking of displaced external values).
func captureSlotPayload(leaf *LeafNode, pos int) slotPayload {
	p := slotPayload{valueHash: leaf.valueHashes[pos]}
	if leaf.inlineMask&(uint32(1)<<uint(pos)) != 0 {
		p.inline = leaf.inlineValues[pos]
	} else {
		p.valueKey = leaf.valueKeys[pos]
	}
	return p
}

// leafInsert inserts into a leaf node (already COW-cloned).
func leafInsert(leaf *LeafNode, key []byte, payload slotPayload) insertResult {
	pos, found := searchLeaf(leaf, key)

	if found {
		// Update existing key — size unchanged. Capture old payload for
		// orphan tracking (the caller orphans only external slots).
		old := captureSlotPayload(leaf, pos)
		applyPayloadToSlot(leaf, pos, payload)
		// Single-slot update. Rehash the slot once, refresh the per-slot
		// hash cache, and walk the mini-merkle path incrementally
		// (5 SHA-256 hashes) rather than paying a full 31-hash rebuild.
		// ensureMiniMerkleBuilt covers the case where an earlier bulk
		// mutation left the tree dirty.
		leaf.ensureMiniMerkleBuilt()
		h := HashLeafSlotFromValueHash(key, payload.valueHash)
		leaf.slotHashes[pos] = h
		leaf.miniTree.SetSlot(pos, h)
		return insertResult{updated: true, oldPayload: old}
	}

	// Insert new key
	if int(leaf.numKeys) < B {
		// Room in this leaf — shift right and insert. The slot-level
		// hash cache (slotHashes) is intentionally NOT shifted: doing so
		// would cost a 32-byte memcpy per shifted slot on every insert,
		// while marking the shifted range dirty defers the actual
		// rehashing until the next Hash() and folds repeated dirtying
		// of the same slot across a burst of inserts into a single
		// recomputation per rebuildMiniMerkleIncremental pass.
		n := int(leaf.numKeys)
		for i := n; i > pos; i-- {
			leaf.keys[i] = leaf.keys[i-1]
			leaf.valueHashes[i] = leaf.valueHashes[i-1]
			leaf.valueKeys[i] = leaf.valueKeys[i-1]
			leaf.inlineValues[i] = leaf.inlineValues[i-1]
		}
		shiftInlineMaskUp(leaf, pos)
		leaf.keys[pos] = key
		// Clear existing state at pos then apply payload.
		leaf.inlineValues[pos] = nil
		leaf.valueKeys[pos] = nil
		leaf.inlineMask &^= uint32(1) << uint(pos)
		applyPayloadToSlot(leaf, pos, payload)
		leaf.numKeys++
		// Slots [pos, numKeys) all hold data that does not match
		// slotHashes[i] (either freshly inserted, or shifted from a
		// neighbour). Mark the entire range dirty so the next merkle
		// rebuild rehashes them.
		leaf.markLeafSlotsDirtyRange(pos, int(leaf.numKeys))
		return insertResult{updated: false}
	}

	// Leaf is full (numKeys == B) — need to split. Scratch arrays for
	// the B+1 overflow entries live on the stack; splitLeaf copies out
	// of them into the returned LeafNodes, so the backing arrays do
	// not escape.
	var allKeys [B + 1][]byte
	var allVH [B + 1]Hash
	var allVK [B + 1][]byte
	var allInline [B + 1][]byte
	// Overflow array has B+1 = 33 slots — one more than a uint32 can
	// represent. Use uint64 locally; splitLeaf produces two uint32
	// halves each covering at most B slots (left ≤ 17, right ≤ 17).
	var allInlineMask uint64
	copy(allKeys[:pos], leaf.keys[:pos])
	allKeys[pos] = key
	copy(allKeys[pos+1:], leaf.keys[pos:B])
	copy(allVH[:pos], leaf.valueHashes[:pos])
	allVH[pos] = payload.valueHash
	copy(allVH[pos+1:], leaf.valueHashes[pos:B])
	copy(allVK[:pos], leaf.valueKeys[:pos])
	copy(allVK[pos+1:], leaf.valueKeys[pos:B])
	copy(allInline[:pos], leaf.inlineValues[:pos])
	copy(allInline[pos+1:], leaf.inlineValues[pos:B])
	// Widen before shifting so the high bit of a full leaf (bit 31)
	// can survive the up-shift into bit 32.
	srcMask := uint64(leaf.inlineMask)
	lowMask := srcMask & ((uint64(1) << uint(pos)) - 1)
	highMask := (srcMask &^ ((uint64(1) << uint(pos)) - 1)) << 1
	allInlineMask = lowMask | highMask
	if payload.inline != nil {
		allInline[pos] = payload.inline
		allInlineMask |= uint64(1) << uint(pos)
	} else {
		allVK[pos] = payload.valueKey
	}

	left, sr := splitLeaf(allKeys[:], allVH[:], allVK[:], allInline[:], allInlineMask, pos)
	*leaf = *left
	leaf.markLeafSlotsDirtyRange(0, int(leaf.numKeys))
	r := sr.right.(*LeafNode)
	r.markLeafSlotsDirtyRange(0, int(r.numKeys))
	return insertResult{updated: false, split: &sr}
}

// shiftInlineMaskUp shifts inlineMask bits at positions [pos, 32) up
// by one. The bit previously at position B-1 is dropped (leaf had
// room by contract, so that position must be zero).
func shiftInlineMaskUp(leaf *LeafNode, pos int) {
	highBits := leaf.inlineMask &^ ((uint32(1) << uint(pos)) - 1)
	leaf.inlineMask = (leaf.inlineMask & ((uint32(1) << uint(pos)) - 1)) | (highBits << 1)
}

// innerInsert inserts into an inner node (already COW-cloned).
func innerInsert(inner *InnerNode, key []byte, payload slotPayload) insertResult {
	childIdx := searchInner(inner, key)

	child := inner.getChild(childIdx)
	if child == nil {
		panic("inner node has nil child")
	}

	// COW-clone the child
	child = cloneNode(child)
	inner.setChild(childIdx, child)

	// Recurse
	res := nodeInsert(child, key, payload)

	if !res.updated {
		inner.childSizes[childIdx]++
	}

	// Update child hash
	inner.childHashes[childIdx] = child.Hash()

	if res.split == nil {
		// Single-slot update (one child hash changed). Prefer the
		// incremental SetSlot (5 hashes up the mini-merkle) over a
		// full 31-hash rebuild, provided the mini-merkle is current.
		inner.ensureMiniMerkleBuilt()
		inner.miniTree.SetSlot(childIdx, inner.childHashes[childIdx])
		return insertResult{updated: res.updated, oldPayload: res.oldPayload}
	}

	// Child split — insert separator and new right child into this inner node
	sr := res.split
	rightChildHash := sr.right.Hash()

	if int(inner.numKeys) < B-1 {
		// Room in this inner node — shift right and insert
		n := int(inner.numKeys)
		for i := n; i > childIdx; i-- {
			inner.keys[i] = inner.keys[i-1]
			inner.childNodes[i+1] = inner.childNodes[i]
			inner.children[i+1] = inner.children[i]
			inner.childHashes[i+1] = inner.childHashes[i]
			inner.childSizes[i+1] = inner.childSizes[i]
		}
		inner.keys[childIdx] = sr.separator
		inner.childNodes[childIdx+1] = sr.right
		inner.children[childIdx+1] = nil
		inner.childHashes[childIdx+1] = rightChildHash
		inner.childSizes[childIdx] = nodeSize(child)
		inner.childSizes[childIdx+1] = nodeSize(sr.right)
		inner.numKeys++
		inner.rebuildChildLoaded()
		inner.miniTreeDirty = true
		return insertResult{updated: res.updated, oldPayload: res.oldPayload}
	}

	// Inner node is full (numKeys == B-1) — need to split.
	// Scratch arrays for the overflow entries live on the stack; the
	// five make() calls that preceded this were five heap allocations
	// per inner split. splitInner copies out of these slices, so the
	// backing arrays do not escape. See Finding #21.
	var allKeys [B][]byte
	var allChildNodes [B + 1]Node
	var allChildHashes [B + 1]Hash
	var allSizes [B + 1]int64

	// Copy existing, inserting at childIdx
	copy(allKeys[:childIdx], inner.keys[:childIdx])
	allKeys[childIdx] = sr.separator
	copy(allKeys[childIdx+1:], inner.keys[childIdx:B-1])

	copy(allChildNodes[:childIdx+1], inner.childNodes[:childIdx+1])
	allChildNodes[childIdx+1] = sr.right
	copy(allChildNodes[childIdx+2:], inner.childNodes[childIdx+1:B])

	copy(allChildHashes[:childIdx+1], inner.childHashes[:childIdx+1])
	allChildHashes[childIdx+1] = rightChildHash
	copy(allChildHashes[childIdx+2:], inner.childHashes[childIdx+1:B])

	// Build sizes from childSizes — the split child's sizes are known from
	// the in-memory nodes; all others come from childSizes (no disk read).
	copy(allSizes[:childIdx], inner.childSizes[:childIdx])
	allSizes[childIdx] = nodeSize(child)      // post-split left (loaded)
	allSizes[childIdx+1] = nodeSize(sr.right) // new right (in memory)
	copy(allSizes[childIdx+2:], inner.childSizes[childIdx+1:B])

	// Build serialized child refs — preserve refs for unloaded children.
	// Copy stops at childIdx (exclusive) because positions childIdx and
	// childIdx+1 are assigned explicitly below; the old `copy(..., [:childIdx+1])`
	// wrote position childIdx twice. See Finding #27.
	var allChildRefs [B + 1][]byte
	copy(allChildRefs[:childIdx], inner.children[:childIdx])
	allChildRefs[childIdx] = nil   // loaded & cloned; ref is stale
	allChildRefs[childIdx+1] = nil // sr.right is new; no serialized ref
	copy(allChildRefs[childIdx+2:], inner.children[childIdx+1:B])
	leftInner, innerSR := splitInner(allKeys[:], allChildRefs[:], allChildHashes[:], inner.height, allSizes[:])

	// Replace *inner's data with leftInner's, preserving the existing
	// childMu (must not be copied) and ndb (leftInner has none). Using
	// explicit field assignment avoids copying sync.Mutex, which vet
	// correctly flags as unsafe even when both locks are idle here.
	// See Finding #24.
	savedNdb := inner.ndb
	inner.nodeKey = leftInner.nodeKey
	inner.numKeys = leftInner.numKeys
	inner.childSizes = leftInner.childSizes
	inner.height = leftInner.height
	inner.keys = leftInner.keys
	inner.children = leftInner.children
	inner.childHashes = leftInner.childHashes
	inner.miniTree = leftInner.miniTree
	// inner.ndb preserved (leftInner was built without ndb).
	// inner.childMu preserved (sync.Mutex must never be copied).
	for i := 0; i < inner.NumChildren(); i++ {
		inner.childNodes[i] = allChildNodes[i]
	}
	// Zero out trailing slots from the pre-split state so the bitmap
	// rebuilt below doesn't accidentally keep them marked loaded.
	for i := inner.NumChildren(); i < B; i++ {
		inner.childNodes[i] = nil
	}
	inner.rebuildChildLoaded()
	inner.miniTreeDirty = true

	// Wire up child nodes for right
	rightInner := innerSR.right.(*InnerNode)
	rightInner.ndb = savedNdb
	splitIdx := int(inner.numKeys) + 1 // separator was consumed
	for i := 0; i < rightInner.NumChildren(); i++ {
		rightInner.childNodes[i] = allChildNodes[splitIdx+i]
	}
	rightInner.rebuildChildLoaded()
	rightInner.miniTreeDirty = true

	return insertResult{updated: res.updated, oldPayload: res.oldPayload, split: &innerSR}
}

func cloneNode(n Node) Node {
	switch n := n.(type) {
	case *InnerNode:
		return n.Clone()
	case *LeafNode:
		return n.Clone()
	default:
		panic("unknown node type")
	}
}

func nodeHeight(n Node) int16 {
	switch n := n.(type) {
	case *InnerNode:
		return n.height
	case *LeafNode:
		return 0
	default:
		panic("unknown node type")
	}
}

func nodeSize(n Node) int64 {
	switch n := n.(type) {
	case *InnerNode:
		var s int64
		for i := 0; i < n.NumChildren(); i++ {
			s += n.childSizes[i]
		}
		return s
	case *LeafNode:
		return int64(n.numKeys)
	default:
		panic("unknown node type")
	}
}
