package bptree

import "encoding/binary"

// NodeKey identifies a node in the database: (version, nonce).
// Version is the tree version when the node was created.
// Nonce is a per-version counter distinguishing nodes within a version.
//
// The serialized bytes are cached on the struct so repeated GetKey()
// calls on the same NodeKey don't re-allocate. Production code paths
// (NextNodeKey, GetNodeKey, NewNodeKey) precompute the bytes at
// construction so GetKey() never allocates on the hot paths. See
// Finding #21.
type NodeKey struct {
	Version int64
	Nonce   uint32
	bytes   []byte // serialized form; populated by constructors or lazily
}

// NewNodeKey constructs a NodeKey and precomputes its serialized bytes.
// Prefer this over the struct literal so GetKey() avoids allocation.
func NewNodeKey(version int64, nonce uint32) *NodeKey {
	return &NodeKey{Version: version, Nonce: nonce, bytes: encodeNodeKeyBytes(version, nonce)}
}

// encodeNodeKeyBytes returns the 12-byte big-endian serialization of
// (version, nonce) without constructing a NodeKey. Callers that only
// need the serialized form (e.g. ValueKey allocation in hot paths)
// save one heap allocation by skipping the wrapping struct.
func encodeNodeKeyBytes(version int64, nonce uint32) []byte {
	b := make([]byte, NodeKeySize)
	binary.BigEndian.PutUint64(b[:8], uint64(version))
	binary.BigEndian.PutUint32(b[8:], nonce)
	return b
}

// GetKey returns the 12-byte big-endian serialization of the NodeKey.
// The returned slice must not be mutated by the caller — it may be the
// cached buffer shared with future GetKey() calls on the same NodeKey.
//
// For NodeKeys constructed via NewNodeKey / NextNodeKey / GetNodeKey the
// bytes are already cached, so this is a single field load. Struct-literal
// NodeKeys (primarily tests) pay a one-time allocation on first call; the
// lazy assignment is not safe against concurrent first callers, but
// production NodeKeys are never in that state because their cache is
// populated before the struct crosses a goroutine boundary.
func (nk *NodeKey) GetKey() []byte {
	if b := nk.bytes; b != nil {
		return b
	}
	b := make([]byte, NodeKeySize)
	binary.BigEndian.PutUint64(b[:8], uint64(nk.Version))
	binary.BigEndian.PutUint32(b[8:], nk.Nonce)
	nk.bytes = b
	return b
}

// GetNodeKey deserializes a NodeKey from a 12-byte slice. The cached
// bytes are populated to a defensive copy of the input (the caller's
// buffer may be reused or mutated).
func GetNodeKey(key []byte) *NodeKey {
	if len(key) != NodeKeySize {
		return nil
	}
	b := make([]byte, NodeKeySize)
	copy(b, key)
	return &NodeKey{
		Version: int64(binary.BigEndian.Uint64(b[:8])),
		Nonce:   binary.BigEndian.Uint32(b[8:]),
		bytes:   b,
	}
}

// GetRootKey is removed. The root node's NodeKey is NOT nonce=1 —
// nonces are assigned bottom-up during SaveVersion, so the root gets
// the last nonce. Use ndb.GetRoot(version) to find the actual root.
