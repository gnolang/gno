package bptree

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"testing"
)

// TestLeafV3_CommonPrefixRoundTrip covers the headline prefix-compression
// case: sorted keys share a long common prefix. Serialize -> ReadNode
// must reconstruct full keys and the hash tree must be identical.
func TestLeafV3_CommonPrefixRoundTrip(t *testing.T) {
	prefix := []byte("vm:realm/very/long/namespace:")
	keys := [][]byte{
		append(append([]byte{}, prefix...), []byte("aaaa")...),
		append(append([]byte{}, prefix...), []byte("bbbb")...),
		append(append([]byte{}, prefix...), []byte("cccc")...),
		append(append([]byte{}, prefix...), []byte("dddd")...),
	}

	leaf := &LeafNode{
		nodeKey:  &NodeKey{Version: 1, Nonce: 1},
		numKeys:  int16(len(keys)),
		miniTree: NewMiniMerkle(),
	}
	for i, k := range keys {
		leaf.keys[i] = k
		leaf.valueHashes[i] = sha256.Sum256([]byte{byte(i)})
	}
	leaf.RebuildMiniMerkle()

	var buf bytes.Buffer
	if err := leaf.Serialize(&buf); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	data := buf.Bytes()
	if data[0] != TypeLeafV3 {
		t.Fatalf("expected TypeLeafV3, got 0x%02x", data[0])
	}

	// Round-trip parses back full keys.
	got, err := ReadNode(&NodeKey{Version: 1, Nonce: 1}, data)
	if err != nil {
		t.Fatalf("ReadNode: %v", err)
	}
	gl := got.(*LeafNode)
	if gl.numKeys != leaf.numKeys {
		t.Fatalf("numKeys mismatch: %d vs %d", gl.numKeys, leaf.numKeys)
	}
	for i := 0; i < int(leaf.numKeys); i++ {
		if !bytes.Equal(gl.keys[i], leaf.keys[i]) {
			t.Fatalf("key %d mismatch: %q vs %q", i, gl.keys[i], leaf.keys[i])
		}
	}
	wh := leaf.Hash()
	gh := gl.Hash()
	if !bytes.Equal(wh[:], gh[:]) {
		t.Fatalf("hash mismatch after round-trip")
	}

	// Second serialize pass must be byte-identical — stable encoding.
	var buf2 bytes.Buffer
	if err := gl.Serialize(&buf2); err != nil {
		t.Fatalf("re-Serialize: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), buf2.Bytes()) {
		t.Fatalf("re-serialization produced different bytes")
	}
}

// TestLeafV3_NoCommonPrefix covers the degenerate case where keys share
// no leading bytes. The emitted commonPrefixLen is 0; each suffix is
// the full key.
func TestLeafV3_NoCommonPrefix(t *testing.T) {
	leaf := &LeafNode{nodeKey: &NodeKey{Version: 1, Nonce: 1}, numKeys: 3}
	leaf.keys[0] = []byte("apple")
	leaf.keys[1] = []byte("banana")
	leaf.keys[2] = []byte("cherry")
	leaf.valueHashes[0] = sha256.Sum256([]byte("0"))
	leaf.valueHashes[1] = sha256.Sum256([]byte("1"))
	leaf.valueHashes[2] = sha256.Sum256([]byte("2"))

	var buf bytes.Buffer
	if err := leaf.Serialize(&buf); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	got, err := ReadNode(&NodeKey{Version: 1, Nonce: 1}, buf.Bytes())
	if err != nil {
		t.Fatalf("ReadNode: %v", err)
	}
	gl := got.(*LeafNode)
	for i, want := range [][]byte{[]byte("apple"), []byte("banana"), []byte("cherry")} {
		if !bytes.Equal(gl.keys[i], want) {
			t.Fatalf("key %d = %q, want %q", i, gl.keys[i], want)
		}
	}
}

// TestLeafV3_SingleKey covers the edge case of a single-key leaf where
// "common prefix of first and last" collapses to the entire key.
func TestLeafV3_SingleKey(t *testing.T) {
	leaf := &LeafNode{nodeKey: &NodeKey{Version: 1, Nonce: 1}, numKeys: 1}
	leaf.keys[0] = []byte("solo")
	leaf.valueHashes[0] = sha256.Sum256([]byte("v"))

	var buf bytes.Buffer
	if err := leaf.Serialize(&buf); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	got, err := ReadNode(&NodeKey{Version: 1, Nonce: 1}, buf.Bytes())
	if err != nil {
		t.Fatalf("ReadNode: %v", err)
	}
	gl := got.(*LeafNode)
	if !bytes.Equal(gl.keys[0], []byte("solo")) {
		t.Fatalf("single-key round-trip failed: got %q", gl.keys[0])
	}
}

// TestLeafV3_EmptyLeaf covers the pathological 0-key leaf — the prefix
// block is omitted entirely so the reader must not attempt to parse it.
func TestLeafV3_EmptyLeaf(t *testing.T) {
	leaf := &LeafNode{nodeKey: &NodeKey{Version: 1, Nonce: 1}, numKeys: 0}
	var buf bytes.Buffer
	if err := leaf.Serialize(&buf); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	got, err := ReadNode(&NodeKey{Version: 1, Nonce: 1}, buf.Bytes())
	if err != nil {
		t.Fatalf("ReadNode: %v", err)
	}
	if got.(*LeafNode).numKeys != 0 {
		t.Fatalf("expected empty leaf, got numKeys=%d", got.(*LeafNode).numKeys)
	}
}

// TestLeafV3_SavesBytesOverV2 verifies that prefix compression actually
// shrinks the on-disk payload for a prefix-heavy key set — the whole
// point of the format bump.
func TestLeafV3_SavesBytesOverV2(t *testing.T) {
	prefix := []byte("vm:realm/very/long/namespace:") // 29 bytes
	numKeys := 10
	leaf := &LeafNode{nodeKey: &NodeKey{Version: 1, Nonce: 1}, numKeys: int16(numKeys)}
	for i := 0; i < numKeys; i++ {
		leaf.keys[i] = append(append([]byte{}, prefix...), byte('a'+i), byte('a'+i), byte('a'+i))
		leaf.valueHashes[i] = sha256.Sum256([]byte{byte(i)})
	}

	var v3 bytes.Buffer
	if err := leaf.Serialize(&v3); err != nil {
		t.Fatalf("v3 Serialize: %v", err)
	}
	// Hand-build v2 bytes for the same leaf for comparison. This is
	// byte-equivalent to what the pre-2.3 writer would have produced.
	var v2 bytes.Buffer
	v2.WriteByte(TypeLeafV2)
	writeUvarintBuf(&v2, uint64(leaf.numKeys))
	for i := 0; i < numKeys; i++ {
		writeBytesBuf(&v2, leaf.keys[i])
	}
	for i := 0; i < numKeys; i++ {
		v2.Write(leaf.valueHashes[i][:])
	}
	var maskBuf [4]byte
	binary.BigEndian.PutUint32(maskBuf[:], leaf.inlineMask)
	v2.Write(maskBuf[:])
	var zeroVK [NodeKeySize]byte
	for i := 0; i < numKeys; i++ {
		v2.Write(zeroVK[:])
	}

	if v3.Len() >= v2.Len() {
		t.Fatalf("v3 (%d bytes) did not shrink vs v2 (%d bytes) for prefix-heavy keys", v3.Len(), v2.Len())
	}
	saved := v2.Len() - v3.Len()
	wantMin := (len(prefix) - 1) * (numKeys - 1) // prefix emitted once vs N times
	if saved < wantMin {
		t.Fatalf("expected at least %d bytes saved, got %d (v2=%d, v3=%d)", wantMin, saved, v2.Len(), v3.Len())
	}
}

// TestLeafV3_ReadsV2Legacy verifies the reader still accepts TypeLeafV2
// payloads, so DBs written by the pre-2.3 writer continue to mount.
func TestLeafV3_ReadsV2Legacy(t *testing.T) {
	keys := [][]byte{[]byte("alpha"), []byte("beta"), []byte("gamma")}
	var buf bytes.Buffer
	buf.WriteByte(TypeLeafV2)
	writeUvarintBuf(&buf, uint64(len(keys)))
	for _, k := range keys {
		writeBytesBuf(&buf, k)
	}
	vhs := make([]Hash, len(keys))
	for i, k := range keys {
		vhs[i] = sha256.Sum256(k)
		buf.Write(vhs[i][:])
	}
	var mask [4]byte // all external
	buf.Write(mask[:])
	var zeroVK [NodeKeySize]byte
	for range keys {
		buf.Write(zeroVK[:])
	}

	got, err := ReadNode(&NodeKey{Version: 1, Nonce: 1}, buf.Bytes())
	if err != nil {
		t.Fatalf("ReadNode(v2): %v", err)
	}
	gl := got.(*LeafNode)
	for i, want := range keys {
		if !bytes.Equal(gl.keys[i], want) {
			t.Fatalf("v2-read key %d = %q, want %q", i, gl.keys[i], want)
		}
		if gl.valueHashes[i] != vhs[i] {
			t.Fatalf("v2-read valueHash %d mismatch", i)
		}
	}
}
