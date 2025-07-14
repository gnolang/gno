package position

import (
	"std"

	"gno.land/p/demo/ufmt"
	u256 "gno.land/p/gnoswap/uint256"

	pl "gno.land/r/gnoswap/v1/pool"
)

// Reposition adjusts the price range and liquidity of an existing position.
// Reposition is executed via multiple messages as follows:
//
// 1. Decrease the liquidity of the existing position.
// 2. Swap tokens as needed to reach the desired token ratio.
// 3. Add new liquidity with the desired price range and amount.
//
// Returns Position ID, liquidity, tickLower, tickUpper, amount0, amount1
// ref: https://docs.gnoswap.io/contracts/position/position.gno#reposition
func Reposition(
	cur realm,
	positionId uint64,
	tickLower int32,
	tickUpper int32,
	amount0DesiredStr string,
	amount1DesiredStr string,
	amount0MinStr string,
	amount1MinStr string,
) (uint64, string, int32, int32, string, string) {
	assertOnlyNotHalted()
	verifyTokenIdAndOwnership(positionId)
	mintAndDistributeGnsCallback()

	// position should be burned to reposition
	position := MustGetPosition(positionId)
	caller := getPrevAddr()
	oldTickLower := position.tickLower
	oldTickUpper := position.tickUpper

	if !(position.isClear()) {
		panic(newErrorWithDetail(
			errNotClear,
			ufmt.Sprintf(
				"position(%d) isn't clear(liquidity:%s, tokensOwed0:%s, tokensOwed1:%s)",
				positionId,
				position.liquidity.ToString(),
				position.tokensOwed0.ToString(),
				position.tokensOwed1.ToString(),
			),
		))
	}

	token0, token1, _ := splitOf(position.poolKey)
	token0, token1, _, _, _ = processTokens(token0, token1, amount0DesiredStr, amount1DesiredStr, caller)

	liquidity, amount0, amount1 := addLiquidity(
		AddLiquidityParams{
			poolKey:        position.poolKey,
			tickLower:      tickLower,
			tickUpper:      tickUpper,
			amount0Desired: u256.MustFromDecimal(amount0DesiredStr),
			amount1Desired: u256.MustFromDecimal(amount1DesiredStr),
			amount0Min:     u256.MustFromDecimal(amount0MinStr),
			amount1Min:     u256.MustFromDecimal(amount1MinStr),
			caller:         caller,
		},
	)

	// update position tickLower, tickUpper to new value
	// because getCurrentFeeGrowth() uses tickLower, tickUpper
	position.tickLower = tickLower
	position.tickUpper = tickUpper

	currentFeeGrowth, err := getCurrentFeeGrowth(position, caller)
	if err != nil {
		panic(newErrorWithDetail(err, "failed to get current fee growth"))
	}
	position.feeGrowthInside0LastX128 = currentFeeGrowth.feeGrowthInside0LastX128
	position.feeGrowthInside1LastX128 = currentFeeGrowth.feeGrowthInside1LastX128

	position.liquidity = liquidity
	// OBS: do not reset feeGrowthInside1LastX128 and feeGrowthInside1LastX128 to zero
	// if so, ( decrease 100% -> reposition )
	// > at this point, that position will have unclaimedFee which isn't intended
	position.tokensOwed0 = u256.Zero()
	position.tokensOwed1 = u256.Zero()
	position.burned = false
	mustUpdatePosition(positionId, position)

	poolSqrtPriceX96 := pl.PoolGetSlot0SqrtPriceX96(position.poolKey)
	token0Balance := pl.PoolGetBalanceToken0(position.poolKey)
	token1Balance := pl.PoolGetBalanceToken1(position.poolKey)
	prevAddr, prevPkgPath := getPrevAsString()

	std.Emit(
		"Reposition",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"lpPositionId", formatUint(positionId),
		"tickLower", formatInt(int64(tickLower)),
		"tickUpper", formatInt(int64(tickUpper)),
		"liquidity", liquidity.ToString(),
		"amount0", amount0.ToString(),
		"amount1", amount1.ToString(),
		"prevTickLower", formatInt(int64(oldTickLower)),
		"prevTickUpper", formatInt(int64(oldTickUpper)),
		"poolPath", position.poolKey,
		"sqrtPriceX96", poolSqrtPriceX96,
		"positionLiquidity", PositionGetPositionLiquidityStr(positionId),
		"token0Balance", token0Balance,
		"token1Balance", token1Balance,
	)

	return positionId, liquidity.ToString(), tickLower, tickUpper, amount0.ToString(), amount1.ToString()
}
