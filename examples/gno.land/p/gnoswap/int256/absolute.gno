package int256

import (
	"gno.land/p/gnoswap/uint256"
)

// Abs returns |z|
func (z *Int) Abs() *uint256.Uint {
	return z.abs.Clone()
}

// AbsGt returns true if |z| > x, where x is a uint256
func (z *Int) AbsGt(x *uint256.Uint) bool {
	return z.abs.Gt(x)
}

// AbsLt returns true if |z| < x, where x is a uint256
func (z *Int) AbsLt(x *uint256.Uint) bool {
	return z.abs.Lt(x)
}

// AbsOverflow returns z = |x|, and true if overflow occurs.
//
// The overflow can be happen when negating a minimum or a maximum value.
func (z *Int) AbsOverflow(x *Int) (*Int, bool) {
	z = z.initiateAbs()

	// overflow can be happen when negating a minimum of int256 value
	if x.neg && x.abs.Eq(MinInt256().abs) {
		z.Set(x) // keep the original value
		return z, true
	}

	z.abs.Set(x.abs)
	z.neg = false

	return z, false
}
