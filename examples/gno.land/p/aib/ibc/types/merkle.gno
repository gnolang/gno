package types

// MerklePath is the path used to verify commitment proofs, which can be an
// arbitrary structured object (defined by a commitment type).
// ICS-23 verification supports membership proofs for nested merkle trees.
// The ICS-24 standard provable keys MUST be stored in the lowest level tree
// with an optional prefix.
// The IC24 provable tree may then be stored in a higher level tree(s) that
// hash up to the root hash stored in the consensus state of the client.
// Each element of the path represents the key of a merkle tree from the root
// to the leaf.
// The elements of the path before the final element must be the path to the
// tree that contains the ICS24 provable store. Thus, it should remain constant
// for all ICS24 proofs.
// The final element of the path is the key of the leaf in the ICS24 provable
// store, thus IBC core will append the ICS24 path to the final element of the
// MerklePath stored in the counterparty to create the full path to the leaf
// for proof verification.
//
// Examples:
// Cosmos SDK:
// The Cosmos SDK commits to a multi-tree where each store is an IAVL tree and
// all store hashes are hashed in a simple merkle tree to get the final root
// hash. Thus, the MerklePath in the counterparty MerklePrefix has the
// following structure: ["ibc", ""]
// The core IBC handler will append the ICS24 path to the final element of the
// MerklePath like so: ["ibc", "{packetCommitmentPath}"] which will then be
// used for final verification.
type MerklePath struct {
	KeyPath [][]byte
}

// BuildMerklePath takes the merkle path prefix and an ICS24 path and builds a
// new path by appending the ICS24 path to the last element of the merkle path
// prefix.
func BuildMerklePath(prefix [][]byte, path []byte) MerklePath {
	prefixLength := len(prefix)
	if prefixLength == 0 {
		panic("cannot build merkle path with empty prefix")
	}

	// copy prefix to avoid modifying the original slice
	fullPath := make([][]byte, len(prefix))
	for i, p := range prefix {
		fullPath[i] = make([]byte, len(p))
		copy(fullPath[i], p)
	}
	// append path to last element
	fullPath[prefixLength-1] = append(fullPath[prefixLength-1], path...)
	return MerklePath{KeyPath: fullPath}
}
