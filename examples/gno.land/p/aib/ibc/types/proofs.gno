package types

// GetSDKSpecs returns the proof specs of an SDK chain.
func GetSDKProofSpecs() []*ProofSpec {
	return sdkSpecs
}

// var representing the proofspecs for an SDK chain
var sdkSpecs = []*ProofSpec{iavlSpec, tendermintSpec}

// IavlSpec constrains the format from proofs-iavl (iavl merkle proofs)
var iavlSpec = &ProofSpec{
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

// TendermintSpec constrains the format from proofs-tendermint (crypto/merkle SimpleProof)
var tendermintSpec = &ProofSpec{
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
	LeafSpec  *LeafOp    `protobuf:"bytes,1,opt,name=leaf_spec,json=leafSpec,proto3" json:"leaf_spec,omitempty"`
	InnerSpec *InnerSpec `protobuf:"bytes,2,opt,name=inner_spec,json=innerSpec,proto3" json:"inner_spec,omitempty"`
	// max_depth (if > 0) is the maximum number of InnerOps allowed (mainly for fixed-depth tries)
	// the max_depth is interpreted as 128 if set to 0
	MaxDepth int32 `protobuf:"varint,3,opt,name=max_depth,json=maxDepth,proto3" json:"max_depth,omitempty"`
	// min_depth (if > 0) is the minimum number of InnerOps allowed (mainly for fixed-depth tries)
	MinDepth int32 `protobuf:"varint,4,opt,name=min_depth,json=minDepth,proto3" json:"min_depth,omitempty"`
	// prehash_key_before_comparison is a flag that indicates whether to use the
	// prehash_key specified by LeafOp to compare lexical ordering of keys for
	// non-existence proofs.
	PrehashKeyBeforeComparison bool `protobuf:"varint,5,opt,name=prehash_key_before_comparison,json=prehashKeyBeforeComparison,proto3" json:"prehash_key_before_comparison,omitempty"`
}

// LeafOp represents the raw key-value data we wish to prove, and
// must be flexible to represent the internal transformation from
// the original key-value pairs into the basis hash, for many existing
// merkle trees.
//
// key and value are passed in. So that the signature of this operation is:
// leafOp(key, value) -> output
//
// To process this, first prehash the keys and values if needed (ANY means no hash in this case):
// hkey = prehashKey(key)
// hvalue = prehashValue(value)
//
// Then combine the bytes, and hash it
// output = hash(prefix || length(hkey) || hkey || length(hvalue) || hvalue)
type LeafOp struct {
	Hash         HashOp   `protobuf:"varint,1,opt,name=hash,proto3,enum=cosmos.ics23.v1.HashOp" json:"hash,omitempty"`
	PrehashKey   HashOp   `protobuf:"varint,2,opt,name=prehash_key,json=prehashKey,proto3,enum=cosmos.ics23.v1.HashOp" json:"prehash_key,omitempty"`
	PrehashValue HashOp   `protobuf:"varint,3,opt,name=prehash_value,json=prehashValue,proto3,enum=cosmos.ics23.v1.HashOp" json:"prehash_value,omitempty"`
	Length       LengthOp `protobuf:"varint,4,opt,name=length,proto3,enum=cosmos.ics23.v1.LengthOp" json:"length,omitempty"`
	// prefix is a fixed bytes that may optionally be included at the beginning to differentiate
	// a leaf node from an inner node.
	Prefix []byte `protobuf:"bytes,5,opt,name=prefix,proto3" json:"prefix,omitempty"`
}

type HashOp int32

const (
	// NO_HASH is the default if no data passed. Note this is an illegal argument some places.
	HashOp_NO_HASH     HashOp = 0
	HashOp_SHA256      HashOp = 1
	HashOp_SHA512      HashOp = 2
	HashOp_KECCAK256   HashOp = 3
	HashOp_RIPEMD160   HashOp = 4
	HashOp_BITCOIN     HashOp = 5
	HashOp_SHA512_256  HashOp = 6
	HashOp_BLAKE2B_512 HashOp = 7
	HashOp_BLAKE2S_256 HashOp = 8
	HashOp_BLAKE3      HashOp = 9
)

// LengthOp defines how to process the key and value of the LeafOp
// to include length information. After encoding the length with the given
// algorithm, the length will be prepended to the key and value bytes.
// (Each one with it's own encoded length)
type LengthOp int32

const (
	// NO_PREFIX don't include any length info
	LengthOp_NO_PREFIX LengthOp = 0
	// VAR_PROTO uses protobuf (and go-amino) varint encoding of the length
	LengthOp_VAR_PROTO LengthOp = 1
	// VAR_RLP uses rlp int encoding of the length
	LengthOp_VAR_RLP LengthOp = 2
	// FIXED32_BIG uses big-endian encoding of the length as a 32 bit integer
	LengthOp_FIXED32_BIG LengthOp = 3
	// FIXED32_LITTLE uses little-endian encoding of the length as a 32 bit integer
	LengthOp_FIXED32_LITTLE LengthOp = 4
	// FIXED64_BIG uses big-endian encoding of the length as a 64 bit integer
	LengthOp_FIXED64_BIG LengthOp = 5
	// FIXED64_LITTLE uses little-endian encoding of the length as a 64 bit integer
	LengthOp_FIXED64_LITTLE LengthOp = 6
	// REQUIRE_32_BYTES is like NONE, but will fail if the input is not exactly 32 bytes (sha256 output)
	LengthOp_REQUIRE_32_BYTES LengthOp = 7
	// REQUIRE_64_BYTES is like NONE, but will fail if the input is not exactly 64 bytes (sha512 output)
	LengthOp_REQUIRE_64_BYTES LengthOp = 8
)

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
	ChildOrder      []int32 `protobuf:"varint,1,rep,packed,name=child_order,json=childOrder,proto3" json:"child_order,omitempty"`
	ChildSize       int32   `protobuf:"varint,2,opt,name=child_size,json=childSize,proto3" json:"child_size,omitempty"`
	MinPrefixLength int32   `protobuf:"varint,3,opt,name=min_prefix_length,json=minPrefixLength,proto3" json:"min_prefix_length,omitempty"`
	MaxPrefixLength int32   `protobuf:"varint,4,opt,name=max_prefix_length,json=maxPrefixLength,proto3" json:"max_prefix_length,omitempty"`
	// empty child is the prehash image that is used when one child is nil (eg. 20 bytes of 0)
	EmptyChild []byte `protobuf:"bytes,5,opt,name=empty_child,json=emptyChild,proto3" json:"empty_child,omitempty"`
	// hash is the algorithm that must be used for each InnerOp
	Hash HashOp `protobuf:"varint,6,opt,name=hash,proto3,enum=cosmos.ics23.v1.HashOp" json:"hash,omitempty"`
}
