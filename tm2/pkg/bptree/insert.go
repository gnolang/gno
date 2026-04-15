package bptree

import "sync"

// insertResult is returned by recursive insert functions.
type insertResult struct {
	updated     bool         // true if key already existed (value replaced)
	split       *splitResult // non-nil if the node split
	oldValueKey []byte       // old valueKey if key was updated (for orphan tracking)
}

// treeInsert inserts a key with a pre-computed value hash and valueKey.
// Returns the (possibly new) root, whether it was an update, and the old
// valueKey if an existing key was overwritten (nil otherwise).
func treeInsert(root Node, key []byte, valueHash Hash, valueKey []byte) (Node, bool, []byte) {
	key = copyKey(key) // defensive copy — caller may reuse the slice

	// COW-clone the root
	root = cloneNode(root)
	res := nodeInsert(root, key, valueHash, valueKey)

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
		newRoot.RebuildMiniMerkle()
		return newRoot, res.updated, res.oldValueKey
	}
	return root, res.updated, res.oldValueKey
}

// nodeInsert recursively inserts into the subtree rooted at node.
// The node must already be COW-cloned by the caller.
func nodeInsert(node Node, key []byte, valueHash Hash, valueKey []byte) insertResult {
	switch n := node.(type) {
	case *LeafNode:
		return leafInsert(n, key, valueHash, valueKey)
	case *InnerNode:
		return innerInsert(n, key, valueHash, valueKey)
	default:
		panic("unknown node type")
	}
}

// leafInsert inserts into a leaf node (already COW-cloned).
func leafInsert(leaf *LeafNode, key []byte, valueHash Hash, valueKey []byte) insertResult {
	pos, found := searchLeaf(leaf, key)

	if found {
		// Update existing key — size unchanged. Capture old valueKey for orphan tracking.
		oldVK := leaf.valueKeys[pos]
		leaf.valueHashes[pos] = valueHash
		leaf.valueKeys[pos] = valueKey
		leaf.miniTree.SetSlot(pos, HashLeafSlotFromValueHash(key, valueHash))
		return insertResult{updated: true, oldValueKey: oldVK}
	}

	// Insert new key
	if int(leaf.numKeys) < B {
		// Room in this leaf — shift right and insert
		n := int(leaf.numKeys)
		for i := n; i > pos; i-- {
			leaf.keys[i] = leaf.keys[i-1]
			leaf.valueHashes[i] = leaf.valueHashes[i-1]
			leaf.valueKeys[i] = leaf.valueKeys[i-1]
		}
		leaf.keys[pos] = key
		leaf.valueHashes[pos] = valueHash
		leaf.valueKeys[pos] = valueKey
		leaf.numKeys++
		leaf.RebuildMiniMerkle()
		return insertResult{updated: false}
	}

	// Leaf is full (numKeys == B) — need to split
	allKeys := make([][]byte, B+1)
	allVH := make([]Hash, B+1)
	allVK := make([][]byte, B+1)
	// Copy existing keys, inserting new key at pos
	copy(allKeys[:pos], leaf.keys[:pos])
	allKeys[pos] = key
	copy(allKeys[pos+1:], leaf.keys[pos:B])
	copy(allVH[:pos], leaf.valueHashes[:pos])
	allVH[pos] = valueHash
	copy(allVH[pos+1:], leaf.valueHashes[pos:B])
	copy(allVK[:pos], leaf.valueKeys[:pos])
	allVK[pos] = valueKey
	copy(allVK[pos+1:], leaf.valueKeys[pos:B])

	left, sr := splitLeaf(allKeys, allVH, allVK, pos)
	*leaf = *left
	leaf.RebuildMiniMerkle()
	sr.right.(*LeafNode).RebuildMiniMerkle()
	return insertResult{updated: false, split: &sr}
}

// innerInsert inserts into an inner node (already COW-cloned).
func innerInsert(inner *InnerNode, key []byte, valueHash Hash, valueKey []byte) insertResult {
	childIdx := searchInner(inner, key)

	child := inner.getChild(childIdx)
	if child == nil {
		panic("inner node has nil child")
	}

	// COW-clone the child
	child = cloneNode(child)
	inner.setChild(childIdx, child)

	// Recurse
	res := nodeInsert(child, key, valueHash, valueKey)

	if !res.updated {
		inner.childSizes[childIdx]++
	}

	// Update child hash
	inner.childHashes[childIdx] = child.Hash()

	if res.split == nil {
		inner.miniTree.SetSlot(childIdx, inner.childHashes[childIdx])
		return insertResult{updated: res.updated, oldValueKey: res.oldValueKey}
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
		inner.RebuildMiniMerkle()
		return insertResult{updated: res.updated, oldValueKey: res.oldValueKey}
	}

	// Inner node is full (numKeys == B-1) — need to split.
	// Build allSizes from childSizes (no child loading needed).
	allKeys := make([][]byte, B)
	allChildNodes := make([]Node, B+1)
	allChildHashes := make([]Hash, B+1)
	allSizes := make([]int64, B+1)

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

	// Build serialized child refs — preserve refs for unloaded children
	allChildRefs := make([][]byte, B+1)
	copy(allChildRefs[:childIdx+1], inner.children[:childIdx+1])
	allChildRefs[childIdx] = nil   // loaded & cloned; ref is stale
	allChildRefs[childIdx+1] = nil // sr.right is new; no serialized ref
	copy(allChildRefs[childIdx+2:], inner.children[childIdx+1:B])
	leftInner, innerSR := splitInner(allKeys, allChildRefs, allChildHashes, inner.height, allSizes)

	// Wire up child nodes for left, preserving ndb from the original node
	savedNdb := inner.ndb
	*inner = *leftInner //nolint:govet // intentional copy; mutex re-initialized below
	inner.childMu = sync.Mutex{}
	inner.ndb = savedNdb
	for i := 0; i < inner.NumChildren(); i++ {
		inner.childNodes[i] = allChildNodes[i]
	}
	inner.RebuildMiniMerkle()

	// Wire up child nodes for right
	rightInner := innerSR.right.(*InnerNode)
	rightInner.ndb = savedNdb
	splitIdx := int(inner.numKeys) + 1 // separator was consumed
	for i := 0; i < rightInner.NumChildren(); i++ {
		rightInner.childNodes[i] = allChildNodes[splitIdx+i]
	}
	rightInner.RebuildMiniMerkle()

	return insertResult{updated: res.updated, oldValueKey: res.oldValueKey, split: &innerSR}
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
