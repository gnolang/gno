package gnsmath

import (
	u256 "gno.land/p/gnoswap/uint256"
)

var (
	msbShifts = []bitShift{
		{new(u256.Uint).Lsh(u256.One(), 128), 128}, // 2^128
		{new(u256.Uint).Lsh(u256.One(), 64), 64},   // 2^64
		{new(u256.Uint).Lsh(u256.One(), 32), 32},   // 2^32
		{new(u256.Uint).Lsh(u256.One(), 16), 16},   // 2^16
		{new(u256.Uint).Lsh(u256.One(), 8), 8},     // 2^8
		{new(u256.Uint).Lsh(u256.One(), 4), 4},     // 2^4
		{new(u256.Uint).Lsh(u256.One(), 2), 2},     // 2^2
		{new(u256.Uint).Lsh(u256.One(), 1), 1},     // 2^1
	}

	lsbShifts = []bitShift{
		{new(u256.Uint).Sub(new(u256.Uint).Lsh(u256.One(), 128), u256.One()), 128}, // 2^128 - 1
		{new(u256.Uint).Sub(new(u256.Uint).Lsh(u256.One(), 64), u256.One()), 64},   // 2^64 - 1
		{new(u256.Uint).Sub(new(u256.Uint).Lsh(u256.One(), 32), u256.One()), 32},   // 2^32 - 1
		{new(u256.Uint).Sub(new(u256.Uint).Lsh(u256.One(), 16), u256.One()), 16},   // 2^16 - 1
		{new(u256.Uint).Sub(new(u256.Uint).Lsh(u256.One(), 8), u256.One()), 8},     // 2^8 - 1
		{u256.NewUint(0xf), 4}, // 2^4 - 1 = 15
		{u256.NewUint(0x3), 2}, // 2^2 - 1 = 3
		{u256.NewUint(0x1), 1}, // 2^1 - 1 = 1
	}
)

type bitShift struct {
	bitPattern *u256.Uint
	shift      uint
}

// BitMathMostSignificantBit finds the highest set bit (0-based) in x.
// If x == 0, it panics.
func BitMathMostSignificantBit(x *u256.Uint) uint8 {
	if x.IsZero() {
		panic(errMSBZeroInput)
	}

	temp := x.Clone()
	r := uint8(0)

	for _, s := range msbShifts {
		if temp.Gte(s.bitPattern) {
			temp.Rsh(temp, s.shift)
			r += uint8(s.shift)
		}
	}

	return r
}

// BitMathLeastSignificantBit finds the lowest set bit (0-based) in x.
// If x == 0, it panics.
func BitMathLeastSignificantBit(x *u256.Uint) uint8 {
	if x.IsZero() {
		panic(errLSBZeroInput)
	}

	temp := x.Clone()
	hasSetBits := new(u256.Uint)
	r := uint8(255)

	for _, s := range lsbShifts {
		if !hasSetBits.And(temp, s.bitPattern).IsZero() {
			r -= uint8(s.shift)
		} else {
			temp.Rsh(temp, s.shift)
		}
	}

	return r
}
