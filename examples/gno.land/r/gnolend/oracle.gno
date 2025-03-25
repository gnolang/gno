package gnolend

import (
	u256 "gno.land/p/demo/uint256"
	"gno.land/r/gnoswap/v1/pool"
)

// Q96 represents 2^96
var Q96 = u256.MustFromDecimal("79228162514264337593543950336") // 2^96

// Q192 represents 2^192
var Q192 = u256.MustFromDecimal("6277101735386680763835789423207666416102355444464034512896") // 2^192

var ORACLE_PRICE_SCALE = u256.MustFromDecimal("1000000000000000000000000000000000000") // 1e36

// GetPrice returns the price from a Gnoswap pool
// The price is returned as a WAD (18 decimals) in terms of loan token per collateral token
func GetPrice(poolPath string) (*u256.Uint, error) {
	// Get market params to determine token ordering
	_, params := GetMarket(poolPath)

	// Get the sqrt price from the pool
	sqrtPriceX96Str := pool.PoolGetSlot0SqrtPriceX96(poolPath)
	if sqrtPriceX96Str == "" {
		return nil, ErrPriceNotAvailable
	}
	println("sqrtPriceX96Str:", sqrtPriceX96Str)

	// Convert string to uint256
	sqrtPriceX96 := u256.MustFromDecimal(sqrtPriceX96Str)
	println("sqrtPriceX96:", sqrtPriceX96.String())

	// First multiply sqrtPriceX96 by WAD to maintain precision
	sqrtPriceX96WAD := new(u256.Uint).Mul(sqrtPriceX96, WAD)
	println("sqrtPriceX96WAD:", sqrtPriceX96WAD.String())

	// Square the price to get the actual price in Q192
	priceQ192 := new(u256.Uint).Mul(sqrtPriceX96WAD, sqrtPriceX96WAD)
	println("priceQ192:", priceQ192.String())

	// Finally divide by Q192 to get the actual price ratio with WAD precision
	price := new(u256.Uint).Div(priceQ192, Q192)
	println("final price:", price.String())

	// If token0 is not the loan token, we need to invert the price
	// because Gnoswap's price is always token1/token0
	if !params.IsToken0Loan {
		// Invert price: WAD² / price
		price = new(u256.Uint).Div(new(u256.Uint).Mul(WAD, WAD), price)
	}

	return price, nil
}