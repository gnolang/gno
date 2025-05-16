package gnolend

import (
	"std"
	u256 "gno.land/p/gnoswap/uint256"
)

// ExpectedMarketBalances returns the expected market balances after interest accrual
func ExpectedMarketBalances(marketId string) (*u256.Uint, *u256.Uint, *u256.Uint, *u256.Uint) {
	// Call the existing AccrueInterest function to perform interest accrual
	AccrueInterest(marketId)
	
	// Get the updated market state
	market, _ := GetMarket(marketId)
	
	// Return the current market balances
	return market.TotalSupplyAssets, market.TotalSupplyShares,
	       market.TotalBorrowAssets, market.TotalBorrowShares
}

// ExpectedTotalSupplyAssets returns the expected total supply assets after interest accrual
func ExpectedTotalSupplyAssets(marketId string) *u256.Uint {
	// Call AccrueInterest first and get the updated value
	AccrueInterest(marketId)
	market, _ := GetMarket(marketId)
	return market.TotalSupplyAssets
}

// ExpectedTotalBorrowAssets returns the expected total borrow assets after interest accrual
func ExpectedTotalBorrowAssets(marketId string) *u256.Uint {
	// Call AccrueInterest first and get the updated value
	AccrueInterest(marketId)
	market, _ := GetMarket(marketId)
	return market.TotalBorrowAssets
}

// ExpectedTotalSupplyShares returns the expected total supply shares after interest accrual
func ExpectedTotalSupplyShares(marketId string) *u256.Uint {
	// Call AccrueInterest first and get the updated value
	AccrueInterest(marketId)
	market, _ := GetMarket(marketId)
	return market.TotalSupplyShares
}

// ExpectedSupplyAssets returns the expected supply assets balance of a user after interest accrual
// Warning: This may be slightly inaccurate for fee recipients due to share increases
func ExpectedSupplyAssets(marketId string, user string) *u256.Uint {
	// Call AccrueInterest first
	AccrueInterest(marketId)
	
	// Get updated position and market data
	position := GetPosition(marketId, user)
	market, _ := GetMarket(marketId)
	
	return ToAssetsDown(position.SupplyShares, market.TotalSupplyAssets, market.TotalSupplyShares)
}

// ExpectedBorrowAssets returns the expected borrow assets balance of a user after interest accrual
// Warning: The expected balance is rounded up, so it may be greater than the market's expected total borrow assets
func ExpectedBorrowAssets(marketId string, user string) *u256.Uint {
	// Call AccrueInterest first
	AccrueInterest(marketId)
	
	// Get updated position and market data
	position := GetPosition(marketId, user)
	market, _ := GetMarket(marketId)
	
	return ToAssetsUp(position.BorrowShares, market.TotalBorrowAssets, market.TotalBorrowShares)
}

// CalculateBorrowRate returns the current borrow rate per second
// The returned value is WAD-scaled (1e18)
func CalculateBorrowRate(marketId string) *u256.Uint {
	// Check if market exists
	if _, exists := markets.Get(marketId); !exists {
		panic(ErrMarketNotCreated)
	}

	market, params := GetMarket(marketId)

	// If no borrows or no IRM, rate is zero
	if market.TotalBorrowAssets.IsZero() || params.IRM == "" {
		return u256.Zero()
	}

	// Get IRM and calculate current borrow rate
	irm := GetIRM(params.IRM)
	return irm.BorrowRate(market.TotalSupplyAssets, market.TotalBorrowAssets)
}

// CalculateSupplyRate returns the current supply rate per second
// The returned value is WAD-scaled (1e18)
func CalculateSupplyRate(marketId string) *u256.Uint {
	// Check if market exists
	if _, exists := markets.Get(marketId); !exists {
		panic(ErrMarketNotCreated)
	}

	market, params := GetMarket(marketId)

	// If no borrows or no supply or no IRM, rate is zero
	if market.TotalBorrowAssets.IsZero() || market.TotalSupplyAssets.IsZero() || params.IRM == "" {
		return u256.Zero()
	}

	// Get borrow rate
	irm := GetIRM(params.IRM)
	borrowRate := irm.BorrowRate(market.TotalSupplyAssets, market.TotalBorrowAssets)
	
	// Calculate utilization rate: totalBorrow / totalSupply
	utilizationRate := WDivDown(market.TotalBorrowAssets, market.TotalSupplyAssets)
	
	// Calculate fee-adjusted borrow rate
	feeFactor := new(u256.Uint).Sub(WAD, market.Fee) // (1 - fee)
	
	// Supply rate = borrow rate * utilization rate * (1 - fee)
	// This formula accounts for:
	// 1. Interest is only earned on the portion of funds being borrowed (utilization)
	// 2. A portion of interest (the fee) goes to the fee recipient
	return WMulDown(WMulDown(borrowRate, utilizationRate), feeFactor)
}

// CalculateBorrowAPR returns the current borrow APR (scaled by WAD)
// This converts the per-second rate to an annual rate
func CalculateBorrowAPR(marketId string) *u256.Uint {
	// Get the per-second borrow rate
	borrowRatePerSecond := CalculateBorrowRate(marketId)
	if borrowRatePerSecond.IsZero() {
		return u256.Zero()
	}
	
	// Multiply by seconds per year to get APR
	secondsPerYear := u256.NewUint(365 * 24 * 60 * 60)
	return new(u256.Uint).Mul(borrowRatePerSecond, secondsPerYear)
}

// CalculateSupplyAPR returns the current supply APR (scaled by WAD)
// This converts the per-second rate to an annual rate
func CalculateSupplyAPR(marketId string) *u256.Uint {
	// Get the per-second supply rate
	supplyRatePerSecond := CalculateSupplyRate(marketId)
	if supplyRatePerSecond.IsZero() {
		return u256.Zero()
	}
	
	// Multiply by seconds per year to get APR
	secondsPerYear := u256.NewUint(365 * 24 * 60 * 60)
	return new(u256.Uint).Mul(supplyRatePerSecond, secondsPerYear)
} 