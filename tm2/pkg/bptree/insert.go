package bptree

import (
	"crypto/sha256"
	"sync"
)

// insertResult is returned by recursive insert functions.
type insertResult struct {
	updated bool         // true if key already existed (value replaced)
	split   *splitResult // non-nil if the node split
}

// treeInsert inserts a key-value pair into the tree rooted at root.
// It returns the (possibly new) root and whether the key was an update.
// All nodes on the modification path are COW-cloned.
func treeInsert(root Node, key, value []byte) (Node, bool) {
	key = copyKey(key) // defensive copy — caller may reuse the slice
	valueHash := sha256.Sum256(value)

	// COW-clone the root
	root = cloneNode(root)
	res := nodeInsert(root, key, valueHash)

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
		newRoot.size = nodeSize(root) + nodeSize(sr.right)
		newRoot.RebuildMiniMerkle()
		return newRoot, res.updated
	}
	return root, res.updated
}

// nodeInsert recursively inserts into the subtree rooted at node.
// The node must already be COW-cloned by the caller.
func nodeInsert(node Node, key []byte, valueHash Hash) insertResult {
	switch n := node.(type) {
	case *LeafNode:
		return leafInsert(n, key, valueHash)
	case *InnerNode:
		return innerInsert(n, key, valueHash)
	default:
		panic("unknown node type")
	}
}

// leafInsert inserts into a leaf node (already COW-cloned).
func leafInsert(leaf *LeafNode, key []byte, valueHash Hash) insertResult {
	pos, found := searchLeaf(leaf, key)

	if found {
		// Update existing key — size unchanged
		leaf.valueHashes[pos] = valueHash
		leaf.miniTree.SetSlot(pos, HashLeafSlotFromValueHash(key, valueHash))
		return insertResult{updated: true}
	}

	// Insert new key
	if int(leaf.numKeys) < B {
		// Room in this leaf — shift right and insert
		n := int(leaf.numKeys)
		for i := n; i > pos; i-- {
			leaf.keys[i] = leaf.keys[i-1]
			leaf.valueHashes[i] = leaf.valueHashes[i-1]
		}
		leaf.keys[pos] = key
		leaf.valueHashes[pos] = valueHash
		leaf.numKeys++
		leaf.RebuildMiniMerkle()
		return insertResult{updated: false}
	}

	// Leaf is full (numKeys == B) — need to split
	allKeys := make([][]byte, B+1)
	allVH := make([]Hash, B+1)
	// Copy existing keys, inserting new key at pos
	copy(allKeys[:pos], leaf.keys[:pos])
	allKeys[pos] = key
	copy(allKeys[pos+1:], leaf.keys[pos:B])
	copy(allVH[:pos], leaf.valueHashes[:pos])
	allVH[pos] = valueHash
	copy(allVH[pos+1:], leaf.valueHashes[pos:B])

	left, sr := splitLeaf(allKeys, allVH, pos)
	*leaf = *left
	leaf.RebuildMiniMerkle()
	sr.right.(*LeafNode).RebuildMiniMerkle()
	return insertResult{updated: false, split: &sr}
}

// innerInsert inserts into an inner node (already COW-cloned).
func innerInsert(inner *InnerNode, key []byte, valueHash Hash) insertResult {
	childIdx := searchInner(inner, key)

	child := inner.getChild(childIdx)
	if child == nil {
		panic("inner node has nil child")
	}

	// COW-clone the child
	child = cloneNode(child)
	inner.setChild(childIdx, child)

	// Recurse
	res := nodeInsert(child, key, valueHash)

	if !res.updated {
		inner.size++
	}

	// Update child hash
	inner.childHashes[childIdx] = child.Hash()

	if res.split == nil {
		inner.miniTree.SetSlot(childIdx, inner.childHashes[childIdx])
		return insertResult{updated: res.updated}
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
		}
		inner.keys[childIdx] = sr.separator
		inner.childNodes[childIdx+1] = sr.right
		inner.children[childIdx+1] = nil
		inner.childHashes[childIdx+1] = rightChildHash
		inner.numKeys++
		inner.RebuildMiniMerkle()
		return insertResult{updated: res.updated}
	}

	// Inner node is full (numKeys == B-1) — need to split
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

	for i := 0; i < B+1; i++ {
		allSizes[i] = nodeSize(allChildNodes[i])
	}

	// Use children=nil for splitInner (in-memory, no serialized refs)
	allChildRefs := make([][]byte, B+1) // all nil
	leftInner, innerSR := splitInner(allKeys, allChildRefs, allChildHashes, inner.height, allSizes)

	// Wire up child nodes for left
	*inner = *leftInner //nolint:govet // intentional copy; mutex re-initialized below
	inner.childMu = sync.Mutex{}
	for i := 0; i < inner.NumChildren(); i++ {
		inner.childNodes[i] = allChildNodes[i]
	}
	inner.RebuildMiniMerkle()

	// Wire up child nodes for right
	rightInner := innerSR.right.(*InnerNode)
	splitIdx := int(inner.numKeys) + 1 // separator was consumed
	for i := 0; i < rightInner.NumChildren(); i++ {
		rightInner.childNodes[i] = allChildNodes[splitIdx+i]
	}
	rightInner.RebuildMiniMerkle()

	return insertResult{updated: res.updated, split: &innerSR}
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
		return n.size
	case *LeafNode:
		return int64(n.numKeys)
	default:
		panic("unknown node type")
	}
}
