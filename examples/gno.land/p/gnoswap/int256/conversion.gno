package int256

import (
	"gno.land/p/gnoswap/uint256"
)

// SetInt64 sets z to x and returns z.
func (z *Int) SetInt64(x int64) *Int {
	z = z.initiateAbs()

	neg := false
	if x < 0 {
		neg = true
		x = -x
	}
	if z.abs == nil {
		panic("int256_SetInt64()__abs is nil")
	}
	z.abs = z.abs.SetUint64(uint64(x))
	z.neg = neg
	return z
}

// SetUint64 sets z to x and returns z.
func (z *Int) SetUint64(x uint64) *Int {
	z = z.initiateAbs()

	if z.abs == nil {
		panic("int256_SetUint64()__abs is nil")
	}
	z.abs = z.abs.SetUint64(x)
	z.neg = false
	return z
}

// Uint64 returns the lower 64-bits of z
func (z *Int) Uint64() uint64 {
	return z.abs.Uint64()
}

// Int64 returns the lower 64-bits of z
func (z *Int) Int64() int64 {
	u := z.abs.Uint64()

	if u == 1<<63 {
		return -1 << 63 // always return int64 min value
	}

	if u < 1<<63 {
		if z.neg {
			return -int64(u)
		}
		return int64(u)
	}

	// convert uint64 to int64 safely using two's complement
	// when the number is greater than 2^63
	res := -int64(^u + 1)
	if z.neg {
		return -res
	}
	return res
}

// Neg sets z to -x and returns z.)
func (z *Int) Neg(x *Int) *Int {
	z.abs.Set(x.abs)
	if z.abs.IsZero() {
		z.neg = false
	} else {
		z.neg = !x.neg
	}
	return z
}

// NegOverflow returns z = -x, and true if overflow occurs.
//
// The overflow can be happen when negating a minimum or a maximum value.
func (z *Int) NegOverflow(x *Int) (*Int, bool) {
	z = z.initiateAbs()

	if x.IsZero() {
		z.abs.Clear()
		z.neg = false
		return z, false
	}

	if x.neg && x.abs.Eq(MinInt256().abs) {
		z.Set(x) // must preserve the original value
		return z, true
	}

	z.abs.Set(x.abs)
	z.neg = !x.neg

	return z, false
}

// Set sets z to x and returns z.
func (z *Int) Set(x *Int) *Int {
	z.abs.Set(x.abs)
	z.neg = x.neg
	return z
}

// SetFromUint256 converts a uint256.Uint to Int and sets the value to z.
func (z *Int) SetUint256(x *uint256.Uint) *Int {
	z.abs.Set(x)
	z.neg = false
	return z
}

// OBS, differs from original mempooler int256
// ToString returns the decimal representation of z.
func (z *Int) ToString() string {
	if z == nil {
		panic("int256: nil pointer to ToString()")
	}

	t := z.abs.Dec()
	if z.neg {
		return "-" + t
	}
	return t
}
