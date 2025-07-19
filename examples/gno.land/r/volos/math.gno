package volos

import (
	"gno.land/p/gnoswap/uint256"
)

// WAD represents a decimal with 18 decimals (i.e. 1e18)
var WAD = uint256.NewUint(1000000000000000000)

// WMulDown returns (x * y) / WAD rounded down
func WMulDown(x, y *uint256.Uint) *uint256.Uint {
	return uint256.MulDiv(x, y, WAD)
}

// WMulUp returns (x * y) / WAD rounded up
func WMulUp(x, y *uint256.Uint) *uint256.Uint {
	return uint256.MulDivRoundingUp(x, y, WAD)
}

// WDivDown returns (x * WAD) / y rounded down
func WDivDown(x, y *uint256.Uint) *uint256.Uint {
	return uint256.MulDiv(x, WAD, y)
}

// WDivUp returns (x * WAD) / y rounded up
func WDivUp(x, y *uint256.Uint) *uint256.Uint {
	return uint256.MulDivRoundingUp(x, WAD, y)
}

// MulDivDown multiplies two numbers and divides by a third, rounding down.
// Uses Gnoswap's FullMath implementation for safe multiplication and division.
func MulDivDown(a, b, denominator *uint256.Uint) *uint256.Uint {
	return uint256.MulDiv(a, b, denominator)
}

// MulDivUp multiplies two numbers and divides by a third, rounding up.
// Uses Gnoswap's FullMath implementation for safe multiplication and division.
func MulDivUp(a, b, denominator *uint256.Uint) *uint256.Uint {
	return uint256.MulDivRoundingUp(a, b, denominator)
}

// WTaylorCompounded returns the sum of the first three non-zero terms of a Taylor expansion of e^(nx) - 1
// Used to approximate a continuous compound interest rate
func WTaylorCompounded(x, n *uint256.Uint) *uint256.Uint {
	// First term: x * n
	firstTerm := new(uint256.Uint).Mul(x, n)

	// Second term: (firstTerm * firstTerm) / (2 * WAD)
	twoWad := new(uint256.Uint).Mul(uint256.NewUint(2), WAD)
	secondTerm := MulDivDown(firstTerm, firstTerm, twoWad)

	// Third term: (secondTerm * firstTerm) / (3 * WAD)
	threeWad := new(uint256.Uint).Mul(uint256.NewUint(3), WAD)
	thirdTerm := MulDivDown(secondTerm, firstTerm, threeWad)

	// Sum all terms
	sum := new(uint256.Uint).Add(firstTerm, secondTerm)
	return new(uint256.Uint).Add(sum, thirdTerm)
}
