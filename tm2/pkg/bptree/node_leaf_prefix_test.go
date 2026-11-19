package bptree

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// appendLeafV1 writes a leaf in the legacy TypeLeaf layout (each key stored in
// full). Writers no longer emit it, so this is the only way left to produce
// one — the reader must keep accepting it for DBs written before
// TypeLeafPrefixed.
func appendLeafV1(n *LeafNode) []byte {
	b := []byte{TypeLeaf}
	b = binary.AppendUvarint(b, uint64(n.numKeys))
	for i := range int(n.numKeys) {
		b = binary.AppendUvarint(b, uint64(len(n.keys[i])))
		b = append(b, n.keys[i]...)
	}
	for i := range int(n.numKeys) {
		b = append(b, n.valueHashes[i][:]...)
	}
	for i := range int(n.numKeys) {
		b = append(b, n.valueKeys[i]...)
	}
	return b
}

func makeLeaf(keys []string) *LeafNode {
	leaf := &LeafNode{numKeys: int16(len(keys)), miniTree: NewMiniMerkle()}
	for i, k := range keys {
		leaf.keys[i] = []byte(k)
		leaf.valueHashes[i] = HashLeafSlot([]byte(k), fmt.Appendf(nil, "v%d", i))
		leaf.valueKeys[i] = (&NodeKey{Version: 1, Nonce: uint32(i + 1)}).GetKey()
	}
	leaf.RebuildMiniMerkle()
	return leaf
}

func assertLeafEqual(t *testing.T, want *LeafNode, got Node, ctx string) {
	t.Helper()
	gl, ok := got.(*LeafNode)
	if !ok {
		t.Fatalf("%s: decoded %T, want *LeafNode", ctx, got)
	}
	if gl.numKeys != want.numKeys {
		t.Fatalf("%s: numKeys %d, want %d", ctx, gl.numKeys, want.numKeys)
	}
	for i := range int(want.numKeys) {
		if !bytes.Equal(gl.keys[i], want.keys[i]) {
			t.Fatalf("%s: key %d = %q, want %q", ctx, i, gl.keys[i], want.keys[i])
		}
		if gl.valueHashes[i] != want.valueHashes[i] {
			t.Fatalf("%s: valueHash %d differs", ctx, i)
		}
		if !bytes.Equal(gl.valueKeys[i], want.valueKeys[i]) {
			t.Fatalf("%s: valueKey %d differs", ctx, i)
		}
	}
	if gl.Hash() != want.Hash() {
		t.Fatalf("%s: hash %x, want %x", ctx, gl.Hash(), want.Hash())
	}
}

// TestLeafPrefixed_RoundTrip covers the prefix edge cases: none shared, all
// shared, one key equal to the whole prefix, empty keys, a single key, and a
// full B-key leaf.
func TestLeafPrefixed_RoundTrip(t *testing.T) {
	cases := map[string][]string{
		"no-common-prefix":   {"alpha", "bravo", "charlie"},
		"long-common-prefix": {"oid:aabbccddeeff00112233445566778899aabbccdd:1", "oid:aabbccddeeff00112233445566778899aabbccdd:2", "oid:aabbccddeeff00112233445566778899aabbccdd:30"},
		"prefix-is-a-key":    {"abc", "abcd", "abcde"},
		"single-key":         {"only"},
		"empty-key":          {""},
		"empty-then-more":    {"", "a", "b"},
		"one-byte-keys":      {"a", "b", "c", "d"},
		"identical-prefixes": {"xxxxxxxxxxxxxxxx1", "xxxxxxxxxxxxxxxx2"},
	}
	// A full leaf of realistically-shaped keys.
	full := make([]string, B)
	for i := range full {
		full[i] = fmt.Sprintf("oid:0123456789abcdef0123456789abcdef01234567:%04d", i)
	}
	cases["full-leaf"] = full

	for name, keys := range cases {
		t.Run(name, func(t *testing.T) {
			leaf := makeLeaf(keys)
			rec, err := leaf.appendSerialized(nil)
			if err != nil {
				t.Fatalf("appendSerialized: %v", err)
			}
			if rec[0] != TypeLeafPrefixed {
				t.Fatalf("writer emitted type 0x%02x, want 0x%02x", rec[0], TypeLeafPrefixed)
			}
			if want := leaf.serializedSize(); want != len(rec) {
				t.Fatalf("serializedSize %d, wrote %d", want, len(rec))
			}
			got, err := ReadNode(&NodeKey{Version: 2, Nonce: 1}, rec)
			if err != nil {
				t.Fatalf("ReadNode(prefixed): %v", err)
			}
			assertLeafEqual(t, leaf, got, "prefixed")

			// Re-serializing the decoded node must be byte-identical.
			rec2, err := got.(*LeafNode).appendSerialized(nil)
			if err != nil {
				t.Fatalf("re-appendSerialized: %v", err)
			}
			if !bytes.Equal(rec, rec2) {
				t.Fatal("prefixed round-trip is not byte-stable")
			}

			// The legacy layout must still decode to the same node.
			v1 := appendLeafV1(leaf)
			gotV1, err := ReadNode(&NodeKey{Version: 2, Nonce: 1}, v1)
			if err != nil {
				t.Fatalf("ReadNode(v1): %v", err)
			}
			assertLeafEqual(t, leaf, gotV1, "v1")
		})
	}
}

// TestLeafPrefixed_IsSmaller pins the point of the format: a leaf of adjacent
// structured keys must serialize substantially smaller than the legacy layout.
func TestLeafPrefixed_IsSmaller(t *testing.T) {
	keys := make([]string, B)
	for i := range keys {
		keys[i] = fmt.Sprintf("oid:0123456789abcdef0123456789abcdef01234567:%04d", i)
	}
	leaf := makeLeaf(keys)
	prefixed, err := leaf.appendSerialized(nil)
	if err != nil {
		t.Fatal(err)
	}
	v1 := appendLeafV1(leaf)
	keyBytesV1, keyBytesPrefixed := 0, 0
	for i := range int(leaf.numKeys) {
		keyBytesV1 += 1 + len(leaf.keys[i])
	}
	p := leaf.commonPrefixLen()
	keyBytesPrefixed = 1 + p
	for i := range int(leaf.numKeys) {
		keyBytesPrefixed += 1 + len(leaf.keys[i]) - p
	}
	t.Logf("record %d -> %d bytes (%.0f%%); key block %d -> %d bytes; common prefix %d",
		len(v1), len(prefixed), 100*float64(len(prefixed))/float64(len(v1)),
		keyBytesV1, keyBytesPrefixed, p)
	if len(prefixed) >= len(v1) {
		t.Fatalf("prefixed record is not smaller: %d >= %d", len(prefixed), len(v1))
	}
	if keyBytesPrefixed*3 > keyBytesV1 {
		t.Fatalf("key block only shrank from %d to %d bytes; expected >3x", keyBytesV1, keyBytesPrefixed)
	}
}

// TestLeafPrefixed_RejectsOversizedHeader covers the decoder's allocation
// bounds. A prefix is charged once per key, so a short record can declare a
// large key block; the caps must reject the shapes that no legitimate writer
// could produce.
func TestLeafPrefixed_RejectsOversizedHeader(t *testing.T) {
	build := func(numKeys, prefixLen uint64, prefix []byte, suffixLens []uint64) []byte {
		b := []byte{TypeLeafPrefixed}
		b = binary.AppendUvarint(b, numKeys)
		b = binary.AppendUvarint(b, prefixLen)
		b = append(b, prefix...)
		for _, sl := range suffixLens {
			b = binary.AppendUvarint(b, sl)
			b = append(b, make([]byte, sl)...)
		}
		return b
	}
	// prefixLen past the per-key cap.
	rec := build(1, maxReadBytesLen+1, nil, nil)
	if _, err := ReadNode(nil, rec); err == nil || !strings.Contains(err.Error(), "prefix length") {
		t.Fatalf("oversized prefixLen: err = %v", err)
	}
	// prefixLen within the cap but longer than the payload.
	rec = build(1, 1<<16, []byte("short"), nil)
	if _, err := ReadNode(nil, rec); err == nil {
		t.Fatal("prefix longer than payload accepted")
	}
	// prefix + suffix past the per-key cap.
	rec = build(1, 8, []byte("12345678"), []uint64{maxReadBytesLen})
	if _, err := ReadNode(nil, rec); err == nil || !strings.Contains(err.Error(), "exceeds maximum") {
		t.Fatalf("oversized key: err = %v", err)
	}
	// A well-formed record with a truncated suffix block.
	rec = build(2, 2, []byte("ab"), []uint64{1})
	if _, err := ReadNode(nil, rec); err == nil {
		t.Fatal("truncated suffix block accepted")
	}
}

// TestLeafPrefixed_MixedFormatTreeOnDisk mounts a tree whose leaves were
// written in the legacy layout, then keeps using it: reads must work, and
// nodes rewritten by later saves must upgrade to the new format in place
// without disturbing the root hash.
func TestLeafPrefixed_MixedFormatTreeOnDisk(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 0, NewNopLogger())
	r := rand.New(rand.NewSource(99))
	model := map[string]string{}
	for i := range 400 {
		k := fmt.Sprintf("oid:0123456789abcdef0123456789abcdef01234567:%05d", i)
		v := fmt.Sprintf("value-%d", i)
		mustSet(t, tr, []byte(k), []byte(v))
		model[k] = v
	}
	v1ver := mustSave(t, tr)
	hashAtV1 := append([]byte(nil), tr.Hash()...)

	// Rewrite every 'B' record that holds a prefixed leaf back into the legacy
	// layout, simulating a DB written by a pre-TypeLeafPrefixed binary.
	itr, err := db.Iterator([]byte{PrefixNode}, []byte{PrefixNode + 1})
	if err != nil {
		t.Fatal(err)
	}
	type rec struct{ k, v []byte }
	var updates []rec
	for ; itr.Valid(); itr.Next() {
		payload, err := verifyChecksum(itr.Value())
		if err != nil || len(payload) == 0 || payload[0] != TypeLeafPrefixed {
			continue
		}
		node, err := ReadNode(GetNodeKey(itr.Key()[1:]), payload)
		if err != nil {
			t.Fatalf("decoding node: %v", err)
		}
		updates = append(updates, rec{
			append([]byte(nil), itr.Key()...),
			stampChecksum(appendLeafV1(node.(*LeafNode))),
		})
	}
	itr.Close()
	for _, u := range updates {
		if err := db.Set(u.k, u.v); err != nil {
			t.Fatal(err)
		}
	}
	if len(updates) == 0 {
		t.Fatal("no prefixed leaves found to downgrade")
	}
	t.Logf("downgraded %d leaf records to the legacy layout", len(updates))

	// Mount the legacy DB: same root hash, same contents.
	tr2 := NewMutableTreeWithDB(db, 0, NewNopLogger())
	defer tr2.Close()
	if _, err := tr2.LoadVersion(v1ver); err != nil {
		t.Fatalf("LoadVersion: %v", err)
	}
	if !bytes.Equal(tr2.Hash(), hashAtV1) {
		t.Fatalf("legacy-layout tree hash %x != %x", tr2.Hash(), hashAtV1)
	}
	for k, want := range model {
		got, err := tr2.Get([]byte(k))
		if err != nil || string(got) != want {
			t.Fatalf("legacy Get(%s) = %q, %v; want %q", k, got, err, want)
		}
	}

	// Keep writing: touched leaves re-serialize as prefixed, untouched ones
	// stay legacy, and the whole mixed-format tree must still reload clean.
	for i := range 50 {
		k := fmt.Sprintf("oid:0123456789abcdef0123456789abcdef01234567:%05d", r.Intn(400))
		v := fmt.Sprintf("updated-%d", i)
		mustSet(t, tr2, []byte(k), []byte(v))
		model[k] = v
	}
	v2 := mustSave(t, tr2)
	hashAtV2 := append([]byte(nil), tr2.Hash()...)

	tr3 := NewMutableTreeWithDB(db, 0, NewNopLogger())
	defer tr3.Close()
	if _, err := tr3.LoadVersion(v2); err != nil {
		t.Fatalf("LoadVersion(v2): %v", err)
	}
	if !bytes.Equal(tr3.Hash(), hashAtV2) {
		t.Fatalf("mixed-format tree hash %x != %x", tr3.Hash(), hashAtV2)
	}
	for k, want := range model {
		got, err := tr3.Get([]byte(k))
		if err != nil || string(got) != want {
			t.Fatalf("mixed Get(%s) = %q, %v; want %q", k, got, err, want)
		}
	}
	if got := tr3.Size(); got != int64(len(model)) {
		t.Fatalf("size %d, want %d", got, len(model))
	}
}

// TestCommonPrefixLen exercises the word-wise scan across word boundaries.
func TestCommonPrefixLen(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"", "a", 0},
		{"a", "a", 1},
		{"abc", "abd", 2},
		{"01234567", "01234567", 8},
		{"01234567x", "01234567y", 8},
		{"0123456789abcdef", "0123456789abcdef", 16},
		{"0123456789abcdefA", "0123456789abcdefB", 16},
		{"0123456789abcde", "0123456789abcdf", 14},
		{"0123456789abcdefghij", "0123456789abcdefghiX", 19},
	}
	for _, c := range cases {
		if got := commonPrefixLen([]byte(c.a), []byte(c.b)); got != c.want {
			t.Errorf("commonPrefixLen(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
	// Differential against a naive implementation.
	r := rand.New(rand.NewSource(5))
	for range 20_000 {
		n := r.Intn(40)
		a := make([]byte, n)
		r.Read(a)
		b := append([]byte(nil), a[:r.Intn(n+1)]...)
		if r.Intn(2) == 0 && len(b) < n {
			b = append(b, a[len(b)]^0xFF)
		}
		naive := 0
		for naive < len(a) && naive < len(b) && a[naive] == b[naive] {
			naive++
		}
		if got := commonPrefixLen(a, b); got != naive {
			t.Fatalf("commonPrefixLen(%x, %x) = %d, want %d", a, b, got, naive)
		}
	}
}
