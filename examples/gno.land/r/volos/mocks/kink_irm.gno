package mocks

import (
	volos "gno.land/r/volos/core"

	u256 "gno.land/p/gnoswap/uint256"
	"gno.land/p/volos/math"
	"gno.land/p/volos/consts"
)

// KinkIRM implements a kinked interest rate model similar to Aave's model
// It has two slopes: a gentle slope below optimal utilization and a steep slope above
type KinkIRM struct {
	// Interest rate parameters (all in WAD format - 18 decimals)
	optimalUtilizationRate *u256.Uint // 80% = 0.8 * WAD
	baseBorrowRate         *u256.Uint // 2% = 0.02 * WAD  
	slope1                 *u256.Uint // 4% = 0.04 * WAD
	slope2                 *u256.Uint // 75% = 0.75 * WAD
}

// NewKinkIRM creates a new KinkIRM with default parameters
func NewKinkIRM() *KinkIRM {
	return &KinkIRM{
		optimalUtilizationRate: u256.NewUint(800000000000000000), // 80% = 0.8 * WAD
		baseBorrowRate:         u256.NewUint(20000000000000000),  // 2% = 0.02 * WAD
		slope1:                 u256.NewUint(40000000000000000),  // 4% = 0.04 * WAD
		slope2:                 u256.NewUint(750000000000000000), // 75% = 0.75 * WAD
	}
}

// Name returns the human readable name of the IRM
func (irm *KinkIRM) Name() string {
	return "kink"
}

// BorrowRate returns the borrow rate per second scaled by WAD
// Uses kinked rate model:
// - Below optimal: gradual increase
// - Above optimal: steep increase to discourage over-borrowing
func (irm *KinkIRM) BorrowRate(totalSupplyAssets, totalBorrowAssets *u256.Uint) *u256.Uint {
	if totalSupplyAssets.IsZero() {
		return u256.Zero()
	}

	// Calculate utilization rate scaled by WAD
	// utilization = (totalBorrow * WAD) / totalSupply
	utilization := math.WDivDown(totalBorrowAssets, totalSupplyAssets)

	// Below optimal utilization
	if utilization.Cmp(irm.optimalUtilizationRate) <= 0 {
		// Rate = Base + (Utilization / Optimal) * Slope1
		utilizationRatio := math.WDivDown(utilization, irm.optimalUtilizationRate)
		rateIncrease := math.WMulDown(utilizationRatio, irm.slope1)
		borrowRate := new(u256.Uint).Add(irm.baseBorrowRate, rateIncrease)
		
		// Convert APR to per-second rate
		secondsPerYear := u256.NewUint(365 * 24 * 60 * 60)
		return new(u256.Uint).Div(borrowRate, secondsPerYear)
	}

	// Above optimal utilization - steep increase
	// Calculate excess utilization
	excessUtilization := new(u256.Uint).Sub(utilization, irm.optimalUtilizationRate)
	maxExcessUtilization := new(u256.Uint).Sub(consts.WAD, irm.optimalUtilizationRate)

	// Rate = Base + Slope1 + (ExcessUtilization / MaxExcess) * Slope2
	excessRatio := math.WDivDown(excessUtilization, maxExcessUtilization)
	excessRateIncrease := math.WMulDown(excessRatio, irm.slope2)
	
	borrowRate := new(u256.Uint).Add(irm.baseBorrowRate, irm.slope1)
	borrowRate.Add(borrowRate, excessRateIncrease)

	// Convert APR to per-second rate
	secondsPerYear := u256.NewUint(365 * 24 * 60 * 60)
	return new(u256.Uint).Div(borrowRate, secondsPerYear)
}

func init() {
	volos.RegisterIRM(cross, NewKinkIRM())
}
