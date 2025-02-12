package pool

import (
	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"
	plp "gno.land/p/gnoswap/gnsmath"
	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"
	"gno.land/r/gnoswap/v1/common"
)

// modifyPosition updates a position in the pool and calculates the amount of tokens
// needed (for minting) or returned (for burning). The calculation depends on the current
// price (tick) relative to the position's price range.
//
// The function handles three cases:
//  1. Current price below range (tick < tickLower): only token0 is used/returned
//  2. Current price in range (tickLower <= tick < tickUpper): both tokens are used/returned
//  3. Current price above range (tick >= tickUpper): only token1 is used/returned
//
// Parameters:
//   - params: ModifyPositionParams containing owner, tickLower, tickUpper, and liquidityDelta
//
// Returns:
//   - PositionInfo: updated position information
//   - *u256.Uint: amount of token0 needed/returned
//   - *u256.Uint: amount of token1 needed/returned
func (p *Pool) modifyPosition(params ModifyPositionParams) (PositionInfo, *u256.Uint, *u256.Uint) {
	checkTicks(params.tickLower, params.tickUpper)

	// get current state and price bounds
	tick := p.Slot0Tick()
	// update position state
	position := p.updatePosition(params, tick)
	liqDelta := params.liquidityDelta
	if liqDelta.IsZero() {
		return position, u256.Zero(), u256.Zero()
	}

	amount0, amount1 := i256.Zero(), i256.Zero()

	// covert ticks to sqrt price to use in amount calculations
	// price = 1.0001^tick, but we use sqrtPriceX96
	sqrtRatioLower := common.TickMathGetSqrtRatioAtTick(params.tickLower)
	sqrtRatioUpper := common.TickMathGetSqrtRatioAtTick(params.tickUpper)
	sqrtPriceX96 := p.Slot0SqrtPriceX96()

	// calculate token amounts based on current price position relative to range
	switch {
	case tick < params.tickLower:
		// case 1
		// full range between lower and upper tick is used for token0
		// current tick is below the passed range; liquidity can only become in range by crossing from left to
		// right, when we'll need _more_ token0 (it's becoming more valuable) so user must provide it
		amount0 = calculateToken0Amount(sqrtRatioLower, sqrtRatioUpper, liqDelta)

	case tick < params.tickUpper:
		// case 2
		liquidityBefore := p.liquidity
		// token0 used from current price to upper tick
		amount0 = calculateToken0Amount(sqrtPriceX96, sqrtRatioUpper, liqDelta)
		// token1 used from lower tick to current price
		amount1 = calculateToken1Amount(sqrtRatioLower, sqrtPriceX96, liqDelta)
		// update pool's active liquidity since price is in range
		p.liquidity = liquidityMathAddDelta(liquidityBefore, liqDelta)

	default:
		// case 3
		// full range between lower and upper tick is used for token1
		// current tick is above the passed range; liquidity can only become in range by crossing from right to
		// left, when we'll need _more_ token1 (it's becoming more valuable) so user must provide it
		amount1 = calculateToken1Amount(sqrtRatioLower, sqrtRatioUpper, liqDelta)
	}

	return position, amount0.Abs(), amount1.Abs()
}

func calculateToken0Amount(sqrtPriceLower, sqrtPriceUpper *u256.Uint, liquidityDelta *i256.Int) *i256.Int {
	res := plp.GetAmount0DeltaStr(sqrtPriceLower, sqrtPriceUpper, liquidityDelta)
	return i256.MustFromDecimal(res)
}

func calculateToken1Amount(sqrtPriceLower, sqrtPriceUpper *u256.Uint, liquidityDelta *i256.Int) *i256.Int {
	res := plp.GetAmount1DeltaStr(sqrtPriceLower, sqrtPriceUpper, liquidityDelta)
	return i256.MustFromDecimal(res)
}

func checkTicks(tickLower, tickUpper int32) {
	assertTickLowerLessThanUpper(tickLower, tickUpper)
	assertValidTickLower(tickLower)
	assertValidTickUpper(tickUpper)
}

func assertTickLowerLessThanUpper(tickLower, tickUpper int32) {
	if tickLower >= tickUpper {
		panic(addDetailToError(
			errInvalidTickRange,
			ufmt.Sprintf("tickLower(%d), tickUpper(%d)", tickLower, tickUpper),
		))
	}
}

func assertValidTickLower(tickLower int32) {
	if tickLower < consts.MIN_TICK {
		panic(addDetailToError(
			errTickLowerInvalid,
			ufmt.Sprintf("tickLower(%d) < MIN_TICK(%d)", tickLower, consts.MIN_TICK),
		))
	}
}

func assertValidTickUpper(tickUpper int32) {
	if tickUpper > consts.MAX_TICK {
		panic(addDetailToError(
			errTickUpperInvalid,
			ufmt.Sprintf("tickUpper(%d) > MAX_TICK(%d)", tickUpper, consts.MAX_TICK),
		))
	}
}
