package int256

import (
	"math"

	"gno.land/p/onbloc/uint256"
)

// SetInt64 sets the Int to the value of the provided int64.
//
// This method allows for easy conversion from standard Go integer types
// to Int, correctly handling both positive and negative values.
func (z *Int) SetInt64(v int64) *Int {
	if v >= 0 {
		z.value.SetUint64(uint64(v))
	} else {
		z.value.SetUint64(uint64(-v)).Neg(&z.value)
	}
	return z
}

// SetUint64 sets the Int to the value of the provided uint64.
func (z *Int) SetUint64(v uint64) *Int {
	z.value.SetUint64(v)
	return z
}

// Uint64 returns the lower 64-bits of z
func (z *Int) Uint64() uint64 {
	if z.Sign() < 0 {
		panic("cannot convert negative int256 to uint64")
	}
	if z.value.Gt(uint256.NewUint(0).SetUint64(math.MaxUint64)) {
		panic("overflow: int256 does not fit in uint64 type")
	}
	return z.value.Uint64()
}

// Int64 returns the lower 64-bits of z
func (z *Int) Int64() int64 {
	if z.Sign() >= 0 {
		if z.value.BitLen() > 64 {
			panic("overflow: int256 does not fit in int64 type")
		}
		return int64(z.value.Uint64())
	}
	var temp uint256.Uint
	temp.Sub(uint256.NewUint(0), &z.value) // temp = -z.value
	if temp.BitLen() > 64 {
		panic("overflow: int256 does not fit in int64 type")
	}
	return -int64(temp.Uint64())
}

// Neg sets z to -x and returns z.)
func (z *Int) Neg(x *Int) *Int {
	if x.IsZero() {
		z.value.Clear()
	} else {
		z.value.Neg(&x.value)
	}
	return z
}

// Set sets z to x and returns z.
func (z *Int) Set(x *Int) *Int {
	z.value.Set(&x.value)
	return z
}

// SetFromUint256 converts a uint256.Uint to Int and sets the value to z.
func (z *Int) SetUint256(x *uint256.Uint) *Int {
	z.value.Set(x)
	return z
}

// ToString returns a string representation of z in base 10.
// The string is prefixed with a minus sign if z is negative.
func (z *Int) String() string {
	if z.value.IsZero() {
		return "0"
	}
	sign := z.Sign()
	var temp uint256.Uint
	if sign >= 0 {
		temp.Set(&z.value)
	} else {
		// temp = -z.value
		temp.Sub(uint256.NewUint(0), &z.value)
	}
	s := temp.Dec()
	if sign < 0 {
		return "-" + s
	}
	return s
}

// NilToZero returns the Int if it's not nil, or a new zero-valued Int otherwise.
//
// This method is useful for safely handling potentially nil Int pointers,
// ensuring that operations always have a valid Int to work with.
func (z *Int) NilToZero() *Int {
	if z == nil {
		return Zero()
	}
	return z
}
