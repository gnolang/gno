package gnolang

import (
	"math"
	"math/big"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang/internal/softfloat"
	"github.com/stretchr/testify/require"
)

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
		{"one-decimal", 6, 5, "1.2"},               // 1.2
		{"two-decimal", 157, 50, "3.14"},           // 3.14
		{"tiny terminating", 1, 1000, "0.001"},     // 1/1000
		{"eighth", 3, 8, "0.375"},                  // 3/8
		{"negative decimal", -6, 5, "-1.2"},        // -1.2
		{"non-terminating", 1, 3, "1/3"},           // 1.0/3.0 style: falls back
		{"non-terminating 2", 22, 7, "22/7"},       // classic
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
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
