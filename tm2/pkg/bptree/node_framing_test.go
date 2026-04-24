package bptree

import (
	"bytes"
	"strings"
	"testing"
)

// TestReadNode_RejectsTrailingBytes verifies ReadNode fails cleanly on
// payloads with extra bytes after the type-specific decode. Silently
// ignoring trailing bytes would mask on-disk corruption as "successfully
// decoded a truncated view".
func TestReadNode_RejectsTrailingBytes(t *testing.T) {
	// Serialize a valid leaf.
	leaf := &LeafNode{miniTree: NewMiniMerkle()}
	leaf.numKeys = 1
	leaf.keys[0] = []byte("k")
	leaf.valueHashes[0] = HashLeafSlot(leaf.keys[0], []byte("v"))
	leaf.valueKeys[0] = (&NodeKey{Version: 1, Nonce: 1}).GetKey()

	var buf bytes.Buffer
	if err := leaf.Serialize(&buf); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	clean := buf.Bytes()

	// Clean round-trip succeeds.
	if _, err := ReadNode(&NodeKey{Version: 1, Nonce: 2}, clean); err != nil {
		t.Fatalf("ReadNode(clean): %v", err)
	}

	// Corrupt by appending garbage.
	corrupt := append(append([]byte(nil), clean...), 0xAA, 0xBB, 0xCC)
	_, err := ReadNode(&NodeKey{Version: 1, Nonce: 2}, corrupt)
	if err == nil {
		t.Fatalf("ReadNode(corrupt) succeeded; expected trailing-bytes error")
	}
	if !strings.Contains(err.Error(), "trailing bytes") {
		t.Fatalf("error does not mention trailing bytes: %v", err)
	}
}

// TestReadNode_RejectsTrailingBytes_Inner does the same for inner nodes.
func TestReadNode_RejectsTrailingBytes_Inner(t *testing.T) {
	inner := &InnerNode{miniTree: NewMiniMerkle()}
	inner.numKeys = 1
	inner.height = 1
	inner.keys[0] = []byte("sep")
	inner.children[0] = (&NodeKey{Version: 1, Nonce: 1}).GetKey()
	inner.children[1] = (&NodeKey{Version: 1, Nonce: 2}).GetKey()
	inner.childSizes[0] = 1
	inner.childSizes[1] = 1

	var buf bytes.Buffer
	if err := inner.Serialize(&buf); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	clean := buf.Bytes()

	if _, err := ReadNode(&NodeKey{Version: 1, Nonce: 3}, clean); err != nil {
		t.Fatalf("ReadNode(clean inner): %v", err)
	}

	corrupt := append(append([]byte(nil), clean...), 0x99)
	_, err := ReadNode(&NodeKey{Version: 1, Nonce: 3}, corrupt)
	if err == nil {
		t.Fatalf("ReadNode(corrupt inner) succeeded; expected trailing-bytes error")
	}
	if !strings.Contains(err.Error(), "trailing bytes") {
		t.Fatalf("error does not mention trailing bytes: %v", err)
	}
}
