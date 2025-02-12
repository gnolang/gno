package pool

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
//	updatedPosition := pool.updatePosition(positionParams, currentTick)
//	fmt.Println("Updated Position Info:", updatedPosition)
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
