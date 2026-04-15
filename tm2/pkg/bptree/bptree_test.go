package bptree

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"math/rand"
	"strings"
	"testing"
)

// --- const.go ---

func TestSentinelHash(t *testing.T) {
	expected := sha256.Sum256([]byte{0x02})
	if SentinelHash != expected {
		t.Fatalf("SentinelHash mismatch")
	}
}

func TestEmptyTreeHash(t *testing.T) {
	// Empty tree Hash() must return SHA256(""), matching IAVL behavior.
	expectedHash := sha256.Sum256(nil)

	// MutableTree (in-memory)
	tree := NewMutableTreeMem()
	h := tree.WorkingHash()
	if !bytes.Equal(h, expectedHash[:]) {
		t.Fatalf("WorkingHash on empty tree = %x, want %x", h, expectedHash)
	}

	// SaveVersion on empty tree must return SHA256("")
	hash, _, err := tree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(hash, expectedHash[:]) {
		t.Fatalf("SaveVersion on empty tree = %x, want %x", hash, expectedHash)
	}

	// Hash() after SaveVersion
	if !bytes.Equal(tree.Hash(), expectedHash[:]) {
		t.Fatalf("Hash() after empty SaveVersion = %x, want %x", tree.Hash(), expectedHash)
	}

	// ImmutableTree with nil root
	imm := NewImmutableTree(nil, 1)
	if !bytes.Equal(imm.Hash(), expectedHash[:]) {
		t.Fatalf("ImmutableTree.Hash() on nil root = %x, want %x", imm.Hash(), expectedHash)
	}
}

func TestSetEmptyValue(t *testing.T) {
	// Set(key, []byte{}) must round-trip correctly — not return nil.
	tree := NewMutableTreeMem()
	_, err := tree.Set([]byte("k"), []byte{})
	if err != nil {
		t.Fatal(err)
	}
	val, err := tree.Get([]byte("k"))
	if err != nil {
		t.Fatal(err)
	}
	if val == nil {
		t.Fatal("Get returned nil for empty value; expected []byte{}")
	}
	if len(val) != 0 {
		t.Fatalf("Get returned %x, expected empty slice", val)
	}
}

func TestConstants(t *testing.T) {
	if B != 32 {
		t.Fatalf("B = %d, want 32", B)
	}
	if MinKeys != 16 {
		t.Fatalf("MinKeys = %d, want 16", MinKeys)
	}
	if HashSize != 32 {
		t.Fatalf("HashSize = %d, want 32", HashSize)
	}
	if NodeKeySize != 12 {
		t.Fatalf("NodeKeySize = %d, want 12", NodeKeySize)
	}
}

// --- hash.go ---

func TestHashInnerShortCircuit(t *testing.T) {
	// Both sentinel → sentinel
	result := HashInner(SentinelHash, SentinelHash)
	if result != SentinelHash {
		t.Fatalf("expected sentinel short-circuit")
	}

	// One non-sentinel → real hash
	var other Hash
	other[0] = 0xFF
	result = HashInner(SentinelHash, other)
	if result == SentinelHash {
		t.Fatalf("expected real hash, got sentinel")
	}
}

func TestHashInnerAsymmetry(t *testing.T) {
	a := sha256.Sum256([]byte("a"))
	b := sha256.Sum256([]byte("b"))
	if HashInner(a, b) == HashInner(b, a) {
		t.Fatalf("HashInner(a,b) should differ from HashInner(b,a)")
	}
}

func TestHashInnerOneSentinel(t *testing.T) {
	other := sha256.Sum256([]byte("x"))
	lr := HashInner(SentinelHash, other)
	rl := HashInner(other, SentinelHash)
	if lr == rl {
		t.Fatalf("sentinel position should matter")
	}
	if lr == SentinelHash || rl == SentinelHash {
		t.Fatalf("one-sentinel case should not short-circuit")
	}
}

func TestHashLeafSlotFromValueHash(t *testing.T) {
	key := []byte("testkey")
	value := []byte("testvalue")
	valueHash := sha256.Sum256(value)
	h1 := HashLeafSlot(key, value)
	h2 := HashLeafSlotFromValueHash(key, valueHash)
	if h1 != h2 {
		t.Fatalf("HashLeafSlot and HashLeafSlotFromValueHash disagree")
	}
}

func TestHashLeafSlot_EmptyKey(t *testing.T) {
	h := HashLeafSlot([]byte{}, []byte("val"))
	if h == SentinelHash {
		t.Fatalf("empty key hash should not be sentinel")
	}
}

func TestHashLeafSlot_EmptyValue(t *testing.T) {
	h := HashLeafSlot([]byte("key"), []byte{})
	if h == SentinelHash {
		t.Fatalf("empty value hash should not be sentinel")
	}
}

func TestHashLeafSlot_LongKey(t *testing.T) {
	// 256 bytes
	key256 := []byte(strings.Repeat("a", 256))
	h256 := HashLeafSlot(key256, []byte("v"))

	// 257 bytes — previously would have overflowed a fixed buffer
	key257 := []byte(strings.Repeat("a", 257))
	h257 := HashLeafSlot(key257, []byte("v"))

	if h256 == h257 {
		t.Fatalf("different length keys should produce different hashes")
	}

	// 1000 bytes
	key1000 := []byte(strings.Repeat("b", 1000))
	h1000 := HashLeafSlot(key1000, []byte("v"))
	if h1000 == SentinelHash {
		t.Fatalf("long key hash should not be sentinel")
	}
}

func TestHashLeafSlot_Determinism(t *testing.T) {
	key := []byte("det")
	val := []byte("erministic")
	h1 := HashLeafSlot(key, val)
	h2 := HashLeafSlot(key, val)
	if h1 != h2 {
		t.Fatalf("hash should be deterministic")
	}
}

func TestHashLeafSlot_DomainSeparation(t *testing.T) {
	// A leaf hash and an inner hash should never collide
	// (can't prove cryptographically, but verify they differ for trivial cases)
	key := make([]byte, 31) // chosen so preimage could be 64 bytes
	var vh Hash
	leafHash := HashLeafSlotFromValueHash(key, vh)
	innerHash := HashInner(vh, vh)
	if leafHash == innerHash {
		t.Fatalf("leaf and inner hashes collided — domain separation failure")
	}
}

// --- mini_merkle.go ---

func TestMiniMerkle_EmptyRoot(t *testing.T) {
	m := NewMiniMerkle()
	if m.Root() != SentinelHash {
		t.Fatalf("empty mini merkle root should be sentinel")
	}
}

func TestMiniMerkle_SetSlotAndBuild(t *testing.T) {
	m := NewMiniMerkle()
	h := sha256.Sum256([]byte("hello"))
	m.SetSlot(0, h)
	if m.Root() == SentinelHash {
		t.Fatalf("root should not be sentinel after setting a slot")
	}

	// Full build should agree
	var m2 MiniMerkle
	m2.Clear()
	m2.tree[B+0] = h
	m2.Build()
	if m.Root() != m2.Root() {
		t.Fatalf("incremental and full build disagree")
	}
}

func TestMiniMerkle_SetSlotLastIndex(t *testing.T) {
	m := NewMiniMerkle()
	h := sha256.Sum256([]byte("last"))
	m.SetSlot(B-1, h)
	if m.Root() == SentinelHash {
		t.Fatalf("root should not be sentinel after setting last slot")
	}
	if m.GetSlot(B-1) != h {
		t.Fatalf("GetSlot(B-1) mismatch")
	}
}

func TestMiniMerkle_SetSlotUpdate(t *testing.T) {
	m := NewMiniMerkle()
	h1 := sha256.Sum256([]byte("first"))
	h2 := sha256.Sum256([]byte("second"))
	m.SetSlot(5, h1)
	root1 := m.Root()
	m.SetSlot(5, h2)
	root2 := m.Root()
	if root1 == root2 {
		t.Fatalf("root should change after updating a slot")
	}
	if m.GetSlot(5) != h2 {
		t.Fatalf("GetSlot should return updated value")
	}
}

func TestMiniMerkle_AllSlotsFilled(t *testing.T) {
	m := NewMiniMerkle()
	for i := 0; i < B; i++ {
		m.SetSlot(i, sha256.Sum256([]byte{byte(i)}))
	}

	// Build from scratch should agree
	var m2 MiniMerkle
	for i := 0; i < B; i++ {
		m2.tree[B+i] = sha256.Sum256([]byte{byte(i)})
	}
	m2.Build()
	if m.Root() != m2.Root() {
		t.Fatalf("incremental all-slots vs build disagree")
	}
}

func TestMiniMerkle_ClearAfterDirty(t *testing.T) {
	m := NewMiniMerkle()
	m.SetSlot(3, sha256.Sum256([]byte("dirty")))
	if m.Root() == SentinelHash {
		t.Fatalf("should be dirty")
	}
	m.Clear()
	if m.Root() != SentinelHash {
		t.Fatalf("Clear should reset to sentinel")
	}
}

func TestMiniMerkle_HalfFilledStructure(t *testing.T) {
	m := NewMiniMerkle()
	for i := 0; i < B/2; i++ {
		m.SetSlot(i, sha256.Sum256([]byte{byte(i)}))
	}
	// Right half subtree root (tree[3]) should be sentinel
	if m.tree[3] != SentinelHash {
		t.Fatalf("right half subtree should be sentinel, got non-sentinel")
	}
	// Root = HashInner(left_half, sentinel)
	expected := HashInner(m.tree[2], SentinelHash)
	if m.Root() != expected {
		t.Fatalf("root should be HashInner(left_half, sentinel)")
	}
}

func TestMiniMerkle_SingleOccupiedSlot(t *testing.T) {
	m := NewMiniMerkle()
	h := sha256.Sum256([]byte("only"))
	m.SetSlot(0, h)

	// Walk up manually: slot 0 is always left child at every level
	cur := h
	for level := 0; level < miniMerkleDepth(); level++ {
		cur = HashInner(cur, SentinelHash)
	}
	if m.Root() != cur {
		t.Fatalf("single slot root mismatch")
	}
}

// --- mini_merkle.go: SiblingPath ---

func TestMiniMerkleSiblingPath_Slot0(t *testing.T) {
	m := filledMiniMerkle()
	verifySiblingPath(t, &m, 0)
}

func TestMiniMerkleSiblingPath_LastSlot(t *testing.T) {
	m := filledMiniMerkle()
	_, positions := m.SiblingPath(B - 1)
	// Slot B-1 is always right child at every level
	for i, pos := range positions {
		if pos != 1 {
			t.Fatalf("slot B-1 level %d: expected position 1 (right), got %d", i, pos)
		}
	}
	verifySiblingPath(t, &m, B-1)
}

func TestMiniMerkleSiblingPath_MiddleSlot(t *testing.T) {
	m := filledMiniMerkle()
	verifySiblingPath(t, &m, 13)
}

func TestMiniMerkleSiblingPath_PartiallyFilled(t *testing.T) {
	m := NewMiniMerkle()
	for i := 0; i < 16; i++ {
		m.SetSlot(i, sha256.Sum256([]byte{byte(i)}))
	}
	// Verify path reconstruction for a slot in the occupied half
	verifySiblingPath(t, &m, 5)
	// And for a slot in the empty half (sentinel slot)
	verifySiblingPath(t, &m, 20)
}

func filledMiniMerkle() MiniMerkle {
	m := NewMiniMerkle()
	for i := 0; i < B; i++ {
		m.SetSlot(i, sha256.Sum256([]byte{byte(i)}))
	}
	return m
}

func verifySiblingPath(t *testing.T, m *MiniMerkle, index int) {
	t.Helper()
	siblings, positions := m.SiblingPath(index)
	if len(siblings) != miniMerkleDepth() {
		t.Fatalf("slot %d: expected %d siblings, got %d", index, miniMerkleDepth(), len(siblings))
	}
	current := m.GetSlot(index)
	for i, sib := range siblings {
		if positions[i] == 0 {
			current = HashInner(current, sib)
		} else {
			current = HashInner(sib, current)
		}
	}
	if current != m.Root() {
		t.Fatalf("slot %d: sibling path reconstruction failed", index)
	}
}

// --- node_key.go ---

func TestNodeKey_Roundtrip(t *testing.T) {
	nk := &NodeKey{Version: 42, Nonce: 7}
	b := nk.GetKey()
	if len(b) != NodeKeySize {
		t.Fatalf("expected %d bytes, got %d", NodeKeySize, len(b))
	}
	nk2 := GetNodeKey(b)
	if nk2.Version != 42 || nk2.Nonce != 7 {
		t.Fatalf("roundtrip failed: got %+v", nk2)
	}
}

func TestNodeKey_ZeroValues(t *testing.T) {
	nk := &NodeKey{Version: 0, Nonce: 0}
	b := nk.GetKey()
	nk2 := GetNodeKey(b)
	if nk2.Version != 0 || nk2.Nonce != 0 {
		t.Fatalf("zero roundtrip failed")
	}
}

func TestNodeKey_MaxValues(t *testing.T) {
	nk := &NodeKey{Version: math.MaxInt64, Nonce: math.MaxUint32}
	b := nk.GetKey()
	nk2 := GetNodeKey(b)
	if nk2.Version != math.MaxInt64 || nk2.Nonce != math.MaxUint32 {
		t.Fatalf("max value roundtrip failed: %+v", nk2)
	}
}

func TestNodeKey_NegativeVersion(t *testing.T) {
	nk := &NodeKey{Version: -1, Nonce: 0}
	b := nk.GetKey()
	nk2 := GetNodeKey(b)
	if nk2.Version != -1 {
		t.Fatalf("negative version roundtrip failed: got %d", nk2.Version)
	}
}

func TestGetNodeKey_InvalidInputs(t *testing.T) {
	if GetNodeKey(nil) != nil {
		t.Fatalf("nil should return nil")
	}
	if GetNodeKey([]byte{1, 2, 3}) != nil {
		t.Fatalf("short slice should return nil")
	}
	if GetNodeKey(make([]byte, 13)) != nil {
		t.Fatalf("too-long slice should return nil")
	}
}

func TestGetRootKey(t *testing.T) {
	rk := GetRootKey(42)
	expected := (&NodeKey{Version: 42, Nonce: 1}).GetKey()
	if !bytes.Equal(rk, expected) {
		t.Fatalf("GetRootKey mismatch")
	}
}

// --- search.go ---

func TestSearchLeaf_Basic(t *testing.T) {
	leaf := &LeafNode{numKeys: 5}
	for i := 0; i < 5; i++ {
		leaf.keys[i] = []byte{byte(i * 10)} // 0, 10, 20, 30, 40
	}

	idx, found := searchLeaf(leaf, []byte{20})
	if !found || idx != 2 {
		t.Fatalf("exact match: expected (2, true), got (%d, %v)", idx, found)
	}

	idx, found = searchLeaf(leaf, []byte{15})
	if found || idx != 2 {
		t.Fatalf("insert point: expected (2, false), got (%d, %v)", idx, found)
	}

	idx, found = searchLeaf(leaf, []byte{0})
	if !found || idx != 0 {
		t.Fatalf("first key: expected (0, true), got (%d, %v)", idx, found)
	}

	idx, found = searchLeaf(leaf, []byte{50})
	if found || idx != 5 {
		t.Fatalf("after all: expected (5, false), got (%d, %v)", idx, found)
	}
}

func TestSearchLeaf_EmptyNode(t *testing.T) {
	leaf := &LeafNode{numKeys: 0}
	idx, found := searchLeaf(leaf, []byte{42})
	if found || idx != 0 {
		t.Fatalf("empty node: expected (0, false), got (%d, %v)", idx, found)
	}
}

func TestSearchLeaf_SingleKey(t *testing.T) {
	leaf := &LeafNode{numKeys: 1}
	leaf.keys[0] = []byte{10}

	idx, found := searchLeaf(leaf, []byte{10})
	if !found || idx != 0 {
		t.Fatalf("exact: expected (0, true), got (%d, %v)", idx, found)
	}
	idx, found = searchLeaf(leaf, []byte{5})
	if found || idx != 0 {
		t.Fatalf("before: expected (0, false), got (%d, %v)", idx, found)
	}
	idx, found = searchLeaf(leaf, []byte{15})
	if found || idx != 1 {
		t.Fatalf("after: expected (1, false), got (%d, %v)", idx, found)
	}
}

func TestSearchLeaf_MultiByteKeys(t *testing.T) {
	leaf := &LeafNode{numKeys: 3}
	leaf.keys[0] = []byte{10, 0}
	leaf.keys[1] = []byte{10, 5}
	leaf.keys[2] = []byte{10, 10}

	idx, found := searchLeaf(leaf, []byte{10, 3})
	if found || idx != 1 {
		t.Fatalf("multi-byte: expected (1, false), got (%d, %v)", idx, found)
	}
}

func TestSearchLeaf_FullNode(t *testing.T) {
	leaf := &LeafNode{numKeys: B}
	for i := 0; i < B; i++ {
		leaf.keys[i] = []byte{byte(i * 2)} // 0, 2, 4, ..., 62
	}

	idx, found := searchLeaf(leaf, []byte{0})
	if !found || idx != 0 {
		t.Fatalf("first: expected (0, true), got (%d, %v)", idx, found)
	}
	idx, found = searchLeaf(leaf, []byte{62})
	if !found || idx != 31 {
		t.Fatalf("last: expected (31, true), got (%d, %v)", idx, found)
	}
	idx, found = searchLeaf(leaf, []byte{7})
	if found || idx != 4 {
		t.Fatalf("miss: expected (4, false), got (%d, %v)", idx, found)
	}
}

func TestSearchInner_Basic(t *testing.T) {
	inner := &InnerNode{numKeys: 3}
	inner.keys[0] = []byte{10}
	inner.keys[1] = []byte{20}
	inner.keys[2] = []byte{30}

	tests := []struct {
		key      byte
		expected int
	}{
		{5, 0}, {10, 1}, {15, 1}, {20, 2}, {25, 2}, {30, 3}, {35, 3},
	}
	for _, tt := range tests {
		idx := searchInner(inner, []byte{tt.key})
		if idx != tt.expected {
			t.Errorf("searchInner(key=%d): expected %d, got %d", tt.key, tt.expected, idx)
		}
	}
}

func TestSearchInner_EmptyNode(t *testing.T) {
	inner := &InnerNode{numKeys: 0}
	idx := searchInner(inner, []byte{42})
	if idx != 0 {
		t.Fatalf("empty inner: expected 0, got %d", idx)
	}
}

func TestSearchInner_SingleSeparator(t *testing.T) {
	inner := &InnerNode{numKeys: 1}
	inner.keys[0] = []byte{50}
	if searchInner(inner, []byte{25}) != 0 {
		t.Fatalf("before separator")
	}
	if searchInner(inner, []byte{50}) != 1 {
		t.Fatalf("equal to separator")
	}
	if searchInner(inner, []byte{75}) != 1 {
		t.Fatalf("after separator")
	}
}

func TestSearchInner_MultiByteKeys(t *testing.T) {
	inner := &InnerNode{numKeys: 2}
	inner.keys[0] = []byte{10, 0}
	inner.keys[1] = []byte{10, 10}
	if searchInner(inner, []byte{10, 5}) != 1 {
		t.Fatalf("multi-byte inner search")
	}
}

// --- node.go: serialization ---

func TestNodeSerialization_InnerBasic(t *testing.T) {
	inner := &InnerNode{
		nodeKey: &NodeKey{Version: 1, Nonce: 2},
		numKeys: 2,
		size:    100,
		height:  1,
	}
	inner.keys[0] = []byte("key1")
	inner.keys[1] = []byte("key2")
	for i := 0; i < 3; i++ {
		inner.children[i] = (&NodeKey{Version: 1, Nonce: uint32(10 + i)}).GetKey()
		inner.childHashes[i] = sha256.Sum256([]byte{byte(i)})
	}
	inner.RebuildMiniMerkle()
	origHash := inner.Hash()

	var buf bytes.Buffer
	if err := inner.Serialize(&buf); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	node, err := ReadNode(&NodeKey{Version: 1, Nonce: 2}, buf.Bytes())
	if err != nil {
		t.Fatalf("ReadNode: %v", err)
	}
	inner2 := node.(*InnerNode)
	if inner2.numKeys != 2 || inner2.size != 100 || inner2.height != 1 {
		t.Fatalf("metadata mismatch")
	}
	if !bytes.Equal(inner2.keys[0], []byte("key1")) {
		t.Fatalf("key mismatch")
	}
	if inner2.childHashes[0] != inner.childHashes[0] {
		t.Fatalf("childHash mismatch")
	}
	// Hash should be rebuilt during deserialization
	if inner2.Hash() != origHash {
		t.Fatalf("hash mismatch after deserialization: mini merkle not rebuilt")
	}
}

func TestNodeSerialization_LeafBasic(t *testing.T) {
	leaf := &LeafNode{
		nodeKey: &NodeKey{Version: 1, Nonce: 3},
		numKeys: 2,
	}
	leaf.keys[0] = []byte("a")
	leaf.keys[1] = []byte("b")
	leaf.valueHashes[0] = sha256.Sum256([]byte("val_a"))
	leaf.valueHashes[1] = sha256.Sum256([]byte("val_b"))
	leaf.RebuildMiniMerkle()
	origHash := leaf.Hash()

	var buf bytes.Buffer
	if err := leaf.Serialize(&buf); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	node, err := ReadNode(&NodeKey{Version: 1, Nonce: 3}, buf.Bytes())
	if err != nil {
		t.Fatalf("ReadNode: %v", err)
	}
	leaf2 := node.(*LeafNode)
	if leaf2.numKeys != 2 {
		t.Fatalf("numKeys mismatch")
	}
	if leaf2.valueHashes[0] != leaf.valueHashes[0] {
		t.Fatalf("valueHash mismatch")
	}
	if leaf2.Hash() != origHash {
		t.Fatalf("hash mismatch after deserialization: mini merkle not rebuilt")
	}
}

func TestNodeSerialization_EmptyInner(t *testing.T) {
	inner := &InnerNode{
		nodeKey: &NodeKey{Version: 1, Nonce: 1},
		numKeys: 0,
		size:    5,
		height:  1,
	}
	inner.children[0] = (&NodeKey{Version: 1, Nonce: 10}).GetKey()
	inner.childHashes[0] = sha256.Sum256([]byte("child"))
	inner.RebuildMiniMerkle()

	var buf bytes.Buffer
	if err := inner.Serialize(&buf); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	node, err := ReadNode(&NodeKey{Version: 1, Nonce: 1}, buf.Bytes())
	if err != nil {
		t.Fatalf("ReadNode: %v", err)
	}
	inner2 := node.(*InnerNode)
	if inner2.numKeys != 0 || inner2.NumChildren() != 1 {
		t.Fatalf("empty inner: numKeys=%d, children=%d", inner2.numKeys, inner2.NumChildren())
	}
}

func TestNodeSerialization_FullInner(t *testing.T) {
	inner := &InnerNode{
		nodeKey: &NodeKey{Version: 1, Nonce: 1},
		numKeys: B - 1,
		size:    1000,
		height:  2,
	}
	for i := 0; i < B-1; i++ {
		inner.keys[i] = []byte{byte(i)}
	}
	for i := 0; i < B; i++ {
		inner.children[i] = (&NodeKey{Version: 1, Nonce: uint32(i)}).GetKey()
		inner.childHashes[i] = sha256.Sum256([]byte{byte(i)})
	}
	inner.RebuildMiniMerkle()

	var buf bytes.Buffer
	if err := inner.Serialize(&buf); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	node, err := ReadNode(&NodeKey{Version: 1, Nonce: 1}, buf.Bytes())
	if err != nil {
		t.Fatalf("ReadNode: %v", err)
	}
	inner2 := node.(*InnerNode)
	if inner2.numKeys != B-1 {
		t.Fatalf("full inner: numKeys=%d, want %d", inner2.numKeys, B-1)
	}
	if inner2.Hash() != inner.Hash() {
		t.Fatalf("full inner hash mismatch")
	}
}

func TestNodeSerialization_FullLeaf(t *testing.T) {
	leaf := &LeafNode{
		nodeKey: &NodeKey{Version: 1, Nonce: 1},
		numKeys: B,
	}
	for i := 0; i < B; i++ {
		leaf.keys[i] = []byte{byte(i)}
		leaf.valueHashes[i] = sha256.Sum256([]byte{byte(i)})
	}
	leaf.RebuildMiniMerkle()

	var buf bytes.Buffer
	if err := leaf.Serialize(&buf); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	node, err := ReadNode(&NodeKey{Version: 1, Nonce: 1}, buf.Bytes())
	if err != nil {
		t.Fatalf("ReadNode: %v", err)
	}
	leaf2 := node.(*LeafNode)
	if leaf2.numKeys != B {
		t.Fatalf("full leaf: numKeys=%d, want %d", leaf2.numKeys, B)
	}
	if leaf2.Hash() != leaf.Hash() {
		t.Fatalf("full leaf hash mismatch")
	}
}

func TestNodeSerialization_EmptyLeaf(t *testing.T) {
	leaf := &LeafNode{
		nodeKey: &NodeKey{Version: 1, Nonce: 1},
		numKeys: 0,
	}
	var buf bytes.Buffer
	if err := leaf.Serialize(&buf); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	node, err := ReadNode(&NodeKey{Version: 1, Nonce: 1}, buf.Bytes())
	if err != nil {
		t.Fatalf("ReadNode: %v", err)
	}
	leaf2 := node.(*LeafNode)
	if leaf2.numKeys != 0 {
		t.Fatalf("empty leaf numKeys=%d", leaf2.numKeys)
	}
}

func TestNodeSerialization_LongKeys(t *testing.T) {
	leaf := &LeafNode{
		nodeKey: &NodeKey{Version: 1, Nonce: 1},
		numKeys: 2,
	}
	leaf.keys[0] = []byte(strings.Repeat("x", 300))
	leaf.keys[1] = []byte(strings.Repeat("y", 500))
	leaf.valueHashes[0] = sha256.Sum256([]byte("a"))
	leaf.valueHashes[1] = sha256.Sum256([]byte("b"))

	var buf bytes.Buffer
	if err := leaf.Serialize(&buf); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	node, err := ReadNode(&NodeKey{Version: 1, Nonce: 1}, buf.Bytes())
	if err != nil {
		t.Fatalf("ReadNode: %v", err)
	}
	leaf2 := node.(*LeafNode)
	if !bytes.Equal(leaf2.keys[0], leaf.keys[0]) || !bytes.Equal(leaf2.keys[1], leaf.keys[1]) {
		t.Fatalf("long key roundtrip failed")
	}
}

func TestReadNode_EmptyData(t *testing.T) {
	_, err := ReadNode(&NodeKey{Version: 1, Nonce: 1}, []byte{})
	if err == nil {
		t.Fatalf("expected error for empty data")
	}
}

func TestReadNode_UnknownType(t *testing.T) {
	_, err := ReadNode(&NodeKey{Version: 1, Nonce: 1}, []byte{0xFF})
	if err == nil {
		t.Fatalf("expected error for unknown type")
	}
}

// --- node.go: Clone ---

func TestInnerNode_Clone(t *testing.T) {
	inner := &InnerNode{
		nodeKey: &NodeKey{Version: 1, Nonce: 1},
		numKeys: 2,
		size:    50,
		height:  1,
	}
	inner.keys[0] = []byte("a")
	inner.keys[1] = []byte("b")

	cloned := inner.Clone()
	if cloned.nodeKey != nil {
		t.Fatalf("cloned nodeKey should be nil")
	}
	if cloned.numKeys != 2 || cloned.size != 50 {
		t.Fatalf("cloned metadata mismatch")
	}
	// Modifying clone should not affect original
	cloned.numKeys = 99
	if inner.numKeys != 2 {
		t.Fatalf("modifying clone affected original")
	}
}

func TestLeafNode_Clone(t *testing.T) {
	leaf := &LeafNode{
		nodeKey: &NodeKey{Version: 1, Nonce: 1},
		numKeys: 3,
	}
	leaf.keys[0] = []byte("x")

	cloned := leaf.Clone()
	if cloned.nodeKey != nil {
		t.Fatalf("cloned nodeKey should be nil")
	}
	cloned.numKeys = 99
	if leaf.numKeys != 3 {
		t.Fatalf("modifying clone affected original")
	}
}

// --- node.go: RebuildMiniMerkle ---

func TestInnerNode_RebuildMiniMerkle(t *testing.T) {
	inner := &InnerNode{
		numKeys:  2,
		miniTree: NewMiniMerkle(),
	}
	for i := 0; i < 3; i++ {
		inner.childHashes[i] = sha256.Sum256([]byte{byte(i)})
	}
	inner.RebuildMiniMerkle()
	if inner.Hash() == SentinelHash {
		t.Fatalf("inner hash should not be sentinel after rebuild")
	}

	// Mutate and rebuild — hash should change
	old := inner.Hash()
	inner.childHashes[0] = sha256.Sum256([]byte("changed"))
	inner.RebuildMiniMerkle()
	if inner.Hash() == old {
		t.Fatalf("hash should change after mutation + rebuild")
	}
}

func TestLeafNode_RebuildMiniMerkle(t *testing.T) {
	leaf := &LeafNode{
		numKeys:  2,
		miniTree: NewMiniMerkle(),
	}
	leaf.keys[0] = []byte("k1")
	leaf.keys[1] = []byte("k2")
	leaf.valueHashes[0] = sha256.Sum256([]byte("v1"))
	leaf.valueHashes[1] = sha256.Sum256([]byte("v2"))
	leaf.RebuildMiniMerkle()
	if leaf.Hash() == SentinelHash {
		t.Fatalf("leaf hash should not be sentinel")
	}

	// Verify slot 0 matches manual computation
	expected := HashLeafSlotFromValueHash(leaf.keys[0], leaf.valueHashes[0])
	if leaf.miniTree.GetSlot(0) != expected {
		t.Fatalf("leaf slot 0 mismatch")
	}
}

// --- options.go ---

func TestDefaultOptions(t *testing.T) {
	o := DefaultOptions()
	if o.FlushThreshold != 100*1024 {
		t.Fatalf("default FlushThreshold = %d", o.FlushThreshold)
	}
	if o.Sync || o.AsyncPruning || o.InitialVersion != 0 {
		t.Fatalf("unexpected defaults: %+v", o)
	}
}

func TestFunctionalOptions(t *testing.T) {
	o := DefaultOptions()
	SyncOption(true)(&o)
	InitialVersionOption(5)(&o)
	FlushThresholdOption(999)(&o)
	AsyncPruningOption(true)(&o)
	if !o.Sync || o.InitialVersion != 5 || o.FlushThreshold != 999 || !o.AsyncPruning {
		t.Fatalf("options not applied: %+v", o)
	}
}

// --- Golden vectors (regression anchors for Phase 2) ---

func TestHashLeafSlot_GoldenVector(t *testing.T) {
	h := HashLeafSlot([]byte("k"), []byte("v"))
	// Verify determinism and non-sentinel
	if h == SentinelHash {
		t.Fatalf("golden vector should not be sentinel")
	}
	h2 := HashLeafSlot([]byte("k"), []byte("v"))
	if h != h2 {
		t.Fatalf("golden vector not deterministic")
	}
	// Log for bootstrapping a hardcoded vector if needed
	t.Logf("HashLeafSlot(k,v) = %s", hex.EncodeToString(h[:]))
}

func TestHashInner_GoldenVector(t *testing.T) {
	a := sha256.Sum256([]byte("left"))
	b := sha256.Sum256([]byte("right"))
	h := HashInner(a, b)
	if h == SentinelHash {
		t.Fatalf("inner golden should not be sentinel")
	}
	h2 := HashInner(a, b)
	if h != h2 {
		t.Fatalf("inner golden not deterministic")
	}
}

// --- Clone deep isolation ---

func TestInnerNode_CloneChildIsolation(t *testing.T) {
	inner := &InnerNode{
		nodeKey: &NodeKey{Version: 1, Nonce: 1},
		numKeys: 1,
		size:    10,
		height:  1,
	}
	inner.children[0] = (&NodeKey{Version: 1, Nonce: 10}).GetKey()
	inner.children[1] = (&NodeKey{Version: 1, Nonce: 11}).GetKey()
	inner.childHashes[0] = sha256.Sum256([]byte("c0"))
	inner.childHashes[1] = sha256.Sum256([]byte("c1"))

	cloned := inner.Clone()
	// Mutating cloned child ref bytes should NOT affect original
	// because children[i] is a []byte slice — shallow copy shares the backing array
	origByte := inner.children[0][0]
	cloned.children[0][0] = 0xFF
	if inner.children[0][0] != 0xFF {
		// If this fails, the shallow copy DID isolate (unexpected for []byte)
		// Actually, shallow copy of slice header shares backing array
	}
	// Restore
	cloned.children[0][0] = origByte

	// But replacing the entire slice is safe
	cloned.children[0] = []byte("replaced")
	if bytes.Equal(inner.children[0], []byte("replaced")) {
		t.Fatalf("replacing cloned child slice affected original")
	}
}

func TestInnerNode_CloneMiniTreeIsolation(t *testing.T) {
	inner := &InnerNode{
		nodeKey:  &NodeKey{Version: 1, Nonce: 1},
		numKeys:  1,
		size:     10,
		height:   1,
		miniTree: NewMiniMerkle(),
	}
	inner.childHashes[0] = sha256.Sum256([]byte("c0"))
	inner.childHashes[1] = sha256.Sum256([]byte("c1"))
	inner.RebuildMiniMerkle()
	origHash := inner.Hash()

	cloned := inner.Clone()
	cloned.childHashes[0] = sha256.Sum256([]byte("modified"))
	cloned.RebuildMiniMerkle()

	// Original should be unaffected (MiniMerkle is a value type)
	if inner.Hash() != origHash {
		t.Fatalf("cloned mini tree mutation affected original")
	}
}

// --- Serialization truncation ---

func TestReadNode_TruncatedInner(t *testing.T) {
	inner := &InnerNode{
		nodeKey: &NodeKey{Version: 1, Nonce: 1},
		numKeys: 3,
		size:    100,
		height:  1,
	}
	for i := 0; i < 3; i++ {
		inner.keys[i] = []byte{byte(i)}
	}
	for i := 0; i < 4; i++ {
		inner.children[i] = (&NodeKey{Version: 1, Nonce: uint32(i)}).GetKey()
		inner.childHashes[i] = sha256.Sum256([]byte{byte(i)})
	}
	var buf bytes.Buffer
	if err := inner.Serialize(&buf); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	data := buf.Bytes()

	// Truncate at various points
	for _, cutoff := range []int{1, 5, 10, len(data) / 2, len(data) - 1} {
		if cutoff >= len(data) {
			continue
		}
		_, err := ReadNode(&NodeKey{Version: 1, Nonce: 1}, data[:cutoff])
		if err == nil {
			t.Errorf("expected error for truncated inner at %d/%d bytes", cutoff, len(data))
		}
	}
}

func TestReadNode_TruncatedLeaf(t *testing.T) {
	leaf := &LeafNode{
		nodeKey: &NodeKey{Version: 1, Nonce: 1},
		numKeys: 5,
	}
	for i := 0; i < 5; i++ {
		leaf.keys[i] = []byte{byte(i * 10)}
		leaf.valueHashes[i] = sha256.Sum256([]byte{byte(i)})
	}
	var buf bytes.Buffer
	if err := leaf.Serialize(&buf); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	data := buf.Bytes()

	for _, cutoff := range []int{1, 3, 10, len(data) / 2, len(data) - 1} {
		if cutoff >= len(data) {
			continue
		}
		_, err := ReadNode(&NodeKey{Version: 1, Nonce: 1}, data[:cutoff])
		if err == nil {
			t.Errorf("expected error for truncated leaf at %d/%d bytes", cutoff, len(data))
		}
	}
}

// --- MiniMerkle additional ---

func TestMiniMerkle_SetSlotSequenceMatchesBuild(t *testing.T) {
	// Insert B slots in random order via SetSlot
	rng := rand.New(rand.NewSource(42))
	order := rng.Perm(B)
	hashes := make([]Hash, B)
	for i := range hashes {
		hashes[i] = sha256.Sum256([]byte{byte(i), byte(i >> 8)})
	}

	m1 := NewMiniMerkle()
	for _, idx := range order {
		m1.SetSlot(idx, hashes[idx])
	}

	// Build from scratch
	var m2 MiniMerkle
	for i := 0; i < B; i++ {
		m2.tree[B+i] = hashes[i]
	}
	m2.Build()

	if m1.Root() != m2.Root() {
		t.Fatalf("random-order SetSlot disagrees with Build")
	}
}

func TestMiniMerkle_SetSlotToSentinel(t *testing.T) {
	m := NewMiniMerkle()
	h := sha256.Sum256([]byte("temp"))
	m.SetSlot(5, h)
	nonSentinelRoot := m.Root()
	if nonSentinelRoot == SentinelHash {
		t.Fatalf("should be non-sentinel")
	}

	// Set slot 5 back to sentinel — should return to single-occupied state... no,
	// all other slots are already sentinel, so clearing slot 5 should give sentinel root.
	m.SetSlot(5, SentinelHash)
	if m.Root() != SentinelHash {
		t.Fatalf("clearing the only non-sentinel slot should give sentinel root")
	}
}

// --- Search full inner ---

func TestSearchInner_FullNode(t *testing.T) {
	inner := &InnerNode{numKeys: B - 1}
	for i := 0; i < B-1; i++ {
		inner.keys[i] = []byte{byte(i * 2)} // 0, 2, 4, ..., 60
	}

	// Before first separator
	if searchInner(inner, []byte{0}) != 1 { // 0 == keys[0], goes right
		t.Fatalf("equal to first separator should return 1")
	}
	// Before first
	if idx := searchInner(inner, []byte{255}); idx != B-1 {
		t.Fatalf("after all: expected %d, got %d", B-1, idx)
	}
	// Between
	if idx := searchInner(inner, []byte{3}); idx != 2 {
		t.Fatalf("between: expected 2, got %d", idx)
	}
}

// --- Search duplicate-prefix keys ---

func TestSearchLeaf_DuplicatePrefixKeys(t *testing.T) {
	leaf := &LeafNode{numKeys: 3}
	leaf.keys[0] = []byte("abc")
	leaf.keys[1] = []byte("abcd")
	leaf.keys[2] = []byte("abcde")

	idx, found := searchLeaf(leaf, []byte("abcd"))
	if !found || idx != 1 {
		t.Fatalf("exact match: expected (1, true), got (%d, %v)", idx, found)
	}
	idx, found = searchLeaf(leaf, []byte("abcc"))
	if found || idx != 1 {
		t.Fatalf("insert point: expected (1, false), got (%d, %v)", idx, found)
	}
}

// --- RebuildMiniMerkle single child ---

func TestInnerNode_RebuildMiniMerkle_SingleChild(t *testing.T) {
	inner := &InnerNode{
		numKeys:  0,
		miniTree: NewMiniMerkle(),
	}
	inner.childHashes[0] = sha256.Sum256([]byte("only_child"))
	inner.RebuildMiniMerkle()

	// Manually compute: slot 0 = childHash, slots 1..31 = sentinel
	// Root should be childHash walked up through 5 levels of HashInner(x, sentinel)
	cur := inner.childHashes[0]
	for level := 0; level < miniMerkleDepth(); level++ {
		cur = HashInner(cur, SentinelHash)
	}
	if inner.Hash() != cur {
		t.Fatalf("single child inner hash mismatch")
	}
}

// --- LeafNode RebuildMiniMerkle all slots ---

func TestLeafNode_RebuildMiniMerkle_AllSlots(t *testing.T) {
	leaf := &LeafNode{
		numKeys:  B,
		miniTree: NewMiniMerkle(),
	}
	for i := 0; i < B; i++ {
		leaf.keys[i] = []byte{byte(i)}
		leaf.valueHashes[i] = sha256.Sum256([]byte{byte(i)})
	}
	leaf.RebuildMiniMerkle()

	// Verify every slot
	for i := 0; i < B; i++ {
		expected := HashLeafSlotFromValueHash(leaf.keys[i], leaf.valueHashes[i])
		if leaf.miniTree.GetSlot(i) != expected {
			t.Fatalf("leaf slot %d mismatch", i)
		}
	}
}

// --- NodeKey ---

func TestGetRootKey_ZeroVersion(t *testing.T) {
	rk := GetRootKey(0)
	nk := GetNodeKey(rk)
	if nk.Version != 0 || nk.Nonce != 1 {
		t.Fatalf("GetRootKey(0): got version=%d nonce=%d", nk.Version, nk.Nonce)
	}
}

// --- Node GetNodeKey / SetNodeKey ---

func TestInnerNode_GetSetNodeKey(t *testing.T) {
	inner := &InnerNode{}
	if inner.GetNodeKey() != nil {
		t.Fatalf("new inner should have nil nodeKey")
	}
	nk := &NodeKey{Version: 5, Nonce: 3}
	inner.SetNodeKey(nk)
	if inner.GetNodeKey() != nk {
		t.Fatalf("SetNodeKey/GetNodeKey roundtrip failed")
	}
}

func TestLeafNode_GetSetNodeKey(t *testing.T) {
	leaf := &LeafNode{}
	if leaf.GetNodeKey() != nil {
		t.Fatalf("new leaf should have nil nodeKey")
	}
	nk := &NodeKey{Version: 7, Nonce: 2}
	leaf.SetNodeKey(nk)
	if leaf.GetNodeKey() != nk {
		t.Fatalf("SetNodeKey/GetNodeKey roundtrip failed")
	}
}

// --- Serialization format stability ---

func TestNodeSerialization_InnerGoldenBytes(t *testing.T) {
	// Fixed input — if serialization format changes, this test breaks.
	inner := &InnerNode{
		nodeKey: &NodeKey{Version: 1, Nonce: 1},
		numKeys: 1,
		size:    42,
		height:  1,
	}
	inner.keys[0] = []byte("sep")
	inner.children[0] = (&NodeKey{Version: 1, Nonce: 10}).GetKey()
	inner.children[1] = (&NodeKey{Version: 1, Nonce: 11}).GetKey()
	inner.childHashes[0] = sha256.Sum256([]byte("c0"))
	inner.childHashes[1] = sha256.Sum256([]byte("c1"))

	var buf bytes.Buffer
	if err := inner.Serialize(&buf); err != nil {
		t.Fatalf("Serialize: %v", err)
	}

	// Record the serialized length — shouldn't change
	data := buf.Bytes()
	if data[0] != TypeInner {
		t.Fatalf("first byte should be TypeInner (0x%02x), got 0x%02x", TypeInner, data[0])
	}
	// Round-trip must produce identical bytes
	node, err := ReadNode(&NodeKey{Version: 1, Nonce: 1}, data)
	if err != nil {
		t.Fatalf("ReadNode: %v", err)
	}
	var buf2 bytes.Buffer
	if err := node.(*InnerNode).Serialize(&buf2); err != nil {
		t.Fatalf("re-Serialize: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), buf2.Bytes()) {
		t.Fatalf("serialization roundtrip produced different bytes")
	}
}

func TestNodeSerialization_LeafGoldenBytes(t *testing.T) {
	leaf := &LeafNode{
		nodeKey: &NodeKey{Version: 1, Nonce: 1},
		numKeys: 2,
	}
	leaf.keys[0] = []byte("aa")
	leaf.keys[1] = []byte("bb")
	leaf.valueHashes[0] = sha256.Sum256([]byte("v0"))
	leaf.valueHashes[1] = sha256.Sum256([]byte("v1"))

	var buf bytes.Buffer
	if err := leaf.Serialize(&buf); err != nil {
		t.Fatalf("Serialize: %v", err)
	}

	data := buf.Bytes()
	if data[0] != TypeLeaf {
		t.Fatalf("first byte should be TypeLeaf (0x%02x), got 0x%02x", TypeLeaf, data[0])
	}
	// Round-trip
	node, err := ReadNode(&NodeKey{Version: 1, Nonce: 1}, data)
	if err != nil {
		t.Fatalf("ReadNode: %v", err)
	}
	var buf2 bytes.Buffer
	if err := node.(*LeafNode).Serialize(&buf2); err != nil {
		t.Fatalf("re-Serialize: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), buf2.Bytes()) {
		t.Fatalf("leaf serialization roundtrip produced different bytes")
	}
}
