package bptree

import (
	"fmt"

	ics23 "github.com/cosmos/ics23/go"
)

// GetMembershipProof generates an ICS23 existence proof for a key.
func (t *ImmutableTree) GetMembershipProof(key []byte) (*ics23.CommitmentProof, error) {
	exist, err := t.createExistenceProof(key)
	if err != nil {
		return nil, err
	}
	return &ics23.CommitmentProof{
		Proof: &ics23.CommitmentProof_Exist{Exist: exist},
	}, nil
}

// GetNonMembershipProof generates an ICS23 non-existence proof for a key.
//
// The previous implementation performed six full tree walks: Has →
// GetWithIndex → GetByIndex(idx-1) → createExistenceProof(leftKey) →
// GetByIndex(idx) → createExistenceProof(rightKey). Each neighbor's
// existence proof re-walked the tree by key even though we had just
// walked to that index. The current implementation collapses the
// membership check, the neighbor lookups, and the path collection into
// three walks: one treeGetWithIndex for the absence check plus at most
// one findPathToIndex per neighbor. See Finding #21.
func (t *ImmutableTree) GetNonMembershipProof(key []byte) (*ics23.CommitmentProof, error) {
	if t.root == nil {
		return nil, ErrEmptyTree
	}

	// One walk: does key exist, and what is its would-be index?
	idx, _, _, found := treeGetWithIndex(t.root, key)
	if found {
		return nil, fmt.Errorf("key exists, cannot create non-membership proof")
	}

	nonexist := &ics23.NonExistenceProof{Key: key}

	// Left neighbor (greatest key < key)
	if idx > 0 {
		proof, err := t.existenceProofAtIndex(idx - 1)
		if err != nil {
			return nil, fmt.Errorf("left neighbor proof: %w", err)
		}
		nonexist.Left = proof
	}

	// Right neighbor (smallest key > key)
	if idx < t.Size() {
		proof, err := t.existenceProofAtIndex(idx)
		if err != nil {
			return nil, fmt.Errorf("right neighbor proof: %w", err)
		}
		nonexist.Right = proof
	}

	return &ics23.CommitmentProof{
		Proof: &ics23.CommitmentProof_Nonexist{Nonexist: nonexist},
	}, nil
}

// VerifyMembership verifies an ICS23 existence proof against the tree's root hash.
// The value is taken from the proof itself (no tree lookup needed).
func (t *ImmutableTree) VerifyMembership(proof *ics23.CommitmentProof, key []byte) (bool, error) {
	exist := proof.GetExist()
	if exist == nil {
		return false, fmt.Errorf("proof is not an existence proof")
	}
	root := t.Hash()
	return ics23.VerifyMembership(BptreeSpec, root, proof, key, exist.Value), nil
}

// VerifyNonMembership verifies an ICS23 non-existence proof.
func (t *ImmutableTree) VerifyNonMembership(proof *ics23.CommitmentProof, key []byte) (bool, error) {
	root := t.Hash()
	return ics23.VerifyNonMembership(BptreeSpec, root, proof, key), nil
}

// createExistenceProof builds the ExistenceProof for a key that exists in the tree.
// The value parameter must be the raw value (NOT the hash). ICS23's LeafOp
// applies PrehashValue=SHA256 to get the value hash during verification.
func (t *ImmutableTree) createExistenceProof(key []byte) (*ics23.ExistenceProof, error) {
	if t.root == nil {
		return nil, ErrEmptyTree
	}

	path, leafSlotIdx, _, err := t.findPathToKey(key)
	if err != nil {
		return nil, err
	}
	return t.buildExistenceProofFromPath(path, leafSlotIdx)
}

// existenceProofAtIndex builds the ExistenceProof for the idx-th key in
// sorted order. Used by non-membership proofs to avoid the lookup-by-key
// second walk: the neighbor's index is already known from the
// treeGetWithIndex call, so we can descend directly by index and build
// the proof in one pass. See Finding #21.
func (t *ImmutableTree) existenceProofAtIndex(idx int64) (*ics23.ExistenceProof, error) {
	path, leafSlotIdx := findPathToIndex(t.root, idx)
	return t.buildExistenceProofFromPath(path, leafSlotIdx)
}

// buildExistenceProofFromPath assembles an ICS23 ExistenceProof given a
// root-to-leaf path and the slot index within the leaf. The path entries
// are consumed leaf-to-root when emitting InnerOps, matching ICS23's
// expected ordering.
func (t *ImmutableTree) buildExistenceProofFromPath(path []pathEntry, leafSlotIdx int) (*ics23.ExistenceProof, error) {
	// For ICS23, we need the raw value. The tree only stores the hash.
	if t.valueResolver == nil {
		return nil, fmt.Errorf("cannot create existence proof without a value resolver")
	}
	leaf := path[len(path)-1].node.(*LeafNode)
	key := leaf.keys[leafSlotIdx]
	vk := leaf.valueKeys[leafSlotIdx]
	rawValue, err := t.valueResolver(vk)
	if err != nil {
		return nil, fmt.Errorf("resolving value for proof: %w", err)
	}

	// Build the ICS23 InnerOps from the path.
	// The path goes from root to leaf. ICS23 expects leaf-to-root order.
	// Each level contributes MiniMerkleDepth ops; preallocate exactly.
	innerOps := make([]*ics23.InnerOp, 0, len(path)*MiniMerkleDepth)

	// 1. Mini merkle ops within the leaf node.
	leafOps := miniMerkleInnerOps(&leaf.miniTree, leafSlotIdx)
	innerOps = append(innerOps, leafOps...)

	// 2. Mini merkle ops for each inner node, from leaf's parent to root.
	for i := len(path) - 2; i >= 0; i-- {
		inner := path[i].node.(*InnerNode)
		childIdx := path[i].childIdx
		ops := miniMerkleInnerOps(&inner.miniTree, childIdx)
		innerOps = append(innerOps, ops...)
	}

	return &ics23.ExistenceProof{
		Key:   key,
		Value: rawValue,
		Leaf:  BptreeSpec.LeafSpec,
		Path:  innerOps,
	}, nil
}

// findPathToIndex descends to the leaf holding the idx-th key in sorted
// order, collecting the root-to-leaf path. Returns the path and the slot
// index within the leaf. The caller guarantees 0 <= idx < nodeSize(root).
func findPathToIndex(root Node, idx int64) ([]pathEntry, int) {
	var path []pathEntry
	node := root
	for {
		switch n := node.(type) {
		case *LeafNode:
			path = append(path, pathEntry{node: n, childIdx: -1})
			return path, int(idx)
		case *InnerNode:
			offset := int64(0)
			for i := 0; i < n.NumChildren(); i++ {
				childSize := n.childSizes[i]
				if idx < offset+childSize {
					path = append(path, pathEntry{node: n, childIdx: i})
					child := n.getChild(i)
					if child == nil {
						panic("nil child at inner node during index descent")
					}
					node = child
					idx -= offset
					goto next
				}
				offset += childSize
			}
			panic("index out of range in findPathToIndex")
		default:
			panic("unknown node type")
		}
	next:
	}
}

// pathEntry records a node and which child was descended into.
type pathEntry struct {
	node     Node
	childIdx int // which child was followed (-1 for leaf)
}

// findPathToKey traverses the tree to find the key, collecting the path.
// Returns the path (root first, leaf last), the slot index within the leaf,
// and the value hash (NOT the raw value — the caller must resolve the raw
// value separately since ICS23 ExistenceProof.Value must be the raw value,
// and the LeafOp applies PrehashValue=SHA256 to get the value hash).
func (t *ImmutableTree) findPathToKey(key []byte) ([]pathEntry, int, Hash, error) {
	var path []pathEntry
	node := t.root

	for {
		switch n := node.(type) {
		case *LeafNode:
			pos, found := searchLeaf(n, key)
			if !found {
				return nil, 0, Hash{}, fmt.Errorf("key not found for existence proof")
			}
			path = append(path, pathEntry{node: n, childIdx: -1})
			return path, pos, n.valueHashes[pos], nil

		case *InnerNode:
			childIdx := searchInner(n, key)
			path = append(path, pathEntry{node: n, childIdx: childIdx})
			child := n.getChild(childIdx)
			if child == nil {
				return nil, 0, Hash{}, fmt.Errorf("nil child at inner node")
			}
			node = child

		default:
			return nil, 0, Hash{}, fmt.Errorf("unknown node type")
		}
	}
}

// miniMerkleInnerOps generates the ICS23 InnerOps for proving that
// slot[index] is part of the mini merkle root. The ops go from the
// leaf level of the mini merkle toward the root (bottom-up).
func miniMerkleInnerOps(m *MiniMerkle, index int) []*ics23.InnerOp {
	siblings, positions := m.SiblingPath(index)
	ops := make([]*ics23.InnerOp, MiniMerkleDepth)

	for i, sib := range siblings {
		op := &ics23.InnerOp{Hash: ics23.HashOp_SHA256}
		if positions[i] == 0 {
			// Proven node is the left child.
			// hash(0x01 || child || sibling)
			// prefix = 0x01, suffix = sibling
			op.Prefix = []byte{DomainInner}
			op.Suffix = sib[:]
		} else {
			// Proven node is the right child.
			// hash(0x01 || sibling || child)
			// prefix = 0x01 || sibling, suffix = empty
			op.Prefix = make([]byte, 1+HashSize)
			op.Prefix[0] = DomainInner
			copy(op.Prefix[1:], sib[:])
		}
		ops[i] = op
	}
	return ops
}

// --- MutableTree wrappers ---

func (t *MutableTree) GetMembershipProof(key []byte) (*ics23.CommitmentProof, error) {
	imm := t.immutableForProof()
	defer imm.Close()
	return imm.GetMembershipProof(key)
}

func (t *MutableTree) GetNonMembershipProof(key []byte) (*ics23.CommitmentProof, error) {
	imm := t.immutableForProof()
	defer imm.Close()
	return imm.GetNonMembershipProof(key)
}

// immutableForProof creates an ImmutableTree with a value resolver for proof
// generation. When backed by a nodeDB, it registers as an active version
// reader so a concurrent PruneVersionsTo(t.version) is rejected for the
// duration of proof generation. Callers MUST Close() the returned tree.
// See Finding #30.
//
// Note: the returned tree still shares t.root in memory, so concurrent
// mutations via Set/Remove on the MutableTree can still tear mini-merkle
// state. See Finding #9.
func (t *MutableTree) immutableForProof() *ImmutableTree {
	if t.ndb != nil {
		// Register as a reader FIRST so that a concurrent
		// PruneVersionsTo(t.version) cannot delete nodes that proof
		// generation will later traverse via getChild. See Findings
		// #30 and #40.
		if t.version > 0 {
			t.ndb.incrVersionReaders(t.version)
		}
		imm := &ImmutableTree{root: t.root, version: t.version}
		imm.ndb = t.ndb
		imm.valueResolver = func(vk []byte) ([]byte, error) {
			return t.ndb.GetValue(vk)
		}
		return imm
	}
	imm := &ImmutableTree{root: t.root, version: t.version}
	if t.memValues != nil {
		imm.valueResolver = func(vk []byte) ([]byte, error) {
			val, ok := t.memValues[string(vk)]
			if !ok {
				return nil, fmt.Errorf("value not found in memValues")
			}
			return val, nil
		}
	}
	return imm
}
