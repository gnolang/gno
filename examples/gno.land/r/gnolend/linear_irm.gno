package irm

import (
	"gno.land/p/demo/uint256"
)

// LinearIRM is a simple interest rate model where utilization rate directly corresponds to APR
// For example: 80% utilization = 80% APR
type LinearIRM struct{}

// BorrowRate returns the borrow rate per second scaled by WAD
// For example: 80% utilization = 80% APR = (0.8 * WAD) / (365 days)
func (irm *LinearIRM) BorrowRate(market Market, _ MarketParams) *uint256.Uint {
	if market.TotalSupplyAssets.IsZero() {
		return uint256.NewInt(0)
	}

	// Calculate utilization rate scaled by WAD
	// utilization = (totalBorrow * WAD) / totalSupply
	utilization := WDivDown(market.TotalBorrowAssets, market.TotalSupplyAssets)

	// Convert APR to per-second rate
	// APR = utilization (e.g. 80% utilization = 80% APR)
	// Per second rate = APR / seconds_per_year
	secondsPerYear := uint256.NewInt(365 * 24 * 60 * 60)
	return WDivDown(utilization, secondsPerYear)
}