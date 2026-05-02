package tests

// binary_fixes_test.go: regression tests pinning behaviors that
// previously diverged between the amino reflection path and the
// genproto2 codegen path. Each test documents the bug it covers
// and the audit that surfaced it.

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- Int / Uint + BinFixed32 round-trip ---------------------------

// Pre-fix state: the reflection encoder emitted 4 fixed bytes for
// `Int + binary:"fixed32"` (binary_encode.go:143) but the decoder had
// no BinFixed32 arm and silently fell through to DecodeVarint
// (binary_decode.go reflect.Int case). Round-trip via reflection
// corrupted. Same bug existed for Uint. Also typeToTyp3 didn't return
// Typ34Byte for Int/Uint+BinFixed32, so even the field-key wire-type
// was wrong. ValidateBasic gate hid it.

type fixed32IntStruct struct {
	A int   `binary:"fixed32"`
	B int32 `binary:"fixed32"`
}

func TestBinFixed32_Int_RoundTrip(t *testing.T) {
	t.Parallel()

	cdc := registerLocal(t, fixed32IntStruct{})
	cases := []fixed32IntStruct{
		{A: 0, B: 0},
		{A: 1, B: 1},
		{A: -1, B: -1},
		{A: math.MaxInt32, B: math.MaxInt32},
		{A: math.MinInt32, B: math.MinInt32},
	}
	for _, c := range cases {
		c := c
		t.Run("", func(t *testing.T) {
			t.Parallel()
			bz, err := cdc.MarshalSized(&c)
			require.NoError(t, err)

			var got fixed32IntStruct
			require.NoError(t, cdc.UnmarshalSized(bz, &got))
			assert.Equal(t, c, got)
		})
	}
}

type fixed32UintStruct struct {
	A uint   `binary:"fixed32"`
	B uint32 `binary:"fixed32"`
}

func TestBinFixed32_Uint_RoundTrip(t *testing.T) {
	t.Parallel()

	cdc := registerLocal(t, fixed32UintStruct{})
	cases := []fixed32UintStruct{
		{A: 0, B: 0},
		{A: 1, B: 1},
		{A: math.MaxUint32, B: math.MaxUint32},
	}
	for _, c := range cases {
		c := c
		t.Run("", func(t *testing.T) {
			t.Parallel()
			bz, err := cdc.MarshalSized(&c)
			require.NoError(t, err)

			var got fixed32UintStruct
			require.NoError(t, cdc.UnmarshalSized(bz, &got))
			assert.Equal(t, c, got)
		})
	}
}

// TestBinFixed32_Int_WireWidth pins that an Int field tagged
// binary:"fixed32" emits exactly 4 body bytes (plus the 1-byte field
// key prefix). Without the typeToTyp3 fix, the field key would carry
// wire-type 0 (varint) while the body was 4 fixed bytes — a
// peer-side decode error or silent corruption.
func TestBinFixed32_Int_WireWidth(t *testing.T) {
	t.Parallel()

	type single struct {
		X int `binary:"fixed32"`
	}
	cdc := registerLocal(t, single{})

	bz, err := cdc.MarshalSized(&single{X: 0x01020304})
	require.NoError(t, err)
	// MarshalSized prepends a varint length. Body = 1-byte field key
	// (field 1, wire-type 5 = (1<<3)|5 = 0x0d) + 4 little-endian bytes.
	require.Equal(t, []byte{0x05, 0x0d, 0x04, 0x03, 0x02, 0x01}, bz,
		"Int + binary:\"fixed32\" must emit a fixed32 wire-type field key followed by 4 LE bytes")
}

// ---- Pointer to zero value: proto3 default-skip --------------------

// Pre-fix state: reflection encoder hard-coded `writeEmpty := true` for
// any non-nil pointer field (binary_encode.go:541), so a `*string("")`
// emitted `<key> 0x00` (3 bytes total in MarshalSized) where codegen
// (gen_marshal.go:215-233) followed proto3 default-skip and produced
// no field at all. Wire-byte divergence on a path with no
// ValidateBasic gate. Codegen-emitted wire bytes are what tm2's chain
// types produce, so reflection was the side that needed to align.

type ptrZeroStruct struct {
	X *string
}

func TestPointerToZero_Reflection_OmitsField(t *testing.T) {
	t.Parallel()

	cdc := registerLocal(t, ptrZeroStruct{})
	empty := ""
	bz, err := cdc.MarshalSized(&ptrZeroStruct{X: &empty})
	require.NoError(t, err)
	// proto3 default-skip: an empty string field is omitted regardless
	// of the surrounding pointer being non-nil. Pre-fix reflection
	// emitted [0x02 0x0a 0x00] (length=2, field 1 length-delim, body 0).
	assert.Equal(t, []byte{0x00}, bz,
		"non-nil pointer to zero-value scalar must be omitted (proto3 default-skip)")
}

func TestPointerToZero_Reflection_NilOmitsToo(t *testing.T) {
	t.Parallel()

	cdc := registerLocal(t, ptrZeroStruct{})
	bz, err := cdc.MarshalSized(&ptrZeroStruct{X: nil})
	require.NoError(t, err)
	assert.Equal(t, []byte{0x00}, bz, "nil pointer field must be omitted")
}

func TestPointerToZero_Reflection_NonZeroEmits(t *testing.T) {
	t.Parallel()

	cdc := registerLocal(t, ptrZeroStruct{})
	s := "hello"
	bz, err := cdc.MarshalSized(&ptrZeroStruct{X: &s})
	require.NoError(t, err)
	// length(7) + field-key(0x0a) + length(5) + "hello"
	want := []byte{0x07, 0x0a, 0x05, 'h', 'e', 'l', 'l', 'o'}
	assert.Equal(t, want, bz)
}

type ptrZeroWriteEmptyStruct struct {
	X *string `amino:"write_empty"`
}

// TestPointerToZero_Reflection_WriteEmptyForcesEmit pins the escape
// hatch: when an operator explicitly opts into write_empty, the
// non-nil-zero pointer IS emitted, recovering the pre-fix behavior
// for callers that depended on it.
func TestPointerToZero_Reflection_WriteEmptyForcesEmit(t *testing.T) {
	t.Parallel()

	cdc := registerLocal(t, ptrZeroWriteEmptyStruct{})
	empty := ""
	bz, err := cdc.MarshalSized(&ptrZeroWriteEmptyStruct{X: &empty})
	require.NoError(t, err)
	// Same shape as a non-nil-zero ptr would have produced pre-fix:
	// length(2) + field-key(0x0a) + length(0).
	assert.Equal(t, []byte{0x02, 0x0a, 0x00}, bz,
		"explicit amino:\"write_empty\" must force-emit the field even when the pointer dereferences to zero")
}

// ---- []*Primitive without nil_elements --------------------------

// An audit hypothesized a divergence where codegen's unpacked-list
// "else" branch (line 622-633 of gen_unmarshal.go) treats 0x00 as
// length-prefix-of-empty-message while reflection treats 0x00 as
// defaultValue. Probing the actual codegen output via
// `genproto2.TestSlicePtrInt_CodegenUsesPackedEncoding` shows the
// branch is never reached for `[]*int`: codegen routes it through
// packed encoding because int's typ3 is varint, not byte-length.
// These tests exercise the reflection path's round-trip for
// completeness, including the non-nil-zero element case.

type slicePtrIntStruct struct {
	Xs []*int
}

func TestSlicePtrNonStruct_Reflection_RoundTrip(t *testing.T) {
	t.Parallel()

	cdc := registerLocal(t, slicePtrIntStruct{})
	five, ten := 5, 10
	val := &slicePtrIntStruct{Xs: []*int{&five, &ten}}

	bz, err := cdc.MarshalSized(val)
	require.NoError(t, err)

	var got slicePtrIntStruct
	require.NoError(t, cdc.UnmarshalSized(bz, &got))

	require.Len(t, got.Xs, 2)
	assert.Equal(t, 5, *got.Xs[0])
	assert.Equal(t, 10, *got.Xs[1])
}

// TestSlicePtrNonStruct_Reflection_ZeroElement is the load-bearing case
// for Subagent C's finding: the element pointer is non-nil but points
// to the zero value. Reflect's decode special-cases 0x00 → defaultValue
// (a non-nil zero pointer), so the round-trip preserves "non-nil zero".
func TestSlicePtrNonStruct_Reflection_ZeroElement(t *testing.T) {
	t.Parallel()

	cdc := registerLocal(t, slicePtrIntStruct{})
	zero := 0
	val := &slicePtrIntStruct{Xs: []*int{&zero}}

	bz, err := cdc.MarshalSized(val)
	require.NoError(t, err)

	var got slicePtrIntStruct
	require.NoError(t, cdc.UnmarshalSized(bz, &got))

	require.Len(t, got.Xs, 1)
	require.NotNil(t, got.Xs[0], "non-nil zero element must round-trip to a non-nil pointer")
	assert.Equal(t, 0, *got.Xs[0])
}

// TestPointerToZero_Reflection_NilAndEmptyAreIndistinguishable pins the
// proto3 default-skip semantic: both `nil` and `&""` produce identical
// wire bytes, and the decoder fills the absent field with a non-nil
// default pointer (`&""`). Callers that need to distinguish "set but
// empty" from "unset" must use `amino:"write_empty"`.
func TestPointerToZero_Reflection_NilAndEmptyAreIndistinguishable(t *testing.T) {
	t.Parallel()

	cdc := registerLocal(t, ptrZeroStruct{})

	empty := ""
	bzEmpty, err := cdc.MarshalSized(&ptrZeroStruct{X: &empty})
	require.NoError(t, err)
	bzNil, err := cdc.MarshalSized(&ptrZeroStruct{X: nil})
	require.NoError(t, err)
	assert.Equal(t, bzNil, bzEmpty, "nil and non-nil-zero pointers must produce identical wire bytes")

	var gotEmpty, gotNil ptrZeroStruct
	require.NoError(t, cdc.UnmarshalSized(bzEmpty, &gotEmpty))
	require.NoError(t, cdc.UnmarshalSized(bzNil, &gotNil))
	// Both decode to the same shape (non-nil pointer to zero string —
	// the decoder's defaultValue init for absent pointer fields).
	assert.Equal(t, gotEmpty, gotNil)
}
