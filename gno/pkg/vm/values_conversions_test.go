package gnolang

import (
	"math"
	"testing"

	"github.com/cockroachdb/apd/v3"
	"github.com/stretchr/testify/require"
)

func TestConvertUntypedBigdecToFloat(t *testing.T) {
	t.Parallel()

	dst := &TypedValue{}

	dec, err := apd.New(-math.MaxInt64, -4).SetFloat64(math.SmallestNonzeroFloat64 / 2)
	require.NoError(t, err)
	bd := BigdecValue{
		V: dec,
	}

	typ := Float64Type

	ConvertUntypedBigdecTo(dst, bd, typ)

	require.Equal(t, float64(0), dst.GetFloat64())
}
