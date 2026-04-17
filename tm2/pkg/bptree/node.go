package bptree

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
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
//
// Concurrency:
//
//   - `childLoaded` is a bitmap (bit i set iff childNodes[i] has been
//     populated) read/written only via atomic ops. It provides the
//     release-acquire ordering that lets a reader observe a fully-written
//     `childNodes[i]` interface value after seeing the corresponding bit
//     set: the writer's non-atomic store of the fat pointer happens-before
//     the atomic.Store that sets the bit, which happens-before the reader's
//     atomic.Load that observes the bit, which happens-before the reader's
//     non-atomic load of the pointer.
//   - `childMu` serialises the slow-path lazy-load (DB fetch +
//     deserialisation + mini-merkle rebuild) so only one goroutine does
//     the work per slot. Fast-path readers never acquire `childMu`.
//
// The bitmap + mutex pair replaces the earlier design where `getChild`
// always took `childMu`, which added a full mutex Lock/Unlock to every
// traversal step even for already-loaded children.
type InnerNode struct {
	nodeKey     *NodeKey
	numKeys     int16
	childSizes  [B]int64 // leaf count per child subtree; total = sum(childSizes[:numKeys+1])
	height      int16    // levels above leaf level (parent of leaves = 1)
	keys        [B - 1][]byte
	children    [B][]byte    // serialized NodeKey references (12 bytes each), used for persistence
	childHashes [B]Hash      // hash of each child subtree
	childNodes  [B]Node      // in-memory child references (nil == unset; read only after childLoaded bit is set)
	childLoaded atomic.Uint32 // bitmap: bit i set iff childNodes[i] is populated
	miniTree    MiniMerkle   // in-memory only, not serialized
	ndb         *nodeDB      // for lazy child loading; nil for in-memory trees
	childMu     sync.Mutex   // serialises the slow-path lazy load in getChild
}

// LeafNode stores sorted key-value hash pairs.
type LeafNode struct {
	nodeKey     *NodeKey
	numKeys     int16
	keys        [B][]byte
	valueHashes [B]Hash   // SHA256 of each value (for Merkle proofs)
	valueKeys   [B][]byte // ValueKey references (12 bytes each, for value DB lookup)
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
//
// Fast path (cache hit): an atomic.Load on `childLoaded` tests whether
// slot idx has been populated. On a hit, we read childNodes[idx]
// directly — the atomic's release-acquire ordering guarantees the
// non-atomic interface read sees the value the loader wrote before
// setting the bit.
//
// Slow path: acquire childMu, re-check the bit (another goroutine may
// have loaded the slot meanwhile), then do the DB fetch. Publishing
// the loaded child is a non-atomic store of childNodes[idx] followed
// by atomic.Or on childLoaded to set the bit — fast-path readers that
// observe the bit therefore observe the complete write.
func (n *InnerNode) getChild(idx int) Node {
	mask := uint32(1) << uint(idx)
	if n.childLoaded.Load()&mask != 0 {
		return n.childNodes[idx]
	}

	n.childMu.Lock()
	defer n.childMu.Unlock()

	if n.childLoaded.Load()&mask != 0 {
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
	n.publishChild(idx, child)
	return child
}

// publishChild stores `child` at slot idx and marks the slot loaded.
// The non-atomic store of the fat interface pointer happens-before the
// atomic.Or that sets the bit; fast-path readers that observe the bit
// therefore observe the complete store.
func (n *InnerNode) publishChild(idx int, child Node) {
	n.childNodes[idx] = child
	n.childLoaded.Or(uint32(1) << uint(idx))
}

// setChild sets the in-memory child node at index and clears the
// serialized NodeKey ref (it will be assigned during SaveVersion).
// Publishes the child via the atomic bit so other goroutines that may
// hold a shared reference to this InnerNode see the new slot value.
func (n *InnerNode) setChild(idx int, child Node) {
	n.children[idx] = nil
	n.publishChild(idx, child)
}

// rebuildChildLoaded recomputes the childLoaded bitmap from the
// current childNodes state. Write paths (insert, remove, split, merge,
// redistribute, import) mutate childNodes via direct array shifts and
// slot assignments; calling rebuildChildLoaded after the bulk
// mutation restores the bit ↔ slot invariant that getChild depends on.
// Cost: one O(B) pass per call, trivial relative to the mini-merkle
// rebuild that usually follows.
func (n *InnerNode) rebuildChildLoaded() {
	var mask uint32
	for i := 0; i < B; i++ {
		if n.childNodes[i] != nil {
			mask |= uint32(1) << uint(i)
		}
	}
	n.childLoaded.Store(mask)
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
			n.miniTree.tree[B+i] = sentinelHash
		}
	}
	n.miniTree.Build()
}

func (n *LeafNode) RebuildMiniMerkle() {
	for i := 0; i < B; i++ {
		if i < int(n.numKeys) {
			n.miniTree.tree[B+i] = HashLeafSlotFromValueHash(n.keys[i], n.valueHashes[i])
		} else {
			n.miniTree.tree[B+i] = sentinelHash
		}
	}
	n.miniTree.Build()
}

// Clone creates a shallow copy of the node with nodeKey set to nil
// (marking it as unsaved/new for COW).
//
// The clone SHARES the following with n: keys slice headers, children
// slice headers, childNodes pointers. These references are COW-safe
// because of the key-ownership invariant (Finding #20): each node.keys[i]
// points to a distinct backing array whose contents are never mutated in
// place — the slot is only ever overwritten by a full slice-header
// assignment during insert shifts, remove shifts, or redistribute/merge
// paths. Therefore reading n.keys[i] bytes from any clone remains
// consistent even while another clone is being mutated.
//
// The childMu is intentionally NOT copied: each clone gets a fresh mutex
// so concurrent lazy-load on n and c do not alias the same lock. The
// ndb reference is preserved so the clone can itself lazy-load children.
//
// miniTree is copied by value (2 KB, compiles to a single memcpy).
// Finding #29 considered storing only the leaf slots and recomputing
// intermediates on demand: it would save ≈ 1 KB per clone but forces
// every subsequent SetSlot to rebuild the full 31-hash walk-up instead
// of the current 5-hash incremental update, turning each Set/Remove
// into a ~6× slower mini-merkle operation. The full-heap layout is
// retained because the memcpy is amortised by those fast incremental
// updates; Option B remains tractable only if the module adopts
// demand-driven hashing everywhere, which is out of scope here.
func (n *InnerNode) Clone() *InnerNode {
	// Explicit field copy avoids copying sync.Mutex and atomic.Uint32
	// (both would be flagged by vet). The clone inherits the parent's
	// childNodes array by value; the childLoaded bitmap is copied
	// explicitly so getChild's fast path on the clone observes the
	// same loaded set.
	//
	// Clone is not atomic with respect to a concurrent publisher on
	// `n` — the childNodes array copy and the childLoaded.Load happen
	// at different instants. Under the single-writer MutableTree
	// contract no such publisher exists, so the pair is coherent; a
	// future relaxation of the contract would need to take a snapshot
	// under childMu.
	c := &InnerNode{
		nodeKey:     nil,
		numKeys:     n.numKeys,
		childSizes:  n.childSizes,
		height:      n.height,
		keys:        n.keys,
		children:    n.children,
		childHashes: n.childHashes,
		childNodes:  n.childNodes,
		miniTree:    n.miniTree,
		ndb:         n.ndb,
		// childMu is zero-init for the fresh clone.
	}
	c.childLoaded.Store(n.childLoaded.Load())
	return c
}

// Clone creates a shallow copy of the leaf with nodeKey set to nil.
//
// The clone SHARES keys and valueKeys slice headers with n. This is safe
// under the key-ownership invariant (Finding #20): key byte contents are
// immutable — insert/remove/redistribute paths only assign slice headers
// (leaf.keys[i] = ...) or nil them, never mutate the underlying bytes
// (e.g. leaf.keys[i][j] = ... or append(leaf.keys[i], ...)). Callers
// MUST uphold this invariant; violating it would corrupt every clone
// that still shares the slot's backing array.
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
	// children (numKeys+1 NodeKey refs) + childHashes are fixed-width
	// fields, packed into a pooled scratch buffer and flushed in one
	// Write per field. The old per-slot loop was 32 × 2 Write calls
	// per full inner node. Finding #18's sanity check on child-ref
	// length is preserved.
	scratch := serializeBufPool.Get().(*serializeScratch)
	defer serializeBufPool.Put(scratch)
	for i := 0; i < nc; i++ {
		if len(n.children[i]) != NodeKeySize {
			return fmt.Errorf("InnerNode.Serialize: child[%d] ref has len %d, want %d (unsaved child reached SaveNode?)", i, len(n.children[i]), NodeKeySize)
		}
		copy(scratch.keyBytes[i*NodeKeySize:], n.children[i])
	}
	if _, err := w.Write(scratch.keyBytes[:nc*NodeKeySize]); err != nil {
		return err
	}
	for i := 0; i < nc; i++ {
		copy(scratch.hashBytes[i*HashSize:], n.childHashes[i][:])
	}
	if _, err := w.Write(scratch.hashBytes[:nc*HashSize]); err != nil {
		return err
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
	// valueHashes + valueKeys are fixed-width. Pack into pooled
	// scratch buffers and flush in one Write per field. Nil valueKey
	// slots leave their region at zero — the scratch pool clears the
	// valueKeys slice on Put, matching the previous "zero-filled
	// placeholder for missing valueKey" path.
	n16 := int(n.numKeys)
	scratch := serializeBufPool.Get().(*serializeScratch)
	defer serializeBufPool.Put(scratch)
	for i := 0; i < n16; i++ {
		copy(scratch.hashBytes[i*HashSize:], n.valueHashes[i][:])
	}
	if _, err := w.Write(scratch.hashBytes[:n16*HashSize]); err != nil {
		return err
	}
	for i := 0; i < n16; i++ {
		if n.valueKeys[i] != nil {
			copy(scratch.keyBytes[i*NodeKeySize:], n.valueKeys[i])
		} else {
			// Explicitly zero the slot so a recycled buffer with
			// stale bytes doesn't leak a non-zero placeholder.
			clear(scratch.keyBytes[i*NodeKeySize : (i+1)*NodeKeySize])
		}
	}
	if _, err := w.Write(scratch.keyBytes[:n16*NodeKeySize]); err != nil {
		return err
	}
	return nil
}

// serializeScratch is a reusable pair of stack-sized scratch buffers
// used by Serialize to batch fixed-width field writes. Pooled because
// a per-call stack buffer escapes to heap when passed to w.Write (the
// compiler can't prove the io.Writer doesn't retain the slice).
type serializeScratch struct {
	keyBytes  [B * NodeKeySize]byte // children refs (inner) or valueKeys (leaf)
	hashBytes [B * HashSize]byte    // childHashes (inner) or valueHashes (leaf)
}

var serializeBufPool = sync.Pool{
	New: func() any { return &serializeScratch{} },
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
	// Finding #23: check the raw uint64 before casting. A crafted value
	// like 0xFFFF would wrap to -1 as int16 and slip past the negative
	// check below.
	if numKeys > B-1 {
		return nil, fmt.Errorf("inner numKeys %d out of range [0,%d]", numKeys, B-1)
	}
	n.numKeys = int16(numKeys)

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
	// Finding #23: check the raw uint64 before casting to int16.
	if numKeys > B {
		return nil, fmt.Errorf("leaf numKeys %d out of range [0,%d]", numKeys, B)
	}
	n.numKeys = int16(numKeys)

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
	for i := 0; i < int(n.numKeys); i++ {
		n.valueKeys[i] = make([]byte, NodeKeySize)
		if _, err := io.ReadFull(r, n.valueKeys[i]); err != nil {
			return nil, fmt.Errorf("reading value key %d: %w", i, err)
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

// maxReadBytesLen caps allocations from untrusted data to prevent OOM
// during deserialization. readBytes is called only for keys (inner
// separators and leaf keys); values are stored separately and reach the
// reader via fixed-size NodeKey refs, not varint-length-prefixed bytes.
// 64 KiB is more than an order of magnitude above any realistic key
// size while still bounding a single malicious length prefix to a
// reasonable allocation. See Finding #22.
const maxReadBytesLen = 1 << 16 // 64 KiB

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
