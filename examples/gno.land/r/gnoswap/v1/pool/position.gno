package pool

import (
	"encoding/base64"
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"
	plp "gno.land/p/gnoswap/gnsmath"
	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/gnoswap/v1/common"
)

var Q128 = u256.MustFromDecimal(consts.Q128)

// getPositionKey generates a unique, encoded key for a liquidity position.
//
// This function creates a unique key for identifying a liquidity position in a pool. The key is based
// on the position's owner address, lower tick, and upper tick values. The generated key is then encoded
// as a base64 string to ensure compatibility and uniqueness.
//
// Parameters:
//   - owner: std.Address, the address of the position's owner.
//   - tickLower: int32, the lower tick boundary for the position.
//   - tickUpper: int32, the upper tick boundary for the position.
//
// Returns:
//   - string: A base64-encoded string representing the unique position key.
//
// Workflow:
//  1. Validates that the `owner` address is valid using `assertOnlyValidAddress`.
//  2. Ensures `tickLower` is less than `tickUpper` using `assertTickLowerLessThanUpper`.
//  3. Constructs the position key as a formatted string:
//     "<owner>__<tickLower>__<tickUpper>"
//  4. Encodes the generated position key into a base64 string for safety and uniqueness.
//  5. Returns the encoded position key.
//
// Example:
//
// ```gno
//
//	owner := std.Address("0x123456789")
//	positionKey := getPositionKey(owner, 100, 200)
//	println("Position Key:", positionKey)
//
//	// Output: base64-encoded string representing "0x123456789__100__200"
//
// ```
//
// Notes:
//   - The base64 encoding ensures that the position key can be safely used as an identifier
//     across different systems or data stores.
//   - The function will panic if:
//   - The `owner` address is invalid.
//   - `tickLower` is greater than or equal to `tickUpper`.
func getPositionKey(
	owner std.Address,
	tickLower int32,
	tickUpper int32,
) string {
	assertOnlyValidAddress(owner)
	assertTickLowerLessThanUpper(tickLower, tickUpper)

	positionKey := ufmt.Sprintf("%s__%d__%d", owner.String(), tickLower, tickUpper)
	encodedPositionKey := base64.StdEncoding.EncodeToString([]byte(positionKey))
	return encodedPositionKey
}

// positionUpdate calculates and returns an updated PositionInfo.
func positionUpdate(
	position PositionInfo,
	liquidityDelta *i256.Int,
	feeGrowthInside0X128 *u256.Uint,
	feeGrowthInside1X128 *u256.Uint,
) PositionInfo {
	position.valueOrZero()

	var liquidityNext *u256.Uint
	if liquidityDelta.IsZero() {
		if position.liquidity.IsZero() {
			panic(newErrorWithDetail(
				errZeroLiquidity,
				"both liquidityDelta and current position's liquidity are zero",
			))
		}

		liquidityNext = position.liquidity
	} else {
		liquidityNext = liquidityMathAddDelta(position.liquidity, liquidityDelta)
	}

	tokensOwed0 := u256.Zero()
	diff0 := new(u256.Uint).Sub(feeGrowthInside0X128, position.feeGrowthInside0LastX128)
	tokensOwed0 = u256.MulDiv(diff0, position.liquidity, Q128)

	tokensOwed1 := u256.Zero()
	diff1 := new(u256.Uint).Sub(feeGrowthInside1X128, position.feeGrowthInside1LastX128)
	tokensOwed1 = u256.MulDiv(diff1, position.liquidity, Q128)

	if !(liquidityDelta.IsZero()) {
		position.liquidity = liquidityNext
	}

	position.feeGrowthInside0LastX128 = feeGrowthInside0X128
	position.feeGrowthInside1LastX128 = feeGrowthInside1X128
	if tokensOwed0.Gt(u256.Zero()) || tokensOwed1.Gt(u256.Zero()) {
		position.tokensOwed0 = position.tokensOwed0.Add(position.tokensOwed0, tokensOwed0)
		position.tokensOwed1 = position.tokensOwed1.Add(position.tokensOwed1, tokensOwed1)
	}

	return position
}

func calculateToken0Amount(sqrtPriceLower, sqrtPriceUpper *u256.Uint, liquidityDelta *i256.Int) *i256.Int {
	res := plp.GetAmount0DeltaStr(sqrtPriceLower, sqrtPriceUpper, liquidityDelta)
	return i256.MustFromDecimal(res)
}

func calculateToken1Amount(sqrtPriceLower, sqrtPriceUpper *u256.Uint, liquidityDelta *i256.Int) *i256.Int {
	res := plp.GetAmount1DeltaStr(sqrtPriceLower, sqrtPriceUpper, liquidityDelta)
	return i256.MustFromDecimal(res)
}

// PositionLiquidity returns the liquidity of a position.
func (p *Pool) PositionLiquidity(key string) *u256.Uint {
	return p.mustGetPosition(key).liquidity
}

// PositionFeeGrowthInside0LastX128 returns the fee growth of token0 inside a position.
func (p *Pool) PositionFeeGrowthInside0LastX128(key string) *u256.Uint {
	return p.mustGetPosition(key).feeGrowthInside0LastX128
}

// PositionFeeGrowthInside1LastX128 returns the fee growth of token1 inside a position.
func (p *Pool) PositionFeeGrowthInside1LastX128(key string) *u256.Uint {
	return p.mustGetPosition(key).feeGrowthInside1LastX128
}

// PositionTokensOwed0 returns the amount of token0 owed by a position.
func (p *Pool) PositionTokensOwed0(key string) *u256.Uint {
	return p.mustGetPosition(key).tokensOwed0
}

// PositionTokensOwed1 returns the amount of token1 owed by a position.
func (p *Pool) PositionTokensOwed1(key string) *u256.Uint {
	return p.mustGetPosition(key).tokensOwed1
}

// GetPosition returns the position info for a given key.
func (p *Pool) GetPosition(key string) (PositionInfo, bool) {
	iPositionInfo, exist := p.positions.Get(key)
	if !exist {
		newPosition := PositionInfo{}
		newPosition.valueOrZero()
		return newPosition, false
	}

	return iPositionInfo.(PositionInfo), true
}

// positionUpdateWithKey updates a position in the pool and returns the updated position.
func (p *Pool) positionUpdateWithKey(
	positionKey string,
	liquidityDelta *i256.Int,
	feeGrowthInside0X128 *u256.Uint,
	feeGrowthInside1X128 *u256.Uint,
) PositionInfo {
	// if pointer is nil, set to zero for calculation
	liquidityDelta = liquidityDelta.NilToZero()
	feeGrowthInside0X128 = feeGrowthInside0X128.NilToZero()
	feeGrowthInside1X128 = feeGrowthInside1X128.NilToZero()

	positionToUpdate, _ := p.GetPosition(positionKey)
	positionAfterUpdate := positionUpdate(positionToUpdate, liquidityDelta, feeGrowthInside0X128, feeGrowthInside1X128)

	p.setPosition(positionKey, positionAfterUpdate)

	return positionAfterUpdate
}

// setPosition sets the position info for a given key.
func (p *Pool) setPosition(posKey string, positionInfo PositionInfo) {
	p.positions.Set(posKey, positionInfo)
}

// mustGetPosition returns the position info for a given key.
func (p *Pool) mustGetPosition(positionKey string) PositionInfo {
	positionInfo, exist := p.GetPosition(positionKey)
	if !exist {
		panic(newErrorWithDetail(
			errDataNotFound,
			ufmt.Sprintf("positionKey(%s) does not exist", positionKey),
		))
	}
	return positionInfo
}

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

// updatePosition modifies the position's liquidity and updates the corresponding tick states.
//
// This function updates the position data based on the specified liquidity delta and tick range.
// It also manages the fee growth, tick state flipping, and cleanup of unused tick data.
//
// Parameters:
//   - positionParams: ModifyPositionParams, the parameters for the position modification, which include:
//   - owner: The address of the position owner.
//   - tickLower: The lower tick boundary of the position.
//   - tickUpper: The upper tick boundary of the position.
//   - liquidityDelta: The change in liquidity (positive or negative).
//   - tick: int32, the current tick position.
//
// Returns:
//   - PositionInfo: The updated position information.
//
// Workflow:
//  1. Clone the global fee growth values (token 0 and token 1).
//  2. If the liquidity delta is non-zero:
//     - Update the lower and upper ticks using `tickUpdate`, flipping their states if necessary.
//     - If a tick's state was flipped, update the tick bitmap to reflect the new state.
//  3. Calculate the fee growth inside the tick range using `getFeeGrowthInside`.
//  4. Generate a unique position key and update the position data using `positionUpdateWithKey`.
//  5. If liquidity is being removed (negative delta), clean up unused tick data by deleting the tick entries.
//  6. Return the updated position.
//
// Notes:
//   - The function flips the tick states and cleans up unused tick data when liquidity is removed.
//   - It ensures fee growth and position data remain accurate after the update.
//
// Example Usage:
//
// ```gno
//
//	updatedPosition := pool.updatePosition(positionParams, currentTick)
//	println("Updated Position Info:", updatedPosition)
//
// ```
func (p *Pool) updatePosition(positionParams ModifyPositionParams, tick int32) PositionInfo {
	feeGrowthGlobal0X128 := p.FeeGrowthGlobal0X128().Clone()
	feeGrowthGlobal1X128 := p.FeeGrowthGlobal1X128().Clone()

	var flippedLower, flippedUpper bool
	if !(positionParams.liquidityDelta.IsZero()) {
		flippedLower = p.tickUpdate(
			positionParams.tickLower,
			tick,
			positionParams.liquidityDelta,
			feeGrowthGlobal0X128,
			feeGrowthGlobal1X128,
			false,
			p.maxLiquidityPerTick,
		)

		flippedUpper = p.tickUpdate(
			positionParams.tickUpper,
			tick,
			positionParams.liquidityDelta,
			feeGrowthGlobal0X128,
			feeGrowthGlobal1X128,
			true,
			p.maxLiquidityPerTick,
		)

		if flippedLower {
			p.tickBitmapFlipTick(positionParams.tickLower, p.tickSpacing)
		}

		if flippedUpper {
			p.tickBitmapFlipTick(positionParams.tickUpper, p.tickSpacing)
		}
	}

	feeGrowthInside0X128, feeGrowthInside1X128 := p.getFeeGrowthInside(
		positionParams.tickLower,
		positionParams.tickUpper,
		tick,
		feeGrowthGlobal0X128,
		feeGrowthGlobal1X128,
	)

	positionKey := getPositionKey(positionParams.owner, positionParams.tickLower, positionParams.tickUpper)
	position := p.positionUpdateWithKey(
		positionKey,
		positionParams.liquidityDelta,
		feeGrowthInside0X128.Clone(),
		feeGrowthInside1X128.Clone(),
	)

	// clear any tick data that is no longer needed
	if positionParams.liquidityDelta.IsNeg() {
		if flippedLower {
			p.deleteTick(positionParams.tickLower)
		}
		if flippedUpper {
			p.deleteTick(positionParams.tickUpper)
		}
	}

	return position
}
