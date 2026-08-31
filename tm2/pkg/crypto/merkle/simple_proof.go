package merkle

import (
	"bytes"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/crypto/tmhash"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

const (
	maxAunts = 100
)

// SimpleProof represents a simple Merkle proof.
// NOTE: The convention for proofs is to include leaf hashes but to
// exclude the root hash.
// This convention is implemented across IAVL range proofs as well.
// Keep this consistent unless there's a very good reason to change
// everything.  This also affects the generalized proof system as
// well.
type SimpleProof struct {
	Total    int      `json:"total"`     // Total number of items.
	Index    int      `json:"index"`     // Index of item to prove.
	LeafHash []byte   `json:"leaf_hash"` // Hash of item value.
	Aunts    [][]byte `json:"aunts"`     // Hashes from leaf's sibling to a root's child.
}

// SimpleProofsFromByteSlices computes inclusion proof for given items.
// proofs[0] is the proof for items[0].
func SimpleProofsFromByteSlices(items [][]byte) (rootHash []byte, proofs []*SimpleProof) {
	trails, rootSPN := trailsFromByteSlices(items)
	rootHash = rootSPN.Hash
	proofs = make([]*SimpleProof, len(items))
	for i, trail := range trails {
		proofs[i] = &SimpleProof{
			Total:    len(items),
			Index:    i,
			LeafHash: trail.Hash,
			Aunts:    trail.FlattenAunts(),
		}
	}
	return
}

// SimpleProofsFromMap generates proofs from a map. The keys/values of the map will be used as the keys/values
// in the underlying key-value pairs.
// The keys are sorted before the proofs are computed.
func SimpleProofsFromMap(m map[string][]byte) (rootHash []byte, proofs map[string]*SimpleProof, keys []string) {
	sm := newSimpleMap()
	for k, v := range m {
		sm.Set(k, v)
	}
	sm.Sort()
	kvs := sm.kvs
	kvsBytes := make([][]byte, len(kvs))
	for i, kvp := range kvs {
		kvsBytes[i] = KVPair(kvp).Bytes()
	}

	rootHash, proofList := SimpleProofsFromByteSlices(kvsBytes)
	proofs = make(map[string]*SimpleProof)
	keys = make([]string, len(proofList))
	for i, kvp := range kvs {
		proofs[string(kvp.Key)] = proofList[i]
		keys[i] = string(kvp.Key)
	}
	return
}

// Verify that the SimpleProof proves the root hash.
// Check sp.Index/sp.Total manually if needed
func (sp *SimpleProof) Verify(rootHash []byte, leaf []byte) error {
	leafHash := leafHash(leaf)
	if sp.Total <= 0 {
		return errors.New("Proof total must be positive, got %d", sp.Total)
	}
	if sp.Index < 0 || sp.Index >= sp.Total {
		return errors.New("Proof index %d out of range for total %d", sp.Index, sp.Total)
	}
	if !bytes.Equal(sp.LeafHash, leafHash) {
		return errors.New("invalid leaf hash: wanted %X got %X", leafHash, sp.LeafHash)
	}
	computedHash, err := sp.ComputeRootHash()
	if err != nil {
		return errors.Wrap(err, "cannot compute root hash")
	}
	if !bytes.Equal(computedHash, rootHash) {
		return errors.New("invalid root hash: wanted %X got %X", rootHash, computedHash)
	}
	return nil
}

// ComputeRootHash computes the root hash implied by the proof's leaf hash and
// sibling hashes. It does not verify the result against an expected root.
//
// It returns an error, rather than a nil hash, when the proof's shape does not
// let a root be computed at all. Callers compare the returned hash against an
// expected root, and bytes.Equal treats a nil hash as equal to an empty one —
// so a nil return would make an unverifiable proof "match" an empty expected
// root.
func (sp *SimpleProof) ComputeRootHash() ([]byte, error) {
	return computeHashFromAunts(
		sp.Index,
		sp.Total,
		sp.LeafHash,
		sp.Aunts,
	)
}

// String implements the stringer interface for SimpleProof.
// It is a wrapper around StringIndented.
func (sp *SimpleProof) String() string {
	return sp.StringIndented("")
}

// StringIndented generates a canonical string representation of a SimpleProof.
func (sp *SimpleProof) StringIndented(indent string) string {
	return fmt.Sprintf(`SimpleProof{
%s  Aunts: %X
%s}`,
		indent, sp.Aunts,
		indent)
}

// ValidateBasic performs basic validation.
// NOTE: - it expects LeafHash and Aunts of tmhash.Size size
//   - it expects no more than 100 aunts
func (sp *SimpleProof) ValidateBasic() error {
	if sp.Total < 0 {
		return errors.New("negative Total")
	}
	if sp.Index < 0 {
		return errors.New("negative Index")
	}
	if len(sp.LeafHash) != tmhash.Size {
		return errors.New("expected LeafHash size to be %d, got %d", tmhash.Size, len(sp.LeafHash))
	}
	if len(sp.Aunts) > maxAunts {
		return errors.New("expected no more than %d aunts, got %d", maxAunts, len(sp.Aunts))
	}
	for i, auntHash := range sp.Aunts {
		if len(auntHash) != tmhash.Size {
			return errors.New("expected Aunts#%d size to be %d, got %d", i, tmhash.Size, len(auntHash))
		}
	}
	return nil
}

// computeHashFromAunts walks the sibling hashes from leaf to root. Each way it
// can fail returns a distinct error: a nil hash would be indistinguishable from
// an empty one at the caller's bytes.Equal, which is what let malformed proofs
// verify against an empty expected root.
func computeHashFromAunts(index int, total int, leafHash []byte, innerHashes [][]byte) ([]byte, error) {
	if total <= 0 {
		return nil, errors.New("total must be positive, got %d", total)
	}
	if index < 0 || index >= total {
		return nil, errors.New("index %d out of range for total %d", index, total)
	}
	switch total {
	case 1:
		if len(innerHashes) != 0 {
			return nil, errors.New("single-leaf proof carries %d sibling hashes, want none", len(innerHashes))
		}
		return leafHash, nil
	default:
		if len(innerHashes) == 0 {
			return nil, errors.New("proof over %d leaves carries no sibling hashes", total)
		}
		numLeft := getSplitPoint(total)
		if index < numLeft {
			leftHash, err := computeHashFromAunts(index, numLeft, leafHash, innerHashes[:len(innerHashes)-1])
			if err != nil {
				return nil, err
			}
			return innerHash(leftHash, innerHashes[len(innerHashes)-1]), nil
		}
		rightHash, err := computeHashFromAunts(index-numLeft, total-numLeft, leafHash, innerHashes[:len(innerHashes)-1])
		if err != nil {
			return nil, err
		}
		return innerHash(innerHashes[len(innerHashes)-1], rightHash), nil
	}
}

// SimpleProofNode is a helper structure to construct merkle proof.
// The node and the tree is thrown away afterwards.
// Exactly one of node.Left and node.Right is nil, unless node is the root, in which case both are nil.
// node.Parent.Hash = hash(node.Hash, node.Right.Hash) or
// hash(node.Left.Hash, node.Hash), depending on whether node is a left/right child.
type SimpleProofNode struct {
	Hash   []byte
	Parent *SimpleProofNode
	Left   *SimpleProofNode // Left sibling  (only one of Left,Right is set)
	Right  *SimpleProofNode // Right sibling (only one of Left,Right is set)
}

// FlattenAunts will return the inner hashes for the item corresponding to the leaf,
// starting from a leaf SimpleProofNode.
func (spn *SimpleProofNode) FlattenAunts() [][]byte {
	// Nonrecursive impl.
	innerHashes := [][]byte{}
	for spn != nil {
		switch {
		case spn.Left != nil:
			innerHashes = append(innerHashes, spn.Left.Hash)
		case spn.Right != nil:
			innerHashes = append(innerHashes, spn.Right.Hash)
		default:
			break
		}
		spn = spn.Parent
	}
	return innerHashes
}

// trails[0].Hash is the leaf hash for items[0].
// trails[i].Parent.Parent....Parent == root for all i.
func trailsFromByteSlices(items [][]byte) (trails []*SimpleProofNode, root *SimpleProofNode) {
	// Recursive impl.
	switch len(items) {
	case 0:
		return nil, nil
	case 1:
		trail := &SimpleProofNode{leafHash(items[0]), nil, nil, nil}
		return []*SimpleProofNode{trail}, trail
	default:
		k := getSplitPoint(len(items))
		lefts, leftRoot := trailsFromByteSlices(items[:k])
		rights, rightRoot := trailsFromByteSlices(items[k:])
		rootHash := innerHash(leftRoot.Hash, rightRoot.Hash)
		root := &SimpleProofNode{rootHash, nil, nil, nil}
		leftRoot.Parent = root
		leftRoot.Right = rightRoot
		rightRoot.Parent = root
		rightRoot.Left = leftRoot
		return append(lefts, rights...), root
	}
}
