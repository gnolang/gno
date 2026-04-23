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
	childLoaded   atomic.Uint32 // bitmap: bit i set iff childNodes[i] is populated
	miniTree      MiniMerkle    // in-memory only, not serialized
	miniTreeDirty bool          // true when slot contents changed since last Build
	ndb           *nodeDB       // for lazy child loading; nil for in-memory trees
	childMu       sync.Mutex    // serialises the slow-path lazy load in getChild
}

// LeafNode stores sorted key-value hash pairs.
//
// Value storage is either inline (for small values; payload held
// directly on the leaf and serialised alongside the node) or external
// (for large values; a ValueKey references a separate record under
// PrefixVal). `inlineMask` bit i set ↔ slot i is inline; in that case
// `inlineValues[i]` holds the raw bytes and `valueKeys[i]` is nil.
// `valueHashes[i]` is computed regardless so proofs share one path.
type LeafNode struct {
	nodeKey       *NodeKey
	numKeys       int16
	keys          [B][]byte
	valueHashes   [B]Hash    // SHA256 of each value (for Merkle proofs)
	valueKeys     [B][]byte  // ValueKey references (12 bytes each, nil when inline)
	inlineValues  [B][]byte  // raw bytes when slot is inline
	inlineMask    uint32     // bit i set ↔ slot i is inline
	miniTree      MiniMerkle // in-memory only, not serialized
	miniTreeDirty bool       // true when slot contents changed since last Build
	// slotHashes caches HashLeafSlotFromValueHash(keys[i], valueHashes[i])
	// per slot so rebuildMiniMerkleIncremental skips slots whose key/value
	// did not change since the previous rebuild. slotsDirty is a bitmap:
	// bit i is set when slotHashes[i] is stale. A rebuild clears the
	// bitmap. RebuildMiniMerkle (the eager full-rebuild used by
	// constructors / deserialisation) ignores slotsDirty and recomputes
	// every occupied slot unconditionally.
	slotHashes [B]Hash
	slotsDirty uint32
}

func (*InnerNode) isNode() {}
func (*LeafNode) isNode()  {}

// valueAt returns the raw value bytes for leaf slot `i`, whether the
// slot is stored inline or externally. For inline slots the bytes are
// returned directly from the leaf's own storage (shared slice —
// caller must not mutate). External slots call `resolver` with the
// slot's valueKey. A nil resolver together with an external slot
// signals misconfiguration via ErrNoValueResolver.
func (n *LeafNode) valueAt(i int, resolver ValueResolver) ([]byte, error) {
	if n.inlineMask&(uint32(1)<<uint(i)) != 0 {
		return n.inlineValues[i], nil
	}
	if resolver == nil {
		return nil, ErrNoValueResolver
	}
	return resolver(n.valueKeys[i])
}

// valueKeyAt returns the external valueKey at slot i, or nil if the
// slot is inline. Used by orphan tracking paths that only care about
// external slots.
func (n *LeafNode) valueKeyAt(i int) []byte {
	if n.inlineMask&(uint32(1)<<uint(i)) != 0 {
		return nil
	}
	return n.valueKeys[i]
}

func (n *InnerNode) GetNodeKey() *NodeKey  { return n.nodeKey }
func (n *LeafNode) GetNodeKey() *NodeKey   { return n.nodeKey }
func (n *InnerNode) SetNodeKey(nk *NodeKey) { n.nodeKey = nk }
func (n *LeafNode) SetNodeKey(nk *NodeKey)  { n.nodeKey = nk }

// Hash returns the mini merkle root of the node. Rebuilds the mini
// merkle lazily if a prior mutation marked it dirty — write paths
// (insert/remove/redistribute/merge/split) set the dirty flag rather
// than rebuilding immediately, so a burst of N writes to the same node
// pays one rebuild at the first Hash observation instead of N rebuilds
// inline.
func (n *InnerNode) Hash() Hash {
	n.ensureMiniMerkleBuilt()
	return n.miniTree.Root()
}

func (n *LeafNode) Hash() Hash {
	n.ensureMiniMerkleBuilt()
	return n.miniTree.Root()
}

// ensureMiniMerkleBuilt forces a rebuild of the mini merkle if a prior
// mutation flagged it dirty. Called from Hash and from any site that
// reads the mini-merkle directly (currently: proof generation, which
// walks mini-merkle sibling paths without going through Hash).
func (n *InnerNode) ensureMiniMerkleBuilt() {
	if n.miniTreeDirty {
		n.RebuildMiniMerkle()
	}
}

func (n *LeafNode) ensureMiniMerkleBuilt() {
	if n.miniTreeDirty {
		n.rebuildMiniMerkleIncremental()
	}
}

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

	// Slow path: short critical section over a known-bounded body, so
	// release the mutex with explicit Unlock calls rather than via
	// defer. Each return point unlocks first; the panic branch also
	// releases the mutex so a recovered caller does not inherit a
	// permanently-held lock.
	n.childMu.Lock()
	if n.childLoaded.Load()&mask != 0 {
		n.childMu.Unlock()
		return n.childNodes[idx]
	}
	if n.ndb == nil || n.children[idx] == nil {
		n.childMu.Unlock()
		return nil
	}
	child, err := n.ndb.GetNode(n.children[idx])
	if err != nil {
		key := n.children[idx]
		n.childMu.Unlock()
		panic(fmt.Sprintf("bptree: failed to load child node %x: %v", key, err))
	}
	// Propagate ndb for recursive lazy loading
	if inner, ok := child.(*InnerNode); ok {
		inner.ndb = n.ndb
	}
	n.publishChild(idx, child)
	n.childMu.Unlock()
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


// RebuildMiniMerkle eagerly recomputes the full mini merkle tree from
// slot-level hashes. For InnerNode, slots are childHashes. For
// LeafNode, slots are HashLeafSlotFromValueHash(key, valueHash). Cost:
// B-1 = 31 SHA256 calls.
//
// Hot mutation paths (insert/remove/redistribute/merge) instead set
// miniTreeDirty = true and let Hash()/ensureMiniMerkleBuilt() rebuild
// on demand — this collapses a burst of writes to the same node into
// one rebuild per Hash observation.
func (n *InnerNode) RebuildMiniMerkle() {
	// Two-pass split: branch-free hot loop over occupied slots, then a
	// tail-fill of the unused slots. With B=32 and typical fill ratio
	// well below the maximum, the per-iteration `i < NumChildren()`
	// check (and its branch-mispredict cost) is gone for both passes.
	nc := n.NumChildren()
	for i := 0; i < nc; i++ {
		n.miniTree.tree[B+i] = n.childHashes[i]
	}
	for i := nc; i < B; i++ {
		n.miniTree.tree[B+i] = sentinelHash
	}
	n.miniTree.Build()
	n.miniTreeDirty = false
}

// RebuildMiniMerkle unconditionally rehashes every occupied slot and
// builds the mini-merkle. Used by constructors and tests where the
// per-slot cache is cold (slotHashes not yet populated). Hot mutation
// paths instead mark specific slots dirty via markLeafSlotDirty /
// markLeafSlotsDirtyRange and let ensureMiniMerkleBuilt call the
// incremental variant.
func (n *LeafNode) RebuildMiniMerkle() {
	nk := int(n.numKeys)
	// Two-pass split: hot loop hashes only occupied slots; tail-fill
	// stamps sentinel hashes into the unused suffix. Avoids the
	// per-iteration branch on `i < nk` for both phases.
	for i := 0; i < nk; i++ {
		n.slotHashes[i] = HashLeafSlotFromValueHash(n.keys[i], n.valueHashes[i])
		n.miniTree.tree[B+i] = n.slotHashes[i]
	}
	for i := nk; i < B; i++ {
		n.miniTree.tree[B+i] = sentinelHash
	}
	n.miniTree.Build()
	n.miniTreeDirty = false
	n.slotsDirty = 0
}

// rebuildMiniMerkleIncremental rehashes only the slots flagged in
// slotsDirty, reusing slotHashes for the rest. Called from
// ensureMiniMerkleBuilt; assumes slotHashes is already populated for
// non-dirty slots (true for any node that has been rebuilt at least
// once).
func (n *LeafNode) rebuildMiniMerkleIncremental() {
	nk := int(n.numKeys)
	dirty := n.slotsDirty
	// Same two-pass split as RebuildMiniMerkle: occupied-slot pass
	// (with per-slot dirty-bit check) followed by sentinel tail-fill.
	for i := 0; i < nk; i++ {
		if dirty&(uint32(1)<<uint(i)) != 0 {
			n.slotHashes[i] = HashLeafSlotFromValueHash(n.keys[i], n.valueHashes[i])
		}
		n.miniTree.tree[B+i] = n.slotHashes[i]
	}
	for i := nk; i < B; i++ {
		n.miniTree.tree[B+i] = sentinelHash
	}
	n.miniTree.Build()
	n.miniTreeDirty = false
	n.slotsDirty = 0
}

// markLeafSlotDirty flags slot i as needing recomputation on the next
// ensureMiniMerkleBuilt. Used by every hot-path site that mutates
// keys[i] or valueHashes[i] so the per-slot hash cache can skip
// untouched slots.
func (n *LeafNode) markLeafSlotDirty(i int) {
	n.slotsDirty |= uint32(1) << uint(i)
	n.miniTreeDirty = true
}

// markLeafSlotsDirtyRange flags slots [lo, hi) dirty. Used after bulk
// shifts (insert/remove) where a contiguous range of slots changed.
func (n *LeafNode) markLeafSlotsDirtyRange(lo, hi int) {
	if lo >= hi {
		return
	}
	var mask uint32
	if hi >= B {
		mask = ^uint32(0) << uint(lo)
	} else {
		mask = ((uint32(1) << uint(hi-lo)) - 1) << uint(lo)
	}
	n.slotsDirty |= mask
	n.miniTreeDirty = true
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
		nodeKey:       nil,
		numKeys:       n.numKeys,
		childSizes:    n.childSizes,
		height:        n.height,
		keys:          n.keys,
		children:      n.children,
		childHashes:   n.childHashes,
		childNodes:    n.childNodes,
		miniTree:      n.miniTree,
		miniTreeDirty: n.miniTreeDirty,
		ndb:           n.ndb,
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

// Serialize takes *bytes.Buffer rather than io.Writer so writes can be
// emitted directly (WriteByte, Write on a concrete type — no interface
// dispatch, no intermediate varint buffer that escapes to the heap).
// SaveNode is the only non-test caller and already holds a buffer.
func (n *InnerNode) Serialize(buf *bytes.Buffer) error {
	buf.WriteByte(TypeInner)
	writeUvarintBuf(buf, uint64(n.numKeys))
	nc := n.NumChildren()
	for i := 0; i < nc; i++ {
		writeVarintBuf(buf, n.childSizes[i])
	}
	writeUvarintBuf(buf, uint64(n.height))
	for i := 0; i < int(n.numKeys); i++ {
		writeBytesBuf(buf, n.keys[i])
	}
	// Fixed-width fields (children refs + child hashes) flush as
	// contiguous regions directly onto the buffer via Write, avoiding
	// 2*nc per-slot WriteByte/copy churn. Finding #18's child-ref
	// length invariant is preserved.
	for i := 0; i < nc; i++ {
		if len(n.children[i]) != NodeKeySize {
			return fmt.Errorf("InnerNode.Serialize: child[%d] ref has len %d, want %d (unsaved child reached SaveNode?)", i, len(n.children[i]), NodeKeySize)
		}
		buf.Write(n.children[i])
	}
	for i := 0; i < nc; i++ {
		buf.Write(n.childHashes[i][:])
	}
	return nil
}

// Serialize writes the leaf in v3 format (TypeLeafV3 = 0x22). Readers
// accept v3, v2 (TypeLeafV2 = 0x12), and legacy v1 (TypeLeaf = 0x02);
// writers always emit v3.
//
// v3 adds on-disk prefix compression: sorted leaf keys share a common
// byte prefix that is emitted once, followed by per-slot suffixes.
// Full keys are reconstructed at deserialise time — the in-memory
// layout is unchanged from v2.
//
// v3 layout:
//
//	type(1) = 0x22
//	numKeys (uvarint)
//	if numKeys > 0:
//	    commonPrefixLen (uvarint)
//	    commonPrefix    (commonPrefixLen bytes)
//	    per slot:
//	        suffixLen (uvarint)
//	        suffix    (suffixLen bytes)
//	valueHashes (numKeys × 32 bytes)
//	inlineMask  (4 bytes, uint32 big-endian; bit i ↔ slot i inline)
//	per-slot value: either inline (uvarint len + bytes) or external
//	                (12 bytes valueKey; all-zero placeholder for nil).
func (n *LeafNode) Serialize(buf *bytes.Buffer) error {
	buf.WriteByte(TypeLeafV3)
	writeUvarintBuf(buf, uint64(n.numKeys))
	n16 := int(n.numKeys)
	if n16 > 0 {
		// Keys are sorted, so the first/last pair bounds the common
		// prefix of every key in between. INVARIANT: callers MUST NOT
		// invoke Serialize on a leaf with a transient unsorted key
		// state (e.g. mid-redistribute, mid-merge, mid-split). The
		// B+tree mutation paths uphold this — Serialize is only
		// reached via SaveNode → saveNode after the post-mutation
		// rebuild has restored canonical sorted order. A violation
		// would emit a wrong common prefix and the v3 reader would
		// reconstruct different bytes, silently corrupting persistent
		// state.
		plen := commonPrefixLen(n.keys[0], n.keys[n16-1])
		writeUvarintBuf(buf, uint64(plen))
		if plen > 0 {
			buf.Write(n.keys[0][:plen])
		}
		for i := 0; i < n16; i++ {
			k := n.keys[i]
			writeUvarintBuf(buf, uint64(len(k)-plen))
			if len(k) > plen {
				buf.Write(k[plen:])
			}
		}
	}
	for i := 0; i < n16; i++ {
		buf.Write(n.valueHashes[i][:])
	}
	// inlineMask — fixed 4 bytes so readers can size the payload block
	// before walking it.
	var maskBuf [4]byte
	binary.BigEndian.PutUint32(maskBuf[:], n.inlineMask)
	buf.Write(maskBuf[:])
	var zeroVK [NodeKeySize]byte
	for i := 0; i < n16; i++ {
		if n.inlineMask&(uint32(1)<<uint(i)) != 0 {
			// Inline: length-prefixed raw bytes.
			writeBytesBuf(buf, n.inlineValues[i])
		} else if n.valueKeys[i] != nil {
			buf.Write(n.valueKeys[i])
		} else {
			buf.Write(zeroVK[:])
		}
	}
	return nil
}

// commonPrefixLen returns the length of the byte prefix shared by a and b.
func commonPrefixLen(a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return n
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
		return readLeafNodeV1(nk, r)
	case TypeLeafV2:
		return readLeafNodeV2(nk, r)
	case TypeLeafV3:
		return readLeafNodeV3(nk, r)
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

// readLeafNodeV1 parses a legacy leaf (TypeLeaf = 0x02) where every
// slot uses the external ValueKey indirection.
func readLeafNodeV1(nk *NodeKey, r *bytes.Reader) (_ *LeafNode, err error) {
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

	var cumulative uint64
	for i := 0; i < int(n.numKeys); i++ {
		n.keys[i], err = readBytes(r)
		if err != nil {
			return nil, fmt.Errorf("reading key %d: %w", i, err)
		}
		cumulative += uint64(len(n.keys[i]))
		if cumulative > maxLeafReadBytes {
			return nil, fmt.Errorf("leaf cumulative key bytes %d exceeds maximum %d", cumulative, maxLeafReadBytes)
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

// readLeafNodeV2 parses the current leaf format (TypeLeafV2 = 0x12)
// which adds a per-slot inline-value option. Layout after the type byte:
//
//	numKeys (uvarint)
//	keys (each: varint-len-prefixed bytes)
//	valueHashes (numKeys × 32 bytes)
//	inlineMask (4 bytes big-endian uint32)
//	per-slot value: inline (uvarint len + bytes) or external (12 B valueKey)
func readLeafNodeV2(nk *NodeKey, r *bytes.Reader) (_ *LeafNode, err error) {
	n := &LeafNode{nodeKey: nk, miniTree: NewMiniMerkle()}

	numKeys, err := binary.ReadUvarint(r)
	if err != nil {
		return nil, fmt.Errorf("reading numKeys: %w", err)
	}
	if numKeys > B {
		return nil, fmt.Errorf("leaf numKeys %d out of range [0,%d]", numKeys, B)
	}
	n.numKeys = int16(numKeys)

	var cumulative uint64
	for i := 0; i < int(n.numKeys); i++ {
		n.keys[i], err = readBytes(r)
		if err != nil {
			return nil, fmt.Errorf("reading key %d: %w", i, err)
		}
		cumulative += uint64(len(n.keys[i]))
		if cumulative > maxLeafReadBytes {
			return nil, fmt.Errorf("leaf cumulative key bytes %d exceeds maximum %d", cumulative, maxLeafReadBytes)
		}
	}
	for i := 0; i < int(n.numKeys); i++ {
		if _, err := io.ReadFull(r, n.valueHashes[i][:]); err != nil {
			return nil, fmt.Errorf("reading value hash %d: %w", i, err)
		}
	}
	var maskBuf [4]byte
	if _, err := io.ReadFull(r, maskBuf[:]); err != nil {
		return nil, fmt.Errorf("reading inlineMask: %w", err)
	}
	n.inlineMask = binary.BigEndian.Uint32(maskBuf[:])
	if err := readLeafValueBlock(n, r, &cumulative); err != nil {
		return nil, err
	}

	n.RebuildMiniMerkle()
	return n, nil
}

// readLeafValueBlock parses the per-slot value section of a v2/v3 leaf
// (numKeys × {inline-value | external valueKey}, mux'd by inlineMask).
// When inlineMask == 0 every slot is external; the entire block is a
// fixed numKeys × NodeKeySize bytes that can be read in one io.ReadFull
// into a single backing array, then carved into per-slot slice headers.
// That avoids numKeys separate allocations and numKeys per-slot
// inlineMask branches in the common all-external case.
func readLeafValueBlock(n *LeafNode, r *bytes.Reader, cumulative *uint64) error {
	if n.inlineMask == 0 {
		nk := int(n.numKeys)
		block := make([]byte, nk*NodeKeySize)
		if nk > 0 {
			if _, err := io.ReadFull(r, block); err != nil {
				return fmt.Errorf("reading external value-key block: %w", err)
			}
		}
		for i := 0; i < nk; i++ {
			lo := i * NodeKeySize
			hi := lo + NodeKeySize
			n.valueKeys[i] = block[lo:hi:hi]
		}
		return nil
	}
	for i := 0; i < int(n.numKeys); i++ {
		if n.inlineMask&(uint32(1)<<uint(i)) != 0 {
			v, err := readBytes(r)
			if err != nil {
				return fmt.Errorf("reading inline value %d: %w", i, err)
			}
			*cumulative += uint64(len(v))
			if *cumulative > maxLeafReadBytes {
				return fmt.Errorf("leaf cumulative bytes %d (key+inline) exceeds maximum %d", *cumulative, maxLeafReadBytes)
			}
			n.inlineValues[i] = v
		} else {
			n.valueKeys[i] = make([]byte, NodeKeySize)
			if _, err := io.ReadFull(r, n.valueKeys[i]); err != nil {
				return fmt.Errorf("reading value key %d: %w", i, err)
			}
		}
	}
	return nil
}

// readLeafNodeV3 parses the current leaf format (TypeLeafV3 = 0x22)
// which prefix-compresses the keys block relative to v2. Layout after
// the type byte:
//
//	numKeys (uvarint)
//	if numKeys > 0:
//	    commonPrefixLen (uvarint) + commonPrefix bytes
//	    per slot: suffixLen (uvarint) + suffix bytes
//	valueHashes (numKeys × 32 bytes)
//	inlineMask (4 bytes big-endian uint32)
//	per-slot value: inline (uvarint len + bytes) or external (12 B valueKey)
func readLeafNodeV3(nk *NodeKey, r *bytes.Reader) (_ *LeafNode, err error) {
	n := &LeafNode{nodeKey: nk, miniTree: NewMiniMerkle()}

	numKeys, err := binary.ReadUvarint(r)
	if err != nil {
		return nil, fmt.Errorf("reading numKeys: %w", err)
	}
	if numKeys > B {
		return nil, fmt.Errorf("leaf numKeys %d out of range [0,%d]", numKeys, B)
	}
	n.numKeys = int16(numKeys)

	// Cumulative leaf-bytes budget: bounds the v3-specific amplification
	// where each reconstructed key includes a fresh copy of the common
	// prefix (B copies of prefixLen extra bytes vs v2). Without this
	// cap, a malicious blob with prefixLen ≈ maxReadBytesLen and
	// numKeys = B can allocate B*prefixLen ≈ 2 MiB per leaf via the
	// per-key bound alone — see maxLeafReadBytes.
	var cumulative uint64
	if n.numKeys > 0 {
		prefixLen, err := binary.ReadUvarint(r)
		if err != nil {
			return nil, fmt.Errorf("reading commonPrefixLen: %w", err)
		}
		if prefixLen > maxReadBytesLen {
			return nil, fmt.Errorf("commonPrefixLen %d exceeds maximum %d", prefixLen, maxReadBytesLen)
		}
		var prefix []byte
		if prefixLen > 0 {
			prefix = make([]byte, prefixLen)
			if _, err := io.ReadFull(r, prefix); err != nil {
				return nil, fmt.Errorf("reading commonPrefix: %w", err)
			}
		}
		for i := 0; i < int(n.numKeys); i++ {
			suffixLen, err := binary.ReadUvarint(r)
			if err != nil {
				return nil, fmt.Errorf("reading key %d suffixLen: %w", i, err)
			}
			keyLen := uint64(prefixLen) + suffixLen
			if keyLen > maxReadBytesLen {
				return nil, fmt.Errorf("key %d length %d exceeds maximum %d", i, keyLen, maxReadBytesLen)
			}
			cumulative += keyLen
			if cumulative > maxLeafReadBytes {
				return nil, fmt.Errorf("leaf cumulative key bytes %d exceeds maximum %d", cumulative, maxLeafReadBytes)
			}
			if suffixLen == 0 {
				// Key equals the common prefix exactly. Alias the prefix
				// slice instead of allocating + copying it again — the
				// key-ownership invariant (Finding #20) makes the bytes
				// immutable post-construction, so sharing the backing
				// array across slots in this leaf is safe. The
				// three-index slice expression caps the length so an
				// accidental append cannot reach into a sibling slot.
				n.keys[i] = prefix[:prefixLen:prefixLen]
				continue
			}
			// Allocate the full key once. This keeps each leaf.keys[i]
			// backed by its own byte array, preserving the key-ownership
			// invariant (Finding #20) that Clone relies on.
			key := make([]byte, keyLen)
			copy(key, prefix)
			if _, err := io.ReadFull(r, key[prefixLen:]); err != nil {
				return nil, fmt.Errorf("reading key %d suffix: %w", i, err)
			}
			n.keys[i] = key
		}
	}

	for i := 0; i < int(n.numKeys); i++ {
		if _, err := io.ReadFull(r, n.valueHashes[i][:]); err != nil {
			return nil, fmt.Errorf("reading value hash %d: %w", i, err)
		}
	}
	var maskBuf [4]byte
	if _, err := io.ReadFull(r, maskBuf[:]); err != nil {
		return nil, fmt.Errorf("reading inlineMask: %w", err)
	}
	n.inlineMask = binary.BigEndian.Uint32(maskBuf[:])
	if err := readLeafValueBlock(n, r, &cumulative); err != nil {
		return nil, err
	}

	n.RebuildMiniMerkle()
	return n, nil
}

// --- encoding helpers ---
//
// Buffer-direct varint emission. Writing directly via buf.WriteByte
// avoids the stack-allocated [MaxVarintLen64]byte intermediate buffer
// that would otherwise escape to the heap once passed through
// io.Writer.Write. Every save in a 100k-key tree emits dozens of
// varints per node × thousands of nodes; the escape was the top
// allocator in the profile.

func writeUvarintBuf(buf *bytes.Buffer, v uint64) {
	for v >= 0x80 {
		buf.WriteByte(byte(v) | 0x80)
		v >>= 7
	}
	buf.WriteByte(byte(v))
}

func writeVarintBuf(buf *bytes.Buffer, v int64) {
	// ZigZag-compatible with binary.PutVarint.
	ux := uint64(v) << 1
	if v < 0 {
		ux = ^ux
	}
	writeUvarintBuf(buf, ux)
}

func writeBytesBuf(buf *bytes.Buffer, b []byte) {
	writeUvarintBuf(buf, uint64(len(b)))
	buf.Write(b)
}

// maxReadBytesLen caps allocations from untrusted data to prevent OOM
// during deserialization. readBytes is called only for keys (inner
// separators and leaf keys); values are stored separately and reach the
// reader via fixed-size NodeKey refs, not varint-length-prefixed bytes.
// 64 KiB is more than an order of magnitude above any realistic key
// size while still bounding a single malicious length prefix to a
// reasonable allocation. See Finding #22.
const maxReadBytesLen = 1 << 16 // 64 KiB

// maxLeafReadBytes caps the cumulative bytes a single leaf reader is
// allowed to allocate for keys + inline values. Without this, the
// per-field maxReadBytesLen bound is multiplied by B = 32 slots,
// letting a malicious DB blob allocate up to B*maxReadBytesLen ≈ 2 MiB
// per leaf via crafted lengths. The v3 reader amplifies further
// because the common prefix is duplicated into every reconstructed
// key (B copies of prefixLen → up to B*64 KiB extra). 256 KiB is
// roughly 30× a realistic leaf footprint (32 slots × (32-byte key +
// 256-byte value) ≈ 9 KiB) while still bounding the worst
// pathological blob.
const maxLeafReadBytes = 1 << 18 // 256 KiB

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
