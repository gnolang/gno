package bptree

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/bits"
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
// Field order is load-bearing, not stylistic. An InnerNode is ~5.4 KiB and the
// COW write path allocates one per level per Set, so these are among the
// hottest objects in the process. Go's collector scans an object only up to
// the last pointer in its type (runtime type PtrBytes), so every
// pointer-bearing field is grouped FIRST and the ~3.3 KiB of pointer-free
// bulk (childSizes, childHashes, miniTree) goes last. With ndb at the end,
// PtrBytes covered the whole struct and the GC walked all 5.4 KiB of it;
// grouped this way it walks ~2 KiB. TestNodePointerBytes pins the layout.
type InnerNode struct {
	nodeKey    *NodeKey
	keys       [B - 1][]byte
	children   [B][]byte // serialized NodeKey references (12 bytes each), used for persistence
	childNodes [B]Node   // in-memory child references (nil = not yet loaded)
	ndb        *nodeDB   // for lazy child loading; nil for freshly-built (unsaved) nodes

	// --- pointer-free from here down; keep it that way ---
	numKeys     int16
	height      int16      // levels above leaf level (parent of leaves = 1)
	childSizes  [B]int64   // leaf count per child subtree; total = sum(childSizes[:numKeys+1])
	childHashes [B]Hash    // hash of each child subtree
	miniTree    MiniMerkle // in-memory only, not serialized
}

// LeafNode stores sorted key-value hash pairs. Pointer fields come first for
// the same reason as InnerNode.
type LeafNode struct {
	nodeKey   *NodeKey
	keys      [B][]byte
	valueKeys [B][]byte // ValueKey references (12 bytes each, for value DB lookup)

	// --- pointer-free from here down; keep it that way ---
	numKeys     int16
	valueHashes [B]Hash // SHA256 of each value (for Merkle proofs)
	// miniBuilt reports whether miniTree holds this node's mini merkle. It is
	// false on a node decoded from the DB: see "Deferred mini merkle" below.
	miniBuilt bool
	miniTree  MiniMerkle // in-memory only, not serialized
}

func (*InnerNode) isNode() {}
func (*LeafNode) isNode()  {}

func (n *InnerNode) GetNodeKey() *NodeKey   { return n.nodeKey }
func (n *LeafNode) GetNodeKey() *NodeKey    { return n.nodeKey }
func (n *InnerNode) SetNodeKey(nk *NodeKey) { n.nodeKey = nk }
func (n *LeafNode) SetNodeKey(nk *NodeKey)  { n.nodeKey = nk }

// Hash returns the mini merkle root of the node. A leaf whose mini merkle is
// not built (one decoded from the DB) folds its root into a local buffer
// instead, leaving the shared node untouched — see "Deferred leaf mini
// merkle".
func (n *InnerNode) Hash() Hash { return n.miniTree.Root() }

func (n *LeafNode) Hash() Hash {
	if n.miniBuilt {
		return n.miniTree.Root()
	}
	var m MiniMerkle
	n.builtMiniMerkleInto(&m)
	return m.Root()
}

// NumChildren returns the number of children (numKeys + 1).
func (n *InnerNode) NumChildren() int { return int(n.numKeys) + 1 }

// getChild returns the child node at index. If an in-memory child is present —
// an unsaved/dirty node, or one set earlier in the current descent — it is
// returned directly; otherwise the child is lazy-loaded from the DB and returned
// WITHOUT being memoized on the node. Reads therefore never mutate the node, and
// the working tree stays bounded by the nodeDB cache instead of pinning every
// node it has touched. (GetNode already sets ndb on loaded inner nodes.)
//
// A load failure (missing or corrupt record) returns an error so a single bad
// record fails the one operation instead of killing the process; (nil, nil) is
// returned only for the defensive no-ndb/no-ref case, which is unreachable on
// a healthy tree.
func (n *InnerNode) getChild(idx int) (Node, error) {
	if n.childNodes[idx] != nil {
		return n.childNodes[idx], nil
	}
	if n.ndb == nil || n.children[idx] == nil {
		return nil, nil
	}
	child, err := n.ndb.GetNode(n.children[idx])
	if err != nil {
		return nil, fmt.Errorf("loading child node %x: %w", n.children[idx], err)
	}
	return child, nil
}

// hasChildRef reports whether ref is one of this node's serialized child
// references. Used by pruning to decide whether a record is shared with the
// successor version: a linear scan over at most B=32 refs, no allocation.
//
// A NodeKey ref is a fixed 12 bytes, so the comparison is two scalar loads
// rather than bytes.Equal's memequal call plus slice-header loads — worth
// spelling out because walkAndPrune does up to B×B of these per visited
// node pair, and the scalar form measured 3.4× faster (a 16% whole-prune
// difference at 100k keys). A wrong-width or absent ref matches nothing.
func (n *InnerNode) hasChildRef(ref []byte) bool {
	if len(ref) != NodeKeySize {
		return false
	}
	version := binary.BigEndian.Uint64(ref[:8])
	nonce := binary.BigEndian.Uint32(ref[8:])
	for i := range n.NumChildren() {
		c := n.children[i]
		if len(c) == NodeKeySize &&
			binary.BigEndian.Uint64(c[:8]) == version &&
			binary.BigEndian.Uint32(c[8:]) == nonce {
			return true
		}
	}
	return false
}

// setChild sets the in-memory child node at index and clears the
// serialized NodeKey ref (it will be assigned during SaveVersion).
func (n *InnerNode) setChild(idx int, child Node) {
	n.childNodes[idx] = child
	n.children[idx] = nil
}

// --- Deferred leaf mini merkle ---
//
// A node's mini merkle is derived state. For a LEAF it costs numKeys + B-1
// SHA256 calls — ~5.7 µs for a full leaf, which was ~95% of the time to
// decode one — and nothing on the READ paths needs it: a Get/Has descent
// searches keys, an iterator walks keys and valueKeys, and a leaf's hash is
// already cached in its parent's childHashes. So decoding a leaf no longer
// builds it; `miniBuilt` records that, and the consumers that do need a hash
// handle the unbuilt case WITHOUT mutating the node:
//
//   - Clone builds it on the COPY. Every mutating path clones first, and a
//     leaf is cloned about once per write that touches it, so the write path
//     pays the same build it paid at decode time.
//   - Hash and builtMiniMerkleInto fold into a caller-owned local buffer.
//
// Not mutating matters: nodeDB.GetNode publishes decoded nodes into a shared
// LRU and hands the same instance to the writer and to concurrent
// committed-snapshot readers. A node is therefore immutable from the moment
// it is decoded, and "build it lazily on first Hash()" — which is what
// PR #5571 did — would be a data race on 32-byte hash writes, with a torn
// read yielding a wrong hash rather than a detected conflict.
//
// INNER nodes stay eager, and that asymmetry is load-bearing. Their build is
// only B-1 hashes (the slots are childHashes verbatim, with no per-slot
// hashing), but a block clones the same inner node once per write that
// descends through it — up to every write in the block for the root. Building
// at decode time lets the node cache amortise that across all of them;
// deferring it made a 100-write block 19% SLOWER even though it made decoding
// faster. A leaf, cloned about once, has nothing to amortise.

// fillSlots writes the leaf's B mini-merkle slot inputs into dst, with the
// unoccupied tail sentinelled. The single definition of what a leaf's slots
// are; every builder below goes through it. (An inner node's slots ARE its
// childHashes, so it needs no equivalent.)
func (n *LeafNode) fillSlots(dst []Hash) {
	nk := int(n.numKeys)
	for i := range nk {
		dst[i] = HashLeafSlotFromValueHash(n.keys[i], n.valueHashes[i])
	}
	for i := nk; i < B; i++ {
		dst[i] = sentinelHash
	}
}

// builtMiniMerkleInto fills dst with a BUILT mini merkle for the leaf without
// mutating the leaf: a copy of its own when it has one, a fresh fold
// otherwise. For the consumers that need the intermediate levels and not just
// the root — proof generation's sibling paths, and Hash on an unbuilt leaf.
// Takes an out-pointer because a MiniMerkle is 2 KiB and returning it by value
// copies that again. (An inner node's is always built, so those callers use
// &inner.miniTree directly and copy nothing.)
func (n *LeafNode) builtMiniMerkleInto(dst *MiniMerkle) {
	if n.miniBuilt {
		*dst = n.miniTree
		return
	}
	n.fillSlots(dst.slots())
	dst.Build()
}

// RebuildMiniMerkle recomputes the full mini merkle tree from the
// slot-level hashes. For InnerNode, slots are childHashes.
// For LeafNode, slots are HashLeafSlotFromValueHash(key, valueHash).
// Cost: B-1 = 31 SHA256 calls (sets leaf slots directly, then Build).
func (n *InnerNode) RebuildMiniMerkle() {
	nc := n.NumChildren()
	slots := n.miniTree.slots()
	copy(slots[:nc], n.childHashes[:nc])
	for i := nc; i < B; i++ {
		slots[i] = sentinelHash
	}
	n.miniTree.Build()
}

func (n *LeafNode) RebuildMiniMerkle() {
	n.fillSlots(n.miniTree.slots())
	n.miniTree.Build()
	n.miniBuilt = true
}

// rebuildMiniMerkleFromSlots rebuilds the inner levels from slot hashes the
// caller has already placed (e.g. carried over from the node a split came
// from), sentinelling the unoccupied tail. Saves the per-slot SHA256 that
// RebuildMiniMerkle would redo.
func (n *LeafNode) rebuildMiniMerkleFromSlots(src []Hash) {
	nk := int(n.numKeys)
	slots := n.miniTree.slots()
	copy(slots[:nk], src[:nk])
	for i := nk; i < B; i++ {
		slots[i] = sentinelHash
	}
	n.miniTree.Build()
	n.miniBuilt = true
}

// Clone creates a shallow copy of the node with nodeKey set to nil
// (marking it as unsaved/new for COW).
// Keys and childNodes are shared slice/pointer references (COW-safe:
// keys are never mutated in-place, only replaced by shifting).
// The ndb reference is preserved for lazy loading.
func (n *InnerNode) Clone() *InnerNode {
	c := *n
	c.nodeKey = nil
	return &c
}

func (n *LeafNode) Clone() *LeafNode {
	c := *n
	c.nodeKey = nil
	// The clone is about to be mutated and its mini merkle maintained
	// incrementally, so build it here if the original (a DB-decoded leaf) has
	// none. Building on the copy keeps the shared original immutable.
	if !c.miniBuilt {
		c.RebuildMiniMerkle()
	}
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
//   | valueKeys[0..numKeys-1] (each: 12 bytes NodeKey)
//
// Both directions work on a []byte rather than an io.Writer/io.Reader. That
// is not cosmetic: routing a varint through io.Writer forces its scratch
// array onto the heap (the compiler cannot prove an interface method does not
// retain it), which made writeUvarint/writeVarint the single largest source
// of allocated OBJECTS in the whole tree — ~44% of them. appendSerialized
// uses binary.AppendUvarint instead, and appends into one exactly-sized
// buffer (serializedSize) so a node save allocates once in total.

// serializedSize returns the exact number of bytes appendSerialized appends.
// Kept exact (not an upper bound) so SaveNode can allocate the record once;
// TestSerializedSizeIsExact pins the two in step.
func (n *InnerNode) serializedSize() int {
	nc := n.NumChildren()
	size := 1 + uvarintLen(uint64(n.numKeys)) + uvarintLen(uint64(n.height))
	for i := range nc {
		size += varintLen(n.childSizes[i])
	}
	for i := 0; i < int(n.numKeys); i++ {
		size += uvarintLen(uint64(len(n.keys[i]))) + len(n.keys[i])
	}
	return size + nc*(NodeKeySize+HashSize)
}

func (n *LeafNode) serializedSize() int {
	return n.serializedSizeWithPrefix(n.commonPrefixLen())
}

// serializedSizeWithPrefix is serializedSize for a common prefix the caller
// has already measured. SaveNode uses this pair so the LCP fold runs once per
// save rather than once in the sizing and again in the append.
func (n *LeafNode) serializedSizeWithPrefix(p int) int {
	nk := int(n.numKeys)
	size := 1 + uvarintLen(uint64(n.numKeys)) + uvarintLen(uint64(p)) + p
	for i := range nk {
		suffix := len(n.keys[i]) - p
		size += uvarintLen(uint64(suffix)) + suffix
	}
	return size + nk*(HashSize+NodeKeySize)
}

// commonPrefixLen is the length of the longest prefix shared by all occupied
// keys — the prefix TypeLeafPrefixed factors out. Folded over every key rather
// than inferred from the first and last: the sorted-keys shortcut would be
// correct, but silently emits a wrong prefix (and so corrupts every key in the
// leaf on reload) if the ordering invariant is ever broken upstream, and the
// word-wise scan makes the full fold cost a couple of hundred comparisons.
func (n *LeafNode) commonPrefixLen() int {
	nk := int(n.numKeys)
	if nk == 0 {
		return 0
	}
	p := len(n.keys[0])
	for i := 1; i < nk && p > 0; i++ {
		if l := commonPrefixLen(n.keys[0][:p], n.keys[i]); l < p {
			p = l
		}
	}
	return p
}

// commonPrefixLen returns the length of the longest common prefix of a and b.
func commonPrefixLen(a, b []byte) int {
	n := min(len(a), len(b))
	i := 0
	for ; i+8 <= n; i += 8 {
		if binary.LittleEndian.Uint64(a[i:]) != binary.LittleEndian.Uint64(b[i:]) {
			break
		}
	}
	for ; i < n && a[i] == b[i]; i++ {
	}
	return i
}

// appendSerialized appends the node's on-disk payload to b.
func (n *InnerNode) appendSerialized(b []byte) ([]byte, error) {
	nc := n.NumChildren()
	b = append(b, TypeInner)
	b = binary.AppendUvarint(b, uint64(n.numKeys))
	// childSizes (numKeys+1 entries)
	for i := range nc {
		b = binary.AppendVarint(b, n.childSizes[i])
	}
	b = binary.AppendUvarint(b, uint64(n.height))
	for i := 0; i < int(n.numKeys); i++ {
		b = binary.AppendUvarint(b, uint64(len(n.keys[i])))
		b = append(b, n.keys[i]...)
	}
	// children (numKeys+1 NodeKey refs). saveNode must have wired these up
	// before calling appendSerialized; a nil here would silently truncate the
	// payload and shift every subsequent read during deserialization.
	for i := range nc {
		if n.children[i] == nil {
			return nil, fmt.Errorf("InnerNode.Serialize: nil child ref at index %d (numChildren=%d)", i, nc)
		}
		if len(n.children[i]) != NodeKeySize {
			return nil, fmt.Errorf("InnerNode.Serialize: child ref at index %d has wrong size %d (want %d)", i, len(n.children[i]), NodeKeySize)
		}
		b = append(b, n.children[i]...)
	}
	for i := range nc {
		b = append(b, n.childHashes[i][:]...)
	}
	return b, nil
}

func (n *LeafNode) appendSerialized(b []byte) ([]byte, error) {
	return n.appendSerializedWithPrefix(b, n.commonPrefixLen())
}

func (n *LeafNode) appendSerializedWithPrefix(b []byte, p int) ([]byte, error) {
	nk := int(n.numKeys)
	b = append(b, TypeLeafPrefixed)
	b = binary.AppendUvarint(b, uint64(n.numKeys))
	b = binary.AppendUvarint(b, uint64(p))
	if p > 0 {
		b = append(b, n.keys[0][:p]...)
	}
	for i := range nk {
		b = binary.AppendUvarint(b, uint64(len(n.keys[i])-p))
		b = append(b, n.keys[i][p:]...)
	}
	for i := range nk {
		b = append(b, n.valueHashes[i][:]...)
	}
	for i := range nk {
		// Invariant: every occupied leaf slot has a non-nil valueKey (set by
		// Set, leaf splits, redistributes, and import). A nil here means
		// upstream code forgot to wire one up — on deserialization the
		// round-tripped NodeKey would be {0,0}, which silently maps to
		// "value not found" on every Get. Fail fast instead.
		if n.valueKeys[i] == nil {
			return nil, fmt.Errorf("LeafNode.Serialize: nil valueKey at slot %d (numKeys=%d)", i, n.numKeys)
		}
		if len(n.valueKeys[i]) != NodeKeySize {
			return nil, fmt.Errorf("LeafNode.Serialize: valueKey at slot %d has wrong size %d (want %d)", i, len(n.valueKeys[i]), NodeKeySize)
		}
		b = append(b, n.valueKeys[i]...)
	}
	return b, nil
}

// appendNodeRecord returns a freshly allocated buffer holding node's payload
// followed by checksumSize zero bytes for the caller to seal. Sized exactly
// from serializedSize, so a node save allocates once in total.
func appendNodeRecord(node Node) ([]byte, error) {
	var (
		b   []byte
		err error
	)
	switch n := node.(type) {
	case *InnerNode:
		b, err = n.appendSerialized(make([]byte, 0, n.serializedSize()+checksumSize))
	case *LeafNode:
		p := n.commonPrefixLen()
		b, err = n.appendSerializedWithPrefix(make([]byte, 0, n.serializedSizeWithPrefix(p)+checksumSize), p)
	default:
		return nil, fmt.Errorf("unknown node type")
	}
	if err != nil {
		return nil, err
	}
	for range checksumSize {
		b = append(b, 0)
	}
	return b, nil
}

// Serialize writes the node's on-disk payload to w. Kept for tests and
// external callers; the persistence path uses appendSerialized directly so it
// can size and seal the record in one buffer.
func (n *InnerNode) Serialize(w io.Writer) error {
	b, err := n.appendSerialized(make([]byte, 0, n.serializedSize()))
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func (n *LeafNode) Serialize(w io.Writer) error {
	b, err := n.appendSerialized(make([]byte, 0, n.serializedSize()))
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

// --- Deserialization ---

// byteCursor is a minimal read cursor over a payload. It replaces
// bytes.Reader so the decoders can slab-allocate (one allocation for all of a
// node's fixed-width refs instead of one per ref) and so no reader object is
// allocated per node load.
type byteCursor struct {
	b []byte
	i int
}

func (c *byteCursor) remaining() int { return len(c.b) - c.i }

// next returns the next n bytes as a sub-slice of the payload. The result
// ALIASES the payload, so callers that retain it past the decode must copy
// (the payload may alias DB-owned storage).
func (c *byteCursor) next(n int) ([]byte, error) {
	if n < 0 || c.remaining() < n {
		return nil, io.ErrUnexpectedEOF
	}
	s := c.b[c.i : c.i+n]
	c.i += n
	return s, nil
}

func (c *byteCursor) uvarint() (uint64, error) {
	v, n := binary.Uvarint(c.b[c.i:])
	if n <= 0 {
		if n == 0 {
			return 0, io.ErrUnexpectedEOF
		}
		return 0, errors.New("uvarint overflows 64 bits")
	}
	c.i += n
	return v, nil
}

func (c *byteCursor) varint() (int64, error) {
	v, n := binary.Varint(c.b[c.i:])
	if n <= 0 {
		if n == 0 {
			return 0, io.ErrUnexpectedEOF
		}
		return 0, errors.New("varint overflows 64 bits")
	}
	c.i += n
	return v, nil
}

// readKeys decodes count length-prefixed keys into dst[0:count], backing them
// all with a single allocation.
//
// The payload frames the block, so the slab can never exceed the record size;
// the per-key cap is maxReadBytesLen, as for a single key.
func (c *byteCursor) readKeys(dst [][]byte, count int) error {
	var src [B][]byte
	if count > len(src) {
		return fmt.Errorf("key count %d exceeds %d", count, len(src))
	}
	total := 0
	for i := range count {
		length, err := c.uvarint()
		if err != nil {
			return fmt.Errorf("reading key %d: %w", i, err)
		}
		if length > maxReadBytesLen {
			return fmt.Errorf("readBytes: length %d exceeds maximum %d", length, maxReadBytesLen)
		}
		if src[i], err = c.next(int(length)); err != nil {
			return fmt.Errorf("reading key %d: %w", i, err)
		}
		total += int(length)
	}
	// One slab for the whole block: 32 slots used to mean 32 allocations per
	// node load. The keys must be copied out of the payload rather than
	// aliasing it — some backends return their internal storage from Get.
	slab := make([]byte, total)
	off := 0
	for i := range count {
		end := off + len(src[i])
		dst[i] = slab[off:end:end]
		copy(dst[i], src[i])
		off = end
	}
	return nil
}

// readPrefixedKeys decodes a TypeLeafPrefixed key block: one shared prefix
// followed by count length-prefixed suffixes, reassembled into whole keys
// backed by a single allocation.
//
// Allocation is bounded two ways. Per key, prefixLen+suffixLen must stay
// within maxReadBytesLen, so no decoded key can exceed what the writer side
// accepts (MaxKeyLen). In total the block must stay within B*maxReadBytesLen,
// the structural maximum for B keys — a corrupt header can therefore demand at
// most the slab a legitimate leaf of the same declared shape would, and
// nothing legal is rejected. Note that the prefix is charged once per key, so
// unlike the unprefixed block the payload is NOT a lower bound on the slab;
// that is what the explicit caps are here for.
func (c *byteCursor) readPrefixedKeys(dst [][]byte, count int) error {
	var suffixes [B][]byte
	if count > len(suffixes) {
		return fmt.Errorf("key count %d exceeds %d", count, len(suffixes))
	}
	prefixLen, err := c.uvarint()
	if err != nil {
		return fmt.Errorf("reading key prefix length: %w", err)
	}
	if prefixLen > maxReadBytesLen {
		return fmt.Errorf("leaf key prefix length %d exceeds maximum %d", prefixLen, maxReadBytesLen)
	}
	prefix, err := c.next(int(prefixLen))
	if err != nil {
		return fmt.Errorf("reading key prefix: %w", err)
	}
	total := 0
	for i := range count {
		suffixLen, err := c.uvarint()
		if err != nil {
			return fmt.Errorf("reading key %d suffix length: %w", i, err)
		}
		// Both bounds, not just the sum: a suffixLen near 2^64 would wrap
		// the addition back under the cap.
		if suffixLen > maxReadBytesLen || suffixLen+prefixLen > maxReadBytesLen {
			return fmt.Errorf("leaf key %d length %d exceeds maximum %d", i, suffixLen+prefixLen, maxReadBytesLen)
		}
		if suffixes[i], err = c.next(int(suffixLen)); err != nil {
			return fmt.Errorf("reading key %d suffix: %w", i, err)
		}
		total += int(prefixLen) + int(suffixLen)
		if total > B*maxReadBytesLen {
			return fmt.Errorf("leaf key bytes %d exceed maximum %d", total, B*maxReadBytesLen)
		}
	}
	slab := make([]byte, total)
	off := 0
	for i := range count {
		end := off + int(prefixLen) + len(suffixes[i])
		key := slab[off:end:end]
		copy(key, prefix)
		copy(key[prefixLen:], suffixes[i])
		dst[i] = key
		off = end
	}
	return nil
}

// readRefs decodes count fixed-width NodeKeySize refs into dst[0:count],
// backing them all with a single allocation. The refs must be copied out of
// the payload (it may alias DB-owned storage) but need not be copied one by
// one — 32 slots used to mean 32 allocations per node load.
func (c *byteCursor) readRefs(dst [][]byte, count int, what string) error {
	src, err := c.next(count * NodeKeySize)
	if err != nil {
		return fmt.Errorf("reading %s: %w", what, err)
	}
	slab := make([]byte, count*NodeKeySize)
	copy(slab, src)
	for i := range count {
		dst[i] = slab[i*NodeKeySize : (i+1)*NodeKeySize : (i+1)*NodeKeySize]
	}
	return nil
}

// ReadNode deserializes a node from bytes. Returns either *InnerNode or *LeafNode.
// Returns an error if the payload has trailing bytes not consumed by the
// type-specific decoder — a tight framing check that catches on-disk
// corruption early instead of silently returning a truncated view.
func ReadNode(nk *NodeKey, data []byte) (Node, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty node data")
	}
	c := byteCursor{b: data, i: 1} // skip type byte
	var (
		node Node
		err  error
	)
	switch data[0] {
	case TypeInner:
		node, err = readInnerNode(nk, &c)
	case TypeLeaf, TypeLeafPrefixed:
		node, err = readLeafNode(nk, &c, data[0])
	default:
		return nil, fmt.Errorf("unknown node type: %d", data[0])
	}
	if err != nil {
		return nil, err
	}
	if c.remaining() != 0 {
		return nil, fmt.Errorf("corrupt node data: %d trailing bytes after decode (type=%d)", c.remaining(), data[0])
	}
	return node, nil
}

func readInnerNode(nk *NodeKey, c *byteCursor) (_ *InnerNode, err error) {
	// RebuildMiniMerkle below fills every slot, so the NewMiniMerkle sentinel
	// pre-fill it used to start from was 2 KiB of redundant writes.
	n := &InnerNode{nodeKey: nk}
	// ndb will be set by the caller (nodeDB.GetNode) after deserialization

	numKeys, err := c.uvarint()
	if err != nil {
		return nil, fmt.Errorf("reading numKeys: %w", err)
	}
	if numKeys > B-1 {
		return nil, fmt.Errorf("inner numKeys %d out of range [0,%d]", numKeys, B-1)
	}
	n.numKeys = int16(numKeys)

	nc := n.NumChildren()
	for i := range nc {
		n.childSizes[i], err = c.varint()
		if err != nil {
			return nil, fmt.Errorf("reading childSize %d: %w", i, err)
		}
	}

	height, err := c.uvarint()
	if err != nil {
		return nil, fmt.Errorf("reading height: %w", err)
	}
	n.height = int16(height)

	if err := c.readKeys(n.keys[:], int(n.numKeys)); err != nil {
		return nil, err
	}
	if err := c.readRefs(n.children[:], nc, "child refs"); err != nil {
		return nil, err
	}
	hashes, err := c.next(nc * HashSize)
	if err != nil {
		return nil, fmt.Errorf("reading child hashes: %w", err)
	}
	for i := range nc {
		copy(n.childHashes[i][:], hashes[i*HashSize:])
	}

	// Inner mini merkles stay eager; see "Deferred leaf mini merkle".
	n.RebuildMiniMerkle()
	return n, nil
}

func readLeafNode(nk *NodeKey, c *byteCursor, format byte) (*LeafNode, error) {
	// Mini merkle deferred; see "Deferred leaf mini merkle".
	n := &LeafNode{nodeKey: nk}

	numKeys, err := c.uvarint()
	if err != nil {
		return nil, fmt.Errorf("reading numKeys: %w", err)
	}
	if numKeys > B {
		return nil, fmt.Errorf("leaf numKeys %d out of range [0,%d]", numKeys, B)
	}
	n.numKeys = int16(numKeys)
	nkeys := int(n.numKeys)

	readKeyBlock := c.readPrefixedKeys
	if format == TypeLeaf {
		readKeyBlock = c.readKeys
	}
	if err := readKeyBlock(n.keys[:], nkeys); err != nil {
		return nil, err
	}
	hashes, err := c.next(nkeys * HashSize)
	if err != nil {
		return nil, fmt.Errorf("reading value hashes: %w", err)
	}
	for i := range nkeys {
		copy(n.valueHashes[i][:], hashes[i*HashSize:])
	}
	if err := c.readRefs(n.valueKeys[:], nkeys, "value keys"); err != nil {
		return nil, err
	}
	return n, nil
}

// --- encoding helpers ---

// uvarintLen is the encoded length of binary.AppendUvarint(nil, v).
//
// amino.UvarintSize / amino.VarintSize are the same arithmetic, but bptree
// depends on nothing but the stdlib and tm2/pkg/db, and coupling the on-disk
// node format to the codec package for six lines is the worse trade.
// TestVarintLenMatchesStdlib pins these against binary.Append* directly.
func uvarintLen(v uint64) int {
	return (bits.Len64(v|1) + 6) / 7
}

// varintLen is the encoded length of binary.AppendVarint(nil, v) (zig-zag).
func varintLen(v int64) int {
	uv := uint64(v) << 1
	if v < 0 {
		uv = ^uv
	}
	return uvarintLen(uv)
}

// maxReadBytesLen caps allocations from untrusted data to prevent OOM.
const maxReadBytesLen = 1 << 20 // 1 MiB — no key or inline field should exceed this
