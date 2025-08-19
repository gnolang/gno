package core

import (
	u256 "gno.land/p/gnoswap/uint256"
	"gno.land/p/volos/consts"
	"gno.land/p/volos/math"
	"gno.land/r/gnoswap/v1/pool"
)

// GetPrice returns the price from a Gnoswap pool
// The price is returned as a sqrtPriceX96e36 (36 decimals) in terms of loan token per collateral token
func GetPrice(marketId string) *u256.Uint {
	// Get market params to determine token ordering
	_, params := GetMarket(marketId)

	// Get the sqrt price from the pool
	sqrtPriceX96Str := pool.PoolGetSlot0SqrtPriceX96(params.PoolPath)
	if sqrtPriceX96Str == "" {
		panic(ErrPriceNotAvailable)
	}

	// Convert string to uint256
	sqrtPriceX96 := u256.MustFromDecimal(sqrtPriceX96Str)

	// Square the price to get the actual price in Q192
	priceQ192 := new(u256.Uint).Mul(sqrtPriceX96, sqrtPriceX96)

	// Finally divide by Q192 to get the actual price ratio with ORACLE_PRICE_SCALE precision
	price := math.MulDivDown(priceQ192, consts.ORACLE_PRICE_SCALE, consts.Q192)

	// If token0 is the loan token, we need to invert the price
	// because Gnoswap's price is always token1/token0
	if params.IsToken0Loan {
		// Invert price: ORACLE_PRICE_SCALE² / price
		price = math.MulDivDown(consts.ORACLE_PRICE_SCALE, consts.ORACLE_PRICE_SCALE, price)
	}

	return price
}
