package gnolend

import (
	"gno.land/p/demo/json"
	u256 "gno.land/p/gnoswap/uint256"
)

// RpcMarket

type RpcMarket struct {
	TotalSupplyAssets string `json:"totalSupplyAssets"`
	TotalSupplyShares string `json:"totalSupplyShares"`
	TotalBorrowAssets string `json:"totalBorrowAssets"`
	TotalBorrowShares string `json:"totalBorrowShares"`
	LastUpdate        int64  `json:"lastUpdate"`
	Fee               string `json:"fee"`
}

func (m Market) ToRpc() RpcMarket {
	return RpcMarket{
		TotalSupplyAssets: m.TotalSupplyAssets.ToString(),
		TotalSupplyShares: m.TotalSupplyShares.ToString(),
		TotalBorrowAssets: m.TotalBorrowAssets.ToString(),
		TotalBorrowShares: m.TotalBorrowShares.ToString(),
		LastUpdate:        m.LastUpdate,
		Fee:               m.Fee.ToString(),
	}
}

func (r RpcMarket) JSON() *json.Node {
	return json.ObjectNode("market", map[string]*json.Node{
		"totalSupplyAssets": json.StringNode("totalSupplyAssets", r.TotalSupplyAssets),
		"totalSupplyShares": json.StringNode("totalSupplyShares", r.TotalSupplyShares),
		"totalBorrowAssets": json.StringNode("totalBorrowAssets", r.TotalBorrowAssets),
		"totalBorrowShares": json.StringNode("totalBorrowShares", r.TotalBorrowShares),
		"lastUpdate":        json.NumberNode("lastUpdate", float64(r.LastUpdate)),
		"fee":               json.StringNode("fee", r.Fee),
	})
}

// RpcPosition

type RpcPosition struct {
	SupplyShares string `json:"supplyShares"`
	BorrowShares string `json:"borrowShares"`
	Collateral   string `json:"collateral"`
}

func (p Position) ToRpc() RpcPosition {
	return RpcPosition{
		SupplyShares: p.SupplyShares.ToString(),
		BorrowShares: p.BorrowShares.ToString(),
		Collateral:   p.Collateral.ToString(),
	}
}

func (r RpcPosition) JSON() *json.Node {
	return json.ObjectNode("position", map[string]*json.Node{
		"supplyShares": json.StringNode("supplyShares", r.SupplyShares),
		"borrowShares": json.StringNode("borrowShares", r.BorrowShares),
		"collateral":   json.StringNode("collateral", r.Collateral),
	})
}

// RpcMarketParams

type RpcMarketParams struct {
	PoolPath     string `json:"poolPath"`
	IRM          string `json:"irm"`
	LLTV         string `json:"lltv"`
	IsToken0Loan bool   `json:"isToken0Loan"`
}

func (mp MarketParams) ToRpc() RpcMarketParams {
	return RpcMarketParams{
		PoolPath:     mp.PoolPath,
		IRM:          mp.IRM,
		LLTV:         mp.LLTV.ToString(),
		IsToken0Loan: mp.IsToken0Loan,
	}
}

func (r RpcMarketParams) JSON() *json.Node {
	return json.ObjectNode("market_params", map[string]*json.Node{
		"poolPath":     json.StringNode("poolPath", r.PoolPath),
		"irm":          json.StringNode("irm", r.IRM),
		"lltv":         json.StringNode("lltv", r.LLTV),
		"isToken0Loan": json.BoolNode("isToken0Loan", r.IsToken0Loan),
	})
}

// RpcIRM

type RpcIRM struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Parameters string `json:"parameters"`
}

func IRMToRpc(irm IRM) RpcIRM {
	return RpcIRM{
		Name:       irm.Name(),
		Type:       "IRM",
		Parameters: "{}",
	}
}

func (r RpcIRM) JSON() *json.Node {
	return json.ObjectNode("irm", map[string]*json.Node{
		"name":       json.StringNode("name", r.Name),
		"type":       json.StringNode("type", r.Type),
		"parameters": json.StringNode("parameters", r.Parameters),
	})
}

// RpcMarketInfo combines all market information into a single flattened structure
type RpcMarketInfo struct {
	// Market fields
	TotalSupplyAssets string `json:"totalSupplyAssets"`
	TotalSupplyShares string `json:"totalSupplyShares"`
	TotalBorrowAssets string `json:"totalBorrowAssets"`
	TotalBorrowShares string `json:"totalBorrowShares"`
	LastUpdate        int64  `json:"lastUpdate"`
	Fee               string `json:"fee"`

	// Params fields
	PoolPath     string `json:"poolPath"`
	IRM          string `json:"irm"`
	LLTV         string `json:"lltv"`
	IsToken0Loan bool   `json:"isToken0Loan"`

	// Additional fields
	LoanToken       string `json:"loanToken"`
	CollateralToken string `json:"collateralToken"`
	CurrentPrice    string `json:"currentPrice"`
	BorrowAPR       string `json:"borrowAPR"`
	SupplyAPR       string `json:"supplyAPR"`
	Utilization     string `json:"utilization"`

	// Token information
	LoanTokenName     string `json:"loanTokenName"`
	LoanTokenSymbol   string `json:"loanTokenSymbol"`
	LoanTokenDecimals uint   `json:"loanTokenDecimals"`

	CollateralTokenName     string `json:"collateralTokenName"`
	CollateralTokenSymbol   string `json:"collateralTokenSymbol"`
	CollateralTokenDecimals uint   `json:"collateralTokenDecimals"`
}

// GetMarketInfo returns comprehensive information about a market
func GetRpcMarketInfo(cur realm, marketId string) RpcMarketInfo {
	market, params := GetMarket(cur, marketId)

	// Get price from oracle
	price, err := GetPrice(marketId)
	priceStr := ""
	if err == nil {
		priceStr = price.ToString()
	}

	// Get APRs using utility functions
	borrowAPR := CalculateBorrowAPR(cur, marketId)
	supplyAPR := CalculateSupplyAPR(cur, marketId)

	// Calculate utilization rate
	var utilization *u256.Uint
	if market.TotalSupplyAssets.IsZero() {
		utilization = u256.Zero()
	} else {
		utilization = WDivDown(market.TotalBorrowAssets, market.TotalSupplyAssets)
	}

	// Get token paths
	loanToken := params.GetLoanToken()
	collateralToken := params.GetCollateralToken()

	// Get token information using GetToken utility
	var loanTokenName, loanTokenSymbol string
	var loanTokenDecimals uint
	var collateralTokenName, collateralTokenSymbol string
	var collateralTokenDecimals uint

	loanTokenInfo := GetToken(loanToken)
	if loanTokenInfo != nil {
		loanTokenName = loanTokenInfo.GetName()
		loanTokenSymbol = loanTokenInfo.GetSymbol()
		loanTokenDecimals = uint(loanTokenInfo.GetDecimals())
	}

	collateralTokenInfo := GetToken(collateralToken)
	if collateralTokenInfo != nil {
		collateralTokenName = collateralTokenInfo.GetName()
		collateralTokenSymbol = collateralTokenInfo.GetSymbol()
		collateralTokenDecimals = uint(collateralTokenInfo.GetDecimals())
	}

	return RpcMarketInfo{
		// Market fields
		TotalSupplyAssets: market.TotalSupplyAssets.ToString(),
		TotalSupplyShares: market.TotalSupplyShares.ToString(),
		TotalBorrowAssets: market.TotalBorrowAssets.ToString(),
		TotalBorrowShares: market.TotalBorrowShares.ToString(),
		LastUpdate:        market.LastUpdate,
		Fee:               market.Fee.ToString(),

		// Params fields
		PoolPath:     params.PoolPath,
		IRM:          params.IRM,
		LLTV:         params.LLTV.ToString(),
		IsToken0Loan: params.IsToken0Loan,

		// Additional fields
		LoanToken:       loanToken,
		CollateralToken: collateralToken,
		CurrentPrice:    priceStr,
		BorrowAPR:       borrowAPR.ToString(),
		SupplyAPR:       supplyAPR.ToString(),
		Utilization:     utilization.ToString(),

		// Token information
		LoanTokenName:     loanTokenName,
		LoanTokenSymbol:   loanTokenSymbol,
		LoanTokenDecimals: loanTokenDecimals,

		CollateralTokenName:     collateralTokenName,
		CollateralTokenSymbol:   collateralTokenSymbol,
		CollateralTokenDecimals: collateralTokenDecimals,
	}
}

func (r RpcMarketInfo) JSON() *json.Node {
	return json.ObjectNode("market_info", map[string]*json.Node{
		// Market fields
		"totalSupplyAssets": json.StringNode("totalSupplyAssets", r.TotalSupplyAssets),
		"totalSupplyShares": json.StringNode("totalSupplyShares", r.TotalSupplyShares),
		"totalBorrowAssets": json.StringNode("totalBorrowAssets", r.TotalBorrowAssets),
		"totalBorrowShares": json.StringNode("totalBorrowShares", r.TotalBorrowShares),
		"lastUpdate":        json.NumberNode("lastUpdate", float64(r.LastUpdate)),
		"fee":               json.StringNode("fee", r.Fee),

		// Params fields
		"poolPath":     json.StringNode("poolPath", r.PoolPath),
		"irm":          json.StringNode("irm", r.IRM),
		"lltv":         json.StringNode("lltv", r.LLTV),
		"isToken0Loan": json.BoolNode("isToken0Loan", r.IsToken0Loan),

		// Additional fields
		"loanToken":       json.StringNode("loanToken", r.LoanToken),
		"collateralToken": json.StringNode("collateralToken", r.CollateralToken),
		"currentPrice":    json.StringNode("currentPrice", r.CurrentPrice),
		"borrowAPR":       json.StringNode("borrowAPR", r.BorrowAPR),
		"supplyAPR":       json.StringNode("supplyAPR", r.SupplyAPR),
		"utilization":     json.StringNode("utilization", r.Utilization),

		// Token information
		"loanTokenName":     json.StringNode("loanTokenName", r.LoanTokenName),
		"loanTokenSymbol":   json.StringNode("loanTokenSymbol", r.LoanTokenSymbol),
		"loanTokenDecimals": json.NumberNode("loanTokenDecimals", float64(r.LoanTokenDecimals)),

		"collateralTokenName":     json.StringNode("collateralTokenName", r.CollateralTokenName),
		"collateralTokenSymbol":   json.StringNode("collateralTokenSymbol", r.CollateralTokenSymbol),
		"collateralTokenDecimals": json.NumberNode("collateralTokenDecimals", float64(r.CollateralTokenDecimals)),
	})
}
