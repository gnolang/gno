package ics23

// GetSDKSpecs returns the proof specs of an SDK chain.
func GetSDKProofSpecs() []*ProofSpec {
	return []*ProofSpec{IavlSpec(), TendermintSpec()}
}

// IavlSpec constrains the format from proofs-iavl (iavl merkle proofs)
func IavlSpec() *ProofSpec {
	return &ProofSpec{
		LeafSpec: &LeafOp{
			Prefix:       []byte{0},
			PrehashKey:   HashOp_NO_HASH,
			Hash:         HashOp_SHA256,
			PrehashValue: HashOp_SHA256,
			Length:       LengthOp_VAR_PROTO,
		},
		InnerSpec: &InnerSpec{
			ChildOrder:      []int32{0, 1},
			MinPrefixLength: 4,
			MaxPrefixLength: 12,
			ChildSize:       33, // (with length byte)
			EmptyChild:      nil,
			Hash:            HashOp_SHA256,
		},
	}
}

// TendermintSpec constrains the format from proofs-tendermint (crypto/merkle SimpleProof)
func TendermintSpec() *ProofSpec {
	return &ProofSpec{
		LeafSpec: &LeafOp{
			Prefix:       []byte{0},
			PrehashKey:   HashOp_NO_HASH,
			Hash:         HashOp_SHA256,
			PrehashValue: HashOp_SHA256,
			Length:       LengthOp_VAR_PROTO,
		},
		InnerSpec: &InnerSpec{
			ChildOrder:      []int32{0, 1},
			MinPrefixLength: 1,
			MaxPrefixLength: 1,
			ChildSize:       32, // (no length byte)
			Hash:            HashOp_SHA256,
		},
	}
}

// ProofSpec defines what the expected parameters are for a given proof type.
// This can be stored in the client and used to validate any incoming proofs.
//
// verify(ProofSpec, Proof) -> Proof | Error
//
// As demonstrated in tests, if we don't fix the algorithm used to calculate the
// LeafHash for a given tree, there are many possible key-value pairs that can
// generate a given hash (by interpretting the preimage differently).
// We need this for proper security, requires client knows a priori what
// tree format server uses. But not in code, rather a configuration object.
type ProofSpec struct {
	// any field in the ExistenceProof must be the same as in this spec.
	// except Prefix, which is just the first bytes of prefix (spec can be longer)
	LeafSpec  *LeafOp
	InnerSpec *InnerSpec
	// max_depth (if > 0) is the maximum number of InnerOps allowed (mainly for fixed-depth tries)
	// the max_depth is interpreted as 128 if set to 0
	MaxDepth int32
	// min_depth (if > 0) is the minimum number of InnerOps allowed (mainly for fixed-depth tries)
	MinDepth int32
	// prehash_key_before_comparison is a flag that indicates whether to use the
	// prehash_key specified by LeafOp to compare lexical ordering of keys for
	// non-existence proofs.
	PrehashKeyBeforeComparison bool
}

// InnerSpec contains all store-specific structure info to determine if two proofs from a
// given store are neighbors.
//
// This enables:
//
// isLeftMost(spec: InnerSpec, op: InnerOp)
// isRightMost(spec: InnerSpec, op: InnerOp)
// isLeftNeighbor(spec: InnerSpec, left: InnerOp, right: InnerOp)
type InnerSpec struct {
	// Child order is the ordering of the children node, must count from 0
	// iavl tree is [0, 1] (left then right)
	// merk is [0, 2, 1] (left, right, here)
	ChildOrder      []int32
	ChildSize       int32
	MinPrefixLength int32
	MaxPrefixLength int32
	// empty child is the prehash image that is used when one child is nil (eg. 20 bytes of 0)
	EmptyChild []byte
	// hash is the algorithm that must be used for each InnerOp
	Hash HashOp
}
