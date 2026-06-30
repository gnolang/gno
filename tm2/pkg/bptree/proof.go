package bptree

import (
	"fmt"

	ics23 "github.com/cosmos/ics23/go"
)

// GetMembershipProof generates an ICS23 existence proof for a key.
// Proofs over empty-valued keys generate but can never verify: ics23's LeafOp
// rejects empty values (the value-side twin of the empty-key constraint;
// IAVL behaves identically). See TestProof_EmptyValueUnprovable.
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
// The proof embeds the gap's neighbor existence proofs, so it cannot verify
// when an adjacent key holds an empty value (see GetMembershipProof).
func (t *ImmutableTree) GetNonMembershipProof(key []byte) (*ics23.CommitmentProof, error) {
	if t.root == nil {
		return nil, ErrEmptyTree
	}

	// Verify the key doesn't exist
	has, err := t.Has(key)
	if err != nil {
		return nil, err
	}
	if has {
		return nil, fmt.Errorf("key exists, cannot create non-membership proof")
	}

	nonexist := &ics23.NonExistenceProof{Key: key}

	// Find the left neighbor (greatest key < key)
	idx, _, err := t.GetWithIndex(key)
	if err != nil {
		return nil, fmt.Errorf("GetWithIndex: %w", err)
	}
	if idx > 0 {
		leftKey, _, err := t.GetByIndex(idx - 1)
		if err != nil {
			return nil, fmt.Errorf("left neighbor GetByIndex(%d): %w", idx-1, err)
		}
		nonexist.Left, err = t.createExistenceProof(leftKey)
		if err != nil {
			return nil, fmt.Errorf("left neighbor proof: %w", err)
		}
	}

	// Find the right neighbor (smallest key > key)
	if idx < t.Size() {
		rightKey, _, err := t.GetByIndex(idx)
		if err != nil {
			return nil, fmt.Errorf("right neighbor GetByIndex(%d): %w", idx, err)
		}
		nonexist.Right, err = t.createExistenceProof(rightKey)
		if err != nil {
			return nil, fmt.Errorf("right neighbor proof: %w", err)
		}
	}

	return &ics23.CommitmentProof{
		Proof: &ics23.CommitmentProof_Nonexist{Nonexist: nonexist},
	}, nil
}

// VerifyMembership verifies an ICS23 existence proof against the tree's root hash.
// The value is taken from the proof itself (no tree lookup needed).
func (t *ImmutableTree) VerifyMembership(proof *ics23.CommitmentProof, key []byte) (bool, error) {
	// Reject a nil / nil-inner / wrong-type proof before ics23 (which would
	// nil-deref on it); this also lets exist.Value be read safely below.
	exist := proof.GetExist()
	if exist == nil {
		return false, fmt.Errorf("proof is not an existence proof")
	}
	root := t.Hash()
	return ics23.VerifyMembership(BptreeSpec, root, proof, key, exist.Value), nil
}

// VerifyNonMembership verifies an ICS23 non-existence proof.
func (t *ImmutableTree) VerifyNonMembership(proof *ics23.CommitmentProof, key []byte) (bool, error) {
	// Reject a non-existence proof with a nil inner (or a nil/wrong-type
	// proof) before handing it to ics23, which would nil-deref on it. Mirrors
	// the GetExist guard in VerifyMembership; not reachable from a decoded
	// (wire) proof, but keeps this independent of ics23's unmarshal details.
	if proof.GetNonexist() == nil {
		return false, fmt.Errorf("proof is not a non-existence proof")
	}
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

	// Find the key and collect the path from root to leaf
	path, leafSlotIdx, _, err := t.findPathToKey(key)
	if err != nil {
		return nil, err
	}

	// For ICS23, we need the raw value. The tree only stores the hash.
	if t.valueResolver == nil {
		return nil, fmt.Errorf("cannot create existence proof without a value resolver")
	}
	leafNode := path[len(path)-1].node.(*LeafNode)
	vk := leafNode.valueKeys[leafSlotIdx]
	rawValue, err := t.valueResolver(vk)
	if err != nil {
		return nil, fmt.Errorf("resolving value for proof: %w", err)
	}

	// Build the ICS23 InnerOps from the path
	// The path goes from root to leaf. ICS23 expects leaf-to-root order.
	var innerOps []*ics23.InnerOp

	// 1. Mini merkle ops within the leaf node
	leaf := path[len(path)-1].node.(*LeafNode)
	leafOps := miniMerkleInnerOps(&leaf.miniTree, leafSlotIdx)
	innerOps = append(innerOps, leafOps...)

	// 2. Mini merkle ops for each inner node, from leaf's parent to root
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
			child, err := n.getChild(childIdx)
			if err != nil {
				return nil, 0, Hash{}, err
			}
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
	ops := make([]*ics23.InnerOp, len(siblings))

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

// GetMembershipProof returns an ICS23 existence proof for key against the last
// committed version (verifiable against MutableTree.Hash()), not the working
// tree. Returns ErrNoCommittedState if no version has been committed yet, or
// ErrEmptyTree if the committed tree is empty.
//
// Single-goroutine: it reads the working tree's lastSaved/version (see the
// MutableTree concurrency contract). For concurrent or historical proofs, use
// GetImmutable(version).GetMembershipProof instead.
func (t *MutableTree) GetMembershipProof(key []byte) (*ics23.CommitmentProof, error) {
	imm, err := t.immutableForProof()
	if err != nil {
		return nil, err
	}
	defer imm.Close()
	return imm.GetMembershipProof(key)
}

// GetNonMembershipProof returns an ICS23 non-existence proof for key against the
// last committed version (verifiable against MutableTree.Hash()), not the working
// tree. Returns ErrNoCommittedState if no version has been committed yet, or
// ErrEmptyTree if the committed tree is empty.
//
// Single-goroutine: it reads the working tree's lastSaved/version (see the
// MutableTree concurrency contract). For concurrent or historical proofs, use
// GetImmutable(version).GetNonMembershipProof instead.
func (t *MutableTree) GetNonMembershipProof(key []byte) (*ics23.CommitmentProof, error) {
	imm, err := t.immutableForProof()
	if err != nil {
		return nil, err
	}
	defer imm.Close()
	return imm.GetNonMembershipProof(key)
}

// immutableForProof creates an ImmutableTree from the last committed state
// with a value resolver for proof generation. Returns an error if no version
// has been committed yet.
func (t *MutableTree) immutableForProof() (*ImmutableTree, error) {
	// lastSaved is nil both when nothing has ever been committed and when the
	// committed tree is empty. Only the former is an error; a committed-but-empty
	// tree (version > 0) falls through, and ImmutableTree returns ErrEmptyTree.
	if t.lastSaved == nil && t.version == 0 {
		return nil, ErrNoCommittedState
	}
	// Register as a reader so a concurrent PruneVersionsTo(t.version) can't
	// delete nodes the proof will traverse (lastSaved's children reload from the
	// DB after clear-on-save). The caller MUST Close the returned snapshot.
	if t.version > 0 {
		t.ndb.incrVersionReaders(t.version)
	}
	// lastSaved is the last COMMITTED root (set after Commit), so its values are
	// durable in the DB. Resolve DB-only (committed=true) so a proof generated
	// concurrently with a later Set cannot race that Set's pendingVals write.
	imm := t.newImmutable(t.lastSaved, t.version, true)
	imm.registered = t.version > 0
	return imm, nil
}
