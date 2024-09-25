package int256

import "gno.land/p/demo/uint256"

// // SetInt64 sets z to x and returns z.
// func (z *Int) SetInt64(x int64) *Int {
// 	z.initiateAbs()

// 	neg := false
// 	if x < 0 {
// 		neg = true
// 		x = -x
// 	}
// 	if z.abs == nil {
// 		panic("abs is nil")
// 	}
// 	z.abs = z.abs.SetUint64(uint64(x))
// 	z.neg = neg
// 	return z
// }

// // SetUint64 sets z to x and returns z.
// func (z *Int) SetUint64(x uint64) *Int {
// 	z.initiateAbs()

// 	if z.abs == nil {
// 		panic("abs is nil")
// 	}
// 	z.abs = z.abs.SetUint64(x)
// 	z.neg = false
// 	return z
// }

// // Uint64 returns the lower 64-bits of z
// func (z *Int) Uint64() uint64 {
// 	return z.abs.Uint64()
// }

// // Int64 returns the lower 64-bits of z
// func (z *Int) Int64() int64 {
// 	_abs := z.abs.Clone()

// 	if z.neg {
// 		return -int64(_abs.Uint64())
// 	}
// 	return int64(_abs.Uint64())
// }

// // Neg sets z to -x and returns z.)
// func (z *Int) Neg(x *Int) *Int {
// 	z.abs.Set(x.abs)
// 	if z.abs.IsZero() {
// 		z.neg = false
// 	} else {
// 		z.neg = !x.neg
// 	}
// 	return z
// }

// // Set sets z to x and returns z.
// func (z *Int) Set(x *Int) *Int {
// 	z.abs.Set(x.abs)
// 	z.neg = x.neg
// 	return z
// }

// // SetFromUint256 converts a uint256.Uint to Int and sets the value to z.
// func (z *Int) SetUint256(x *uint256.Uint) *Int {
// 	z.abs.Set(x)
// 	z.neg = false
// 	return z
// }

// // OBS, differs from original mempooler int256
// // ToString returns the decimal representation of z.
// func (z *Int) ToString() string {
// 	if z == nil {
// 		panic("int256: nil pointer to ToString()")
// 	}

// 	t := z.abs.Dec()
// 	if z.neg {
// 		return "-" + t
// 	}

// 	return t
// }

// ToString returns a string representation of z in base 10.
// The string is prefixed with a minus sign if z is negative.
//
// TODO: fix
func (z *Int) ToString() string {
	// Check if the number is zero
	if z.value.IsZero() {
		return "0"
	}

	// Check if the number is negative
	isNegative := z.Sign() < 0

	// Create a copy of z.value to avoid modifying the original
	var absValue uint256.Uint
	if isNegative {
		// For negative numbers, we need to compute the absolute value
		// by performing two's complement negation
		absValue.Sub(uint256.Zero(), &z.value)
	} else {
		absValue.Set(&z.value)
	}

	// Convert the absolute value to a decimal string
	str := absValue.ToString()

	// Add a minus sign for negative numbers
	if isNegative {
		return "-" + str
	}

	return str
}
