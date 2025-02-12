package math

import (
	"math"
	
	"gno.land/p/gnolend/consts"
	"gno.land/p/gnolend/types"
)

// CalculateLinearInterest calculates the interest using a linear model
// rate is in basis points (1 basis point = 0.01%)
// timeDelta is in seconds
func CalculateLinearInterest(rate uint64, timeDelta int64) uint64 {
	interest := uint64(float64(rate) * float64(timeDelta) / float64(consts.BLOCKS_PER_YEAR))
	return interest
}

// CalculateCompoundedInterest calculates the interest using a compound model
// rate is in basis points
// timeDelta is in seconds
func CalculateCompoundedInterest(rate uint64, timeDelta int64) uint64 {
	ratePerSecond := float64(rate) / float64(consts.BLOCKS_PER_YEAR)
	compoundedRate := math.Pow(1+ratePerSecond, float64(timeDelta))
	return uint64(compoundedRate * consts.PRECISION)
}

// CalculateUtilizationRate calculates the utilization rate of a pool
func CalculateUtilizationRate(totalBorrow, totalSupply uint64) uint64 {
	if totalSupply == 0 {
		return 0
	}
	return (totalBorrow * 10000) / totalSupply // Returns basis points (0.01%)
}

// CalculateBorrowRate calculates the borrow interest rate based on utilization
func CalculateBorrowRate(params types.InterestRateParams, utilizationRate uint64) uint64 {
	if utilizationRate <= params.OptimalUtilization {
		// Rate increases linearly from base rate to optimal rate
		rateIncrease := (utilizationRate * params.Slope1) / params.OptimalUtilization
		return params.BaseRate + rateIncrease
	} else {
		// Rate increases more steeply above optimal utilization
		excessUtilization := utilizationRate - params.OptimalUtilization
		maxExcessUtilization := 10000 - params.OptimalUtilization // 10000 = 100%
		rateIncrease := (excessUtilization * params.Slope2) / maxExcessUtilization
		optimalRate := params.BaseRate + params.Slope1
		return optimalRate + rateIncrease
	}
}

// CalculateSupplyRate calculates the supply interest rate from the borrow rate
func CalculateSupplyRate(borrowRate, utilizationRate uint64) uint64 {
	// Supply rate = Borrow rate * Utilization rate * (1 - Protocol fee)
	supplyRate := (borrowRate * utilizationRate * (10000 - uint64(consts.PROTOCOL_FEE * 10000))) / (10000 * 10000)
	return supplyRate
}

// CalculateHealthFactor calculates a position's health factor
func CalculateHealthFactor(collateralValue, borrowValue, liquidationThreshold uint64) uint64 {
	if borrowValue == 0 {
		return math.MaxUint64 // Position has no debt
	}
	return (collateralValue * liquidationThreshold) / (borrowValue * 10000)
}

// CalculateLiquidationPrice calculates the price at which liquidation would occur
func CalculateLiquidationPrice(borrowValue, collateralAmount, liquidationThreshold uint64) uint64 {
	return (borrowValue * 10000) / (collateralAmount * liquidationThreshold)
}

// CalculateAvailableBorrowingPower calculates how much more a user can borrow
func CalculateAvailableBorrowingPower(collateralValue, borrowValue, ltv uint64) uint64 {
	maxBorrow := (collateralValue * ltv) / 10000
	if maxBorrow <= borrowValue {
		return 0
	}
	return maxBorrow - borrowValue
} 