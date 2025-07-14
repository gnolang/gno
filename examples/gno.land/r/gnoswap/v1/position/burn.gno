package position

import (
	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/demo/wugnot"
	"gno.land/r/gnoswap/v1/common"
	pl "gno.land/r/gnoswap/v1/pool"
)

// decreaseLiquidity reduces the liquidity of a given position and collects the corresponding tokens.
// If unwrapResult is true and the position involves WUGNOT, any leftover WUGNOT will be
// unwrapped to GNOT at the end of the operation.
// Returns position ID, liquidity, fee0, fee1, amount0, amount1, poolPath
func decreaseLiquidity(params DecreaseLiquidityParams) (uint64, *u256.Uint, *u256.Uint, *u256.Uint, *u256.Uint, *u256.Uint, string) {
	verifyTokenIdAndOwnership(params.positionId)

	caller := params.caller

	// BEFORE DECREASE LIQUIDITY, COLLECT FEE FIRST
	_, fee0Str, fee1Str, _, _, _ := collectFee(cross, params.positionId, params.unwrapResult, params.caller)
	fee0 := u256.MustFromDecimal(fee0Str)
	fee1 := u256.MustFromDecimal(fee1Str)

	position := MustGetPosition(params.positionId)
	positionLiquidity := position.liquidity
	if positionLiquidity.IsZero() {
		panic(newErrorWithDetail(
			errZeroLiquidity,
			ufmt.Sprintf("position(position ID:%d) has 0 liquidity", params.positionId),
		))
	}

	liquidityToRemove := u256.MustFromDecimal(params.liquidity)
	if liquidityToRemove.Gt(positionLiquidity) {
		panic(newErrorWithDetail(
			errInvalidLiquidity,
			ufmt.Sprintf("Liquidity requested(%s) is greater than liquidity held(%s)", liquidityToRemove.ToString(), positionLiquidity.ToString()),
		))
	}

	beforeWugnotBalance := wugnot.BalanceOf(caller) // before unwrap

	pToken0, pToken1, pFee := splitOf(position.poolKey)
	burn0, burn1 := pl.Burn(cross, pToken0, pToken1, pFee, position.tickLower, position.tickUpper, liquidityToRemove.ToString(), caller)

	burnedAmount0 := u256.MustFromDecimal(burn0)
	burnedAmount1 := u256.MustFromDecimal(burn1)
	verifySlippageAmounts(burnedAmount0, burnedAmount1, params.amount0Min, params.amount1Min)

	positionKey := computePositionKey(caller, position.tickLower, position.tickUpper)
	pool := pl.GetPoolFromPoolPath(position.poolKey)
	feeGrowthInside0LastX128 := new(u256.Uint).Set(pool.PositionFeeGrowthInside0LastX128(positionKey))
	feeGrowthInside1LastX128 := new(u256.Uint).Set(pool.PositionFeeGrowthInside1LastX128(positionKey))

	currentSqrtPriceX96 := pool.Slot0SqrtPriceX96()
	lowerTickSqrtPriceX96 := common.TickMathGetSqrtRatioAtTick(position.tickLower)
	upperTickSqrtPriceX96 := common.TickMathGetSqrtRatioAtTick(position.tickUpper)
	calculatedToken0BalanceStr, calculatedToken1BalanceStr := common.GetAmountsForLiquidity(
		currentSqrtPriceX96,
		lowerTickSqrtPriceX96,
		upperTickSqrtPriceX96,
		position.liquidity,
	)
	calculatedToken0Balance := u256.MustFromDecimal(calculatedToken0BalanceStr)
	calculatedToken1Balance := u256.MustFromDecimal(calculatedToken1BalanceStr)

	position.tokensOwed0 = updateTokensOwed(
		feeGrowthInside0LastX128,
		position.feeGrowthInside0LastX128,
		position.liquidity,
		burnedAmount0,
		position.tokensOwed0,
	)

	position.tokensOwed1 = updateTokensOwed(
		feeGrowthInside1LastX128,
		position.feeGrowthInside1LastX128,
		position.liquidity,
		burnedAmount1,
		position.tokensOwed1,
	)

	position.feeGrowthInside0LastX128 = feeGrowthInside0LastX128
	position.feeGrowthInside1LastX128 = feeGrowthInside1LastX128
	position.liquidity = new(u256.Uint).Sub(positionLiquidity, liquidityToRemove)
	position.token0Balance = new(u256.Uint).Sub(calculatedToken0Balance, burnedAmount0)
	position.token1Balance = new(u256.Uint).Sub(calculatedToken1Balance, burnedAmount1)
	mustUpdatePosition(params.positionId, position)

	collect0, collect1 := pl.Collect(
		cross,
		pToken0,
		pToken1,
		pFee,
		caller,
		position.tickLower,
		position.tickUpper,
		burn0,
		burn1,
	)

	collectAmount0 := u256.MustFromDecimal(collect0)
	collectAmount1 := u256.MustFromDecimal(collect1)

	underflow := false
	position.tokensOwed0, underflow = new(u256.Uint).SubOverflow(position.tokensOwed0, collectAmount0)
	if underflow {
		panic(newErrorWithDetail(
			errUnderflow,
			"tokensOwed0 underflow",
		))
	}
	position.tokensOwed1, underflow = new(u256.Uint).SubOverflow(position.tokensOwed1, collectAmount1)
	if underflow {
		panic(newErrorWithDetail(
			errUnderflow,
			"tokensOwed1 underflow",
		))
	}

	if position.isClear() {
		position.burned = true // just update flag (we don't want to burn actual position)
	}
	mustUpdatePosition(params.positionId, position)

	// NO UNWRAP
	if !params.unwrapResult {
		return params.positionId, liquidityToRemove, fee0, fee1, collectAmount0, collectAmount1, position.poolKey
	}

	handleUnwrap(pToken0, pToken1, params.unwrapResult, beforeWugnotBalance, caller)

	return params.positionId, liquidityToRemove, fee0, fee1, collectAmount0, collectAmount1, position.poolKey
}

func updateTokensOwed(
	feeGrowthInsideLastX128 *u256.Uint,
	positionFeeGrowthInsideLastX128 *u256.Uint,
	positionLiquidity *u256.Uint,
	burnedAmount *u256.Uint,
	tokensOwed *u256.Uint,
) *u256.Uint {
	additionalTokensOwed := calculateTokensOwed(feeGrowthInsideLastX128, positionFeeGrowthInsideLastX128, positionLiquidity)
	add := new(u256.Uint).Add(burnedAmount, additionalTokensOwed)
	return new(u256.Uint).Add(tokensOwed, add)
}

// calculateFees calculates the fees for the current position.
func calculateFees(position Position, currentFeeGrowth FeeGrowthInside) (*u256.Uint, *u256.Uint) {
	fee0 := calculateTokensOwed(
		currentFeeGrowth.feeGrowthInside0LastX128,
		position.feeGrowthInside0LastX128,
		position.liquidity,
	)

	fee1 := calculateTokensOwed(
		currentFeeGrowth.feeGrowthInside1LastX128,
		position.feeGrowthInside1LastX128,
		position.liquidity,
	)

	tokensOwed0 := new(u256.Uint).Add(new(u256.Uint).Set(position.tokensOwed0), fee0)
	tokensOwed1 := new(u256.Uint).Add(new(u256.Uint).Set(position.tokensOwed1), fee1)

	return tokensOwed0, tokensOwed1
}

func calculateTokensOwed(
	feeGrowthInsideLastX128 *u256.Uint,
	positionFeeGrowthInsideLastX128 *u256.Uint,
	positionLiquidity *u256.Uint,
) *u256.Uint {
	diff := new(u256.Uint).Sub(feeGrowthInsideLastX128, positionFeeGrowthInsideLastX128)
	return u256.MulDiv(diff, positionLiquidity, u256.MustFromDecimal(consts.Q128))
}
