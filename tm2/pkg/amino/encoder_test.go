package amino

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUvarintSize(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		u    uint64
		want int
	}{
		{"0 bit", 0, 1},
		{"1 bit", 1 << 0, 1},
		{"6 bits", 1 << 5, 1},
		{"7 bits", 1 << 6, 1},
		{"8 bits", 1 << 7, 2},
		{"62 bits", 1 << 61, 9},
		{"63 bits", 1 << 62, 9},
		{"64 bits", 1 << 63, 10},
	}
	for i, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, UvarintSize(tc.u), "#%d", i) //nolint:scopelint
		})
	}
}

// helper: encode with Append*Reversed + reverseBytes, compare to Encode* forward path.
func reversedEncode(fn func(buf []byte) []byte) []byte {
	buf := fn(make([]byte, 0, 64))
	reverseBytes(buf)
	return buf
}

func forwardEncode(fn func(w *bytes.Buffer)) []byte {
	var buf bytes.Buffer
	fn(&buf)
	return buf.Bytes()
}

func TestAppendReversed_Uvarint(t *testing.T) {
	t.Parallel()
	for _, u := range []uint64{0, 1, 127, 128, 255, 256, 16383, 16384, 1<<63 - 1, 1 << 63} {
		got := reversedEncode(func(buf []byte) []byte { return AppendUvarintReversed(buf, u) })
		want := forwardEncode(func(w *bytes.Buffer) { EncodeUvarint(w, u) })
		require.Equal(t, want, got, "uvarint %d", u)
	}
}

func TestAppendReversed_Varint(t *testing.T) {
	t.Parallel()
	for _, i := range []int64{0, 1, -1, 63, -64, 127, -128, 1<<31 - 1, -(1 << 31), 1<<63 - 1, -(1 << 63)} {
		got := reversedEncode(func(buf []byte) []byte { return AppendVarintReversed(buf, i) })
		want := forwardEncode(func(w *bytes.Buffer) { EncodeVarint(w, i) })
		require.Equal(t, want, got, "varint %d", i)
	}
}

func TestAppendReversed_String(t *testing.T) {
	t.Parallel()
	for _, s := range []string{"", "a", "hello", "hello world 1234567890"} {
		got := reversedEncode(func(buf []byte) []byte { return AppendStringReversed(buf, s) })
		want := forwardEncode(func(w *bytes.Buffer) { EncodeString(w, s) })
		require.Equal(t, want, got, "string %q", s)
	}
}

func TestAppendReversed_ByteSlice(t *testing.T) {
	t.Parallel()
	for _, bz := range [][]byte{nil, {}, {0x00}, {0x01, 0x02, 0x03}, make([]byte, 300)} {
		got := reversedEncode(func(buf []byte) []byte { return AppendByteSliceReversed(buf, bz) })
		want := forwardEncode(func(w *bytes.Buffer) { EncodeByteSlice(w, bz) })
		require.Equal(t, want, got, "byteslice len=%d", len(bz))
	}
}

func TestAppendReversed_Bool(t *testing.T) {
	t.Parallel()
	for _, b := range []bool{false, true} {
		got := reversedEncode(func(buf []byte) []byte { return AppendBoolReversed(buf, b) })
		want := forwardEncode(func(w *bytes.Buffer) { EncodeBool(w, b) })
		require.Equal(t, want, got, "bool %v", b)
	}
}

func TestAppendReversed_Fixed(t *testing.T) {
	t.Parallel()
	// Int32
	for _, v := range []int32{0, 1, -1, 1<<31 - 1, -(1 << 31)} {
		got := reversedEncode(func(buf []byte) []byte { return AppendInt32Reversed(buf, v) })
		want := forwardEncode(func(w *bytes.Buffer) { EncodeInt32(w, v) })
		require.Equal(t, want, got, "int32 %d", v)
	}
	// Int64
	for _, v := range []int64{0, 1, -1, 1<<63 - 1, -(1 << 63)} {
		got := reversedEncode(func(buf []byte) []byte { return AppendInt64Reversed(buf, v) })
		want := forwardEncode(func(w *bytes.Buffer) { EncodeInt64(w, v) })
		require.Equal(t, want, got, "int64 %d", v)
	}
	// Uint32
	for _, v := range []uint32{0, 1, 1<<32 - 1} {
		got := reversedEncode(func(buf []byte) []byte { return AppendUint32Reversed(buf, v) })
		want := forwardEncode(func(w *bytes.Buffer) { EncodeUint32(w, v) })
		require.Equal(t, want, got, "uint32 %d", v)
	}
	// Uint64
	for _, v := range []uint64{0, 1, 1<<64 - 1} {
		got := reversedEncode(func(buf []byte) []byte { return AppendUint64Reversed(buf, v) })
		want := forwardEncode(func(w *bytes.Buffer) { EncodeUint64(w, v) })
		require.Equal(t, want, got, "uint64 %d", v)
	}
}

func TestAppendReversed_Float(t *testing.T) {
	t.Parallel()
	for _, f := range []float32{0, 1.0, -1.0, 3.14} {
		got := reversedEncode(func(buf []byte) []byte { return AppendFloat32Reversed(buf, f) })
		want := forwardEncode(func(w *bytes.Buffer) { EncodeFloat32(w, f) })
		require.Equal(t, want, got, "float32 %v", f)
	}
	for _, f := range []float64{0, 1.0, -1.0, 3.14159265358979} {
		got := reversedEncode(func(buf []byte) []byte { return AppendFloat64Reversed(buf, f) })
		want := forwardEncode(func(w *bytes.Buffer) { EncodeFloat64(w, f) })
		require.Equal(t, want, got, "float64 %v", f)
	}
}

func TestAppendReversed_FieldNumberAndTyp3(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		num uint32
		typ Typ3
	}{
		{1, Typ3Varint}, {1, Typ3ByteLength}, {2, Typ3Varint},
		{15, Typ3ByteLength}, {16, Typ3Varint}, {100, Typ3ByteLength},
	} {
		got := reversedEncode(func(buf []byte) []byte { return AppendFieldNumberAndTyp3Reversed(buf, tc.num, tc.typ) })
		want := forwardEncode(func(w *bytes.Buffer) { encodeFieldNumberAndTyp3(w, tc.num, tc.typ) })
		require.Equal(t, want, got, "field %d typ %d", tc.num, tc.typ)
	}
}

func TestAppendReversed_Time(t *testing.T) {
	t.Parallel()
	times := []time.Time{
		time.Unix(0, 0).UTC(),
		time.Unix(1, 0).UTC(),
		time.Unix(0, 1).UTC(),
		time.Unix(1622505600, 123456789).UTC(),
	}
	for _, tm := range times {
		got := reversedEncode(func(buf []byte) []byte { buf, _ = AppendTimeReversed(buf, tm); return buf })
		// Use Prepend path as reference.
		size := TimeSize(tm)
		ref := make([]byte, size)
		PrependTime(ref, size, tm)
		require.Equal(t, ref, got, "time %v", tm)
	}
}

func TestAppendReversed_Duration(t *testing.T) {
	t.Parallel()
	durations := []time.Duration{0, time.Second, time.Millisecond, 5*time.Hour + 3*time.Nanosecond}
	for _, d := range durations {
		got := reversedEncode(func(buf []byte) []byte { buf, _ = AppendDurationReversed(buf, d); return buf })
		size := DurationSize(d)
		ref := make([]byte, size)
		PrependDuration(ref, size, d)
		require.Equal(t, ref, got, "duration %v", d)
	}
}

func TestAppendReversed_Composite(t *testing.T) {
	t.Parallel()
	// Simulate a simple struct: field 2 (string "hi"), field 1 (varint 42).
	// Written in reverse field order, reversed bytes, then final reverse.
	got := reversedEncode(func(buf []byte) []byte {
		// Field 2: string "hi" (ByteLength)
		buf = AppendStringReversed(buf, "hi")
		buf = AppendFieldNumberAndTyp3Reversed(buf, 2, Typ3ByteLength)
		// Field 1: varint 42
		buf = AppendUvarintReversed(buf, 42)
		buf = AppendFieldNumberAndTyp3Reversed(buf, 1, Typ3Varint)
		return buf
	})

	// Build expected via forward encoding.
	want := forwardEncode(func(w *bytes.Buffer) {
		// Field 1: varint 42
		encodeFieldNumberAndTyp3(w, 1, Typ3Varint)
		EncodeUvarint(w, 42)
		// Field 2: string "hi"
		encodeFieldNumberAndTyp3(w, 2, Typ3ByteLength)
		EncodeString(w, "hi")
	})

	require.Equal(t, want, got)
}
