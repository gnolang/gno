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
