package gnoxchange

import (
	"math"
	"std"

	"gno.land/r/matijamarjanovic/tokenhub"
)

const (
	FEE_NUMERATOR   = 3 // 0.3%
	FEE_DENOMINATOR = 1000
)

// Swap executes a token swap in a given pool. Pool keys are made
// by concatenating the token keys with a semicolon (alphabetical order).
func Swap(poolKey string, tokenInKey string, amountIn uint64, minAmountOut uint64) (uint64, error) {
	poolInterface, exists := poolRegistry.pools.Get(poolKey)
	if !exists {
		return 0, ErrPoolNotFound
	}
	pool := poolInterface.(*Pool)

	if tokenInKey != pool.tokenA && tokenInKey != pool.tokenB {
		return 0, ErrInvalidInputToken
	}

	tokenIn := tokenhub.GetToken(tokenInKey)
	if tokenIn == nil {
		return 0, ErrTokenNotFound(tokenInKey)
	}

	var tokenOutKey string
	if tokenInKey == pool.tokenA {
		tokenOutKey = pool.tokenB
	} else {
		tokenOutKey = pool.tokenA
	}

	tokenOut := tokenhub.GetToken(tokenOutKey)
	if tokenOut == nil {
		return 0, ErrTokenNotFound(tokenOutKey)
	}

	decimalsIn := tokenIn.GetDecimals()
	decimalsOut := tokenOut.GetDecimals()

	// Ensure no token has more than 6 decimals
	if decimalsIn > 6 {
		return 0, ErrTokenDecimalsExceedRange(tokenInKey, decimalsIn)
	}
	if decimalsOut > 6 {
		return 0, ErrTokenDecimalsExceedRange(tokenOutKey, decimalsOut)
	}

	minDecimals := decimalsIn
	if decimalsOut < decimalsIn {
		minDecimals = decimalsOut
	}

	if decimalsIn > minDecimals {
		expectedScale := uint64(math.Pow10(int(decimalsIn - minDecimals)))
		if amountIn%expectedScale != 0 {
			return 0, ErrInvalidInputAmount(tokenInKey, expectedScale, decimalsIn, minDecimals)
		}
	}

	amountOut, err := calculateOutputAmount(pool, tokenInKey, amountIn)
	if err != nil {
		return 0, err
	}

	if amountOut < minAmountOut {
		return 0, ErrInsufficientOutputAmount(minAmountOut, amountOut)
	}

	caller := std.PreviousRealm().Address()

	allowance := tokenIn.Allowance(caller, std.CurrentRealm().Address())
	if allowance < amountIn {
		return 0, ErrInsufficientAllowance(tokenInKey, allowance, amountIn)
	}

	tellerIn := tokenIn.RealmTeller()
	if err := tellerIn.TransferFrom(
		caller,
		std.CurrentRealm().Address(),
		amountIn,
	); err != nil {
		return 0, ErrFailedTokenTransfer(tokenInKey, err)
	}

	tellerOut := tokenOut.RealmTeller()
	if err := tellerOut.Transfer(caller, amountOut); err != nil {
		if refundErr := tellerIn.Transfer(caller, amountIn); refundErr != nil {
			return 0, ErrFailedTokenTransfer(tokenInKey, refundErr)
		}
		return 0, ErrFailedTokenTransfer(tokenOutKey, err)
	}

	if tokenInKey == pool.tokenA {
		pool.reserveA += amountIn
		pool.reserveB -= amountOut
	} else {
		pool.reserveA -= amountOut
		pool.reserveB += amountIn
	}

	return amountOut, nil
}

// calculateOutputAmount calculates the output amount for a given input amount
// using the constant product formula (x * y = k) with a 0.3% fee
func calculateOutputAmount(pool *Pool, tokenInKey string, amountIn uint64) (uint64, error) {
	var reserveIn, reserveOut uint64

	if tokenInKey == pool.tokenA {
		reserveIn = pool.reserveA
		reserveOut = pool.reserveB
	} else {
		reserveIn = pool.reserveB
		reserveOut = pool.reserveA
	}

	if reserveIn == 0 || reserveOut == 0 {
		return 0, ErrInsufficientLiquidity
	}

	// calculate amount with fee (0.3%)
	amountInWithFee := amountIn * (FEE_DENOMINATOR - FEE_NUMERATOR)
	numerator := amountInWithFee * reserveOut
	denominator := (reserveIn * FEE_DENOMINATOR) + amountInWithFee

	amountOut := numerator / denominator

	if amountOut == 0 {
		return 0, ErrInsufficientOutputAmount(1, amountOut) // Using 1 as a placeholder for minimum expected
	}

	if amountOut >= reserveOut {
		return 0, ErrInsufficientLiquidity
	}

	return amountOut, nil
}

// GetSwapEstimate returns the estimated output amount for a given input amount
// This is useful for UI to show the expected output before executing the swap
func GetSwapEstimate(poolKey string, tokenInKey string, amountIn uint64) (uint64, error) {
	poolInterface, exists := poolRegistry.pools.Get(poolKey)
	if !exists {
		return 0, ErrPoolNotFound
	}
	pool := poolInterface.(*Pool)

	if tokenInKey != pool.tokenA && tokenInKey != pool.tokenB {
		return 0, ErrInvalidInputToken
	}

	return calculateOutputAmount(pool, tokenInKey, amountIn)
}
