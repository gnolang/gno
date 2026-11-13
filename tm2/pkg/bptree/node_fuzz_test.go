package bptree

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
)

// FuzzReadNode fuzzes the node decoder. The decoder reads from a byteCursor
// over the raw payload with hand-written bounds and caps, where it used to
// read through bytes.Reader + io.ReadFull and inherit the stdlib's; this is
// the enforcement for that whole surface rather than for the handful of
// malformed shapes the table-driven tests name.
//
// Two properties are asserted. Nothing may panic, however malformed the
// payload. And whatever decodes must re-serialize to a payload that decodes
// to the same thing — the writer emits TypeLeafPrefixed for every leaf, so a
// legacy TypeLeaf record is normalised on the first round trip and compared
// at the second.
func FuzzReadNode(f *testing.F) {
	// Seed with well-formed records of every shape the writers emit, so the
	// fuzzer mutates around valid framing rather than only rejecting noise.
	seed := func(n Node) {
		payload, err := appendNodeRecord(n)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(payload[:len(payload)-checksumSize]) // ReadNode sees the payload, not the record
	}
	seed(makeLeaf([]string{"a"}))
	seed(makeLeaf([]string{"", "a", "ab"}))
	seed(makeLeaf([]string{
		"oid:0123456789abcdef0123456789abcdef01234567:0001",
		"oid:0123456789abcdef0123456789abcdef01234567:0002",
	}))
	full := make([]string, B)
	for i := range full {
		full[i] = fmt.Sprintf("oid:0123456789abcdef0123456789abcdef01234567:%04d", i)
	}
	seed(makeLeaf(full))
	// The legacy leaf layout, which the reader must keep accepting.
	f.Add(appendLeafV1(makeLeaf([]string{"aa", "ab", "ac"})))

	inner := &InnerNode{numKeys: 2, height: 1}
	inner.keys[0] = []byte("sep-a")
	inner.keys[1] = []byte("sep-b")
	for i := range 3 {
		inner.children[i] = (&NodeKey{Version: int64(i + 1), Nonce: uint32(i + 1)}).GetKey()
		inner.childSizes[i] = int64(i)
	}
	inner.RebuildMiniMerkle()
	seed(inner)

	nk := &NodeKey{Version: 7, Nonce: 9}
	f.Fuzz(func(t *testing.T, data []byte) {
		node, err := ReadNode(nk, data)
		if err != nil {
			return // rejecting malformed input is the expected outcome
		}
		if node == nil {
			t.Fatal("ReadNode returned (nil, nil)")
		}
		// Bounds the decoder promises its callers.
		switch n := node.(type) {
		case *LeafNode:
			if n.numKeys < 0 || n.numKeys > B {
				t.Fatalf("leaf numKeys %d out of range", n.numKeys)
			}
			for i := range int(n.numKeys) {
				if len(n.keys[i]) > maxReadBytesLen {
					t.Fatalf("leaf key %d is %d bytes", i, len(n.keys[i]))
				}
				if len(n.valueKeys[i]) != NodeKeySize {
					t.Fatalf("leaf valueKey %d is %d bytes", i, len(n.valueKeys[i]))
				}
			}
		case *InnerNode:
			if n.numKeys < 0 || n.numKeys > B-1 {
				t.Fatalf("inner numKeys %d out of range", n.numKeys)
			}
			for i := range n.NumChildren() {
				if len(n.children[i]) != NodeKeySize {
					t.Fatalf("inner child ref %d is %d bytes", i, len(n.children[i]))
				}
			}
		}

		// Round trip. The first re-serialization normalises the format; from
		// there it must be a fixed point, byte for byte.
		once, err := appendNodeSerializedForFuzz(node)
		if err != nil {
			return // a decoded node may still be unserializable (e.g. an inner with a nil ref)
		}
		reread, err := ReadNode(nk, once)
		if err != nil {
			t.Fatalf("re-reading our own serialization failed: %v", err)
		}
		twice, err := appendNodeSerializedForFuzz(reread)
		if err != nil {
			t.Fatalf("re-serializing our own round trip failed: %v", err)
		}
		if !bytes.Equal(once, twice) {
			t.Fatalf("serialization is not a fixed point:\n %x\n %x", once, twice)
		}
		if reread.Hash() != node.Hash() {
			t.Fatalf("hash changed across a round trip: %x != %x", reread.Hash(), node.Hash())
		}
	})
}

func appendNodeSerializedForFuzz(node Node) ([]byte, error) {
	rec, err := appendNodeRecord(node)
	if err != nil {
		return nil, err
	}
	return rec[:len(rec)-checksumSize], nil
}

// TestByteCursorRejectsTruncation covers the cursor primitives directly: each
// must report EOF rather than reading past the payload or panicking.
func TestByteCursorRejectsTruncation(t *testing.T) {
	c := byteCursor{b: nil}
	if _, err := c.uvarint(); err == nil {
		t.Fatal("uvarint on an empty cursor succeeded")
	}
	if _, err := c.varint(); err == nil {
		t.Fatal("varint on an empty cursor succeeded")
	}
	if _, err := c.next(1); err == nil {
		t.Fatal("next(1) on an empty cursor succeeded")
	}
	if got, err := c.next(0); err != nil || len(got) != 0 {
		t.Fatalf("next(0) = %v, %v; want empty, nil", got, err)
	}
	if c.remaining() != 0 {
		t.Fatalf("remaining = %d, want 0", c.remaining())
	}
	// A uvarint that overflows 64 bits must be an error, not a wrapped value.
	c = byteCursor{b: bytes.Repeat([]byte{0xFF}, 11)}
	if _, err := c.uvarint(); err == nil {
		t.Fatal("overflowing uvarint succeeded")
	}
	// A length prefix that outruns the payload must not allocate it.
	c = byteCursor{b: binary.AppendUvarint(nil, 1<<19)}
	var dst [B][]byte
	if err := c.readKeys(dst[:], 1); err == nil {
		t.Fatal("readKeys past the payload succeeded")
	}
}
