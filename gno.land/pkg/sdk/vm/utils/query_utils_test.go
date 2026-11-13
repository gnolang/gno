package utils

import (
	"math"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/stretchr/testify/assert"
)

// TestSliceWindowClampNegative exhaustively covers the chain-supplied edge
// cases the slice/array windows guard against: negative or out-of-range
// Offset/Length and offset+length overflow. The pre-clamp code would panic
// on av.List[offset:end] when offset > len.
func TestSliceWindowClampNegative(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                    string
		offset, length, listLen int
		wantOffset, wantEnd     int
	}{
		{"happy_path", 1, 2, 5, 1, 3},
		{"length_zero", 0, 0, 5, 0, 0},
		{"empty_list", 0, 3, 0, 0, 0},
		{"negative_offset", -1, 2, 5, 0, 2},
		{"negative_length", 0, -1, 5, 0, 0},
		{"offset_past_end", 10, 2, 5, 5, 5},
		{"length_past_end", 2, 100, 5, 2, 5},
		{"overflow_offset_plus_length", 2, math.MaxInt - 1, 5, 2, 5},
		{"both_negative", -10, -10, 5, 0, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			offset, end := ClampSliceWindow(tc.offset, tc.length, tc.listLen)
			assert.Equal(t, tc.wantOffset, offset, "offset")
			assert.Equal(t, tc.wantEnd, end, "end")
			assert.LessOrEqual(t, offset, end, "offset must never exceed end")
			assert.LessOrEqual(t, end, tc.listLen, "end must never exceed listLen")
		})
	}
}

func TestPrimitiveString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		pt   gno.PrimitiveType
		set  func(tv *gno.TypedValue)
		want string
	}{
		{"bool_true", gno.BoolType, func(tv *gno.TypedValue) { tv.SetBool(true) }, "true"},
		{"bool_zero", gno.BoolType, func(tv *gno.TypedValue) {}, "false"},
		{"untyped_bool", gno.UntypedBoolType, func(tv *gno.TypedValue) { tv.SetBool(true) }, "true"},
		{"int", gno.IntType, func(tv *gno.TypedValue) { tv.SetInt(-42) }, "-42"},
		{"int8", gno.Int8Type, func(tv *gno.TypedValue) { tv.SetInt8(-1) }, "-1"},
		{"int16", gno.Int16Type, func(tv *gno.TypedValue) { tv.SetInt16(math.MinInt16) }, "-32768"},
		{"int32", gno.Int32Type, func(tv *gno.TypedValue) { tv.SetInt32(math.MaxInt32) }, "2147483647"},
		{"rune", gno.UntypedRuneType, func(tv *gno.TypedValue) { tv.SetInt32('a') }, "97"},
		{"int64", gno.Int64Type, func(tv *gno.TypedValue) { tv.SetInt64(math.MinInt64) }, "-9223372036854775808"},
		{"uint", gno.UintType, func(tv *gno.TypedValue) { tv.SetUint(math.MaxUint64) }, "18446744073709551615"},
		{"uint8", gno.Uint8Type, func(tv *gno.TypedValue) { tv.SetUint8(255) }, "255"},
		{"uint16", gno.Uint16Type, func(tv *gno.TypedValue) { tv.SetUint16(65535) }, "65535"},
		{"uint32", gno.Uint32Type, func(tv *gno.TypedValue) { tv.SetUint32(math.MaxUint32) }, "4294967295"},
		{"uint64", gno.Uint64Type, func(tv *gno.TypedValue) { tv.SetUint64(math.MaxUint64) }, "18446744073709551615"},
		{"float32", gno.Float32Type, func(tv *gno.TypedValue) { tv.SetFloat32(math.Float32bits(1.5)) }, "1.5"},
		{"float64", gno.Float64Type, func(tv *gno.TypedValue) { tv.SetFloat64(math.Float64bits(-0.25)) }, "-0.25"},
		{"zero", gno.Int64Type, func(tv *gno.TypedValue) {}, "0"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// tv.T is a RefType, as in a value of a declared type which is decoded from Amino JSON.
			tv := gno.TypedValue{T: gno.RefType{ID: "gno.land/r/demo/x.T"}}
			tc.set(&tv)
			assert.Equal(t, tc.want, PrimitiveString(tc.pt, tv))
		})
	}
}

func TestGetTypeIDQueryString(t *testing.T) {
	t.Parallel()

	declared := &gno.DeclaredType{PkgPath: "gno.land/r/demo/x", Name: "User", Base: &gno.StructType{}}
	local := &gno.DeclaredType{
		PkgPath:   "gno.land/r/demo/x",
		Name:      "User",
		Base:      &gno.StructType{},
		ParentLoc: gno.Location{PkgPath: "gno.land/r/demo/x", File: "x.gno", Span: gno.Span{Pos: gno.Pos{Line: 3, Column: 1}}},
	}

	assert.Equal(t, "gno.land/r/demo/x.User", GetTypeIDQueryString(gno.RefType{ID: "gno.land/r/demo/x.User"}))
	assert.Equal(t, "gno.land/r/demo/x.User", GetTypeIDQueryString(declared))
	// The same as the TypeID of the type, including a type which is not declared at the package level.
	assert.Equal(t, declared.TypeID().String(), GetTypeIDQueryString(declared))
	assert.Equal(t, local.TypeID().String(), GetTypeIDQueryString(local))
	assert.NotEqual(t, declared.TypeID(), local.TypeID())
	// Drill through a pointer.
	assert.Equal(t, "gno.land/r/demo/x.User", GetTypeIDQueryString(&gno.PointerType{Elt: &gno.PointerType{Elt: declared}}))

	assert.Equal(t, "", GetTypeIDQueryString(nil))
	assert.Equal(t, "", GetTypeIDQueryString(gno.IntType))
	assert.Equal(t, "", GetTypeIDQueryString(&gno.StructType{}))
	assert.Equal(t, "", GetTypeIDQueryString(&gno.SliceType{Elt: declared}))
}

func TestPrimitiveTypeName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "bool", PrimitiveTypeName(gno.BoolType))
	assert.Equal(t, "string", PrimitiveTypeName(gno.StringType))
	assert.Equal(t, "int", PrimitiveTypeName(gno.IntType))
	assert.Equal(t, "uint8", PrimitiveTypeName(gno.Uint8Type))
	assert.Equal(t, "uint64", PrimitiveTypeName(gno.Uint64Type))
	assert.Equal(t, "float64", PrimitiveTypeName(gno.Float64Type))
	assert.Equal(t, "untyped bool", PrimitiveTypeName(gno.UntypedBoolType))
	assert.Equal(t, "rune", PrimitiveTypeName(gno.UntypedRuneType))
	assert.Equal(t, "databyte", PrimitiveTypeName(gno.DataByteType))
	// A primitive type which is not valid doesn't panic.
	assert.Equal(t, "prim(1073741824)", PrimitiveTypeName(gno.PrimitiveType(1<<30)))
}
