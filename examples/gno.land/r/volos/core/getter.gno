package core

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
		panic(ErrIRMNotRegistered)
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

// GetUserMarkets returns a list of market IDs where a user has any position
// (supply, borrow, or collateral)
func GetUserMarkets(userAddr string) []string {
	var userMarkets []string

	// Get list of all markets
	allMarkets := GetMarketList()

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
	price := GetPrice(marketId)
	return price.ToString()
}

// Other getters

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

// GetExpectedMarketBalances returns the expected market balances after interest accrual as strings
func GetExpectedMarketBalances(marketId string) (string, string, string, string) {
	totalSupplyAssets, totalSupplyShares, totalBorrowAssets, totalBorrowShares := ExpectedMarketBalances(marketId)
	return totalSupplyAssets.ToString(), totalSupplyShares.ToString(),
		totalBorrowAssets.ToString(), totalBorrowShares.ToString()
}

// GetExpectedTotalSupplyAssets returns the expected total supply assets after interest accrual as a string
func GetExpectedTotalSupplyAssets(marketId string) string {
	return ExpectedTotalSupplyAssets(marketId).ToString()
}

// GetExpectedTotalBorrowAssets returns the expected total borrow assets after interest accrual as a string
func GetExpectedTotalBorrowAssets(marketId string) string {
	return ExpectedTotalBorrowAssets(marketId).ToString()
}

// GetExpectedTotalSupplyShares returns the expected total supply shares after interest accrual as a string
func GetExpectedTotalSupplyShares(marketId string) string {
	return ExpectedTotalSupplyShares(marketId).ToString()
}

// GetExpectedSupplyAssets returns the expected supply assets balance of a user after interest accrual as a string
func GetExpectedSupplyAssets(marketId string, user string) string {
	return ExpectedSupplyAssets(marketId, user).ToString()
}

// GetExpectedBorrowAssets returns the expected borrow assets balance of a user after interest accrual as a string
func GetExpectedBorrowAssets(marketId string, user string) string {
	return ExpectedBorrowAssets(marketId, user).ToString()
}

// GetSupplyRate returns the current supply rate per second as a string (WAD-scaled)
func GetSupplyRate(marketId string) string {
	supplyRate := CalculateSupplyRate(marketId)
	return supplyRate.ToString()
}

// GetBorrowAPR returns the current annualized borrow rate as a string (WAD-scaled)
func GetBorrowAPR(marketId string) string {
	borrowAPR := CalculateBorrowAPR(marketId)
	return borrowAPR.ToString()
}

// GetSupplyAPR returns the current annualized supply rate as a string (WAD-scaled)
func GetSupplyAPR(marketId string) string {
	supplyAPR := CalculateSupplyAPR(marketId)
	return supplyAPR.ToString()
}

// GetLoanAmount returns the current loan amount for a user in a market as a string
func GetLoanAmount(marketId string, user string) string {
	amount := CalculateLoanAmount(marketId, user)
	return amount.ToString()
}

// GetUserLoans returns a list of all loans for a user in the format:
// [token1, amount1, token2, amount2, ...]
func GetUserLoans(user string) []string {
	loans := CalculateUserLoans(user)
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
func GetHealthFactor(marketId string, userAddr string) string {
	healthFactor := CalculateHealthFactor(marketId, userAddr)
	return healthFactor.ToString()
}

// GetOwner returns the current owner of the volos realm
func GetOwner() string {
	return Ownable.Owner().String()
}

// GetAddress returns the address of this volos realm derived from its path
func GetAddress() string {
	return std.DerivePkgAddr("gno.land/r/volos/core").String()
}

// GetIsAuthorized checks if authorized can act on behalf of authorizer
func GetIsAuthorized(authorizer string, authorized string) bool {
	return IsAuthorized(std.Address(authorizer), std.Address(authorized))
}

// getBorrowRate returns the current borrow rate per second
// The returned value is WAD-scaled (1e18)
func GetBorrowRate(marketId string) *u256.Uint {
	// Check if market exists
	if _, exists := markets.Get(marketId); !exists {
		panic(ErrMarketNotCreated)
	}

	market, params := GetMarket(marketId)

	// Get IRM and calculate current borrow rate
	irm := GetIRM(params.IRM)
	return irm.BorrowRate(market.TotalSupplyAssets, market.TotalBorrowAssets)
}
