package gnsmath

import (
	"gno.land/p/gnoswap/consts"
	u256 "gno.land/p/gnoswap/uint256"
)

var (
	msbShifts = []bitShift{
		{u256.MustFromDecimal(consts.Q128), 128}, // 2^128
		{u256.MustFromDecimal(consts.Q64), 64},   // 2^64
		{u256.NewUint(0x100000000), 32},          // 2^32
		{u256.NewUint(0x10000), 16},              // 2^16
		{u256.NewUint(0x100), 8},                 // 2^8
		{u256.NewUint(0x10), 4},                  // 2^4
		{u256.NewUint(0x4), 2},                   // 2^2
		{u256.NewUint(0x2), 1},                   // 2^1
	}

	lsbShifts = []bitShift{
		{u256.MustFromDecimal(consts.MAX_UINT128), 128},
		{u256.MustFromDecimal(consts.MAX_UINT64), 64},
		{u256.MustFromDecimal(consts.MAX_UINT32), 32},
		{u256.MustFromDecimal(consts.MAX_UINT16), 16},
		{u256.MustFromDecimal(consts.MAX_UINT8), 8},
		{u256.NewUint(0xf), 4},
		{u256.NewUint(0x3), 2},
		{u256.NewUint(0x1), 1},
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

	r := uint8(0)
	for _, s := range msbShifts {
		if x.Gte(s.bitPattern) {
			x = new(u256.Uint).Rsh(x, s.shift)
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

	r := uint8(255)
	for _, s := range lsbShifts {
		if new(u256.Uint).And(x, s.bitPattern).Gt(u256.Zero()) {
			r -= uint8(s.shift)
		} else {
			x = new(u256.Uint).Rsh(x, s.shift)
		}
	}

	return r
}
