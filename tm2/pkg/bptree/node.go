package bptree

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
)

// Node is the interface implemented by both InnerNode and LeafNode.
type Node interface {
	isNode()
	GetNodeKey() *NodeKey
	SetNodeKey(nk *NodeKey)
	Hash() Hash
}

// InnerNode stores separator keys and child references.
// It has numKeys separator keys and numKeys+1 children.
type InnerNode struct {
	nodeKey     *NodeKey
	numKeys     int16
	childSizes  [B]int64 // leaf count per child subtree; total = sum(childSizes[:numKeys+1])
	height      int16 // levels above leaf level (parent of leaves = 1)
	keys        [B - 1][]byte
	children    [B][]byte    // serialized NodeKey references (12 bytes each), used for persistence
	childHashes [B]Hash      // hash of each child subtree
	childNodes  [B]Node      // in-memory child references (nil = not yet loaded)
	miniTree    MiniMerkle   // in-memory only, not serialized
	ndb         *nodeDB      // for lazy child loading; nil for in-memory trees
	childMu     sync.Mutex   // guards lazy loading in getChild for concurrent reads
}

// LeafNode stores sorted key-value hash pairs.
type LeafNode struct {
	nodeKey     *NodeKey
	numKeys     int16
	keys        [B][]byte
	valueHashes [B]Hash // SHA256 of each value (out-of-line)
	miniTree    MiniMerkle // in-memory only, not serialized
}

func (*InnerNode) isNode() {}
func (*LeafNode) isNode()  {}

func (n *InnerNode) GetNodeKey() *NodeKey  { return n.nodeKey }
func (n *LeafNode) GetNodeKey() *NodeKey   { return n.nodeKey }
func (n *InnerNode) SetNodeKey(nk *NodeKey) { n.nodeKey = nk }
func (n *LeafNode) SetNodeKey(nk *NodeKey)  { n.nodeKey = nk }

// Hash returns the mini merkle root of the node.
func (n *InnerNode) Hash() Hash { return n.miniTree.Root() }
func (n *LeafNode) Hash() Hash  { return n.miniTree.Root() }

// NumChildren returns the number of children (numKeys + 1).
func (n *InnerNode) NumChildren() int { return int(n.numKeys) + 1 }

// getChild returns the child node at index, lazy-loading from DB if needed.
// Thread-safe: uses a mutex so concurrent reads on ImmutableTree don't race.
func (n *InnerNode) getChild(idx int) Node {
	n.childMu.Lock()
	defer n.childMu.Unlock()

	if n.childNodes[idx] != nil {
		return n.childNodes[idx]
	}
	if n.ndb == nil || n.children[idx] == nil {
		return nil
	}
	// Lazy load from DB
	child, err := n.ndb.GetNode(n.children[idx])
	if err != nil {
		panic(fmt.Sprintf("bptree: failed to load child node %x: %v", n.children[idx], err))
	}
	// Propagate ndb for recursive lazy loading
	if inner, ok := child.(*InnerNode); ok {
		inner.ndb = n.ndb
	}
	n.childNodes[idx] = child
	return child
}

// setChild sets the in-memory child node at index and clears the
// serialized NodeKey ref (it will be assigned during SaveVersion).
func (n *InnerNode) setChild(idx int, child Node) {
	n.childNodes[idx] = child
	n.children[idx] = nil
}


// RebuildMiniMerkle recomputes the full mini merkle tree from the
// slot-level hashes. For InnerNode, slots are childHashes.
// For LeafNode, slots are HashLeafSlotFromValueHash(key, valueHash).
// Cost: B-1 = 31 SHA256 calls (sets leaf slots directly, then Build).
func (n *InnerNode) RebuildMiniMerkle() {
	for i := 0; i < B; i++ {
		if i < n.NumChildren() {
			n.miniTree.tree[B+i] = n.childHashes[i]
		} else {
			n.miniTree.tree[B+i] = SentinelHash
		}
	}
	n.miniTree.Build()
}

func (n *LeafNode) RebuildMiniMerkle() {
	for i := 0; i < B; i++ {
		if i < int(n.numKeys) {
			n.miniTree.tree[B+i] = HashLeafSlotFromValueHash(n.keys[i], n.valueHashes[i])
		} else {
			n.miniTree.tree[B+i] = SentinelHash
		}
	}
	n.miniTree.Build()
}

// Clone creates a shallow copy of the node with nodeKey set to nil
// (marking it as unsaved/new for COW).
// Keys and childNodes are shared slice/pointer references (COW-safe:
// keys are never mutated in-place, only replaced by shifting).
// The ndb reference is preserved for lazy loading.
func (n *InnerNode) Clone() *InnerNode {
	c := *n //nolint:govet // intentional copy; mutex is re-initialized below
	c.nodeKey = nil
	c.childMu = sync.Mutex{} // fresh mutex for the clone
	return &c
}

func (n *LeafNode) Clone() *LeafNode {
	c := *n
	c.nodeKey = nil
	return &c
}

// --- Serialization ---
//
// InnerNode on-disk format:
//   type(1) | numKeys(varint) | size(varint) | height(varint)
//   | keys[0..numKeys-1] (each: varint-len-prefixed bytes)
//   | children[0..numKeys] (each: 12 bytes NodeKey)
//   | childHashes[0..numKeys] (each: 32 bytes Hash)
//
// LeafNode on-disk format:
//   type(1) | numKeys(varint)
//   | keys[0..numKeys-1] (each: varint-len-prefixed bytes)
//   | valueHashes[0..numKeys-1] (each: 32 bytes Hash)

func (n *InnerNode) Serialize(w io.Writer) error {
	// Type byte
	if _, err := w.Write([]byte{TypeInner}); err != nil {
		return err
	}
	// numKeys
	if err := writeUvarint(w, uint64(n.numKeys)); err != nil {
		return err
	}
	// childSizes (numKeys+1 entries)
	nc := n.NumChildren()
	for i := 0; i < nc; i++ {
		if err := writeVarint(w, n.childSizes[i]); err != nil {
			return err
		}
	}
	// height
	if err := writeUvarint(w, uint64(n.height)); err != nil {
		return err
	}
	// keys
	for i := 0; i < int(n.numKeys); i++ {
		if err := writeBytes(w, n.keys[i]); err != nil {
			return err
		}
	}
	// children (numKeys+1 NodeKey refs)
	for i := 0; i < nc; i++ {
		if _, err := w.Write(n.children[i]); err != nil {
			return err
		}
	}
	// childHashes
	for i := 0; i < nc; i++ {
		if _, err := w.Write(n.childHashes[i][:]); err != nil {
			return err
		}
	}
	return nil
}

func (n *LeafNode) Serialize(w io.Writer) error {
	if _, err := w.Write([]byte{TypeLeaf}); err != nil {
		return err
	}
	if err := writeUvarint(w, uint64(n.numKeys)); err != nil {
		return err
	}
	for i := 0; i < int(n.numKeys); i++ {
		if err := writeBytes(w, n.keys[i]); err != nil {
			return err
		}
	}
	for i := 0; i < int(n.numKeys); i++ {
		if _, err := w.Write(n.valueHashes[i][:]); err != nil {
			return err
		}
	}
	return nil
}

// ReadNode deserializes a node from bytes. Returns either *InnerNode or *LeafNode.
func ReadNode(nk *NodeKey, data []byte) (Node, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty node data")
	}
	r := bytes.NewReader(data[1:]) // skip type byte
	switch data[0] {
	case TypeInner:
		return readInnerNode(nk, r)
	case TypeLeaf:
		return readLeafNode(nk, r)
	default:
		return nil, fmt.Errorf("unknown node type: %d", data[0])
	}
}

func readInnerNode(nk *NodeKey, r *bytes.Reader) (_ *InnerNode, err error) {
	n := &InnerNode{nodeKey: nk, miniTree: NewMiniMerkle()}
	// ndb will be set by the caller (nodeDB.GetNode) after deserialization

	numKeys, err := binary.ReadUvarint(r)
	if err != nil {
		return nil, fmt.Errorf("reading numKeys: %w", err)
	}
	n.numKeys = int16(numKeys)
	if n.numKeys < 0 || n.numKeys > B-1 {
		return nil, fmt.Errorf("inner numKeys %d out of range [0,%d]", n.numKeys, B-1)
	}

	nc := n.NumChildren()
	for i := 0; i < nc; i++ {
		n.childSizes[i], err = binary.ReadVarint(r)
		if err != nil {
			return nil, fmt.Errorf("reading childSize %d: %w", i, err)
		}
	}

	height, err := binary.ReadUvarint(r)
	if err != nil {
		return nil, fmt.Errorf("reading height: %w", err)
	}
	n.height = int16(height)

	for i := 0; i < int(n.numKeys); i++ {
		n.keys[i], err = readBytes(r)
		if err != nil {
			return nil, fmt.Errorf("reading key %d: %w", i, err)
		}
	}

	for i := 0; i < nc; i++ {
		n.children[i] = make([]byte, NodeKeySize)
		if _, err := io.ReadFull(r, n.children[i]); err != nil {
			return nil, fmt.Errorf("reading child ref %d: %w", i, err)
		}
	}
	for i := 0; i < nc; i++ {
		if _, err := io.ReadFull(r, n.childHashes[i][:]); err != nil {
			return nil, fmt.Errorf("reading child hash %d: %w", i, err)
		}
	}

	n.RebuildMiniMerkle()
	return n, nil
}

func readLeafNode(nk *NodeKey, r *bytes.Reader) (_ *LeafNode, err error) {
	n := &LeafNode{nodeKey: nk, miniTree: NewMiniMerkle()}

	numKeys, err := binary.ReadUvarint(r)
	if err != nil {
		return nil, fmt.Errorf("reading numKeys: %w", err)
	}
	n.numKeys = int16(numKeys)
	if n.numKeys < 0 || n.numKeys > B {
		return nil, fmt.Errorf("leaf numKeys %d out of range [0,%d]", n.numKeys, B)
	}

	for i := 0; i < int(n.numKeys); i++ {
		n.keys[i], err = readBytes(r)
		if err != nil {
			return nil, fmt.Errorf("reading key %d: %w", i, err)
		}
	}
	for i := 0; i < int(n.numKeys); i++ {
		if _, err := io.ReadFull(r, n.valueHashes[i][:]); err != nil {
			return nil, fmt.Errorf("reading value hash %d: %w", i, err)
		}
	}

	n.RebuildMiniMerkle()
	return n, nil
}

// --- encoding helpers ---

func writeUvarint(w io.Writer, v uint64) error {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], v)
	_, err := w.Write(buf[:n])
	return err
}

func writeVarint(w io.Writer, v int64) error {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutVarint(buf[:], v)
	_, err := w.Write(buf[:n])
	return err
}

func writeBytes(w io.Writer, b []byte) error {
	if err := writeUvarint(w, uint64(len(b))); err != nil {
		return err
	}
	_, err := w.Write(b)
	return err
}

// maxReadBytesLen caps allocations from untrusted data to prevent OOM.
const maxReadBytesLen = 1 << 20 // 1 MiB — no key or inline field should exceed this

func readBytes(r *bytes.Reader) ([]byte, error) {
	length, err := binary.ReadUvarint(r)
	if err != nil {
		return nil, err
	}
	if length > maxReadBytesLen {
		return nil, fmt.Errorf("readBytes: length %d exceeds maximum %d", length, maxReadBytesLen)
	}
	b := make([]byte, length)
	if _, err := io.ReadFull(r, b); err != nil {
		return nil, err
	}
	return b, nil
}
