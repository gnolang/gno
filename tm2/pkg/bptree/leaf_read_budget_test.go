package bptree

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"
)

// Tests for the cumulative per-leaf allocation cap (maxLeafReadBytes)
// that bounds OOM-class allocations from crafted DB blobs in all three
// leaf readers (v1/v2/v3). Without the cap, the per-field
// maxReadBytesLen bound is multiplied by B = 32 slots, letting a blob
// allocate up to ~2 MiB per leaf via the per-key bound alone. v3
// amplifies further: each reconstructed key includes a fresh copy of
// the common prefix.

// TestLeafReadV3_RejectsPrefixAmplification crafts a v3 leaf whose
// per-key length passes the per-field cap but whose cumulative
// (B copies of prefix + suffixes) exceeds the per-leaf budget, and
// asserts the reader rejects it with the cumulative-budget error.
func TestLeafReadV3_RejectsPrefixAmplification(t *testing.T) {
	const (
		numKeys   = 32       // = B
		prefixLen = 16 << 10 // 16 KiB — comfortably below per-field cap (64 KiB)
		suffixLen = 1        // tiny suffix per key
		// Cumulative key bytes ≈ 32 × (16 KiB + 1) = 512 KiB ≫ 256 KiB
		// budget. The cumulative check trips before the per-key check.
	)

	var buf bytes.Buffer
	buf.WriteByte(TypeLeafV3)
	tmp := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(tmp, uint64(numKeys))
	buf.Write(tmp[:n])
	n = binary.PutUvarint(tmp, uint64(prefixLen))
	buf.Write(tmp[:n])
	prefix := strings.Repeat("a", prefixLen)
	buf.WriteString(prefix)
	for i := 0; i < numKeys; i++ {
		n := binary.PutUvarint(tmp, uint64(suffixLen))
		buf.Write(tmp[:n])
		buf.WriteByte(byte(i)) // unique 1-byte suffix
	}
	// Trailing valueHashes / inlineMask / payloads are unnecessary —
	// the cumulative-budget check fails before the reader gets there.

	nk := &NodeKey{Version: 1, Nonce: 1}
	_, err := ReadNode(nk, buf.Bytes())
	if err == nil {
		t.Fatal("expected ReadNode to reject crafted v3 blob exceeding maxLeafReadBytes, got nil error")
	}
	if !strings.Contains(err.Error(), "leaf cumulative") {
		t.Fatalf("expected cumulative-budget error, got: %v", err)
	}
}

// TestLeafReadV2_RejectsInlineAmplification crafts a v2 leaf whose
// inline-value payloads alone exceed the per-leaf budget, and asserts
// the reader rejects it.
func TestLeafReadV2_RejectsInlineAmplification(t *testing.T) {
	const (
		numKeys   = 32
		keyLen    = 4
		inlineLen = 16 << 10 // 16 KiB each — 32 × 16 KiB = 512 KiB > 256 KiB budget
	)

	var buf bytes.Buffer
	buf.WriteByte(TypeLeafV2)
	tmp := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(tmp, uint64(numKeys))
	buf.Write(tmp[:n])
	for i := 0; i < numKeys; i++ {
		n := binary.PutUvarint(tmp, uint64(keyLen))
		buf.Write(tmp[:n])
		buf.Write([]byte{byte(i), 0, 0, 0})
	}
	for i := 0; i < numKeys; i++ {
		var hash [HashSize]byte
		buf.Write(hash[:])
	}
	var mask [4]byte
	binary.BigEndian.PutUint32(mask[:], 0xFFFFFFFF)
	buf.Write(mask[:])
	payload := make([]byte, inlineLen)
	for i := 0; i < numKeys; i++ {
		n := binary.PutUvarint(tmp, uint64(inlineLen))
		buf.Write(tmp[:n])
		buf.Write(payload)
	}

	nk := &NodeKey{Version: 1, Nonce: 1}
	_, err := ReadNode(nk, buf.Bytes())
	if err == nil {
		t.Fatal("expected ReadNode to reject crafted v2 blob exceeding maxLeafReadBytes, got nil error")
	}
	if !strings.Contains(err.Error(), "leaf cumulative") {
		t.Fatalf("expected cumulative-budget error, got: %v", err)
	}
}
