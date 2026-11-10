package bptree

import "crypto/sha256"

const (
	// B is the branching factor. Inner nodes have up to B children
	// and B-1 separator keys. Leaf nodes have up to B key-value pairs.
	B = 32

	// MinKeys is the minimum occupancy for non-root nodes (B/2).
	MinKeys = B / 2

	// miniMerkleDepth is log₂(B), the depth of the per-node mini-merkle heap.
	// B is a compile-time constant, so this is too; the init() below asserts
	// 1<<miniMerkleDepth == B so the two can't silently drift.
	miniMerkleDepth = 5

	// HashSize is the size of a SHA256 hash in bytes.
	HashSize = sha256.Size // 32

	// NodeKeySize is the size of a serialized NodeKey (version:8 + nonce:4).
	NodeKeySize = 12

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
	// PrefixFast keys the optional latest-version fast index (user-key →
	// version‖value). It is an unauthenticated read accelerator outside the
	// Merkle commitment; see fast_index.go. PrefixMeta‖"fastidx" stamps the
	// version it is complete for.
	PrefixFast byte = 'F'

	// Node type bytes for serialization. The type byte doubles as the record
	// format discriminator: readers accept every type ever written, writers
	// emit the newest. It is NOT part of the Merkle commitment (node hashes
	// come from the mini-merkle over keys and value hashes, never from the
	// serialized bytes), so adding a format changes no app hash and no proof.
	TypeInner byte = 0x01
	// TypeLeaf is the original leaf layout, with each key stored in full.
	// Still read; superseded by TypeLeafPrefixed for writes.
	TypeLeaf byte = 0x02
	// TypeLeafPrefixed stores a leaf's keys as one common prefix plus a
	// per-slot suffix. Leaf keys are sorted and, at B=32, a leaf holds 32
	// ADJACENT keys — which in a structured keyspace means they share a long
	// prefix. gno.land's main store keys are "oid:<40 hex pkgid>:<n>", so a
	// leaf covering one realm's objects repeats the same ~44 bytes 32 times
	// under TypeLeaf. This is what writers emit.
	TypeLeafPrefixed byte = 0x03
)

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
	if B&(B-1) != 0 {
		panic("B must be a power of 2 (required for mini-merkle heap layout)")
	}
	if 1<<miniMerkleDepth != B {
		panic("miniMerkleDepth must equal log₂(B)")
	}
	// DB key prefixes must be pairwise distinct: a collision would let one
	// record type masquerade as another (e.g. a fast-index entry parsed as a
	// node). Asserting here means a future prefix addition can't silently alias.
	seen := map[byte]bool{}
	for _, p := range []byte{PrefixNode, PrefixVal, PrefixRoot, PrefixMeta, PrefixOrphan, PrefixFast} {
		if seen[p] {
			panic("bptree: duplicate DB key prefix")
		}
		seen[p] = true
	}
	sentinelHash = sha256.Sum256([]byte{DomainEmpty})
	emptyTreeHash = sha256.Sum256(nil)
}

// emptyHash returns a fresh copy of the empty tree hash (SHA256("")).
func emptyHash() []byte {
	h := emptyTreeHash
	return h[:]
}
