package tests_test

import (
	"fmt"
	"math"
	"testing"

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
}
