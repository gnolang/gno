package position

import (
	"std"

	"gno.land/p/demo/ufmt"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"
	pl "gno.land/r/gnoswap/v1/pool"
)

type AddLiquidityParams struct {
	poolKey        string      // poolPath of the pool which has the position
	tickLower      int32       // lower end of the tick range for the position
	tickUpper      int32       // upper end of the tick range for the position
	amount0Desired *u256.Uint  // desired amount of token0 to be minted
	amount1Desired *u256.Uint  // desired amount of token1 to be minted
	amount0Min     *u256.Uint  // minimum amount of token0 to be minted
	amount1Min     *u256.Uint  // minimum amount of token1 to be minted
	caller         std.Address // address to call the function
}

// addLiquidity calculates the liquidity to be added to a pool and mints the corresponding tokens.
//
// This function interacts with the specified pool to add liquidity for a given price range, specified by
// `tickLower` and `tickUpper`, and desired token amounts. It ensures that the resulting token amounts meet
// minimum thresholds to prevent excessive slippage.
//
// Parameters:
// - params (AddLiquidityParams): Contains the following fields:
//   - poolKey: The unique identifier for the pool (string).
//   - tickLower: The lower tick boundary of the liquidity range (int32).
//   - tickUpper: The upper tick boundary of the liquidity range (int32).
//   - amount0Desired: The desired amount of token0 to provide as liquidity (*u256.Uint).
//   - amount1Desired: The desired amount of token1 to provide as liquidity (*u256.Uint).
//   - amount0Min: The minimum acceptable amount of token0 to prevent slippage (*u256.Uint).
//   - amount1Min: The minimum acceptable amount of token1 to prevent slippage (*u256.Uint).
//   - caller: The address of the entity adding liquidity (std.Address).
//
// Returns:
// - *u256.Uint: The calculated liquidity amount to be added.
// - *u256.Uint: The actual amount of token0 used.
// - *u256.Uint: The actual amount of token1 used.
//
// Behavior:
// 1. Retrieves the pool information and current square root price (`sqrtPriceX96`).
// 2. Calculates the square root ratios (`sqrtRatioAX96` and `sqrtRatioBX96`) for the given tick boundaries.
// 3. Computes the liquidity to be added based on desired token amounts and the square root ratios.
// 4. Mints liquidity tokens to the position contract and determines the actual amounts of token0 and token1 used.
// 5. Ensures the actual token amounts used meet or exceed the specified minimum thresholds (`amount0Min` and `amount1Min`).
//   - If the conditions are not met, the function panics with a slippage error.
//
// Panics:
// - If the actual token amounts used do not meet the minimum thresholds, a slippage error is raised.
//
// Notes:
//   - The function relies on the `GetLiquidityForAmounts` function to calculate liquidity based on token amounts
//     and price ratios.
//   - Ensures the pool interactions use the caller-provided parameters to add liquidity safely.
//
// Example:
//
//	liquidity, usedAmount0, usedAmount1 := addLiquidity(AddLiquidityParams{
//	  poolKey:       "gno.land/r/demo/pool",
//	  tickLower:     -60000,
//	  tickUpper:     60000,
//	  amount0Desired: u256.MustFromDecimal("1000000000"),
//	  amount1Desired: u256.MustFromDecimal("2000000000"),
//	  amount0Min:    u256.MustFromDecimal("950000000"),
//	  amount1Min:    u256.MustFromDecimal("1900000000"),
//	  caller:        userAddress,
//	})
func addLiquidity(params AddLiquidityParams) (*u256.Uint, *u256.Uint, *u256.Uint) {
	pool := pl.GetPoolFromPoolPath(params.poolKey)

	sqrtPriceX96 := new(u256.Uint).Set(pool.Slot0SqrtPriceX96())
	sqrtRatioAX96 := common.TickMathGetSqrtRatioAtTick(params.tickLower)
	sqrtRatioBX96 := common.TickMathGetSqrtRatioAtTick(params.tickUpper)

	liquidity := common.GetLiquidityForAmounts(
		sqrtPriceX96,
		sqrtRatioAX96,
		sqrtRatioBX96,
		params.amount0Desired,
		params.amount1Desired,
	)

	token0, token1, fee := splitOf(params.poolKey)
	amount0Str, amount1Str := pl.Mint(
		token0,
		token1,
		fee,
		consts.POSITION_ADDR,
		params.tickLower,
		params.tickUpper,
		liquidity.ToString(),
		params.caller,
	)

	amount0 := u256.MustFromDecimal(amount0Str)
	amount1 := u256.MustFromDecimal(amount1Str)

	amount0Cond := amount0.Gte(params.amount0Min)
	amount1Cond := amount1.Gte(params.amount1Min)

	if !(amount0Cond && amount1Cond) {
		panic(newErrorWithDetail(
			errSlippage,
			ufmt.Sprintf(
				"Price Slippage Check(amount0(%s) >= amount0Min(%s), amount1(%s) >= amount1Min(%s))",
				amount0Str, params.amount0Min.ToString(), amount1Str, params.amount1Min.ToString()),
		))
	}

	return liquidity, amount0, amount1
}
