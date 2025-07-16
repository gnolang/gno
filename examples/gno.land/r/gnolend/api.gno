package gnolend

import (
	"gno.land/p/demo/json"
	u256 "gno.land/p/gnoswap/uint256"
)

func ApiGetMarketInfo(cur realm, marketId string) string {
	marketInfo := GetRpcMarketInfo(cur, marketId)
	return marshal(marketInfo.JSON())
}

func ApiGetMarket(cur realm, marketId string) string {
	market, _ := GetMarket(cur, marketId)
	return marshal(market.ToRpc().JSON())
}

func ApiGetMarketParams(cur realm, marketId string) string {
	_, params := GetMarket(cur, marketId)
	return marshal(params.ToRpc().JSON())
}

func ApiGetPosition(cur realm, marketId, userAddr string) string {
	position := GetPosition(cur, marketId, userAddr)
	return marshal(position.ToRpc().JSON())
}

func ApiListMarketsInfo(cur realm) string {
	marketList := GetMarketList(cur)
	markets := json.ArrayNode("", []*json.Node{})

	for _, marketId := range marketList {
		marketInfo := GetRpcMarketInfo(cur, marketId)
		marketWrapper := json.ObjectNode("", map[string]*json.Node{
			marketId: marketInfo.JSON(),
		})
		markets.AppendArray(marketWrapper)
	}

	return marshal(markets)
}

func ApiGetLoanAmount(cur realm, marketId string, user string) string {
	amount := GetLoanAmount(cur, marketId, user)
	amountNode := json.ObjectNode("", map[string]*json.Node{
		"amount": json.StringNode("", amount),
	})
	return marshal(amountNode)
}

func ApiGetUserLoans(cur realm, user string) string {
	loans := CalculateUserLoans(cur, user)
	result := json.ArrayNode("", []*json.Node{})

	loans.Iterate("", "", func(key string, value interface{}) bool {
		token := key
		amount := value.(*u256.Uint)

		loanNode := json.ObjectNode("", map[string]*json.Node{
			"token":  json.StringNode("", token),
			"amount": json.StringNode("", amount.ToString()),
		})
		result.AppendArray(loanNode)
		return false
	})

	return marshal(result)
}

// ApiGetHealthFactor returns the health factor for a user's position as JSON
func ApiGetHealthFactor(cur realm, marketId, userAddr string) string {
	healthFactor := GetHealthFactor(cur, marketId, userAddr)
	healthFactorNode := json.ObjectNode("", map[string]*json.Node{
		"healthFactor": json.StringNode("", healthFactor),
	})
	return marshal(healthFactorNode)
}

// Helper function to marshal JSON
func marshal(node *json.Node) string {
	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}
	return string(b)
}