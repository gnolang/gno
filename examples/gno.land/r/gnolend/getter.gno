package gnolend

import (
	"std"

	"gno.land/p/demo/avl"
	u256 "gno.land/p/gnoswap/uint256"
)

// Position getters
// Main struct getter
func GetPosition(cur realm, marketId string, userAddr string) Position {
	marketPositions, exists := positions.Get(marketId)
	if !exists {
		return Position{
			SupplyShares: new(u256.Uint),
			BorrowShares: new(u256.Uint),
			Collateral:   new(u256.Uint),
		} // Market doesn't exist, return empty position with initialized fields
	}

	position, exists := marketPositions.(*avl.Tree).Get(userAddr)
	if !exists {
		return Position{
			SupplyShares: new(u256.Uint),
			BorrowShares: new(u256.Uint),
			Collateral:   new(u256.Uint),
		} // Position doesn't exist, return empty position with initialized fields
	}
	return position.(Position)
}

// Individual field getters
func GetPositionSupplyShares(cur realm, marketId string, userAddr string) string {
	position := GetPosition(cur, marketId, userAddr)
	return position.SupplyShares.ToString()
}

func GetPositionBorrowShares(cur realm, marketId string, userAddr string) string {
	position := GetPosition(cur, marketId, userAddr)
	return position.BorrowShares.ToString()
}

func GetPositionCollateral(cur realm, marketId string, userAddr string) string {
	position := GetPosition(cur, marketId, userAddr)
	return position.Collateral.ToString()
}

// Market getters
// Main struct getter
func GetMarket(cur realm, marketId string) (Market, MarketParams) {
	market, exists := markets.Get(marketId)
	if !exists {
		panic(ErrMarketNotCreated)
	}

	params, exists := marketParams.Get(marketId)
	if !exists {
		panic(ErrMarketNotCreated)
	}

	return market.(Market), params.(MarketParams)
}

// Individual field getters
func GetMarketTotalSupplyAssets(cur realm, marketId string) string {
	market, _ := GetMarket(cur, marketId)
	return market.TotalSupplyAssets.ToString()
}

func GetMarketTotalSupplyShares(cur realm, marketId string) string {
	market, _ := GetMarket(cur, marketId)
	return market.TotalSupplyShares.ToString()
}

func GetMarketTotalBorrowAssets(cur realm, marketId string) string {
	market, _ := GetMarket(cur, marketId)
	return market.TotalBorrowAssets.ToString()
}

func GetMarketTotalBorrowShares(cur realm, marketId string) string {
	market, _ := GetMarket(cur, marketId)
	return market.TotalBorrowShares.ToString()
}

func GetMarketLastUpdate(cur realm, marketId string) int64 {
	market, _ := GetMarket(cur, marketId)
	return market.LastUpdate
}

func GetMarketFee(cur realm, marketId string) string {
	market, _ := GetMarket(cur, marketId)
	return market.Fee.ToString()
}

// MarketParams getters
func GetMarketParamsLoanToken(cur realm, marketId string) string {
	_, params := GetMarket(cur, marketId)
	return params.GetLoanToken()
}

func GetMarketParamsCollateralToken(cur realm, marketId string) string {
	_, params := GetMarket(cur, marketId)
	return params.GetCollateralToken()
}

func GetMarketParamsPoolPath(cur realm, marketId string) string {
	_, params := GetMarket(cur, marketId)
	return params.PoolPath
}

func GetMarketParamsIRM(cur realm, marketId string) string {
	_, params := GetMarket(cur, marketId)
	return params.IRM
}

func GetMarketParamsLLTV(cur realm, marketId string) string {
	_, params := GetMarket(cur, marketId)
	return params.LLTV.ToString()
}

// IRM getters
// Main getter
func GetIRM(cur realm, name string) IRM {
	irm, exists := irmRegistry.Get(name)
	if !exists {
		panic("IRM not registered: " + name)
	}
	return irm.(IRM)
}

// List getters
func GetMarketList(cur realm) []string {
	var marketList []string
	markets.Iterate("", "", func(key string, _ interface{}) bool {
		marketList = append(marketList, key)
		return false
	})
	return marketList
}

// GetUserMarkets returns a list of market IDs where a user has any position
// (supply, borrow, or collateral)
func GetUserMarkets(cur realm, userAddr string) []string {
	var userMarkets []string

	// Get list of all markets
	allMarkets := GetMarketList(cur)

	// Check each market to see if user has a position
	for _, marketId := range allMarkets {
		marketPositions, exists := positions.Get(marketId)
		if !exists {
			continue
		}

		position, exists := marketPositions.(*avl.Tree).Get(userAddr)
		if !exists {
			continue
		}

		pos := position.(Position)
		if !pos.SupplyShares.IsZero() || !pos.BorrowShares.IsZero() || !pos.Collateral.IsZero() {
			userMarkets = append(userMarkets, marketId)
		}
	}

	return userMarkets
}

func GetPositionList(cur realm, marketId string) []string {
	var positionList []string
	marketPositions, exists := positions.Get(marketId)
	if !exists {
		return positionList
	}

	marketPositions.(*avl.Tree).Iterate("", "", func(key string, _ interface{}) bool {
		positionList = append(positionList, key)
		return false
	})
	return positionList
}

func GetIRMList(cur realm) []string {
	var irmList []string
	irmRegistry.Iterate("", "", func(key string, _ interface{}) bool {
		irmList = append(irmList, key)
		return false
	})
	return irmList
}

func GetEnabledIRMList(cur realm) []string {
	var enabledList []string
	enabledIRMs.Iterate("", "", func(key string, value interface{}) bool {
		if value.(bool) {
			enabledList = append(enabledList, key)
		}
		return false
	})
	return enabledList
}

func GetEnabledLLTVList(cur realm) []string {
	var enabledList []string
	enabledLLTVs.Iterate("", "", func(key string, value interface{}) bool {
		if value.(bool) {
			enabledList = append(enabledList, key)
		}
		return false
	})
	return enabledList
}

// Count getters
func GetMarketCount(cur realm) int {
	return markets.Size()
}

func GetPositionCount(cur realm, marketId string) int {
	marketPositions, exists := positions.Get(marketId)
	if !exists {
		return 0
	}
	return marketPositions.(*avl.Tree).Size()
}

func GetIRMCount(cur realm) int {
	return irmRegistry.Size()
}

// GetMarketPrice returns the current price from the Gnoswap pool
// Returns price in terms of loan token per collateral token
func GetMarketPrice(cur realm, marketId string) string {
	price, err := GetPrice(marketId)
	if err != nil {
		panic(err)
	}

	return price.ToString()
}

// Other getters

func GetSupplyShares(cur realm, marketId string, user std.Address) string {
	position := GetPosition(cur, marketId, user.String())
	return position.SupplyShares.ToString()
}

func GetBorrowShares(cur realm, marketId string, user std.Address) string {
	position := GetPosition(cur, marketId, user.String())
	return position.BorrowShares.ToString()
}

func GetCollateral(cur realm, marketId string, user std.Address) string {
	position := GetPosition(cur, marketId, user.String())
	return position.Collateral.ToString()
}

func GetTotalSupplyAssets(cur realm, marketId string) string {
	market, _ := GetMarket(cur, marketId)
	return market.TotalSupplyAssets.ToString()
}

func GetTotalSupplyShares(cur realm, marketId string) string {
	market, _ := GetMarket(cur, marketId)
	return market.TotalSupplyShares.ToString()
}

func GetTotalBorrowAssets(cur realm, marketId string) string {
	market, _ := GetMarket(cur, marketId)
	return market.TotalBorrowAssets.ToString()
}

func GetTotalBorrowShares(cur realm, marketId string) string {
	market, _ := GetMarket(cur, marketId)
	return market.TotalBorrowShares.ToString()
}

func GetFee(cur realm, marketId string) string {
	market, _ := GetMarket(cur, marketId)
	return market.Fee.ToString()
}

func GetLLTV(cur realm, marketId string) string {
	_, params := GetMarket(cur, marketId)
	return params.LLTV.ToString()
}

func GetFeeRecipient(cur realm) string {
	return feeRecipient.String()

}

// GetExpectedMarketBalances returns the expected market balances after interest accrual as strings
func GetExpectedMarketBalances(cur realm, marketId string) (string, string, string, string) {
	totalSupplyAssets, totalSupplyShares, totalBorrowAssets, totalBorrowShares := ExpectedMarketBalances(cur, marketId)
	return totalSupplyAssets.ToString(), totalSupplyShares.ToString(),
		totalBorrowAssets.ToString(), totalBorrowShares.ToString()
}

// GetExpectedTotalSupplyAssets returns the expected total supply assets after interest accrual as a string
func GetExpectedTotalSupplyAssets(cur realm, marketId string) string {
	return ExpectedTotalSupplyAssets(cur, marketId).ToString()
}

// GetExpectedTotalBorrowAssets returns the expected total borrow assets after interest accrual as a string
func GetExpectedTotalBorrowAssets(cur realm, marketId string) string {
	return ExpectedTotalBorrowAssets(cur, marketId).ToString()
}

// GetExpectedTotalSupplyShares returns the expected total supply shares after interest accrual as a string
func GetExpectedTotalSupplyShares(cur realm, marketId string) string {
	return ExpectedTotalSupplyShares(cur, marketId).ToString()
}

// GetExpectedSupplyAssets returns the expected supply assets balance of a user after interest accrual as a string
func GetExpectedSupplyAssets(cur realm, marketId string, user string) string {
	return ExpectedSupplyAssets(cur, marketId, user).ToString()
}

// GetExpectedBorrowAssets returns the expected borrow assets balance of a user after interest accrual as a string
func GetExpectedBorrowAssets(cur realm, marketId string, user string) string {
	return ExpectedBorrowAssets(cur, marketId, user).ToString()
}

// GetBorrowRate returns the current borrow rate per second as a string (WAD-scaled)
func GetBorrowRate(cur realm, marketId string) string {
	borrowRate := CalculateBorrowRate(cur, marketId)
	return borrowRate.ToString()
}

// GetSupplyRate returns the current supply rate per second as a string (WAD-scaled)
func GetSupplyRate(cur realm, marketId string) string {
	supplyRate := CalculateSupplyRate(cur, marketId)
	return supplyRate.ToString()
}

// GetBorrowAPR returns the current annualized borrow rate as a string (WAD-scaled)
func GetBorrowAPR(cur realm, marketId string) string {
	borrowAPR := CalculateBorrowAPR(cur, marketId)
	return borrowAPR.ToString()
}

// GetSupplyAPR returns the current annualized supply rate as a string (WAD-scaled)
func GetSupplyAPR(cur realm, marketId string) string {
	supplyAPR := CalculateSupplyAPR(cur, marketId)
	return supplyAPR.ToString()
}

// GetLoanAmount returns the current loan amount for a user in a market as a string
func GetLoanAmount(cur realm, marketId string, user string) string {
	amount := CalculateLoanAmount(cur, marketId, user)
	return amount.ToString()
}

// GetUserLoans returns a list of all loans for a user in the format:
// [token1, amount1, token2, amount2, ...]
func GetUserLoans(cur realm, user string) []string {
	loans := CalculateUserLoans(cur, user)
	result := make([]string, 0, loans.Size()*2) // pre-allocate for pairs

	loans.Iterate("", "", func(key string, value interface{}) bool {
		token := key
		amount := value.(*u256.Uint)
		result = append(result, token, amount.ToString())
		return false // continue iteration
	})

	return result
}

// GetHealthFactor returns the health factor for a user's position as a string
// A health factor > 1 means the position is healthy
// A health factor < 1 means the position is eligible for liquidation
func GetHealthFactor(cur realm, marketId string, userAddr string) string {
	healthFactor := CalculateHealthFactor(cur, marketId, userAddr)
	return healthFactor.ToString()
}
