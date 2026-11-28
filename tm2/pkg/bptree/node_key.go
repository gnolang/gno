package bptree

import "encoding/binary"

// NodeKey identifies a node in the database: (version, nonce).
// Version is the tree version when the node was created.
// Nonce is a per-version counter distinguishing nodes within a version.
type NodeKey struct {
	Version int64
	Nonce   uint32
}

// GetKey serializes the NodeKey to a 12-byte slice (big-endian).
func (nk *NodeKey) GetKey() []byte {
	b := make([]byte, NodeKeySize)
	nk.putKey(b)
	return b
}

// putKey writes the serialized NodeKey into the first NodeKeySize bytes of b.
// The single definition of the wire layout, for callers that already have a
// buffer (a DB key with its prefix byte, say) and would otherwise pay an
// allocation for GetKey.
func (nk *NodeKey) putKey(b []byte) {
	binary.BigEndian.PutUint64(b[:8], uint64(nk.Version))
	binary.BigEndian.PutUint32(b[8:], nk.Nonce)
}

// GetNodeKey deserializes a NodeKey from a 12-byte slice.
func GetNodeKey(key []byte) *NodeKey {
	if len(key) != NodeKeySize {
		return nil
	}
	return &NodeKey{
		Version: int64(binary.BigEndian.Uint64(key[:8])),
		Nonce:   binary.BigEndian.Uint32(key[8:]),
	}
}

// GetRootKey is removed. The root node's NodeKey is NOT nonce=1 —
// nonces are assigned bottom-up during SaveVersion, so the root gets
// the last nonce. Use ndb.GetRoot(version) to find the actual root.

// vkVersion extracts the version a valueKey was allocated at — its first 8 bytes,
// big-endian (NodeKey.Version). allocValueKey stamps the WorkingVersion, so this
// is the version in which the key was last written. Used by orphan handling and
// the fast index (to reject entries newer than a reader's snapshot).
func vkVersion(vk []byte) int64 {
	return int64(binary.BigEndian.Uint64(vk[:8]))
}
