package ics23

import "encoding/hex"

// CommitmentRoot is a byte slice that represents the merkle root of a tree
// that can be used to validate proofs
type CommitmentRoot []byte

// VerifyMembership returns true iff
// proof is an ExistenceProof for the given key and value AND
// calculating the root for the ExistenceProof matches the provided
// CommitmentRoot.
func VerifyMembership(spec *ProofSpec, root CommitmentRoot,
	proof *CommitmentProof_Exist, key []byte, value []byte) bool {
	if proof == nil {
		return false
	}
	if proof.Exist == nil {
		return false
	}
	return proof.Exist.Verify(spec, root, key, value) == nil
}

// VerifyNonMembership returns true iff
// proof is (contains) a NonExistenceProof
// both left and right sub-proofs are valid existence proofs (see above) or nil
// left and right proofs are neighbors (or left/right most if one is nil)
// provided key is between the keys of the two proofs
func VerifyNonMembership(spec *ProofSpec, root CommitmentRoot,
	proof *CommitmentProof_Nonexist, key []byte) bool {
	if proof == nil {
		return false
	}
	if proof.Nonexist == nil {
		return false
	}
	return proof.Nonexist.Verify(spec, root, key) == nil
}

// TODO move this elsewhere
func NewExistenceProof(spec *ProofSpec, key []byte, value, leaf string, suffixes, prefixes []string) *ExistenceProof {
	ep := &ExistenceProof{
		Key:   key,
		Value: hexDec(value),
		Leaf: &LeafOp{
			Hash:         spec.LeafSpec.Hash,
			PrehashKey:   spec.LeafSpec.PrehashKey,
			PrehashValue: spec.LeafSpec.PrehashValue,
			Length:       spec.LeafSpec.Length,
			Prefix:       hexDec(leaf),
		},
	}
	for i, _ := range prefixes {
		ep.Path = append(ep.Path, &InnerOp{
			Hash:   spec.InnerSpec.Hash,
			Suffix: hexDec(suffixes[i]),
			Prefix: hexDec(prefixes[i]),
		})
	}
	return ep
}

func hexDec(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}
