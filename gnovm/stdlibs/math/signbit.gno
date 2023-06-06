package math

import (
	imath "internal/math"
)

// Signbit reports whether x is negative or negative zero.
func Signbit(x float64) bool {
	return imath.Float64bits(x)&(1<<63) != 0
}
