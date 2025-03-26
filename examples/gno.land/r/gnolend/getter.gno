package gnolend

import (
	"std"

	"gno.land/p/demo/avl"
	u256 "gno.land/p/gnoswap/uint256"
)

// Position getters
// Main struct getter
func GetPosition(marketId string, userAddr string) Position {
	marketPositions, exists := positions.Get(marketId)
	if !exists {
		return Position{
			SupplyShares: new(u256.Uint),
			BorrowShares: new(u256.Uint),
			Collateral: new(u256.Uint),
		} // Market doesn't exist, return empty position with initialized fields
	}

	position, exists := marketPositions.(*avl.Tree).Get(userAddr)
	if !exists {
		return Position{
			SupplyShares: new(u256.Uint),
			BorrowShares: new(u256.Uint),
			Collateral: new(u256.Uint),
		} // Position doesn't exist, return empty position with initialized fields
	}
	return position.(Position)
}

// Individual field getters
func GetPositionSupplyShares(marketId string, userAddr string) string {
	position := GetPosition(marketId, userAddr)
	return position.SupplyShares.ToString()
}

func GetPositionBorrowShares(marketId string, userAddr string) string {
	position := GetPosition(marketId, userAddr)
	return position.BorrowShares.ToString()
}

func GetPositionCollateral(marketId string, userAddr string) string {
	position := GetPosition(marketId, userAddr)
	return position.Collateral.ToString()
}

// Market getters
// Main struct getter
func GetMarket(marketId string) (Market, MarketParams) {
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
func GetMarketTotalSupplyAssets(marketId string) string {
	market, _ := GetMarket(marketId)
	return market.TotalSupplyAssets.ToString()
}

func GetMarketTotalSupplyShares(marketId string) string {
	market, _ := GetMarket(marketId)
	return market.TotalSupplyShares.ToString()
}

func GetMarketTotalBorrowAssets(marketId string) string {
	market, _ := GetMarket(marketId)
	return market.TotalBorrowAssets.ToString()
}

func GetMarketTotalBorrowShares(marketId string) string {
	market, _ := GetMarket(marketId)
	return market.TotalBorrowShares.ToString()
}

func GetMarketLastUpdate(marketId string) int64 {
	market, _ := GetMarket(marketId)
	return market.LastUpdate
}

func GetMarketFee(marketId string) string {
	market, _ := GetMarket(marketId)
	return market.Fee.ToString()
}

// MarketParams getters
func GetMarketParamsLoanToken(marketId string) string {
	_, params := GetMarket(marketId)
	return params.GetLoanToken()
}

func GetMarketParamsCollateralToken(marketId string) string {
	_, params := GetMarket(marketId)
	return params.GetCollateralToken()
}

func GetMarketParamsPoolPath(marketId string) string {
	_, params := GetMarket(marketId)
	return params.PoolPath
}

func GetMarketParamsIRM(marketId string) string {
	_, params := GetMarket(marketId)
	return params.IRM
}

func GetMarketParamsLLTV(marketId string) string {
	_, params := GetMarket(marketId)
	return params.LLTV.ToString()
}

// IRM getters
// Main getter
func GetIRM(name string) IRM {
	irm, exists := irmRegistry.Get(name)
	if !exists {
		panic("IRM not registered: " + name)
	}
	return irm.(IRM)
}

// List getters
func GetMarketList() []string {
	var marketList []string
	markets.Iterate("", "", func(key string, _ interface{}) bool {
		marketList = append(marketList, key)
		return false
	})
	return marketList
}

func GetPositionList(marketId string) []string {
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

func GetIRMList() []string {
	var irmList []string
	irmRegistry.Iterate("", "", func(key string, _ interface{}) bool {
		irmList = append(irmList, key)
		return false
	})
	return irmList
}

func GetEnabledIRMList() []string {
	var enabledList []string
	enabledIRMs.Iterate("", "", func(key string, value interface{}) bool {
		if value.(bool) {
			enabledList = append(enabledList, key)
		}
		return false
	})
	return enabledList
}

func GetEnabledLLTVList() []string {
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
func GetMarketCount() int {
	return markets.Size()
}

func GetPositionCount(marketId string) int {
	marketPositions, exists := positions.Get(marketId)
	if !exists {
		return 0
	}
	return marketPositions.(*avl.Tree).Size()
}

func GetIRMCount() int {
	return irmRegistry.Size()
}

// GetMarketPrice returns the current price from the Gnoswap pool
// Returns price in terms of loan token per collateral token
func GetMarketPrice(marketId string) string {
	// Get price from pool
	price, err := GetPrice(marketId)
	if err != nil {
		panic(err)
	}

	return price.ToString()
}

// Other getters

func GetSupplyShares(marketId string, user std.Address) string {
	position := GetPosition(marketId, user.String())
	return position.SupplyShares.ToString()
}

func GetBorrowShares(marketId string, user std.Address) string {
	position := GetPosition(marketId, user.String())
	return position.BorrowShares.ToString()
}

func GetCollateral(marketId string, user std.Address) string {
	position := GetPosition(marketId, user.String())
	return position.Collateral.ToString()
}

func GetTotalSupplyAssets(marketId string) string {
	market, _ := GetMarket(marketId)
	return market.TotalSupplyAssets.ToString()
}

func GetTotalSupplyShares(marketId string) string {
	market, _ := GetMarket(marketId)
	return market.TotalSupplyShares.ToString()
}

func GetTotalBorrowAssets(marketId string) string {
	market, _ := GetMarket(marketId)
	return market.TotalBorrowAssets.ToString()
}

func GetTotalBorrowShares(marketId string) string {
	market, _ := GetMarket(marketId)
	return market.TotalBorrowShares.ToString()
}

func GetFee(marketId string) string {
	market, _ := GetMarket(marketId)
	return market.Fee.ToString()
}

func GetLLTV(marketId string) string {
	_, params := GetMarket(marketId)
	return params.LLTV.ToString()
}

func GetFeeRecipient() string {
	return feeRecipient.String()
}