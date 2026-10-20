package gnolang

import (
	"fmt"
	"math"
	"math/big"
	"testing"
	"unicode/utf8"

	"github.com/gnolang/gno/gnovm/pkg/gnolang/internal/softfloat"
	"github.com/stretchr/testify/require"
)

func TestConvertStringToRunes(t *testing.T) {
	t.Parallel()

	runeSliceType := &SliceType{Elt: Int32Type}
	tests := []struct {
		name  string
		input string
	}{
		{"ASCII", "hello"},
		{"multibyte", "héllo, 世界"},
		{"malformed UTF-8", "\xffa\xc0"},
		{"empty", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			alloc := NewAllocator(math.MaxInt64)
			tv := TypedValue{T: StringType, V: StringValue(tc.input)}
			ConvertTo(alloc, nil, &tv, runeSliceType, false)

			slice := tv.V.(*SliceValue)
			base := slice.Base.(*ArrayValue)
			n := utf8.RuneCountInString(tc.input)
			require.Equal(t, n, slice.Length)
			require.Equal(t, n, slice.Maxcap)
			require.Len(t, base.List, n)
			require.Equal(t, n, cap(base.List))
			require.NotNil(t, base.List)
			require.Nil(t, base.Data)

			got := make([]rune, n)
			for i := range base.List {
				got[i] = base.List[i].GetInt32()
			}
			require.Equal(t, []rune(tc.input), got)

			_, bytes := alloc.Status()
			require.Equal(t, int64(allocArray+allocArrayItem*n+allocSlice), bytes)
		})
	}

	t.Run("nil allocator", func(t *testing.T) {
		tv := TypedValue{T: StringType, V: StringValue("é")}
		require.NotPanics(t, func() {
			ConvertTo(nil, nil, &tv, runeSliceType, false)
		})
		require.Equal(t, int32('é'), tv.V.(*SliceValue).Base.(*ArrayValue).List[0].GetInt32())
	})

	t.Run("allocation limit", func(t *testing.T) {
		const input = "ab"
		arrayCost := int64(allocArray + allocArrayItem*utf8.RuneCountInString(input))
		alloc := NewAllocator(arrayCost - 1)
		tv := TypedValue{T: StringType, V: StringValue(input)}

		var recovered any
		func() {
			defer func() { recovered = recover() }()
			ConvertTo(alloc, nil, &tv, runeSliceType, false)
		}()
		require.Contains(t, fmt.Sprint(recovered), "allocation limit exceeded")

		// The array budget check must happen before any allocation.
		_, bytes := alloc.Status()
		require.Zero(t, bytes, "budget must be checked before the backing allocation")
	})
}

func TestConvertUntypedBigdecToFloat(t *testing.T) {
	t.Parallel()

	dst := &TypedValue{}

	// Smallest nonzero float64 / 2 rounds to zero when converted to float64.
	r := new(big.Rat).SetFloat64(math.SmallestNonzeroFloat64 / 2)
	bd := BigdecValue{
		V: r,
	}

	typ := Float64Type

	ConvertUntypedBigdecTo(dst, bd, typ)

	require.True(t, softfloat.Feq64(dst.GetFloat64(), 0))
}

func TestBigdecErrString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		num  int64
		den  int64
		want string
	}{
		{"integer", 42, 1, "42"},
		{"negative integer", -7, 1, "-7"},
		{"one-decimal", 6, 5, "1.2"},           // 1.2
		{"two-decimal", 157, 50, "3.14"},       // 3.14
		{"tiny terminating", 1, 1000, "0.001"}, // 1/1000
		{"eighth", 3, 8, "0.375"},              // 3/8
		{"negative decimal", -6, 5, "-1.2"},    // -1.2
		{"non-terminating", 1, 3, "1/3"},       // 1.0/3.0 style: falls back
		{"non-terminating 2", 22, 7, "22/7"},   // classic
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := big.NewRat(tc.num, tc.den)
			require.Equal(t, tc.want, bigdecErrString(r))
		})
	}
}

func TestConvertUntypedBigdecToFloat32(t *testing.T) {
	t.Parallel()

	// A representative finite value: 1.5 has an exact float32 encoding.
	dst := &TypedValue{}
	bd := BigdecValue{V: new(big.Rat).SetFloat64(1.5)}
	ConvertUntypedBigdecTo(dst, bd, Float32Type)
	require.Equal(t, math.Float32bits(1.5), dst.GetFloat32())

	// A value below the smallest float32 subnormal must round to zero
	// via softfloat, not become an "implementation-defined" result.
	dst = &TypedValue{}
	tiny := new(big.Rat).SetFloat64(float64(math.SmallestNonzeroFloat32) / 4)
	ConvertUntypedBigdecTo(dst, BigdecValue{V: tiny}, Float32Type)
	require.Equal(t, uint32(0), dst.GetFloat32())

	// A value above MaxFloat32 must panic (would narrow to ±Inf).
	huge := new(big.Rat).SetFloat64(math.MaxFloat64)
	require.PanicsWithValue(t,
		"cannot convert untyped bigdec to float32 -- too close to +-Inf",
		func() {
			ConvertUntypedBigdecTo(&TypedValue{}, BigdecValue{V: huge}, Float32Type)
		})
}

func TestUntypedRuneRepresentation(t *testing.T) {
	t.Parallel()

	newRune := func(bi *big.Int) TypedValue {
		tv := TypedValue{T: UntypedRuneType}
		tv.SetUntypedRune(bi)
		return tv
	}

	// A value within int32 is laid out exactly as a rune literal is.
	lit := TypedValue{T: UntypedRuneType}
	lit.SetInt32('a')
	require.Equal(t, lit, newRune(big.NewInt('a')))
	for _, v := range []int64{math.MinInt32, -1, 0, math.MaxInt32} {
		tv := newRune(big.NewInt(v))
		require.Nil(t, tv.V, "value %d", v)
		require.False(t, tv.isBigUntypedRune())
		require.Equal(t, int32(v), tv.GetInt32())
	}

	// A value beyond int32 is held exactly, in V.
	for _, v := range []int64{math.MinInt32 - 1, math.MaxInt32 + 1, 1241243320321} {
		tv := newRune(big.NewInt(v))
		require.True(t, tv.isBigUntypedRune(), "value %d", v)
		require.Equal(t, [8]byte{}, tv.N)
		require.Equal(t, v, tv.GetUntypedRune().Int64())
	}
}

func TestConvertUntypedRuneTo(t *testing.T) {
	t.Parallel()

	big1 := big.NewInt(1241243320321) // '\U0010ffff' * '\U0010ffff'
	newRune := func(bi *big.Int) *TypedValue {
		tv := &TypedValue{T: UntypedRuneType}
		tv.SetUntypedRune(bi)
		return tv
	}

	tv := newRune(big1)
	ConvertUntypedTo(tv, Int64Type)
	require.Equal(t, int64(1241243320321), tv.GetInt64())

	tv = newRune(big1)
	ConvertUntypedTo(tv, Uint64Type)
	require.Equal(t, uint64(1241243320321), tv.GetUint64())

	tv = newRune(big1)
	ConvertUntypedTo(tv, Float64Type)
	require.Equal(t, float64(1241243320321), math.Float64frombits(tv.GetFloat64()))

	tv = newRune(big1)
	ConvertUntypedTo(tv, UntypedBigintType)
	require.Equal(t, 0, big1.Cmp(tv.GetBigInt()))

	// An untyped bigint converted to an untyped rune is range checked
	// against int32, as before: only folding produces big-form runes.
	tv = &TypedValue{T: UntypedBigintType, V: BigintValue{V: big.NewInt('a')}}
	ConvertUntypedTo(tv, UntypedRuneType)
	require.False(t, tv.isBigUntypedRune())
	require.Equal(t, int32('a'), tv.GetInt32())
	require.PanicsWithValue(t, "bigint overflows target kind", func() {
		ConvertUntypedTo(&TypedValue{T: UntypedBigintType, V: BigintValue{V: big1}}, UntypedRuneType)
	})

	// An in-range rune to a narrower type leaves no stale bytes behind.
	tv = newRune(big.NewInt(0x141))
	ConvertUntypedTo(tv, Uint16Type)
	require.Equal(t, uint16(0x141), tv.GetUint16())
	tv = newRune(big.NewInt(0x41))
	ConvertUntypedTo(tv, Int8Type)
	require.Equal(t, [8]byte{0x41}, tv.N)

	// Out of range targets, including rune itself (the default type).
	for _, typ := range []Type{nil, Int32Type, Uint32Type, Int8Type} {
		require.PanicsWithValue(t, "rune overflows target kind", func() {
			ConvertUntypedTo(newRune(big1), typ)
		}, "type %v", typ)
	}
	require.PanicsWithValue(t, "rune underflows target kind", func() {
		ConvertUntypedTo(newRune(big.NewInt(-97)), Uint64Type)
	})
}
