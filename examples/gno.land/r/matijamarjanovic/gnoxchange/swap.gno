package gnoxchange

import (
	"errors"
	"math"
	"std"

	"gno.land/p/demo/ufmt"
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
		return 0, errors.New("pool not found")
	}
	pool := poolInterface.(*Pool)

	if tokenInKey != pool.tokenA && tokenInKey != pool.tokenB {
		return 0, errors.New("invalid input token")
	}

	tokenIn := tokenhub.GetToken(tokenInKey)
	if tokenIn == nil {
		return 0, errors.New("input token not found in tokenhub")
	}

	var tokenOutKey string
	if tokenInKey == pool.tokenA {
		tokenOutKey = pool.tokenB
	} else {
		tokenOutKey = pool.tokenA
	}

	tokenOut := tokenhub.GetToken(tokenOutKey)
	if tokenOut == nil {
		return 0, errors.New("output token not found in tokenhub")
	}

	decimalsIn := tokenIn.GetDecimals()
	decimalsOut := tokenOut.GetDecimals()

	minDecimals := decimalsIn
	if decimalsOut < decimalsIn {
		minDecimals = decimalsOut
	}

	if decimalsIn > minDecimals {
		expectedScale := uint64(math.Pow10(int(decimalsIn - minDecimals)))
		if amountIn%expectedScale != 0 {
			return 0, errors.New(ufmt.Sprintf(
				"invalid input amount for %s: amount must be in units of %d (token has %d decimals vs %d decimals for the other token)",
				tokenInKey, expectedScale, decimalsIn, minDecimals,
			))
		}
	}

	amountOut, err := calculateOutputAmount(pool, tokenInKey, amountIn)
	if err != nil {
		return 0, err
	}

	if amountOut < minAmountOut {
		return 0, errors.New(ufmt.Sprintf("insufficient output amount: expected minimum %d, got %d", minAmountOut, amountOut))
	}

	caller := std.PreviousRealm().Address()

	allowance := tokenIn.Allowance(caller, std.CurrentRealm().Address())
	if allowance < amountIn {
		return 0, errors.New(ufmt.Sprintf("insufficient allowance for token %s: have %d, need %d", tokenInKey, allowance, amountIn))
	}

	tellerIn := tokenIn.RealmTeller()
	if err := tellerIn.TransferFrom(
		caller,
		std.CurrentRealm().Address(),
		amountIn,
	); err != nil {
		return 0, errors.New(ufmt.Sprintf("failed to transfer input token: %v", err))
	}

	tellerOut := tokenOut.RealmTeller()
	if err := tellerOut.Transfer(caller, amountOut); err != nil {
		if refundErr := tellerIn.Transfer(caller, amountIn); refundErr != nil {
			panic(ufmt.Sprintf("CRITICAL: failed to refund input token after failed output transfer: input error: %v, refund error: %v",
				err, refundErr))
		}
		return 0, errors.New(ufmt.Sprintf("failed to transfer output token: %v", err))
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
		return 0, errors.New("insufficient liquidity")
	}

	// calculate amount with fee (0.3%)
	amountInWithFee := amountIn * (FEE_DENOMINATOR - FEE_NUMERATOR)
	numerator := amountInWithFee * reserveOut
	denominator := (reserveIn * FEE_DENOMINATOR) + amountInWithFee

	amountOut := numerator / denominator

	if amountOut == 0 {
		return 0, errors.New("insufficient output amount")
	}

	if amountOut >= reserveOut {
		return 0, errors.New("insufficient liquidity")
	}

	return amountOut, nil
}

// GetSwapEstimate returns the estimated output amount for a given input amount
// This is useful for UI to show the expected output before executing the swap
func GetSwapEstimate(poolKey string, tokenInKey string, amountIn uint64) (uint64, error) {
	poolInterface, exists := poolRegistry.pools.Get(poolKey)
	if !exists {
		return 0, errors.New("pool not found")
	}
	pool := poolInterface.(*Pool)

	if tokenInKey != pool.tokenA && tokenInKey != pool.tokenB {
		return 0, errors.New("invalid input token")
	}

	return calculateOutputAmount(pool, tokenInKey, amountIn)
}
