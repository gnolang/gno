package int256

import (
	"gno.land/p/demo/uint256"
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
