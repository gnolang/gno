package gnolend

import (
    "gno.land/p/demo/json"
    "gno.land/p/demo/u256"
)

func ApiGetMarketInfo(marketId string) string {
    marketInfo := GetRpcMarketInfo(marketId)
    return marshal(marketInfo.JSON())
}

func ApiGetMarket(marketId string) string {
    market, _ := GetMarket(marketId)
    return marshal(market.ToRpc().JSON())
}

func ApiGetMarketParams(marketId string) string {
    _, params := GetMarket(marketId)
    return marshal(params.ToRpc().JSON())
}

func ApiGetPosition(marketId, userAddr string) string {
    position := GetPosition(marketId, userAddr)
    return marshal(position.ToRpc().JSON())
}

func ApiListMarketsInfo() string {
    marketList := GetMarketList()
    markets := json.ArrayNode("", []*json.Node{})

    for _, marketId := range marketList {
        marketInfo := GetMarketInfo(marketId)
        marketWrapper := json.ObjectNode("", map[string]*json.Node{
            marketId: marketInfo.JSON(),
        })
        markets.AppendArray(marketWrapper)
    }

    return marshal(markets)
}

func ApiGetLoanAmount(marketId string, user string) string {
    amount := GetLoanAmount(marketId, user)
    return marshal(json.StringNode("", amount))
}

func ApiGetUserLoans(user string) string {
    loans := CalculateUserLoans(user)
    result := json.ArrayNode("", []*json.Node{})

    loans.Iterate(func(key interface{}, value interface{}) bool {
        token := key.(string)
        amount := value.(*u256.Uint)

        loanNode := json.ObjectNode("", map[string]*json.Node{
            "token":  json.StringNode("", token),
            "amount": json.StringNode("", amount.String()),
        })
        result.AppendArray(loanNode)
        return false
    })

    return marshal(result)
}

// Helper function to marshal JSON
func marshal(node *json.Node) string {
    b, err := json.Marshal(node)
    if err != nil {
        panic(err.Error())
    }
    return string(b)
}