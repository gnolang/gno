package bptree

import "crypto/sha256"

const (
	// B is the branching factor. Inner nodes have up to B children
	// and B-1 separator keys. Leaf nodes have up to B key-value pairs.
	B = 32

	// MinKeys is the minimum occupancy for non-root nodes (B/2).
	MinKeys = B / 2

	// HashSize is the size of a SHA256 hash in bytes.
	HashSize = sha256.Size // 32

	// NodeKeySize is the size of a serialized NodeKey (version:8 + nonce:4).
	NodeKeySize = 12

	// MiniMerkleDepth is log₂(B) — the height of the mini-merkle tree
	// above each node's B slots. Compile-time constant that replaces
	// the former runtime miniMerkleDepth() log₂ loop (Finding #21).
	// init() verifies B == 1<<MiniMerkleDepth so this stays in sync if B
	// is ever adjusted.
	MiniMerkleDepth = 5

	// Domain separator prefix bytes (RFC 6962).
	DomainLeaf  byte = 0x00
	DomainInner byte = 0x01
	DomainEmpty byte = 0x02

	// DB key prefixes.
	PrefixNode   byte = 'B'
	PrefixVal    byte = 'V'
	PrefixRoot   byte = 'R'
	PrefixMeta   byte = 'M'
	PrefixOrphan byte = 'O'

	// Node type bytes for serialization.
	// v1 types (legacy; still readable):
	TypeInner byte = 0x01
	TypeLeaf  byte = 0x02
	// v2 types (current writer output):
	//   - TypeLeafV2 extends leaves with per-slot inline values
	//     (values <= Options.InlineValueThreshold stored directly in
	//     the leaf rather than via an external ValueKey indirection).
	TypeLeafV2 byte = 0x12
)

// DefaultInlineValueThreshold is the default cutoff at which a value
// is stored inline within its leaf rather than via an external
// ValueKey indirection. Values of this size or smaller inline. Tuned
// for gno.land's typical small-value workload; larger values keep the
// external-storage path so leaf serialisation stays bounded.
const DefaultInlineValueThreshold = 64

// Hash is a fixed-size SHA256 hash.
type Hash = [HashSize]byte

// sentinelHash is SHA256(0x02). Used for empty mini-merkle slots.
// Provably distinct from any 0x00-prefixed (leaf) or 0x01-prefixed (inner) hash.
// Unexported to prevent accidental mutation.
var sentinelHash Hash

// emptyTreeHash is SHA256(""). Used by Hash() for empty trees, matching IAVL behavior.
// Stored as a fixed array; callers get a fresh slice via emptyHash().
var emptyTreeHash Hash

func init() {
	// B == 1<<MiniMerkleDepth implies B is a power of two (required for
	// the mini-merkle heap layout); the single identity catches both
	// invariants.
	if B != 1<<MiniMerkleDepth {
		panic("MiniMerkleDepth out of sync with B (must satisfy B == 1<<MiniMerkleDepth, and B must be a power of 2)")
	}
	sentinelHash = sha256.Sum256([]byte{DomainEmpty})
	emptyTreeHash = sha256.Sum256(nil)
	// Pre-fill the template used by MiniMerkle.Clear() now that
	// sentinelHash is finalised (Finding #21).
	for i := range emptyMiniMerkle.tree {
		emptyMiniMerkle.tree[i] = sentinelHash
	}
}

// emptyHash returns a fresh copy of the empty tree hash (SHA256("")).
func emptyHash() []byte {
	h := emptyTreeHash
	return h[:]
}
