package ics23

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
