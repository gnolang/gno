package gnolang

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/cockroachdb/apd/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── fixture helpers ───

func tvBool(b bool) TypedValue {
	tv := TypedValue{T: BoolType}
	tv.SetBool(b)
	return tv
}

func tvInt(n int) TypedValue {
	tv := TypedValue{T: IntType}
	tv.SetInt(int64(n))
	return tv
}

func tvInt8(n int8) TypedValue {
	tv := TypedValue{T: Int8Type}
	tv.SetInt8(n)
	return tv
}

func tvInt16(n int16) TypedValue {
	tv := TypedValue{T: Int16Type}
	tv.SetInt16(n)
	return tv
}

func tvInt32(n int32) TypedValue {
	tv := TypedValue{T: Int32Type}
	tv.SetInt32(n)
	return tv
}

func tvInt64(n int64) TypedValue {
	tv := TypedValue{T: Int64Type}
	tv.SetInt64(n)
	return tv
}

func tvUint(n uint) TypedValue {
	tv := TypedValue{T: UintType}
	tv.SetUint(uint64(n))
	return tv
}

func tvUint8(n uint8) TypedValue {
	tv := TypedValue{T: Uint8Type}
	tv.SetUint8(n)
	return tv
}

func tvUint16(n uint16) TypedValue {
	tv := TypedValue{T: Uint16Type}
	tv.SetUint16(n)
	return tv
}

func tvUint32(n uint32) TypedValue {
	tv := TypedValue{T: Uint32Type}
	tv.SetUint32(n)
	return tv
}

func tvUint64(n uint64) TypedValue {
	tv := TypedValue{T: Uint64Type}
	tv.SetUint64(n)
	return tv
}

func tvFloat32(f float32) TypedValue {
	tv := TypedValue{T: Float32Type}
	tv.SetFloat32(math.Float32bits(f))
	return tv
}

func tvFloat64(f float64) TypedValue {
	tv := TypedValue{T: Float64Type}
	tv.SetFloat64(math.Float64bits(f))
	return tv
}

func tvString(s string) TypedValue {
	return TypedValue{T: StringType, V: StringValue(s)}
}

// tvSlice builds a non-byte slice of the given elements with element
// type eltT. All elements share eltT.
func tvSlice(eltT Type, elems ...TypedValue) TypedValue {
	av := &ArrayValue{List: append([]TypedValue(nil), elems...)}
	sv := &SliceValue{Base: av, Offset: 0, Length: len(elems), Maxcap: len(elems)}
	return TypedValue{T: &SliceType{Elt: eltT}, V: sv}
}

// tvArray builds a non-byte array of the given elements with element type eltT.
func tvArray(eltT Type, elems ...TypedValue) TypedValue {
	av := &ArrayValue{List: append([]TypedValue(nil), elems...)}
	return TypedValue{
		T: &ArrayType{Len: len(elems), Elt: eltT},
		V: av,
	}
}

// tvByteArray builds a byte-data array, exercising the hex path.
func tvByteArray(data []byte) TypedValue {
	av := &ArrayValue{Data: append([]byte(nil), data...)}
	return TypedValue{
		T: &ArrayType{Len: len(data), Elt: Uint8Type},
		V: av,
	}
}

// tvStruct builds a struct whose fields are typed value pairs.
func tvStruct(fields ...TypedValue) TypedValue {
	st := &StructType{}
	for i, f := range fields {
		st.Fields = append(st.Fields, FieldType{
			Name: Name("f" + string(rune('0'+i))),
			Type: f.T,
		})
	}
	sv := &StructValue{Fields: append([]TypedValue(nil), fields...)}
	return TypedValue{T: st, V: sv}
}

// tvByteSlice builds a byte-data slice (exercises the slice hex path,
// distinct from the array hex path).
func tvByteSlice(data []byte) TypedValue {
	av := &ArrayValue{Data: append([]byte(nil), data...)}
	sv := &SliceValue{Base: av, Offset: 0, Length: len(data), Maxcap: len(data)}
	return TypedValue{T: &SliceType{Elt: Uint8Type}, V: sv}
}

// tvNilSlice builds a slice with Base == nil ("nil-slice" path).
func tvNilSlice(eltT Type) TypedValue {
	return TypedValue{T: &SliceType{Elt: eltT}, V: &SliceValue{}}
}

// tvMap builds a map fixture by walking pairs (k0, v0, k1, v1, ...).
func tvMap(keyT, valT Type, kvPairs ...TypedValue) TypedValue {
	mv := &MapValue{}
	mv.MakeMap()
	var prev *MapListItem
	for i := 0; i < len(kvPairs); i += 2 {
		item := &MapListItem{Key: kvPairs[i], Value: kvPairs[i+1]}
		if prev == nil {
			mv.List.Head = item
		} else {
			prev.Next = item
			item.Prev = prev
		}
		mv.List.Tail = item
		mv.List.Size++
		prev = item
	}
	return TypedValue{T: &MapType{Key: keyT, Value: valT}, V: mv}
}

// tvZeroMap returns the "zero-map" fixture (mv.List == nil).
func tvZeroMap(keyT, valT Type) TypedValue {
	return TypedValue{T: &MapType{Key: keyT, Value: valT}, V: &MapValue{}}
}

// tvPtrTo wraps tv as a PointerValue whose TV points at it.
func tvPtrTo(tv TypedValue) TypedValue {
	return TypedValue{T: &PointerType{Elt: tv.T}, V: PointerValue{TV: &tv}}
}

// tvNilPtr returns a pointer fixture with TV == nil ("&<nil>" path).
func tvNilPtr(eltT Type) TypedValue {
	return TypedValue{T: &PointerType{Elt: eltT}, V: PointerValue{}}
}

// tvBigint builds an untyped Bigint fixture.
func tvBigint(s string) TypedValue {
	v, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("invalid bigint literal: " + s)
	}
	return TypedValue{T: UntypedBigintType, V: BigintValue{V: v}}
}

// tvSelfReferentialSlice builds a 1-element slice whose only element is
// itself, exercising the seen.IndexOf "ref@N" cycle branch.
func tvSelfReferentialSlice() TypedValue {
	av := &ArrayValue{List: make([]TypedValue, 1)}
	sv := &SliceValue{Base: av, Offset: 0, Length: 1, Maxcap: 1}
	// Use a *SliceType where Elt is interface{}-like — the inner TV
	// just needs to round-trip the cycle.
	st := &SliceType{Elt: &InterfaceType{}}
	tv := TypedValue{T: st, V: sv}
	av.List[0] = tv
	return tv
}

// tvBigdec builds an untyped Bigdec fixture.
func tvBigdec(s string) TypedValue {
	v, _, err := apd.NewFromString(s)
	if err != nil {
		panic("invalid bigdec literal: " + s)
	}
	return TypedValue{T: UntypedBigdecType, V: BigdecValue{V: v}}
}

// tvNestedSlice nests depth levels of single-element []int slices,
// exercising the nestedLimit "..." truncation when depth > nestedLimit.
func tvNestedSlice(depth int) TypedValue {
	inner := tvSlice(IntType, tvInt(0))
	for i := 1; i < depth; i++ {
		t := inner.T
		inner = tvSlice(t, inner)
	}
	return inner
}

// ─── fixture corpus ───

type fixture struct {
	name string
	tv   TypedValue
}

func fixtureCorpus() []fixture {
	return []fixture{
		// Primitives — bool
		{"bool_true", tvBool(true)},
		{"bool_false", tvBool(false)},

		// Primitives — signed int
		{"int_0", tvInt(0)},
		{"int_42", tvInt(42)},
		{"int_neg", tvInt(-42)},
		{"int8_max", tvInt8(127)},
		{"int8_min", tvInt8(-128)},
		{"int16_max", tvInt16(32767)},
		{"int32_max", tvInt32(2147483647)},
		{"int64_max", tvInt64(9223372036854775807)},
		{"int64_min", tvInt64(-9223372036854775808)},

		// Primitives — unsigned int
		{"uint_0", tvUint(0)},
		{"uint8_max", tvUint8(255)},
		{"uint16_max", tvUint16(65535)},
		{"uint32_max", tvUint32(4294967295)},
		{"uint64_max", tvUint64(18446744073709551615)},

		// Primitives — float edge cases
		{"float32_zero", tvFloat32(0)},
		{"float32_one", tvFloat32(1.0)},
		{"float32_neg", tvFloat32(-1.5)},
		{"float32_small", tvFloat32(0.1)},
		{"float32_nan", tvFloat32(float32(math.NaN()))},
		{"float32_pos_inf", tvFloat32(float32(math.Inf(1)))},
		{"float32_neg_inf", tvFloat32(float32(math.Inf(-1)))},
		{"float32_smallest", tvFloat32(math.SmallestNonzeroFloat32)},
		{"float32_largest", tvFloat32(math.MaxFloat32)},
		{"float64_zero", tvFloat64(0)},
		{"float64_one", tvFloat64(1.0)},
		{"float64_pi", tvFloat64(3.141592653589793)},
		{"float64_nan", tvFloat64(math.NaN())},
		{"float64_pos_inf", tvFloat64(math.Inf(1))},
		{"float64_neg_inf", tvFloat64(math.Inf(-1))},
		{"float64_smallest", tvFloat64(math.SmallestNonzeroFloat64)},
		{"float64_largest", tvFloat64(math.MaxFloat64)},

		// Primitives — string
		{"string_empty", tvString("")},
		{"string_ascii", tvString("hello")},
		{"string_unicode", tvString("héllo世界")},
		{"string_special", tvString("line1\nline2\t\"quoted\"")},

		// Slices
		{"slice_empty_int", tvSlice(IntType)},
		{"slice_int_1", tvSlice(IntType, tvInt(42))},
		{"slice_int_5", tvSlice(IntType, tvInt(1), tvInt(2), tvInt(3), tvInt(4), tvInt(5))},
		{"slice_string_3", tvSlice(StringType, tvString("a"), tvString("b"), tvString("c"))},
		{"slice_bool_2", tvSlice(BoolType, tvBool(true), tvBool(false))},

		// Arrays (non-byte)
		{"array_empty_int", tvArray(IntType)},
		{"array_int_3", tvArray(IntType, tvInt(10), tvInt(20), tvInt(30))},

		// Byte arrays — hex path
		{"array_bytes_small", tvByteArray([]byte{0xde, 0xad, 0xbe, 0xef})},
		{"array_bytes_257", tvByteArray(bytes.Repeat([]byte{0xab}, 257))}, // over 256, triggers truncation

		// Structs
		{"struct_empty", tvStruct()},
		{"struct_int_field", tvStruct(tvInt(1))},
		{"struct_mixed", tvStruct(tvInt(1), tvString("x"), tvBool(true))},

		// Nested
		{"slice_of_slice", tvSlice(
			&SliceType{Elt: IntType},
			tvSlice(IntType, tvInt(1), tvInt(2)),
			tvSlice(IntType, tvInt(3), tvInt(4)),
		)},
		{"struct_of_slice", tvStruct(tvSlice(IntType, tvInt(1), tvInt(2)))},

		// Byte slice (distinct from byte array — exercises the slice hex path)
		{"slice_bytes_small", tvByteSlice([]byte{0xde, 0xad, 0xbe, 0xef})},
		{"slice_bytes_257", tvByteSlice(bytes.Repeat([]byte{0xab}, 257))}, // slice >256: "slice[0x..(257)]"

		// Multi-flush: >1024 bytes of output exercises the meteredWriter
		// flush boundary inside WriteString chunking.
		{"slice_int_300", tvSlice(IntType, func() []TypedValue {
			xs := make([]TypedValue, 300)
			for i := range xs {
				xs[i] = tvInt(i)
			}
			return xs
		}()...)},

		// Deep nesting beyond nestedLimit=10 → "..." truncation.
		{"nested_over_limit", tvNestedSlice(12)},

		// Map
		{"map_zero", tvZeroMap(StringType, IntType)},
		{"map_2_entries", tvMap(
			StringType, IntType,
			tvString("a"), tvInt(1),
			tvString("b"), tvInt(2),
		)},

		// Pointer
		{"pointer_nil", tvNilPtr(IntType)},
		{"pointer_to_int", tvPtrTo(tvInt(42))},

		// Bigint / Bigdec (untyped)
		{"bigint_zero", tvBigint("0")},
		{"bigint_big", tvBigint("123456789012345678901234567890")},
		{"bigdec_simple", tvBigdec("3.14159")},

		// Recursive cycle (seen.IndexOf "ref@N" branch)
		{"slice_self_referential", tvSelfReferentialSlice()},

		// Nil slice (Base == nil → "nil-slice")
		{"slice_nil_base", tvNilSlice(IntType)},
	}
}

// ─── inline golden corpus ───
//
// Format contract: every fixture in fixtureCorpus() has an entry here
// with its expected byte-identical output. Both the preserved
// ProtectedString path and the new WriteProtected path must match.
// To re-capture after an intentional format change, run the test with
// the regen helper below (see TestSprintMatchesGolden_regen).

// sliceInt300Golden is built with fmt (independent of meteredWriter) so it
// is a valid oracle for the multi-flush boundary in slice_int_300.
var sliceInt300Golden = func() string {
	var b strings.Builder
	b.WriteString("(slice[")
	for i := 0; i < 300; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, "(%d int)", i)
	}
	b.WriteString("] []int)")
	return b.String()
}()

var sprintGoldens = map[string]string{
	// 256 "AB" pairs = first 256 bytes hex-encoded, then "...]" truncation marker.
	"array_bytes_257":   "(array[0x" + strings.Repeat("AB", 256) + "...] [257]uint8)",
	"slice_bytes_257":   "(slice[0x" + strings.Repeat("AB", 256) + "...(257)] []uint8)",
	"slice_int_300":     sliceInt300Golden,
	"nested_over_limit": "(slice[(slice[(slice[(slice[(slice[(slice[(slice[(slice[(slice[(slice[(... [][]int)] [][][]int)] [][][][]int)] [][][][][]int)] [][][][][][]int)] [][][][][][][]int)] [][][][][][][][]int)] [][][][][][][][][]int)] [][][][][][][][][][]int)] [][][][][][][][][][][]int)] [][][][][][][][][][][][]int)",
	"bigdec_simple":     "(3.14159 <untyped> bigdec)",
	"array_bytes_small": "(array[0xDEADBEEF] [4]uint8)",
	"array_empty_int":   "(array[] [0]int)",
	"array_int_3":       "(array[(10 int),(20 int),(30 int)] [3]int)",
	"bool_false":        "(false bool)",
	"bool_true":         "(true bool)",
	"float32_largest":   "(3.4028235e+38 float32)",
	"float32_nan":       "(NaN float32)",
	"float32_neg":       "(-1.5 float32)",
	"float32_neg_inf":   "(-Inf float32)",
	"float32_one":       "(1 float32)",
	"float32_pos_inf":   "(+Inf float32)",
	"float32_small":     "(0.1 float32)",
	"float32_smallest":  "(1e-45 float32)",
	"float32_zero":      "(0 float32)",
	"float64_largest":   "(1.7976931348623157e+308 float64)",
	"float64_nan":       "(NaN float64)",
	"float64_neg_inf":   "(-Inf float64)",
	"float64_one":       "(1 float64)",
	"float64_pi":        "(3.141592653589793 float64)",
	"float64_pos_inf":   "(+Inf float64)",
	"float64_smallest":  "(5e-324 float64)",
	"float64_zero":      "(0 float64)",
	"int16_max":         "(32767 int16)",
	"int32_max":         "(2147483647 int32)",
	"int64_max":         "(9223372036854775807 int64)",
	"int64_min":         "(-9223372036854775808 int64)",
	"int8_max":          "(127 int8)",
	"int8_min":          "(-128 int8)",
	"int_0":             "(0 int)",
	"int_42":            "(42 int)",
	"int_neg":           "(-42 int)",
	"slice_bool_2":      "(slice[(true bool),(false bool)] []bool)",
	"slice_empty_int":   "(slice[] []int)",
	"slice_int_1":       "(slice[(42 int)] []int)",
	"slice_int_5":       "(slice[(1 int),(2 int),(3 int),(4 int),(5 int)] []int)",
	"slice_of_slice":    "(slice[(slice[(1 int),(2 int)] []int),(slice[(3 int),(4 int)] []int)] [][]int)",
	"slice_string_3":    "(slice[(\"a\" string),(\"b\" string),(\"c\" string)] []string)",
	"string_ascii":      "(\"hello\" string)",
	"string_empty":      "(\"\" string)",
	"string_special":    "(\"line1\\nline2\\t\\\"quoted\\\"\" string)",
	"string_unicode":    "(\"héllo世界\" string)",
	"struct_empty":      "(struct{} struct{})",
	"struct_int_field":  "(struct{(1 int)} struct{f0 int})",
	"struct_mixed":      "(struct{(1 int),(\"x\" string),(true bool)} struct{f0 int; f1 string; f2 bool})",
	"struct_of_slice":   "(struct{(slice[(1 int),(2 int)] []int)} struct{f0 []int})",
	"uint16_max":        "(65535 uint16)",
	"uint32_max":        "(4294967295 uint32)",
	"uint64_max":        "(18446744073709551615 uint64)",
	"uint8_max":         "(255 uint8)",
	"uint_0":            "(0 uint)",

	// Byte slice (slice hex path, distinct from array hex path)
	"slice_bytes_small": "(slice[0xDEADBEEF] []uint8)",

	// Map
	"map_zero":      "(zero-map map[string]int)",
	"map_2_entries": "(map{(\"a\" string):(1 int),(\"b\" string):(2 int)} map[string]int)",

	// Pointer
	"pointer_nil":    "(&<nil> *int)",
	"pointer_to_int": "(&(42 int) *int)",

	// Untyped bigint
	"bigint_zero": "(0 <untyped> bigint)",
	"bigint_big":  "(123456789012345678901234567890 <untyped> bigint)",

	// Recursive cycle (seen.IndexOf "ref@N" branch)
	"slice_self_referential": "(slice[(ref@0 []interface {})] []interface {})",

	// Nil slice (Base == nil)
	"slice_nil_base": "(nil-slice []int)",
}

// ─── tests ───

func TestSprintMatchesGolden(t *testing.T) {
	corpus := fixtureCorpus()
	require.Equal(t, len(corpus), len(sprintGoldens),
		"fixture corpus and sprintGoldens drifted: %d fixtures vs %d goldens",
		len(corpus), len(sprintGoldens))

	for _, fx := range corpus {
		t.Run(fx.name, func(t *testing.T) {
			want, ok := sprintGoldens[fx.name]
			if !ok {
				t.Fatalf("no golden entry for fixture %q", fx.name)
			}

			gotOld := fx.tv.ProtectedString(newSeenValues())

			var buf bytes.Buffer
			mw := newMeteredWriter(&buf, nil)
			fx.tv.WriteProtected(mw, newSeenValues())
			mw.Flush()
			gotNew := buf.String()

			assert.Equal(t, want, gotOld,
				"OLD path (ProtectedString) drifted from golden")
			assert.Equal(t, want, gotNew,
				"NEW path (WriteProtected) doesn't match golden")
		})
	}
}

// ─── benchmarks ───
//
// Apples-to-apples comparison of the ProtectedString hot path against
// the pre-refactor implementation. Uses only the public receiver method
// (TypedValue.String) so the same file runs unmodified against master,
// for direct allocs/op + ns/op comparison via benchstat.

func benchPrintSlice(b *testing.B, n int) {
	b.Helper()
	list := make([]TypedValue, n)
	for i := range list {
		list[i] = TypedValue{T: IntType}
		list[i].SetInt(int64(i))
	}
	av := &ArrayValue{List: list}
	sv := &SliceValue{Base: av, Offset: 0, Length: n, Maxcap: n}
	tv := TypedValue{T: &SliceType{Elt: IntType}, V: sv}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tv.String()
	}
}

func BenchmarkProtectedString_IntSlice_10(b *testing.B)   { benchPrintSlice(b, 10) }
func BenchmarkProtectedString_IntSlice_100(b *testing.B)  { benchPrintSlice(b, 100) }
func BenchmarkProtectedString_IntSlice_1000(b *testing.B) { benchPrintSlice(b, 1000) }

func benchPrintStruct(b *testing.B, n int) {
	b.Helper()
	st := &StructType{}
	fields := make([]TypedValue, n)
	for i := range fields {
		fields[i] = TypedValue{T: IntType}
		fields[i].SetInt(int64(i))
		st.Fields = append(st.Fields, FieldType{
			Name: Name("f"),
			Type: IntType,
		})
	}
	sv := &StructValue{Fields: fields}
	tv := TypedValue{T: st, V: sv}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tv.String()
	}
}

func BenchmarkProtectedString_Struct_10(b *testing.B)   { benchPrintStruct(b, 10) }
func BenchmarkProtectedString_Struct_100(b *testing.B)  { benchPrintStruct(b, 100) }
func BenchmarkProtectedString_Struct_1000(b *testing.B) { benchPrintStruct(b, 1000) }

func BenchmarkProtectedString_ByteArray_256(b *testing.B) {
	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(i)
	}
	av := &ArrayValue{Data: data}
	tv := TypedValue{T: &ArrayType{Len: 256, Elt: Uint8Type}, V: av}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tv.String()
	}
}

func BenchmarkProtectedString_ByteArray_4096(b *testing.B) {
	data := make([]byte, 4096)
	for i := range data {
		data[i] = byte(i)
	}
	av := &ArrayValue{Data: data}
	tv := TypedValue{T: &ArrayType{Len: 4096, Elt: Uint8Type}, V: av}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tv.String()
	}
}

func BenchmarkProtectedString_Nested_StructOfSlices(b *testing.B) {
	st := &StructType{}
	fields := make([]TypedValue, 10)
	for i := range fields {
		inner := make([]TypedValue, 10)
		for j := range inner {
			inner[j] = TypedValue{T: IntType}
			inner[j].SetInt(int64(j))
		}
		av := &ArrayValue{List: inner}
		sv := &SliceValue{Base: av, Offset: 0, Length: 10, Maxcap: 10}
		fields[i] = TypedValue{T: &SliceType{Elt: IntType}, V: sv}
		st.Fields = append(st.Fields, FieldType{
			Name: Name("f"),
			Type: &SliceType{Elt: IntType},
		})
	}
	sv := &StructValue{Fields: fields}
	tv := TypedValue{T: st, V: sv}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tv.String()
	}
}

func BenchmarkProtectedString_Bigint_Large(b *testing.B) {
	v, _ := new(big.Int).SetString("123456789012345678901234567890123456789012345678901234567890", 10)
	tv := TypedValue{T: UntypedBigintType, V: BigintValue{V: v}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tv.String()
	}
}

func BenchmarkProtectedString_Primitive_Int(b *testing.B) {
	tv := TypedValue{T: IntType}
	tv.SetInt(42)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tv.String()
	}
}

func BenchmarkProtectedString_Primitive_String(b *testing.B) {
	tv := TypedValue{T: StringType, V: StringValue("hello world")}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tv.String()
	}
}
