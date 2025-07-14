package gnsmath

import (
	"gno.land/p/gnoswap/consts"
	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"
)

const (
	Q96_RESOLUTION  uint = 96
	Q160_RESOLUTION uint = 160
)

var (
	q96    = u256.MustFromDecimal(consts.Q96)
	max160 = u256.MustFromDecimal(consts.MAX_UINT160)
)

// getNextPriceAmount0Add calculates the next sqrt price when we are adding token0.
// Preserves the rounding-up logic. No in-place mutation of input arguments.
func getNextPriceAmount0Add(
	currentSqrtPriceX96, liquidity, amountToAdd *u256.Uint,
) *u256.Uint {
	// Shift liquidity left by Q96 bits to increase precision
	liquidityShifted := new(u256.Uint).Lsh(liquidity, Q96_RESOLUTION)
	// Multiply the amount to add by the current square root price
	amountTimesSqrtPrice := new(u256.Uint).Mul(amountToAdd, currentSqrtPriceX96)

	// Overflow check: Ensure (amountTimesSqrtPrice / amountToAdd) == currentSqrtPriceX96
	if new(u256.Uint).Div(amountTimesSqrtPrice, amountToAdd).Eq(currentSqrtPriceX96) {
		// Compute denominator: liquidityShifted + (amountToAdd * currentSqrtPriceX96)
		denominator := new(u256.Uint).Add(liquidityShifted, amountTimesSqrtPrice)
		if denominator.Gte(liquidityShifted) {
			return u256.MulDivRoundingUp(liquidityShifted, currentSqrtPriceX96, denominator)
		}
	}

	// Alternative computation path: (liquidityShifted / currentSqrtPriceX96) + amountToAdd
	divValue := new(u256.Uint).Div(liquidityShifted, currentSqrtPriceX96)
	addValue := new(u256.Uint).Add(divValue, amountToAdd)

	// Compute the next square root price using division rounding up
	return u256.DivRoundingUp(liquidityShifted, addValue)
}

// getNextPriceAmount0Remove calculates the next sqrt price when we are removing token0.
// Preserves the rounding-up logic. No in-place mutation of input arguments.
func getNextPriceAmount0Remove(
	currentSqrtPriceX96, liquidity, amountToRemove *u256.Uint,
) *u256.Uint {
	// Shift liquidity left by Q96 bits to increase precision
	liquidityShifted := new(u256.Uint).Lsh(liquidity, Q96_RESOLUTION)

	// Multiply the amount to remove by the current square root price
	amountTimesSqrtPrice := new(u256.Uint).Mul(amountToRemove, currentSqrtPriceX96)

	// Conditions must hold: product/amount == sqrtPX96 and numerator1 > product
	if !new(u256.Uint).Div(amountTimesSqrtPrice, amountToRemove).Eq(currentSqrtPriceX96) ||
		!liquidityShifted.Gt(amountTimesSqrtPrice) {
		panic(errInvalidPoolSqrtPrice)
	}

	// Compute denominator: liquidityShifted - (amountToRemove * currentSqrtPriceX96)
	denominator := new(u256.Uint).Sub(liquidityShifted, amountTimesSqrtPrice)
	// Calculate next square root price: (liquidityShifted * currentSqrtPriceX96) / denominator (with rounding up)
	nextSqrtPrice := u256.MulDivRoundingUp(liquidityShifted, currentSqrtPriceX96, denominator)

	// Check for overflow
	if nextSqrtPrice.Gt(max160) {
		panic(errNextSqrtPriceOverflow)
	}

	return nextSqrtPrice
}

// getNextSqrtPriceFromAmount0RoundingUp calculates the next square root price
// based on the amount of token0 added or removed from the pool.
// NOTE: Always rounds up, because in the exact output case (increasing price) we need to move the price at least
// far enough to get the desired output amount, and in the exact input case (decreasing price) we need to move the
// price less in order to not send too much output.
// The most precise formula for this is liquidity * sqrtPX96 / (liquidity +- amount * sqrtPX96),
// if this is impossible because of overflow, we calculate liquidity / (liquidity / sqrtPX96 +- amount).
//
// Parameters:
// - sqrtPX96: The current square root price as a Q96 fixed-point number (uint160).
// - liquidity: The pool's active liquidity as a Q128 fixed-point number (uint128).
// - amount: The amount of token0 to be added or removed from the pool (uint256).
// - add: A boolean indicating whether the amount of token0 is being added (true) or removed (false).
//
// Returns:
// - The price after adding or removing amount, depending on add
//
// Notes:
// - When `add` is true, the function calculates the new square root price after adding `amount` of token0.
// - When `add` is false, the function calculates the new square root price after removing `amount` of token0.
// - The function uses high-precision math (MulDivRoundingUp, DivRoundingUp) to handle division rounding issues.
// - The function validates input conditions and panics if the state is invalid.
func getNextSqrtPriceFromAmount0RoundingUp(
	sqrtPX96 *u256.Uint,
	liquidity *u256.Uint,
	amount *u256.Uint,
	add bool,
) *u256.Uint {
	// Shortcut: if no amount, return original price
	if amount.IsZero() {
		return sqrtPX96
	}

	if add {
		return getNextPriceAmount0Add(sqrtPX96, liquidity, amount)
	}
	return getNextPriceAmount0Remove(sqrtPX96, liquidity, amount)
}

// getNextPriceAmount1Add calculates the next sqrt price when adding token1.
// Preserves rounding-down logic for the final result.
func getNextPriceAmount1Add(
	sqrtPX96, liquidity, amount *u256.Uint,
) *u256.Uint {
	quotient := u256.Zero()

	if amount.Lte(max160) {
		value := u256.Zero().Lsh(amount, Q96_RESOLUTION)
		quotient = quotient.Div(value, liquidity)
	} else {
		quotient = u256.MulDiv(amount, q96, liquidity)
	}

	res := new(u256.Uint).Add(sqrtPX96, quotient)
	if res.Gt(max160) {
		panic(errSqrtPriceQuotientOverflow)
	}

	return res
}

// getNextPriceAmount1Remove calculates the next sqrt price when removing token1.
// Preserves rounding-down logic for the final result.
func getNextPriceAmount1Remove(
	sqrtPX96, liquidity, amount *u256.Uint,
) *u256.Uint {
	quotient := u256.Zero()

	if amount.Lte(max160) {
		value := new(u256.Uint).Lsh(amount, Q96_RESOLUTION)
		quotient = u256.DivRoundingUp(value, liquidity)
	} else {
		quotient = u256.MulDivRoundingUp(amount, q96, liquidity)
	}

	if !sqrtPX96.Gt(quotient) {
		panic(errSqrtPriceExceedsQuotient)
	}

	res := new(u256.Uint).Sub(sqrtPX96, quotient)
	if res.Gt(max160) {
		// Force mask into uint160
		mask := new(u256.Uint).Lsh(u256.One(), Q160_RESOLUTION)
		mask = mask.Sub(mask, u256.One())
		res = res.And(res, mask)
	}
	return res
}

// getNextSqrtPriceFromAmount1RoundingDown calculates the next square root price
// based on the amount of token1 added or removed from the pool, with rounding down.
// NOTE: Always rounds down, because in the exact output case (decreasing price) we need to move the price at least
// far enough to get the desired output amount, and in the exact input case (increasing price) we need to move the
// price less in order to not send too much output.
// The formula we compute is within <1 wei of the lossless version: sqrtPX96 +- amount / liquidity
//
// Parameters:
// - sqrtPX96: The current square root price as a Q96 fixed-point number (uint160).
// - liquidity: The pool's active liquidity as a Q128 fixed-point number (uint128).
// - amount: The amount of token1 to be added or removed from the pool (uint256).
// - add: A boolean indicating whether the amount of token1 is being added (true) or removed (false).
//
// Returns:
// - The next square root price as a Q96 fixed-point number (uint160).
//
// Notes:
// - When `add` is true, the function calculates the new square root price after adding `amount` of token1.
// - When `add` is false, the function calculates the new square root price after removing `amount` of token1.
// - The function uses high-precision math (MulDiv and DivRoundingUp) to handle division and prevent precision loss.
// - The function validates input conditions and panics if the state is invalid.
func getNextSqrtPriceFromAmount1RoundingDown(
	sqrtPX96,
	liquidity,
	amount *u256.Uint,
	add bool,
) *u256.Uint {
	// Shortcut: if no amount, return original price
	if amount.IsZero() {
		return sqrtPX96
	}

	if add {
		return getNextPriceAmount1Add(sqrtPX96, liquidity, amount)
	}
	return getNextPriceAmount1Remove(sqrtPX96, liquidity, amount)
}

// getNextSqrtPriceFromInput calculates the next square root price
// based on the amount of token0 or token1 added to the pool.
// NOTE: Always rounds up, because in the exact output case (increasing price) we need to move the price at least
// far enough to get the desired output amount, and in the exact input case (decreasing price) we need to move the
// price less in order to not send too much output.
// The most precise formula for this is liquidity * sqrtPX96 / (liquidity +- amount * sqrtPX96),
// if this is impossible because of overflow, we calculate liquidity / (liquidity / sqrtPX96 +- amount).
//
// Parameters:
// - sqrtPX96: The current square root price as a Q96 fixed-point number (uint160).
// - liquidity: The pool's active liquidity as a Q128 fixed-point number (uint128).
// - amountIn: The amount of token0 or token1 to be added to the pool (uint256).
// - zeroForOne: A boolean indicating whether the amount is being added to token0 (true) or token1 (false).
//
// Returns:
// - The price after adding amountIn, depending on zeroForOne
func getNextSqrtPriceFromInput(
	sqrtPX96, liquidity, amountIn *u256.Uint,
	zeroForOne bool,
) *u256.Uint {
	if sqrtPX96.IsZero() {
		panic(errSqrtPriceZero)
	}

	if liquidity.IsZero() {
		panic(errLiquidityZero)
	}

	if zeroForOne {
		return getNextSqrtPriceFromAmount0RoundingUp(sqrtPX96, liquidity, amountIn, true)
	}

	return getNextSqrtPriceFromAmount1RoundingDown(sqrtPX96, liquidity, amountIn, true)
}

// getNextSqrtPriceFromOutput calculates the next square root price
// based on the amount of token0 or token1 removed from the pool.
//
// NOTE:
// - For zeroForOne == true (Token0 -> Token1): The calculation uses rounding down.
// - For zeroForOne == false (Token1 -> Token0): The calculation uses rounding up.
//
// The most precise formula for this is:
// - liquidity * sqrtPX96 / (liquidity ± amount * sqrtPX96)
// If overflow occurs, it falls back to:
// - liquidity / (liquidity / sqrtPX96 ± amount)
//
// Parameters:
// - sqrtPX96: The current square root price as a Q96 fixed-point number (uint160).
// - liquidity: The pool's active liquidity as a Q128 fixed-point number (uint128).
// - amountOut: The amount of token0 or token1 to be removed from the pool (uint256).
// - zeroForOne: A boolean indicating whether the amount is being removed from token0 (true) or token1 (false).
//
// Returns:
// - The price after removing amountOut, depending on zeroForOne.
//
// Notes:
// - Rounding direction depends on the swap direction (zeroForOne).
// - Relies on helper functions:
//   - `getNextSqrtPriceFromAmount1RoundingDown` for Token0 -> Token1.
//   - `getNextSqrtPriceFromAmount0RoundingUp` for Token1 -> Token0.
func getNextSqrtPriceFromOutput(
	sqrtPX96, liquidity, amountOut *u256.Uint,
	zeroForOne bool,
) *u256.Uint {
	if sqrtPX96.IsZero() {
		panic(errSqrtPriceZero)
	}

	if liquidity.IsZero() {
		panic(errLiquidityZero)
	}

	if zeroForOne {
		return getNextSqrtPriceFromAmount1RoundingDown(sqrtPX96, liquidity, amountOut, false)
	}

	return getNextSqrtPriceFromAmount0RoundingUp(sqrtPX96, liquidity, amountOut, false)
}

// getAmount0DeltaHelper calculates the absolute difference between the amounts of token0 in two
// liquidity ranges defined by the square root prices sqrtRatioAX96 and sqrtRatioBX96. The difference is
// calculated relative to the range [sqrtRatioAX96, sqrtRatioBX96].
//
// If sqrtRatioAX96 > sqrtRatioBX96, their values are swapped to ensure sqrtRatioAX96 is the lower bound.
//
// Parameters:
// - sqrtRatioAX96: The lower bound of the range as a Q96 fixed-point number (uint160).
// - sqrtRatioBX96: The upper bound of the range as a Q96 fixed-point number (uint160).
// - liquidity: The pool's active liquidity as a Q128 fixed-point number (uint128).
// - roundUp: A boolean indicating whether the result should be rounded up (true) or down (false).
//
// Returns:
// - The absolute difference between the amounts of token0 in the two ranges as a uint256.
//
// Notes:
// - If sqrtRatioAX96 is zero or negative, the function panics.
// - The result is calculated using high-precision fixed-point arithmetic.
// - Rounding is applied based on the roundUp parameter.
func getAmount0DeltaHelper(
	sqrtRatioAX96, sqrtRatioBX96, liquidity *u256.Uint,
	roundUp bool,
) *u256.Uint {
	if sqrtRatioAX96.Gt(sqrtRatioBX96) {
		sqrtRatioAX96, sqrtRatioBX96 = sqrtRatioBX96, sqrtRatioAX96
	}

	scaledLiquidity := new(u256.Uint).Lsh(liquidity, Q96_RESOLUTION)
	ratioDiff := new(u256.Uint).Sub(sqrtRatioBX96, sqrtRatioAX96)

	if !(sqrtRatioAX96.Gt(u256.Zero())) {
		panic(errSqrtRatioAX96NotPositive)
	}

	if roundUp {
		value := u256.MulDivRoundingUp(scaledLiquidity, ratioDiff, sqrtRatioBX96)
		return u256.DivRoundingUp(value, sqrtRatioAX96)
	}

	value := u256.MulDiv(scaledLiquidity, ratioDiff, sqrtRatioBX96)
	return new(u256.Uint).Div(value, sqrtRatioAX96)
}

// getAmount1DeltaHelper calculates the absolute difference between the amounts of token1 in two
// liquidity ranges defined by the square root prices sqrtRatioAX96 and sqrtRatioBX96. The difference is
// calculated relative to the range [sqrtRatioAX96, sqrtRatioBX96].
//
// If sqrtRatioAX96 > sqrtRatioBX96, their values are swapped to ensure sqrtRatioAX96 is the lower bound.
//
// Parameters:
// - sqrtRatioAX96: The lower bound of the range as a Q96 fixed-point number (uint160).
// - sqrtRatioBX96: The upper bound of the range as a Q96 fixed-point number (uint160).
// - liquidity: The pool's active liquidity as a Q128 fixed-point number (uint128).
// - roundUp: A boolean indicating whether the result should be rounded up (true) or down (false).
//
// Returns:
// - The absolute difference between the amounts of token1 in the two ranges as a uint256.
//
// Notes:
// - Rounding is applied based on the roundUp parameter.
// - The function swaps sqrtRatioAX96 and sqrtRatioBX96 if sqrtRatioAX96 > sqrtRatioBX96.
func getAmount1DeltaHelper(
	sqrtRatioAX96, sqrtRatioBX96, liquidity *u256.Uint,
	roundUp bool,
) *u256.Uint {
	if sqrtRatioAX96.Gt(sqrtRatioBX96) {
		sqrtRatioAX96, sqrtRatioBX96 = sqrtRatioBX96, sqrtRatioAX96
	}

	diff := new(u256.Uint).Sub(sqrtRatioBX96, sqrtRatioAX96)
	if roundUp {
		return u256.MulDivRoundingUp(liquidity, diff, q96)
	}

	return u256.MulDiv(liquidity, diff, q96)
}

// GetAmount0DeltaStr calculates the difference in the amount of token0
// within a specified liquidity range defined by two square root prices (sqrtRatioAX96 and sqrtRatioBX96).
// This function returns the result as a string representation of an int256 value.
//
// If the liquidity is negative, the result is also negative.
//
// Parameters:
// - sqrtRatioAX96: The lower bound of the range as a Q96 fixed-point number (uint160).
// - sqrtRatioBX96: The upper bound of the range as a Q96 fixed-point number (uint160).
// - liquidity: The pool's active liquidity as a signed Q128 fixed-point number (int128).
//
// Returns:
//   - A string representation of the int256 value representing the difference in token0 amounts
//     within the specified range. The value is negative if the liquidity is negative.
//
// Notes:
// - This function relies on the helper function `getAmount0DeltaHelper` to perform the core calculation.
//   - The helper function calculates the absolute difference between token0 amounts within the range.
//   - If the computed result exceeds the maximum allowable value for int256 (2**255 - 1), the function will panic
//     with an appropriate overflow error.
//   - The rounding behavior of the result is controlled by the `roundUp` parameter passed to the helper function:
//   - For negative liquidity, rounding is always down.
//   - For positive liquidity, rounding is always up.
func GetAmount0DeltaStr(
	sqrtRatioAX96, sqrtRatioBX96 *u256.Uint,
	liquidity *i256.Int,
) string {
	if liquidity.IsNeg() {
		u := getAmount0DeltaHelper(sqrtRatioAX96, sqrtRatioBX96, liquidity.Abs(), false)
		if u.Gt(u256.MustFromDecimal(consts.MAX_INT256)) {
			// if u > (2**255 - 1), cannot cast to int256
			panic(errAmount0DeltaOverflow)
		}
		i := i256.FromUint256(u)
		return i256.Zero().Neg(i).ToString()
	}

	u := getAmount0DeltaHelper(sqrtRatioAX96, sqrtRatioBX96, liquidity.Abs(), true)
	if u.Gt(u256.MustFromDecimal(consts.MAX_INT256)) {
		// if u > (2**255 - 1), cannot cast to int256
		panic(errAmount0DeltaOverflow)
	}

	return i256.FromUint256(u).ToString()
}

// GetAmount1DeltaStr calculates the difference in the amount of token1
// within a specified liquidity range defined by two square root prices (sqrtRatioAX96 and sqrtRatioBX96).
// This function returns the result as a string representation of an int256 value.
//
// If the liquidity is negative, the result is also negative.
//
// Parameters:
// - sqrtRatioAX96: The lower bound of the range as a Q96 fixed-point number (uint160).
// - sqrtRatioBX96: The upper bound of the range as a Q96 fixed-point number (uint160).
// - liquidity: The pool's active liquidity as a signed Q128 fixed-point number (int128).
//
// Returns:
//   - A string representation of the int256 value representing the difference in token1 amounts
//     within the specified range. The value is negative if the liquidity is negative.
//
// Notes:
// - This function relies on the helper function `getAmount1DeltaHelper` to perform the core calculation.
// - The rounding behavior of the result is controlled by the `roundUp` parameter passed to the helper function:
//   - For negative liquidity, rounding is always down.
//   - For positive liquidity, rounding is always up.
func GetAmount1DeltaStr(
	sqrtRatioAX96, sqrtRatioBX96 *u256.Uint,
	liquidity *i256.Int,
) string {
	if liquidity.IsNeg() {
		u := getAmount1DeltaHelper(sqrtRatioAX96, sqrtRatioBX96, liquidity.Abs(), false)
		if u.Gt(u256.MustFromDecimal(consts.MAX_INT256)) {
			// if u > (2**255 - 1), cannot cast to int256
			panic(errAmount1DeltaOverflow)
		}
		i := i256.FromUint256(u)
		return i256.Zero().Neg(i).ToString()
	}

	u := getAmount1DeltaHelper(sqrtRatioAX96, sqrtRatioBX96, liquidity.Abs(), true)
	if u.Gt(u256.MustFromDecimal(consts.MAX_INT256)) {
		// if u > (2**255 - 1), cannot cast to int256
		panic(errAmount1DeltaOverflow)
	}

	return i256.FromUint256(u).ToString()
}
