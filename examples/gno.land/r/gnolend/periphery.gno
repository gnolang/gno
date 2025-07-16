package gnolend

import (
	"gno.land/p/demo/avl"
	u256 "gno.land/p/gnoswap/uint256"
)

// ExpectedMarketBalances returns the expected market balances after interest accrual
func ExpectedMarketBalances(cur realm, marketId string) (*u256.Uint, *u256.Uint, *u256.Uint, *u256.Uint) {
	// Call the existing AccrueInterest function to perform interest accrual
	AccrueInterest(cur, marketId)

	// Get the updated market state
	market, _ := GetMarket(cur, marketId)

	// Return the current market balances
	return market.TotalSupplyAssets, market.TotalSupplyShares,
		market.TotalBorrowAssets, market.TotalBorrowShares
}

// ExpectedTotalSupplyAssets returns the expected total supply assets after interest accrual
func ExpectedTotalSupplyAssets(cur realm, marketId string) *u256.Uint {
	// Call AccrueInterest first and get the updated value
	AccrueInterest(cur, marketId)
	market, _ := GetMarket(cur, marketId)
	return market.TotalSupplyAssets
}

// ExpectedTotalBorrowAssets returns the expected total borrow assets after interest accrual
func ExpectedTotalBorrowAssets(cur realm, marketId string) *u256.Uint {
	// Call AccrueInterest first and get the updated value
	AccrueInterest(cur, marketId)
	market, _ := GetMarket(cur, marketId)
	return market.TotalBorrowAssets
}

// ExpectedTotalSupplyShares returns the expected total supply shares after interest accrual
func ExpectedTotalSupplyShares(cur realm, marketId string) *u256.Uint {
	// Call AccrueInterest first and get the updated value
	AccrueInterest(cur, marketId)
	market, _ := GetMarket(cur, marketId)
	return market.TotalSupplyShares
}

// ExpectedSupplyAssets returns the expected supply assets balance of a user after interest accrual
// Warning: This may be slightly inaccurate for fee recipients due to share increases
func ExpectedSupplyAssets(cur realm, marketId string, user string) *u256.Uint {
	// Call AccrueInterest first
	AccrueInterest(cur, marketId)

	// Get updated position and market data
	position := GetPosition(cur, marketId, user)
	market, _ := GetMarket(cur, marketId)

	return ToAssetsDown(position.SupplyShares, market.TotalSupplyAssets, market.TotalSupplyShares)
}

// ExpectedBorrowAssets returns the expected borrow assets balance of a user after interest accrual
// Warning: The expected balance is rounded up, so it may be greater than the market's expected total borrow assets
func ExpectedBorrowAssets(cur realm, marketId string, user string) *u256.Uint {
	// Call AccrueInterest first
	AccrueInterest(cur, marketId)

	// Get updated position and market data
	position := GetPosition(cur, marketId, user)
	market, _ := GetMarket(cur, marketId)

	return ToAssetsUp(position.BorrowShares, market.TotalBorrowAssets, market.TotalBorrowShares)
}

// CalculateBorrowRate returns the current borrow rate per second
// The returned value is WAD-scaled (1e18)
func CalculateBorrowRate(cur realm, marketId string) *u256.Uint {
	// Check if market exists
	if _, exists := markets.Get(marketId); !exists {
		panic(ErrMarketNotCreated)
	}

	market, params := GetMarket(cur, marketId)

	// If no borrows or no IRM, rate is zero
	if market.TotalBorrowAssets.IsZero() || params.IRM == "" {
		return u256.Zero()
	}

	// Get IRM and calculate current borrow rate
	irm := GetIRM(cur, params.IRM)
	return irm.BorrowRate(market.TotalSupplyAssets, market.TotalBorrowAssets)
}

// CalculateSupplyRate returns the current supply rate per second
// The returned value is WAD-scaled (1e18)
func CalculateSupplyRate(cur realm, marketId string) *u256.Uint {
	// Check if market exists
	if _, exists := markets.Get(marketId); !exists {
		panic(ErrMarketNotCreated)
	}

	market, params := GetMarket(cur, marketId)

	// If no borrows or no supply or no IRM, rate is zero
	if market.TotalBorrowAssets.IsZero() || market.TotalSupplyAssets.IsZero() || params.IRM == "" {
		return u256.Zero()
	}

	// Get borrow rate
	irm := GetIRM(cur, params.IRM)
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
func CalculateBorrowAPR(cur realm, marketId string) *u256.Uint {
	// Get the per-second borrow rate
	borrowRatePerSecond := CalculateBorrowRate(cur, marketId)
	if borrowRatePerSecond.IsZero() {
		return u256.Zero()
	}

	// Multiply by seconds per year to get APR
	secondsPerYear := u256.NewUint(365 * 24 * 60 * 60)
	return new(u256.Uint).Mul(borrowRatePerSecond, secondsPerYear)
}

// CalculateSupplyAPR returns the current supply APR (scaled by WAD)
// This converts the per-second rate to an annual rate
func CalculateSupplyAPR(cur realm, marketId string) *u256.Uint {
	// Get the per-second supply rate
	supplyRatePerSecond := CalculateSupplyRate(cur, marketId)
	if supplyRatePerSecond.IsZero() {
		return u256.Zero()
	}

	// Multiply by seconds per year to get APR
	secondsPerYear := u256.NewUint(365 * 24 * 60 * 60)
	return new(u256.Uint).Mul(supplyRatePerSecond, secondsPerYear)
}

// CalculateLoanAmount returns the current loan amount for a user in a market
// This is the amount of assets the user has borrowed, including accrued interest
func CalculateLoanAmount(cur realm, marketId string, user string) *u256.Uint {
	// Call AccrueInterest first to ensure we have the latest state
	AccrueInterest(cur, marketId)

	// Get updated position and market data
	position := GetPosition(cur, marketId, user)
	market, _ := GetMarket(cur, marketId)

	// Calculate current borrowed value using ToAssetsUp for safety
	return ToAssetsUp(
		position.BorrowShares,
		market.TotalBorrowAssets,
		market.TotalBorrowShares,
	)
}

// CalculateAllLoans returns a tree of all loans for a user
// The tree maps token addresses to total borrowed amounts
// Loans in different markets but with the same token are summed up
func CalculateUserLoans(cur realm, user string) *avl.Tree {
	// Create a new tree to store token -> amount mapping
	loans := avl.NewTree()

	// Get all market IDs
	marketIds := GetMarketList(cur)

	// Iterate through all markets using a regular for loop
	for _, marketId := range marketIds {
		// Calculate loan amount for this market
		loanAmount := CalculateLoanAmount(cur, marketId, user)

		// If user has a loan in this market
		if !loanAmount.IsZero() {
			// Get the borrow token from market parameters
			_, params := GetMarket(cur, marketId)
			borrowToken := params.GetLoanToken()

			// Get existing amount for this token if any
			existingAmount, exists := loans.Get(borrowToken)
			if exists {
				// Add to existing amount
				totalAmount := new(u256.Uint).Add(existingAmount.(*u256.Uint), loanAmount)
				loans.Set(borrowToken, totalAmount)
			} else {
				// Set new amount
				loans.Set(borrowToken, loanAmount)
			}
		}
	}

	return loans
}

// CalculateHealthFactor returns the health factor for a user's position in a market
// - A health factor > 1 means the position is healthy
// - A health factor < 1 means the position is eligible for liquidation
// - The lower the health factor, the higher the risk of liquidation
func CalculateHealthFactor(cur realm, marketId string, userAddr string) *u256.Uint {
	position := GetPosition(cur, marketId, userAddr)

	// If no borrows, position is maximally healthy (return a very large number)
	if position.BorrowShares.IsZero() {
		return new(u256.Uint).Mul(WAD, u256.NewUint(100)) // 100 * WAD
	}

	market, params := GetMarket(cur, marketId)

	// Calculate current borrowed value
	borrowed := ToAssetsUp(
		position.BorrowShares,
		market.TotalBorrowAssets,
		market.TotalBorrowShares,
	)

	// If no collateral, health factor is 0
	if position.Collateral.IsZero() {
		return u256.Zero()
	}

	collateralPrice, err := GetPrice(params.PoolPath)
	if err != nil {
		panic(err)
	}

	// Calculate max borrow allowed
	maxBorrow := WMulDown(MulDivDown(position.Collateral, collateralPrice, ORACLE_PRICE_SCALE), params.LLTV)

	// Health factor = maxBorrow / borrowed (WAD scaled)
	if borrowed.IsZero() {
		return new(u256.Uint).Mul(WAD, u256.NewUint(100)) // 100 * WAD
	}

	return WDivDown(maxBorrow, borrowed)
}
