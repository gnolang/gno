package mocks

import (
	volos "gno.land/r/volos/core"

	u256 "gno.land/p/gnoswap/uint256"
	"gno.land/p/volos/math"
)

// LinearIRM is a simple interest rate model where utilization rate directly corresponds to APR
// For example: 80% utilization = 80% APR
type LinearIRM struct{}

// Name returns the human readable name of the IRM
func (irm *LinearIRM) Name() string {
	return "linear"
}

// BorrowRate returns the borrow rate per second scaled by WAD
// For example: 80% utilization = 80% APR = (0.8 * WAD) / (365 days)
func (irm *LinearIRM) BorrowRate(totalSupplyAssets, totalBorrowAssets *u256.Uint) *u256.Uint {
	if totalSupplyAssets.IsZero() {
		return u256.Zero()
	}

	// Calculate utilization rate scaled by WAD
	// utilization = (totalBorrow * WAD) / totalSupply
	utilization := math.WDivDown(totalBorrowAssets, totalSupplyAssets)

	// Convert APR to per-second rate
	// APR = utilization (e.g. 80% utilization = 80% APR)
	// Per second rate = APR / seconds_per_year
	secondsPerYear := u256.NewUint(365 * 24 * 60 * 60)
	return new(u256.Uint).Div(utilization, secondsPerYear)
}

func init() {
	volos.RegisterIRM(cross, &LinearIRM{})
}
