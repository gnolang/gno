package common

import (
	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"
	u256 "gno.land/p/gnoswap/uint256"
)

// toAscendingOrder checks if the first value is greater than
// the second then swaps two values.
func toAscendingOrder(a, b *u256.Uint) (*u256.Uint, *u256.Uint) {
	if a.Gt(b) {
		return b, a
	}

	return a, b
}

// toUint128 ensures a *u256.Uint value fits within the uint128 range.
//
// This function validates that the given `value` is properly initialized (not nil) and checks whether
// it exceeds the maximum value of uint128. If the value exceeds the uint128 range,
// it applies a masking operation to truncate the value to fit within the uint128 limit.
//
// Parameters:
//   - value: *u256.Uint, the value to be checked and possibly truncated.
//
// Returns:
//   - *u256.Uint: A value guaranteed to fit within the uint128 range.
//
// Notes:
//   - The function first checks if the value is not nil to avoid potential runtime errors.
//   - The mask ensures that only the lower 128 bits of the value are retained.
//   - If the input value is already within the uint128 range, it is returned unchanged.
//   - If masking is required, a new instance is returned without modifying the input.
//   - MAX_UINT128 is a constant representing `2^128 - 1`.
func toUint128(value *u256.Uint) *u256.Uint {
	assertOnlyNotNil(value)
	if value.Gt(u256.MustFromDecimal(consts.MAX_UINT128)) {
		mask := new(u256.Uint).Lsh(u256.One(), consts.Q128_RESOLUTION)
		mask = new(u256.Uint).Sub(mask, u256.One())
		return new(u256.Uint).And(value, mask)
	}
	return value
}

// safeConvertToUint128 safely ensures a *u256.Uint value fits within the uint128 range.
//
// This function verifies that the provided unsigned 256-bit integer does not exceed the maximum value for uint128 (`2^128 - 1`).
// If the value is within the uint128 range, it is returned as is; otherwise, the function triggers a panic.
//
// Parameters:
// - value (*u256.Uint): The unsigned 256-bit integer to be checked.
//
// Returns:
// - *u256.Uint: The same value if it is within the uint128 range.
//
// Panics:
//   - If the value exceeds the maximum uint128 value (`2^128 - 1`), the function will panic with a descriptive error
//     indicating the overflow and the original value.
//
// Notes:
// - The constant `MAX_UINT128` is defined as `340282366920938463463374607431768211455` (the largest uint128 value).
// - No actual conversion occurs since the function works directly with *u256.Uint types.
//
// Example:
// validUint128 := safeConvertToUint128(u256.MustFromDecimal("340282366920938463463374607431768211455")) // Valid
// safeConvertToUint128(u256.MustFromDecimal("340282366920938463463374607431768211456")) // Panics due to overflow
func safeConvertToUint128(value *u256.Uint) *u256.Uint {
	if value.Gt(u256.MustFromDecimal(consts.MAX_UINT128)) {
		panic(ufmt.Sprintf(
			"%v: amount(%s) overflows uint128 range",
			errOverFlow, value.ToString()))
	}
	return value
}

// computeLiquidityForAmount0 calculates the liquidity for a given amount of token0.
//
// This function computes the maximum possible liquidity that can be provided for `token0`
// based on the provided price boundaries (sqrtRatioAX96 and sqrtRatioBX96) in Q64.96 format.
//
// Parameters:
//   - sqrtRatioAX96: *u256.Uint - The square root price at the lower tick boundary (Q64.96).
//   - sqrtRatioBX96: *u256.Uint - The square root price at the upper tick boundary (Q64.96).
//   - amount0: *u256.Uint - The amount of token0 to be converted to liquidity.
//
// Returns:
//   - *u256.Uint: The calculated liquidity, represented as an unsigned 128-bit integer (uint128).
//
// Panics:
//   - If the resulting liquidity exceeds the uint128 range, `safeConvertToUint128` will trigger a panic.
func computeLiquidityForAmount0(sqrtRatioAX96, sqrtRatioBX96, amount0 *u256.Uint) *u256.Uint {
	sqrtRatioAX96, sqrtRatioBX96 = toAscendingOrder(sqrtRatioAX96, sqrtRatioBX96)
	intermediate := u256.MulDiv(sqrtRatioAX96, sqrtRatioBX96, u256.MustFromDecimal(consts.Q96))

	diff := new(u256.Uint).Sub(sqrtRatioBX96, sqrtRatioAX96)
	if diff.IsZero() {
		panic(newErrorWithDetail(
			errIdenticalTicks,
			ufmt.Sprintf("sqrtRatioAX96 (%s) and sqrtRatioBX96 (%s) are identical", sqrtRatioAX96.ToString(), sqrtRatioBX96.ToString()),
		))
	}
	res := u256.MulDiv(amount0, intermediate, diff)
	return safeConvertToUint128(res)
}

// computeLiquidityForAmount1 calculates liquidity based on the provided token1 amount and price range.
//
// This function computes the liquidity for a given amount of token1 by using the difference
// between the upper and lower square root price ratios. The calculation uses Q96 fixed-point
// arithmetic to maintain precision.
//
// Parameters:
//   - sqrtRatioAX96: *u256.Uint - The square root ratio of price at the lower tick, represented in Q96 format.
//   - sqrtRatioBX96: *u256.Uint - The square root ratio of price at the upper tick, represented in Q96 format.
//   - amount1: *u256.Uint - The amount of token1 to calculate liquidity for.
//
// Returns:
//   - *u256.Uint: The calculated liquidity based on the provided amount of token1 and price range.
//
// Notes:
//   - The result is not directly limited to uint128, as liquidity values can exceed uint128 bounds.
//   - If `sqrtRatioAX96 == sqrtRatioBX96`, the function will panic due to division by zero.
//   - Q96 is a constant representing `2^96`, ensuring that precision is maintained during division.
//
// Panics:
//   - If the resulting liquidity exceeds the uint128 range, `safeConvertToUint128` will trigger a panic.
func computeLiquidityForAmount1(sqrtRatioAX96, sqrtRatioBX96, amount1 *u256.Uint) *u256.Uint {
	sqrtRatioAX96, sqrtRatioBX96 = toAscendingOrder(sqrtRatioAX96, sqrtRatioBX96)

	diff := new(u256.Uint).Sub(sqrtRatioBX96, sqrtRatioAX96)
	if diff.IsZero() {
		panic(newErrorWithDetail(
			errIdenticalTicks,
			ufmt.Sprintf("sqrtRatioAX96 (%s) and sqrtRatioBX96 (%s) are identical", sqrtRatioAX96.ToString(), sqrtRatioBX96.ToString()),
		))
	}
	res := u256.MulDiv(amount1, u256.MustFromDecimal(consts.Q96), diff)
	return safeConvertToUint128(res)
}

// GetLiquidityForAmounts calculates the maximum liquidity given the current price (sqrtRatioX96),
// upper and lower price bounds (sqrtRatioAX96 and sqrtRatioBX96), and token amounts (amount0, amount1).
//
// This function evaluates how much liquidity can be obtained for specified amounts of token0 and token1
// within the provided price range. It returns the lesser liquidity based on available token0 or token1
// to ensure the pool remains balanced.
//
// Parameters:
// - sqrtRatioX96: The current price as a square root ratio in Q64.96 format (*u256.Uint).
// - sqrtRatioAX96: The lower bound of the price range as a square root ratio in Q64.96 format (*u256.Uint).
// - sqrtRatioBX96: The upper bound of the price range as a square root ratio in Q64.96 format (*u256.Uint).
// - amount0: The amount of token0 available to provide liquidity (*u256.Uint).
// - amount1: The amount of token1 available to provide liquidity (*u256.Uint).
//
// Returns:
// - *u256.Uint: The maximum possible liquidity that can be minted.
//
// Notes:
//   - The `Clone` method is used to prevent modification of the original values during computation.
//   - The function ensures that liquidity calculations handle edge cases when the current price
//     is outside the specified range by returning liquidity based on the dominant token.
func GetLiquidityForAmounts(sqrtRatioX96, sqrtRatioAX96, sqrtRatioBX96, amount0, amount1 *u256.Uint) *u256.Uint {
	sqrtRatioAX96, sqrtRatioBX96 = toAscendingOrder(sqrtRatioAX96.Clone(), sqrtRatioBX96.Clone())
	var liquidity *u256.Uint

	if sqrtRatioX96.Lte(sqrtRatioAX96) {
		liquidity = computeLiquidityForAmount0(sqrtRatioAX96.Clone(), sqrtRatioBX96.Clone(), amount0.Clone())
	} else if sqrtRatioX96.Lt(sqrtRatioBX96) {
		liquidity0 := computeLiquidityForAmount0(sqrtRatioX96.Clone(), sqrtRatioBX96.Clone(), amount0.Clone())
		liquidity1 := computeLiquidityForAmount1(sqrtRatioAX96.Clone(), sqrtRatioX96.Clone(), amount1.Clone())

		if liquidity0.Lt(liquidity1) {
			liquidity = liquidity0
		} else {
			liquidity = liquidity1
		}
	} else {
		liquidity = computeLiquidityForAmount1(sqrtRatioAX96.Clone(), sqrtRatioBX96.Clone(), amount1.Clone())
	}
	return liquidity
}

// computeAmount0ForLiquidity calculates the required amount of token0 for a given liquidity level
// within a specified price range (represented by sqrt ratios).
//
// This function determines the amount of token0 needed to provide a specified amount of liquidity
// within a price range defined by sqrtRatioAX96 (lower bound) and sqrtRatioBX96 (upper bound).
//
// Parameters:
// - sqrtRatioAX96: The lower bound of the price range as a square root ratio in Q64.96 format (*u256.Uint).
// - sqrtRatioBX96: The upper bound of the price range as a square root ratio in Q64.96 format (*u256.Uint).
// - liquidity: The liquidity to be provided (*u256.Uint).
//
// Returns:
// - *u256.Uint: The amount of token0 required to achieve the specified liquidity level.
//
// Notes:
// - This function assumes the price bounds are expressed in Q64.96 fixed-point format.
// - The function returns 0 if the liquidity is 0 or the price bounds are invalid.
// - Handles edge cases where sqrtRatioAX96 equals sqrtRatioBX96 by returning 0 (to prevent division by zero).
func computeAmount0ForLiquidity(sqrtRatioAX96, sqrtRatioBX96, liquidity *u256.Uint) *u256.Uint {
	sqrtRatioAX96, sqrtRatioBX96 = toAscendingOrder(sqrtRatioAX96, sqrtRatioBX96)
	if sqrtRatioAX96.IsZero() || sqrtRatioBX96.IsZero() || liquidity.IsZero() || sqrtRatioAX96.Eq(sqrtRatioBX96) {
		return u256.Zero()
	}

	val1 := new(u256.Uint).Lsh(liquidity, consts.Q96_RESOLUTION)
	val2 := new(u256.Uint).Sub(sqrtRatioBX96, sqrtRatioAX96)

	res := u256.MulDiv(val1, val2, sqrtRatioBX96)
	res = new(u256.Uint).Div(res, sqrtRatioAX96)

	return res
}

// computeAmount1ForLiquidity calculates the required amount of token1 for a given liquidity level
// within a specified price range (represented by sqrt ratios).
//
// This function determines the amount of token1 needed to provide liquidity between the
// lower (sqrtRatioAX96) and upper (sqrtRatioBX96) price bounds. The calculation is performed
// in Q64.96 fixed-point format, which is standard for many liquidity calculations.
//
// Parameters:
// - sqrtRatioAX96: The lower bound of the price range as a square root ratio in Q64.96 format (*u256.Uint).
// - sqrtRatioBX96: The upper bound of the price range as a square root ratio in Q64.96 format (*u256.Uint).
// - liquidity: The liquidity amount to be used in the calculation (*u256.Uint).
//
// Returns:
// - *u256.Uint: The amount of token1 required to achieve the specified liquidity level.
//
// Notes:
//   - This function handles edge cases where the liquidity is zero or when sqrtRatioAX96 equals sqrtRatioBX96
//     to prevent division by zero.
//   - The calculation assumes sqrtRatioAX96 is always less than or equal to sqrtRatioBX96 after the initial
//     ascending order sorting.
func computeAmount1ForLiquidity(sqrtRatioAX96, sqrtRatioBX96, liquidity *u256.Uint) *u256.Uint {
	sqrtRatioAX96, sqrtRatioBX96 = toAscendingOrder(sqrtRatioAX96, sqrtRatioBX96)
	if liquidity.IsZero() || sqrtRatioAX96.Eq(sqrtRatioBX96) {
		return u256.Zero()
	}

	diff := new(u256.Uint).Sub(sqrtRatioBX96, sqrtRatioAX96)
	res := u256.MulDiv(liquidity, diff, u256.MustFromDecimal(consts.Q96))

	return res
}

// GetAmountsForLiquidity calculates the amounts of token0 and token1 required
// to provide a specified liquidity within a price range.
//
// This function determines the quantities of token0 and token1 necessary to achieve
// a given liquidity level, depending on the current price (sqrtRatioX96) and the
// bounds of the price range (sqrtRatioAX96 and sqrtRatioBX96). The function returns
// the calculated amounts of token0 and token1 as strings.
//
// If the current price is below the lower bound of the price range, only token0 is required.
// If the current price is above the upper bound, only token1 is required. When the
// price is within the range, both token0 and token1 are calculated.
//
// Parameters:
// - sqrtRatioX96: The current price represented as a square root ratio in Q64.96 format (*u256.Uint).
// - sqrtRatioAX96: The lower bound of the price range as a square root ratio in Q64.96 format (*u256.Uint).
// - sqrtRatioBX96: The upper bound of the price range as a square root ratio in Q64.96 format (*u256.Uint).
// - liquidity: The amount of liquidity to be provided (*u256.Uint).
//
// Returns:
// - string: The calculated amount of token0 required to achieve the specified liquidity.
// - string: The calculated amount of token1 required to achieve the specified liquidity.
//
// Notes:
//   - If liquidity is zero, the function returns "0" for both token0 and token1.
//   - The function guarantees that sqrtRatioAX96 is always the lower bound and
//     sqrtRatioBX96 is the upper bound by calling toAscendingOrder().
//   - Edge cases where the current price is exactly on the bounds are handled without division by zero.
//
// Example:
// ```
// amount0, amount1 := GetAmountsForLiquidity(
//
//	u256.MustFromDecimal("79228162514264337593543950336"),  // sqrtRatioX96 (1.0 in Q64.96)
//	u256.MustFromDecimal("39614081257132168796771975168"),  // sqrtRatioAX96 (0.5 in Q64.96)
//	u256.MustFromDecimal("158456325028528675187087900672"), // sqrtRatioBX96 (2.0 in Q64.96)
//	u256.MustFromDecimal("1000000"),                        // Liquidity
//
// )
//
// println("Token0:", amount0, "Token1:", amount1)
//
// // Output:
// Token0: 500000, Token1: 250000
// ```
func GetAmountsForLiquidity(sqrtRatioX96, sqrtRatioAX96, sqrtRatioBX96, liquidity *u256.Uint) (string, string) {
	if liquidity.IsZero() {
		return "0", "0"
	}

	sqrtRatioAX96, sqrtRatioBX96 = toAscendingOrder(sqrtRatioAX96, sqrtRatioBX96)

	amount0 := u256.Zero()
	amount1 := u256.Zero()

	if sqrtRatioX96.Lte(sqrtRatioAX96) {
		amount0 = computeAmount0ForLiquidity(sqrtRatioAX96, sqrtRatioBX96, liquidity)
	} else if sqrtRatioX96.Lt(sqrtRatioBX96) {
		amount0 = computeAmount0ForLiquidity(sqrtRatioX96, sqrtRatioBX96, liquidity)
		amount1 = computeAmount1ForLiquidity(sqrtRatioAX96, sqrtRatioX96, liquidity)
	} else {
		amount1 = computeAmount1ForLiquidity(sqrtRatioAX96, sqrtRatioBX96, liquidity)
	}

	return amount0.ToString(), amount1.ToString()
}
