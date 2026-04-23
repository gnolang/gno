package amino

import (
	"strings"
	"testing"
)

// buildWire composes a wire-format byte slice by appending forward.
// Tests use this instead of the Prepend* backward writers for legibility.
type wireBuilder struct {
	bz []byte
}

func (w *wireBuilder) fieldKey(fnum uint32, typ Typ3) *wireBuilder {
	var tmp [10]byte
	v := uint64(fnum)<<3 | uint64(typ)
	n := putUvarint(tmp[:], v)
	w.bz = append(w.bz, tmp[:n]...)
	return w
}

func (w *wireBuilder) uvarint(u uint64) *wireBuilder {
	var tmp [10]byte
	n := putUvarint(tmp[:], u)
	w.bz = append(w.bz, tmp[:n]...)
	return w
}

func (w *wireBuilder) raw(b ...byte) *wireBuilder {
	w.bz = append(w.bz, b...)
	return w
}

func (w *wireBuilder) lengthPrefixed(payload []byte) *wireBuilder {
	w.uvarint(uint64(len(payload)))
	w.bz = append(w.bz, payload...)
	return w
}

// putUvarint — inline so this file has no dependency on internal helpers.
func putUvarint(bz []byte, u uint64) int {
	n := 0
	for u >= 0x80 {
		bz[n] = byte(u) | 0x80
		u >>= 7
		n++
	}
	bz[n] = byte(u)
	return n + 1
}

func TestDumpWire_EmptyInput(t *testing.T) {
	if s := DumpWire(nil); s != "" {
		t.Fatalf("expected empty string for nil input, got %q", s)
	}
	if s := DumpWire([]byte{}); s != "" {
		t.Fatalf("expected empty string for zero-length input, got %q", s)
	}
}

func TestDumpWire_Varint(t *testing.T) {
	// field 1, Typ3Varint, value = 42.
	b := (&wireBuilder{}).fieldKey(1, Typ3Varint).uvarint(42).bz
	got := DumpWire(b)
	wantSubs := []string{"field 1", "Typ3Varint", "= 42"}
	for _, sub := range wantSubs {
		if !strings.Contains(got, sub) {
			t.Errorf("missing %q in output:\n%s", sub, got)
		}
	}
}

func TestDumpWire_Fixed32(t *testing.T) {
	// field 2, Typ34Byte, 4 bytes.
	b := (&wireBuilder{}).fieldKey(2, Typ34Byte).raw(0xde, 0xad, 0xbe, 0xef).bz
	got := DumpWire(b)
	for _, sub := range []string{"field 2", "Typ34Byte", "deadbeef"} {
		if !strings.Contains(got, sub) {
			t.Errorf("missing %q in output:\n%s", sub, got)
		}
	}
}

func TestDumpWire_Fixed64(t *testing.T) {
	b := (&wireBuilder{}).fieldKey(3, Typ38Byte).raw(0, 1, 2, 3, 4, 5, 6, 7).bz
	got := DumpWire(b)
	for _, sub := range []string{"field 3", "Typ38Byte", "0001020304050607"} {
		if !strings.Contains(got, sub) {
			t.Errorf("missing %q in output:\n%s", sub, got)
		}
	}
}

func TestDumpWire_EmptyByteLength_NilSentinel(t *testing.T) {
	// A ByteLength field with len=0 is the positional-nil sentinel used
	// by unpacked lists with amino:"nil_elements". DumpWire must flag it
	// so reviewers diffing two traces can SEE the nil entry.
	b := (&wireBuilder{}).fieldKey(1, Typ3ByteLength).uvarint(0).bz
	got := DumpWire(b)
	for _, sub := range []string{"len=0", "nil sentinel"} {
		if !strings.Contains(got, sub) {
			t.Errorf("missing %q in output:\n%s", sub, got)
		}
	}
}

func TestDumpWire_NestedMessage_Recurses(t *testing.T) {
	// Inner message: field 1 Varint = 7.
	inner := (&wireBuilder{}).fieldKey(1, Typ3Varint).uvarint(7).bz
	// Outer: field 2 ByteLength wrapping inner.
	outer := (&wireBuilder{}).fieldKey(2, Typ3ByteLength).lengthPrefixed(inner).bz

	got := DumpWire(outer)

	// Outer line.
	if !strings.Contains(got, "field 2 Typ3ByteLength") {
		t.Errorf("missing outer field line:\n%s", got)
	}
	// Inner line should be indented and show the recursive decode.
	if !strings.Contains(got, "  @") {
		t.Errorf("expected indented nested line in output:\n%s", got)
	}
	if !strings.Contains(got, "field 1 Typ3Varint = 7") {
		t.Errorf("missing inner field line:\n%s", got)
	}
}

func TestDumpWire_RawByteLength_NotAMessage(t *testing.T) {
	// A ByteLength payload whose bytes can't parse as a nested message.
	// 0xff has the uvarint continuation bit set but no follow-on byte,
	// so DecodeFieldNumberAndTyp3 fails and couldBeMessage falls through
	// to the raw-bytes rendering path.
	//
	// Note the schema-less limitation: if the payload happens to be
	// valid TLV (e.g. a short string whose bytes coincide with a field
	// key + value), DumpWire will render it as a nested message. That
	// is a known honest limitation — the dumper has no schema to
	// distinguish a 2-byte string from a 2-byte message.
	outer := (&wireBuilder{}).fieldKey(3, Typ3ByteLength).lengthPrefixed([]byte{0xff, 0xff}).bz
	got := DumpWire(outer)

	if !strings.Contains(got, "bytes=") {
		t.Errorf("expected raw bytes= rendering, got:\n%s", got)
	}
	if !strings.Contains(got, "ffff") {
		t.Errorf("expected hex ffff in output:\n%s", got)
	}
}

func TestDumpWire_TruncatedFieldKey(t *testing.T) {
	// A uvarint byte with the continuation bit set but no follow-on.
	// Must degrade gracefully.
	got := DumpWire([]byte{0x80})
	if !strings.Contains(got, "bad field key") {
		t.Errorf("expected bad-field-key diagnostic, got:\n%s", got)
	}
}

func TestDumpWire_TruncatedFixed32Value(t *testing.T) {
	// Declares Typ34Byte but provides only 2 bytes.
	b := (&wireBuilder{}).fieldKey(4, Typ34Byte).raw(0x01, 0x02).bz
	got := DumpWire(b)
	if !strings.Contains(got, "truncated") {
		t.Errorf("expected truncated diagnostic, got:\n%s", got)
	}
}

func TestDumpWire_MultipleFields(t *testing.T) {
	// Two fields at the same level.
	b := (&wireBuilder{}).
		fieldKey(1, Typ3Varint).uvarint(5).
		fieldKey(2, Typ3Varint).uvarint(9).
		bz
	got := DumpWire(b)

	// Both fields present at the same (unindented) level.
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d:\n%s", len(lines), got)
	}
	for _, ln := range lines {
		if strings.HasPrefix(ln, " ") {
			t.Errorf("top-level line unexpectedly indented: %q", ln)
		}
	}
	if !strings.Contains(lines[0], "field 1") || !strings.Contains(lines[0], "= 5") {
		t.Errorf("line 0 wrong: %q", lines[0])
	}
	if !strings.Contains(lines[1], "field 2") || !strings.Contains(lines[1], "= 9") {
		t.Errorf("line 1 wrong: %q", lines[1])
	}
}

// TestDumpWire_PrecommitsShape reproduces the shape of bft/types.Commit
// that PR #5569 fixed: field 1 (Precommits) ByteLength entries, one of
// which is an empty (nil-sentinel) ByteLength. DumpWire should make the
// nil entry glaringly visible in the trace.
func TestDumpWire_PrecommitsShape(t *testing.T) {
	// Entry 1 payload: field 1 (Height) Varint = 100.
	entry1 := (&wireBuilder{}).fieldKey(1, Typ3Varint).uvarint(100).bz
	// Entry 2 payload: empty (nil sentinel).
	// Entry 3 payload: field 1 (Height) Varint = 102.
	entry3 := (&wireBuilder{}).fieldKey(1, Typ3Varint).uvarint(102).bz

	b := (&wireBuilder{}).
		fieldKey(1, Typ3ByteLength).lengthPrefixed(entry1).
		fieldKey(1, Typ3ByteLength).lengthPrefixed(nil). // nil sentinel
		fieldKey(1, Typ3ByteLength).lengthPrefixed(entry3).
		bz

	got := DumpWire(b)

	// Three top-level entries.
	if c := strings.Count(got, "field 1 Typ3ByteLength"); c != 3 {
		t.Errorf("expected 3 top-level Precommits entries, got %d:\n%s", c, got)
	}
	// Middle one is flagged as nil sentinel.
	if !strings.Contains(got, "nil sentinel") {
		t.Errorf("expected nil sentinel annotation for empty entry:\n%s", got)
	}
	// Heights 100 and 102 should appear in nested output.
	if !strings.Contains(got, "= 100") || !strings.Contains(got, "= 102") {
		t.Errorf("expected nested height values 100 and 102:\n%s", got)
	}
}
