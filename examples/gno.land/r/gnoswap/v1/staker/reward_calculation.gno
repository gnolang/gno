package staker

import (
	"gno.land/p/demo/ufmt"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"
)

// liquidityMathAddDelta calculates the new liquidity by applying the delta liquidity to the current liquidity.
// If delta liquidity is negative, it subtracts the absolute value of delta liquidity from the current liquidity.
// If delta liquidity is positive, it adds the absolute value of delta liquidity to the current liquidity.
//
// Parameters:
// - x: The current liquidity as a uint256 value.
// - y: The delta liquidity as a signed int256 value.
//
// Returns:
// - The new liquidity as a uint256 value.
//
// Notes:
// - If `x` or `y` is nil, the function panics with an appropriate error message.
// - If `y` is negative, its absolute value is subtracted from `x`.
//   - The result must be less than `x`. Otherwise, the function panics to prevent underflow.
//
// - If `y` is positive, it is added to `x`.
//   - The result must be greater than or equal to `x`. Otherwise, the function panics to prevent overflow.
//
// - The function ensures correctness by validating the results of the arithmetic operations.
func liquidityMathAddDelta(x *u256.Uint, y *i256.Int) *u256.Uint {
	if x == nil || y == nil {
		panic(addDetailToError(
			errInvalidInput,
			"x or y is nil",
		))
	}

	var z *u256.Uint

	// Subtract or add based on the sign of y
	if y.Lt(i256.Zero()) {
		absDelta := y.Abs()
		z = new(u256.Uint).Sub(x, absDelta)
		if z.Gte(x) {
			panic(addDetailToError(
				errCalculationError,
				ufmt.Sprintf("Condition failed: (z must be < x) (x: %s, y: %s, z:%s)", x.ToString(), y.ToString(), z.ToString()),
			))
		}
	} else {
		z = new(u256.Uint).Add(x, y.Abs())
		if z.Lt(x) {
			panic(addDetailToError(
				errCalculationError,
				ufmt.Sprintf("Condition failed: (z must be >= x) (x: %s, y: %s, z:%s)", x.ToString(), y.ToString(), z.ToString()),
			))
		}
	}

	return z
}
