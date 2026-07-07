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
