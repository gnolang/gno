package router

import (
	"std"

	"gno.land/p/demo/ufmt"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"

	pl "gno.land/r/gnoswap/v1/pool"
)

// swapInner executes the core swap logic by interacting with the pool contract.
// This is the main implementation of token swapping that handles both exact input and output swaps.
//
// Expected behavior:
//   - Forexact input swaps: First return value is the exact input amount
//   - For exact output swaps: Second return value is the exact output amount
//   - Both return values are always positive, regardless of swap direction
//
// Parameters:
//   - amountSpecified: Amount specified for the swap (positive for exact input, negative for exact output)
//   - recipient: Address that will receive the output tokens
//   - sqrtPriceLimitX96: Optional price limit for the swap operation
//   - data: SwapCallbackData containing additional swap information
//
// Returns:
//   - *u256.Uint: Total amount of input tokens used
//   - *u256.Uint: Total amount of output tokens received
func swapInner(
	cur realm,
	amountSpecified *i256.Int,
	recipient std.Address,
	sqrtPriceLimitX96 *u256.Uint,
	data SwapCallbackData,
) (poolRecv, poolOut *u256.Uint) {
	zeroForOne := data.tokenIn < data.tokenOut

	sqrtPriceLimitX96 = calculateSqrtPriceLimitForSwap(zeroForOne, data.fee, sqrtPriceLimitX96)

	// approve only the required amount
	approveAmount := amountSpecified.Abs().Int64()

	// approves pool as spender
	common.GetTokenTeller(data.tokenIn).Approve(poolAddr, approveAmount)
	common.GetTokenTeller(data.tokenOut).Approve(poolAddr, approveAmount)

	amount0Str, amount1Str := pl.Swap(
		cross,
		data.tokenIn,
		data.tokenOut,
		data.fee,
		recipient,
		zeroForOne,
		amountSpecified.ToString(),
		sqrtPriceLimitX96.ToString(),
		data.payer,
	)

	amount0 := i256.MustFromDecimal(amount0Str)
	amount1 := i256.MustFromDecimal(amount1Str)

	poolOut, poolRecv = i256MinMax(amount0, amount1)
	if poolRecv.IsOverflow() || poolOut.IsOverflow() {
		panic("overflow in swapInner")
	}

	return poolRecv, poolOut
}

// RealSwapExecutor implements SwapExecutor for actual swaps
type RealSwapExecutor struct{}

func (e *RealSwapExecutor) execute(p *SingleSwapParams) (amountIn, amountOut *u256.Uint) {
	return swapInner(
		cross,
		p.amountSpecified,
		getPrevAddr(), // if single swap => user will receive
		zero,          // sqrtPriceLimitX96
		newSwapCallbackData(p, getPrevAddr()),
	)
}

// swapDryInner performs a dry-run of a swap operation, calculating the potential
// received and output amounts without actually executing the swap.
//
// Parameters:
//   - amountSpecified: The amount specified for the swap operation. It is an `i256.Int`
//     representing the input token amount or output token amount (depending on the swap direction).
//   - sqrtPriceLimitX96: The price limit of the swap expressed as a square root price in X96 format.
//     This can be a user-defined limit or set to a default min or max price if not provided.
//   - data: A `SwapCallbackData` structure containing details about the tokens being swapped,
//     the fee structure, and other context.
//
// Returns:
// - poolRecv: The absolute value of the maximum amount received by the pool during the swap.
// - poolOut: The absolute value of the minimum amount output from the pool during the swap.
func swapDryInner(
	cur realm,
	amountSpecified *i256.Int,
	sqrtPriceLimitX96 *u256.Uint,
	data SwapCallbackData,
) (poolRecv, poolOut *u256.Uint) {
	zeroForOne := data.tokenIn < data.tokenOut
	sqrtPriceLimitX96 = calculateSqrtPriceLimitForSwap(zeroForOne, data.fee, sqrtPriceLimitX96)

	// check possible
	amount0Str, amount1Str, ok := pl.DrySwap(
		cross,
		data.tokenIn,
		data.tokenOut,
		data.fee,
		zeroForOne,
		amountSpecified.ToString(),
		sqrtPriceLimitX96.ToString(),
	)
	if !ok {
		return zero, zero
	}

	amount0 := i256.MustFromDecimal(amount0Str)
	amount1 := i256.MustFromDecimal(amount1Str)

	poolOut, poolRecv = i256MinMax(amount0, amount1)
	if poolRecv.IsOverflow() || poolOut.IsOverflow() {
		panic("overflow in swapDryInner")
	}

	return poolRecv, poolOut
}

// DrySwapExecutor implements SwapExecutor for dry swaps
type DrySwapExecutor struct{}

func (e *DrySwapExecutor) execute(p *SingleSwapParams) (amountIn, amountOut *u256.Uint) {
	return swapDryInner(
		cross,
		p.amountSpecified,
		zero,
		newSwapCallbackData(p, getPrevAddr()),
	)
}

// calculateSqrtPriceLimitForSwap calculates the price limit for a swap operation.
// This function uses the tick ranges defined by `getMinTick` and `getMaxTick` to set price boundaries.
//
// Price Boundary Visualization:
// ```
//
//	MIN_TICK                    MAX_TICK
//	   v                           v
//	<--|---------------------------|-->
//	   ^                           ^
//	zeroForOne               oneForZero
//	limit + 1                limit - 1
//
// ```
//
// Implementation details:
//   - If a non-zero sqrtPriceLimitX96 is provided, it's used as-is
//   - For zeroForOne swaps (tokenIn < tokenOut):
//   - Uses the minimum tick for the fee tier
//   - Adds 1 to avoid hitting the exact boundary
//   - For oneForZero swaps (tokenIn > tokenOut):
//   - Uses the maximum tick for the fee tier
//   - Subtracts 1 to avoid hitting the exact boundary
//
// Parameters:
//   - zeroForOne: Boolean indicating the swap direction (true for zeroForOne, false for oneForZero)
//   - fee: Fee tier of the pool in basis points
//   - sqrtPriceLimitX96: Optional price limit for the swap operation
//
// Returns:
//   - *u256.Uint: Calculated price limit for the swap operation
func calculateSqrtPriceLimitForSwap(zeroForOne bool, fee uint32, sqrtPriceLimitX96 *u256.Uint) *u256.Uint {
	if !sqrtPriceLimitX96.IsZero() {
		return sqrtPriceLimitX96
	}

	if zeroForOne {
		minTick := getMinTick(fee) + 1
		sqrtPriceLimitX96 = new(u256.Uint).Set(common.TickMathGetSqrtRatioAtTick(minTick))
		if sqrtPriceLimitX96.IsZero() {
			sqrtPriceLimitX96 = u256.MustFromDecimal(consts.MIN_SQRT_RATIO)
		}
		return new(u256.Uint).Add(sqrtPriceLimitX96, one)
	}

	maxTick := getMaxTick(fee) - 1
	sqrtPriceLimitX96 = new(u256.Uint).Set(common.TickMathGetSqrtRatioAtTick(maxTick))
	if sqrtPriceLimitX96.IsZero() {
		sqrtPriceLimitX96 = u256.MustFromDecimal(consts.MAX_SQRT_RATIO)
	}
	return new(u256.Uint).Sub(sqrtPriceLimitX96, one)
}

// getMinTick returns the minimum tick value for a given fee tier.
// The implementation follows Uniswap V3's tick spacing rules where
// lower fee tiers allows for finer price granularity.
//
// Fee tier to min tick mapping demonstrates varying levels of price granularity:
//
// ## How these values are calculated?
//
// The Tick bounds in Uniswap V3 are derived from the desired price range and precisions:
//  1. Price Range: Uniswap V3 uses the formula price = 1.0001^tick
//  2. The minimum tick is calculated to represent a very small but non-zero price:
//     - Let min_tick = log(minimum_price) / log(1.0001)
//     - The minimum price is chosen to be 2^-128 ≈ 2.9387e-39
//     - Therefor, min_tick = log(2^-128) / log(1.0001) ≈ -887272
//
// ### Tick Spacing Adjustment
//
//   - Each fee tier has different tick spacing for efficiency
//   - The actual minimum tick is rounded to the nearest tick spacing:
//   - 0.01% fee -> spacing of 1   -> -887272
//   - 0.05% fee -> spacing of 10  -> -887270
//   - 0.30% fee -> spacing of 60  -> -887220
//   - 1.00% fee -> spacing of 200 -> -887200
//
// ## Tick Range Visualization:
//
// ```
//
//	                     0
//	Fee Tier   Min Tick  |  Max Tick   Tick Spacing
//
// 0.01% (100)  -887272   |   887272         1       finest
//
//	|
//
// 0.05% (500)  -887270   |   887270        10
//
//	|
//
// 0.3% (3000)  -887220   |   887220        60
//
//	|
//
// 1% (10000)   -887200   |   887200        200     coarsest
//
//	|
//
// Price Range:           |
// <cheaper            Price            more expensive>
// ```
//
// Tick spacing determines the granularity of price points:
//
//   - Smaller tick spacing (1) = More precise price points
//     Example for 0.01% fee tier:
//     ```
//     Tick: -887272 [...] -2, -1, 0, 1, 2 [...] 887272
//     Steps:   1           1   1  1  1  1         1
//     ```
//
//   - Larger tick spacing (200) = Fewer, more spread out price points
//     Example for 1% fee tier:
//     ```
//     Tick: -887200 [...] -400, -200, 0, 200, 400 [...] 887200
//     Steps:  200          200   200 200 200  200        200
//     ```
//
// This function returns the minimum tick value for a given fee tier.
//
// Parameters:
//   - fee: Fee tier in basis points
//
// Returns:
//   - int32: Minimum tick value for the given fee tier
//
// Panic:
//   - If the fee tier is not supported
//
// Reference:
//   - https://blog.uniswap.org/uniswap-v3-math-primer
func getMinTick(fee uint32) int32 {
	switch fee {
	case 100:
		return -887272
	case 500:
		return -887270
	case 3000:
		return -887220
	case 10000:
		return -887200
	default:
		panic(addDetailToError(
			errInvalidPoolFeeTier,
			ufmt.Sprintf("unknown fee(%d)", fee),
		))
	}
}

// getMaxTick returns the maximum tick value for a given fee tier.
//
// ## How these values are calculated?
//
// The max tick values are the exact negatives of min tick values because:
//  1. Price symmetry: If min_price = 2^-128, then max_price = 2^128
//  2. Using the same formula: max_tick = log(2^128) / log(1.0001) ≈ 887272
//
// ### Tick Spacing Relationship:
//
// The max ticks follow the same spacing rules as min ticks:
//   - 0.01% fee -> +887272 (finest granularity)
//   - 0.05% fee -> +887270 (10-tick spacing)
//   - 0.30% fee -> +887220 (60-tick spacing)
//   - 1.00% fee -> +887200 (coarsest granularity)
//
// Parameters:
//   - fee: Fee tier in basis points
//
// Returns:
//   - int32: Maximum tick value for the given fee tier
//
// Panic:
//   - If the fee tier is not supported
//
// Reference:
//   - https://blog.uniswap.org/uniswap-v3-math-primer
func getMaxTick(fee uint32) int32 {
	switch fee {
	case 100:
		return 887272
	case 500:
		return 887270
	case 3000:
		return 887220
	case 10000:
		return 887200
	default:
		panic(addDetailToError(
			errInvalidPoolFeeTier,
			ufmt.Sprintf("unknown fee(%d)", fee),
		))
	}
}
