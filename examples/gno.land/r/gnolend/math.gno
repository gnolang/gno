package gnolend

import (
	"gno.land/p/demo/uint256"
)

// WAD represents a decimal with 18 decimals (i.e. 1e18)
var WAD = uint256.NewUint(1000000000000000000)

// WMulDown returns (x * y) / WAD rounded down
func WMulDown(x, y *uint256.Uint) *uint256.Uint {
	return MulDivDown(x, y, WAD)
}

// WDivDown returns (x * WAD) / y rounded down
func WDivDown(x, y *uint256.Uint) *uint256.Uint {
	return MulDivDown(x, WAD, y)
}

// WDivUp returns (x * WAD) / y rounded up
func WDivUp(x, y *uint256.Uint) *uint256.Uint {
	return MulDivUp(x, WAD, y)
}

// MulDivDown returns (x * y) / d rounded down
func MulDivDown(x, y, d *uint256.Uint) *uint256.Uint {
	// Multiply
	mul := new(uint256.Uint).Mul(x, y)
	// Divide and round down
	return new(uint256.Uint).Div(mul, d)
}

// MulDivUp returns (x * y) / d rounded up
func MulDivUp(x, y, d *uint256.Uint) *uint256.Uint {
	// Multiply
	mul := new(uint256.Uint).Mul(x, y)
	// Add (d - 1) for rounding up
	dMinusOne := new(uint256.Uint).Sub(d, uint256.NewUint(1))
	mulPlusD := new(uint256.Uint).Add(mul, dMinusOne)
	// Divide and round up
	return new(uint256.Uint).Div(mulPlusD, d)
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
