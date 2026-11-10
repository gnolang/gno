package bptree

import (
	"encoding/binary"
	"math"
	"math/rand"
	"testing"
)

// TestSerializedSizeIsExact pins serializedSize against appendSerialized.
// SaveNode allocates the record once, sized by serializedSize; if the two
// ever drift the only symptom is a silent extra allocation (too small) or a
// wasted tail (too large), so nothing else would catch it.
func TestSerializedSizeIsExact(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	for iter := range 500 {
		// Inner node with a random shape, including the varint edge cases
		// (0, negative, and multi-byte childSizes).
		nkeys := r.Intn(B) // 0..B-1 separator keys
		inner := &InnerNode{numKeys: int16(nkeys), height: int16(r.Intn(1 << 14)), miniTree: NewMiniMerkle()}
		for i := range nkeys {
			inner.keys[i] = make([]byte, r.Intn(300))
			r.Read(inner.keys[i])
		}
		for i := 0; i <= nkeys; i++ {
			switch iter % 4 {
			case 0:
				inner.childSizes[i] = 0
			case 1:
				inner.childSizes[i] = int64(r.Intn(1 << 20))
			case 2:
				inner.childSizes[i] = -int64(r.Intn(1 << 20))
			default:
				inner.childSizes[i] = math.MaxInt64
			}
			inner.children[i] = (&NodeKey{Version: r.Int63(), Nonce: r.Uint32()}).GetKey()
		}
		got, err := inner.appendSerialized(nil)
		if err != nil {
			t.Fatalf("inner appendSerialized: %v", err)
		}
		if want := inner.serializedSize(); want != len(got) {
			t.Fatalf("inner serializedSize=%d, appendSerialized wrote %d", want, len(got))
		}

		nkeys = 1 + r.Intn(B)
		leaf := &LeafNode{numKeys: int16(nkeys), miniTree: NewMiniMerkle()}
		for i := range nkeys {
			leaf.keys[i] = make([]byte, r.Intn(300))
			r.Read(leaf.keys[i])
			leaf.valueKeys[i] = (&NodeKey{Version: r.Int63(), Nonce: r.Uint32()}).GetKey()
		}
		got, err = leaf.appendSerialized(nil)
		if err != nil {
			t.Fatalf("leaf appendSerialized: %v", err)
		}
		if want := leaf.serializedSize(); want != len(got) {
			t.Fatalf("leaf serializedSize=%d, appendSerialized wrote %d", want, len(got))
		}
	}
}

// TestVarintLenMatchesStdlib pins the two length helpers against the stdlib
// encoders they predict.
func TestVarintLenMatchesStdlib(t *testing.T) {
	vals := []uint64{0, 1, 2, 127, 128, 129, 16383, 16384, 1 << 20, 1 << 56, math.MaxUint64}
	for _, v := range vals {
		if got, want := uvarintLen(v), len(binary.AppendUvarint(nil, v)); got != want {
			t.Fatalf("uvarintLen(%d)=%d want %d", v, got, want)
		}
	}
	svals := []int64{0, 1, -1, 63, 64, -64, -65, math.MaxInt64, math.MinInt64, 1 << 40, -(1 << 40)}
	for _, v := range svals {
		if got, want := varintLen(v), len(binary.AppendVarint(nil, v)); got != want {
			t.Fatalf("varintLen(%d)=%d want %d", v, got, want)
		}
	}
	r := rand.New(rand.NewSource(7))
	for range 10_000 {
		v := int64(r.Uint64())
		if got, want := varintLen(v), len(binary.AppendVarint(nil, v)); got != want {
			t.Fatalf("varintLen(%d)=%d want %d", v, got, want)
		}
		u := r.Uint64()
		if got, want := uvarintLen(u), len(binary.AppendUvarint(nil, u)); got != want {
			t.Fatalf("uvarintLen(%d)=%d want %d", u, got, want)
		}
	}
}

// TestDecodedKeysDoNotAliasPayload guards the slab decode: keys and refs are
// copied out of the payload, which may alias DB-owned storage that the
// backend is free to reuse after the Get returns.
func TestDecodedKeysDoNotAliasPayload(t *testing.T) {
	leaf := &LeafNode{numKeys: 2, miniTree: NewMiniMerkle()}
	leaf.keys[0] = []byte("alpha")
	leaf.keys[1] = []byte("bravo")
	leaf.valueKeys[0] = (&NodeKey{Version: 1, Nonce: 1}).GetKey()
	leaf.valueKeys[1] = (&NodeKey{Version: 1, Nonce: 2}).GetKey()
	payload, err := leaf.appendSerialized(nil)
	if err != nil {
		t.Fatal(err)
	}
	node, err := ReadNode(&NodeKey{Version: 2, Nonce: 1}, payload)
	if err != nil {
		t.Fatal(err)
	}
	got := node.(*LeafNode)
	// Scribble over the payload the way a reused DB buffer would.
	for i := range payload {
		payload[i] = 0xFF
	}
	if string(got.keys[0]) != "alpha" || string(got.keys[1]) != "bravo" {
		t.Fatalf("keys alias the payload: %q %q", got.keys[0], got.keys[1])
	}
	if nk := GetNodeKey(got.valueKeys[1]); nk == nil || nk.Nonce != 2 {
		t.Fatalf("valueKeys alias the payload: %x", got.valueKeys[1])
	}
}
