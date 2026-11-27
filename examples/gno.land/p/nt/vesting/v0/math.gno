package vesting

import (
	"math"
	"math/bits"
	"math/overflow"
)

// mulDivFloor returns floor(a*b/d) for non-negative a and b and positive d.
// The product is computed in 128 bits, so a value near MaxInt64 does not wrap.
// Vesting only calls this with b < d, which makes the quotient fit in a.
func mulDivFloor(a, b, d int64) (int64, error) {
	if a < 0 || b < 0 || d <= 0 {
		return 0, ErrMath
	}
	if a == 0 || b == 0 {
		return 0, nil
	}
	hi, lo := bits.Mul64(uint64(a), uint64(b))
	den := uint64(d)
	// bits.Div64 panics when the quotient does not fit in uint64.
	if hi >= den {
		return 0, ErrMath
	}
	quo, _ := bits.Div64(hi, lo, den)
	if quo > uint64(math.MaxInt64) {
		return 0, ErrMath
	}
	return int64(quo), nil
}

func addInt(a, b int64) (int64, error) {
	c, ok := overflow.Add64(a, b)
	if !ok {
		return 0, ErrMath
	}
	return c, nil
}

func subInt(a, b int64) (int64, error) {
	c, ok := overflow.Sub64(a, b)
	if !ok || c < 0 {
		return 0, ErrMath
	}
	return c, nil
}
