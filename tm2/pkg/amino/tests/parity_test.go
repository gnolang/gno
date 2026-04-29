package tests_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/aminotest"
	"github.com/gnolang/gno/tm2/pkg/amino/tests"
)

// TestCodecParity_AminoFixtures asserts that every hand-crafted value below
// round-trips byte-identically through both the reflect codec and the
// genproto2 fast path, and that SizeBinary2 matches the encoded length.
//
// To add coverage for a new edge case, append an entry to parityCasesAmino
// below — no new test function needed.
func TestCodecParity_AminoFixtures(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(tests.Package)
	cdc.Seal()

	for i, c := range parityCasesAmino {
		c := c
		name := fmt.Sprintf("%d/%s", i, c.name)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			aminotest.AssertCodecParity(t, cdc, c.v)
		})
	}
}

var parityCasesAmino = []struct {
	name string
	v    any
}{
	// bare int/uint + binary:"fixed64" edge values (surface of the recent
	// typeToTyp3 + binary_decode.go fix).
	{"FuzzFixedInt/min-max", &tests.FuzzFixedInt{I64: math.MinInt64, U64: math.MaxUint64}},
	{"FuzzFixedInt/neg-one", &tests.FuzzFixedInt{I64: -1, U64: 1}},
	{"FuzzFixedInt/zero", &tests.FuzzFixedInt{}}, // zero-valued field-omission path

	// nil_elements on []*Struct — the consensus-wedging surface. Mixed nil
	// and non-nil entries, non-zero field content so the values round-trip
	// cleanly without collapsing to nil.
	{"FuzzNilElements/mixed", &tests.FuzzNilElements{
		Entries: []*tests.FuzzFieldInfo{
			{Name: "first", Embedded: true, Tag: "x:1", Index: 1},
			nil,
			{Name: "third", Index: 3},
		},
		Poses: []*tests.GnoVMPos{
			{Line: 7, Column: 9},
			nil,
		},
		Name: "nil-elements",
	}},

	// Fixed-width integer combinations via PrimitivesStruct — covers
	// int32/int64/uint32/uint64 with and without fixed32/fixed64 at
	// their extreme values.
	{"PrimitivesStruct/extreme", &tests.PrimitivesStruct{
		Int8:        math.MinInt8,
		Int16:       math.MinInt16,
		Int32:       math.MinInt32,
		Int32Fixed:  math.MinInt32,
		Int64:       math.MinInt64,
		Int64Fixed:  math.MinInt64,
		Int:         -1,
		Byte:        255,
		Uint8:       255,
		Uint16:      math.MaxUint16,
		Uint32:      math.MaxUint32,
		Uint32Fixed: math.MaxUint32,
		Uint64:      math.MaxUint64,
		Uint64Fixed: math.MaxUint64,
		Uint:        math.MaxUint64,
		Str:         "edge case",
		Bytes:       []byte{0x00, 0xff, 0x55, 0xaa},
	}},
	{"PrimitivesStruct/all-zero", &tests.PrimitivesStruct{}},

	// Fixed-width tagged slices.
	{"SlicesStruct/fixed-mixed", &tests.SlicesStruct{
		Int32FixedSl:  []int32{0, -1, math.MaxInt32, math.MinInt32},
		Int64FixedSl:  []int64{0, -1, math.MaxInt64, math.MinInt64},
		Uint32FixedSl: []uint32{0, 1, math.MaxUint32},
		Uint64FixedSl: []uint64{0, 1, math.MaxUint64},
		StrSl:         []string{"a", "", "bc"},
		BytesSl:       [][]byte{{0x01}, nil, {0x02, 0x03}},
	}},

	// AminoMarshaler with struct repr (ReprStruct1). Exercises the
	// AminoMarshaler-field marshal path at gen_marshal.go.
	{"AminoMarshalerStruct1", &tests.AminoMarshalerStruct1{A: 42, B: -1}},

	// AminoMarshaler with string repr (matches the shape of crypto.Address).
	{"AminoMarshalerInt5/nonzero", func() *tests.AminoMarshalerInt5 {
		v := tests.AminoMarshalerInt5(7)
		return &v
	}()},
	{"AminoMarshalerInt5/zero", func() *tests.AminoMarshalerInt5 {
		v := tests.AminoMarshalerInt5(0)
		return &v
	}()},

	// AminoMarshaler whose MarshalAmino returns "" for the zero value.
	// Exercises the zero-repr skip branch of the gen_marshal.go emission
	// guard (production types rarely hit this — bech32 / Sprintf / etc.
	// always produce non-empty strings even at zero).
	{"EmptyReprOnZero/zero", &tests.EmptyReprOnZero{Val: 0}},
	{"EmptyReprOnZero/nonzero", &tests.EmptyReprOnZero{Val: 42}},

	// FuzzNilEmptyRepr: same AminoMarshaler under amino:"nil_elements".
	// Only values that round-trip losslessly (no non-nil-zero-valued
	// entries, since those would normalize to nil on decode under the
	// lossy nil_elements + empty-repr intersection, which is tested
	// separately by TestParity_FuzzNilEmptyRepr_LossyListEncode below).
	{"FuzzNilEmptyRepr/all-nil", &tests.FuzzNilEmptyRepr{
		Vals: []*tests.EmptyReprOnZero{nil, nil},
	}},
	{"FuzzNilEmptyRepr/mixed-nonzero", &tests.FuzzNilEmptyRepr{
		Vals: []*tests.EmptyReprOnZero{{Val: 1}, nil, {Val: 7}},
	}},
}

// TestParity_FuzzNilEmptyRepr_LossyListEncode covers the one parity
// invariant that AssertCodecParity can't express: a pointer slice with
// amino:"nil_elements" whose elements include an AminoMarshaler whose
// MarshalAmino returns "" for some input. Both nil AND &{Val:0} encode
// to a zero-length element on the wire (indistinguishable), and both
// decode to nil — which violates strict DeepEqual roundtrip even though
// the codec is behaving correctly.
//
// This test asserts the weaker but still meaningful invariants:
//  1. MarshalReflect and MarshalBinary2 produce byte-identical output.
//  2. SizeBinary2 matches len(MarshalBinary2).
//  3. Both decode paths produce the same value (nil-normalized).
//  4. Re-encoding the decoded value reproduces the original bytes
//     (byte-stability).
func TestParity_FuzzNilEmptyRepr_LossyListEncode(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(tests.Package)
	cdc.Seal()

	// Input mixes nil, &{Val:0} (empty repr), and non-zero. After roundtrip
	// we expect the first two to normalize to nil — the "lossy but
	// byte-stable" property of nil_elements.
	orig := &tests.FuzzNilEmptyRepr{
		Vals: []*tests.EmptyReprOnZero{
			nil,
			{Val: 0},
			{Val: 5},
			nil,
			{Val: 99},
		},
	}

	msg, ok := any(orig).(amino.PBMessager2)
	require.True(t, ok)

	// (1) Encoder parity.
	bzReflect, err := cdc.MarshalReflect(orig)
	require.NoError(t, err)
	bzBinary2, err := cdc.MarshalBinary2(msg)
	require.NoError(t, err)
	require.Equal(t, bzReflect, bzBinary2, "encoder parity (reflect vs genproto2)")

	// (2) Size invariant.
	size, err := msg.SizeBinary2(cdc)
	require.NoError(t, err)
	require.Equal(t, len(bzBinary2), size, "size invariant")

	// (3) Cross-decoder agreement.
	var viaReflect tests.FuzzNilEmptyRepr
	require.NoError(t, cdc.UnmarshalReflect(bzReflect, &viaReflect))
	var viaBinary2 tests.FuzzNilEmptyRepr
	require.NoError(t, viaBinary2.UnmarshalBinary2(cdc, bzBinary2, 0))
	require.Equal(t, viaReflect, viaBinary2, "cross-decoder parity")

	// The decoded value collapses &{Val:0} and nil entries into nil — this
	// is the documented lossiness of nil_elements with empty-repr elements.
	require.Len(t, viaReflect.Vals, 5)
	require.Nil(t, viaReflect.Vals[0])
	require.Nil(t, viaReflect.Vals[1], "&{Val:0} with empty repr must normalize to nil on decode")
	require.NotNil(t, viaReflect.Vals[2])
	require.Equal(t, int32(5), viaReflect.Vals[2].Val)
	require.Nil(t, viaReflect.Vals[3])
	require.NotNil(t, viaReflect.Vals[4])
	require.Equal(t, int32(99), viaReflect.Vals[4].Val)

	// (4) Byte-stability: re-encode the (lossy) decoded value and assert
	// bytes match the original. This is the strongest correctness check
	// available when strict DeepEqual is impossible.
	bzReflectRT, err := cdc.MarshalReflect(&viaReflect)
	require.NoError(t, err)
	require.Equal(t, bzReflect, bzReflectRT, "byte-stability via reflect")

	bzBinary2RT, err := cdc.MarshalBinary2(&viaBinary2)
	require.NoError(t, err)
	require.Equal(t, bzBinary2, bzBinary2RT, "byte-stability via genproto2")
}
