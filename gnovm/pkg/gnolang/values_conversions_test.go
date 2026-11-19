package gnolang

import (
	"fmt"
	"math"
	"math/big"
	"runtime"
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
			tv := TypedValue{T: StringType, V: StringValue{Str: tc.input}}
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
		tv := TypedValue{T: StringType, V: StringValue{Str: "é"}}
		require.NotPanics(t, func() {
			ConvertTo(nil, nil, &tv, runeSliceType, false)
		})
		require.Equal(t, int32('é'), tv.V.(*SliceValue).Base.(*ArrayValue).List[0].GetInt32())
	})

	t.Run("allocation limit", func(t *testing.T) {
		const input = "ab"
		arrayCost := int64(allocArray + allocArrayItem*utf8.RuneCountInString(input))
		alloc := NewAllocator(arrayCost - 1)
		tv := TypedValue{T: StringType, V: StringValue{Str: input}}

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

// bigdecFloat returns mant * 2^exp in the float form.
func bigdecFloat(mant float64, exp int) BigdecValue {
	m := new(big.Float).SetPrec(BigdecFloatPrec).SetFloat64(mant)
	return BigdecValue{F: m.SetMantExp(m, exp)}
}

func TestBigdecFloatString(t *testing.T) {
	t.Parallel()

	// At the bound the exact rendering stays, so in-range output is unchanged.
	for _, exp := range []int{ratOverflowBits, -ratOverflowBits} {
		bd := bigdecFloat(0.7, exp)
		require.Equal(t, bd.F.Text('g', -1), bd.String())
	}
	cases := []struct {
		bd   BigdecValue
		want string
	}{
		{bigdecFloat(0.7, ratOverflowBits+1), "1.46214e+1233"},
		{bigdecFloat(0.7, -ratOverflowBits-1), "3.35124e-1234"},
		{parseBigdecLiteral("1.23456789e-5000", "float"), "1.23457e-5000"},
		{parseBigdecLiteral("-9.9999999e-100000", "float"), "-1e-99999"},
		{parseBigdecLiteral("1e-646456992", "float"), "1e-646456992"},
		{parseBigdecLiteral("1e646456992", "float"), "1e+646456992"},
		{bigdecFloat(0.7, math.MaxInt32), "6.16565e+646456992"},
		{bigdecFloat(0.7, math.MinInt32), "3.97363e-646456994"},
	}
	for _, tc := range cases {
		require.Equal(t, tc.want, tc.bd.String())
		require.Equal(t, tc.want, bigdecValueErrString(tc.bd))
		tv := TypedValue{T: UntypedBigdecType, V: tc.bd}
		require.Equal(t, tc.want, BoundedSprintTV(tv, nil, 1024))
	}
}

func TestConvertUntypedBigdecExtremeExponent(t *testing.T) {
	t.Parallel()

	for _, typ := range []Type{UntypedBigintType, IntType, Uint64Type} {
		for _, bd := range []BigdecValue{bigdecFloat(1, ratOverflowBits), parseBigdecLiteral("1e646456992", "float")} {
			require.PanicsWithValue(t, "bigint overflows target kind", func() {
				ConvertUntypedBigdecTo(&TypedValue{}, bd, typ)
			})
		}
		for _, bd := range []BigdecValue{bigdecFloat(-1, ratOverflowBits), parseBigdecLiteral("-1e646456992", "float")} {
			require.PanicsWithValue(t, "bigint underflows target kind", func() {
				ConvertUntypedBigdecTo(&TypedValue{}, bd, typ)
			})
		}
	}
	// A float-form integer at the bound still converts, as a rat that size does.
	dst := &TypedValue{}
	ConvertUntypedBigdecTo(dst, bigdecFloat(1, ratOverflowBits-1), UntypedBigintType)
	require.Equal(t, ratOverflowBits, dst.V.(BigintValue).V.BitLen())
}

// Not parallel: TotalAlloc counts every goroutine's allocations.
func TestConvertUntypedBigdecExtremeExponentAlloc(t *testing.T) {
	// A typed target's own checks give the same panic once the integer is
	// built, so only the allocation tells the guard ran first.
	for _, lit := range []string{"1e646456992", "-1e646456992"} {
		huge := parseBigdecLiteral(lit, "float")
		for _, typ := range []Type{UntypedBigintType, IntType, Uint64Type} {
			var before, after runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&before)
			require.Panics(t, func() {
				ConvertUntypedBigdecTo(&TypedValue{}, huge, typ)
			})
			runtime.ReadMemStats(&after)
			require.Less(t, after.TotalAlloc-before.TotalAlloc, uint64(1<<20), lit+" "+typ.String())
		}
	}
}
