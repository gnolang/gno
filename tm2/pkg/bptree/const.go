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
	// v2 types (legacy; readable):
	//   - TypeLeafV2 extends leaves with per-slot inline values
	//     (values <= Options.InlineValueThreshold stored directly in
	//     the leaf rather than via an external ValueKey indirection).
	TypeLeafV2 byte = 0x12
	// v3 types (current writer output):
	//   - TypeLeafV3 keeps the v2 inline-values semantics and adds
	//     on-disk prefix compression of the keys block. Sorted leaf
	//     keys share a common byte prefix (emitted once) with per-slot
	//     suffixes; the in-memory layout is unchanged, so readers
	//     reconstruct full keys at deserialise time. Saves disk bytes
	//     on workloads whose keys cluster under common prefixes
	//     (gno.land realm/path patterns).
	TypeLeafV3 byte = 0x22
)

// InlineThreshold is the byte-length cutoff at which a value stored
// via Set is written inline into the leaf rather than via an external
// ValueKey indirection. The named type self-documents the meaning of
// values at call sites — InlineDisabled vs DefaultInlineValueThreshold
// vs an explicit byte cutoff are visibly distinct, where a bare int
// would conflate them.
//
// Semantics:
//   - InlineDisabled (or any value <= 0): inline storage is disabled;
//     every value goes external regardless of size.
//   - 1 .. MaxInlineValueThreshold: values of this size or smaller
//     inline; larger values use the external path.
//   - Anything above MaxInlineValueThreshold is silently clamped to
//     MaxInlineValueThreshold; see that constant for the rationale.
type InlineThreshold int

// InlineDisabled, when used as Options.InlineValueThreshold or passed
// to InlineValueThresholdOption, turns off inline-value storage so
// every value takes the external ValueKey path.
const InlineDisabled InlineThreshold = -1

// DefaultInlineValueThreshold is the recommended cutoff at which a
// value is stored inline within its leaf rather than via an external
// ValueKey indirection. Values of this size or smaller inline. Tuned
// for gno.land's typical small-value workload; larger values keep the
// external-storage path so leaf serialisation stays bounded.
const DefaultInlineValueThreshold InlineThreshold = 64

// MaxInlineValueThreshold caps the per-value byte-length that may be
// inlined regardless of the configured InlineValueThreshold. The cap
// exists so that a leaf full of inline values can never exceed the
// reader's per-leaf cumulative budget (maxLeafReadBytes = 256 KiB):
// even with B = 32 inline slots all at the maximum, the resulting
// leaf payload (32 * 4 KiB ≈ 128 KiB plus keys + headers) stays
// safely below the read cap. Without this bound a caller could write
// a leaf whose serialised form exceeds the read budget and is
// permanently un-mountable on the next LoadVersion.
const MaxInlineValueThreshold InlineThreshold = 4 << 10 // 4 KiB

// Hash is a fixed-size SHA256 hash.
type Hash = [HashSize]byte

// sentinelHash is SHA256(0x02). Used for empty mini-merkle slots.
// Provably distinct from any 0x00-prefixed (leaf) or 0x01-prefixed (inner) hash.
// Unexported to prevent accidental mutation.
var sentinelHash Hash

// emptyTreeHash is SHA256(""). Used by Hash() for empty trees, matching IAVL behavior.
// Stored as a fixed array so emptyHashSlice can take a stable view of it.
var emptyTreeHash Hash

// emptyHashSlice is a package-level []byte view of emptyTreeHash,
// initialised at package init so emptyHash() can return it without
// allocating per call. Callers MUST treat the returned slice as
// read-only — mutating it would corrupt the sentinel for every other
// reader.
var emptyHashSlice []byte

func init() {
	// B == 1<<MiniMerkleDepth implies B is a power of two (required for
	// the mini-merkle heap layout); the single identity catches both
	// invariants.
	if B != 1<<MiniMerkleDepth {
		panic("MiniMerkleDepth out of sync with B (must satisfy B == 1<<MiniMerkleDepth, and B must be a power of 2)")
	}
	sentinelHash = sha256.Sum256([]byte{DomainEmpty})
	emptyTreeHash = sha256.Sum256(nil)
	emptyHashSlice = emptyTreeHash[:]
	// Pre-fill the template used by MiniMerkle.Clear() now that
	// sentinelHash is finalised (Finding #21).
	for i := range emptyMiniMerkle.tree {
		emptyMiniMerkle.tree[i] = sentinelHash
	}
}

// emptyHash returns the package-level slice view of SHA256(""). The
// returned slice is shared across callers and MUST NOT be mutated;
// see emptyHashSlice's doc for the corruption hazard a caller-side
// write would cause.
func emptyHash() []byte {
	return emptyHashSlice
}
