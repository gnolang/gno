package gnolend

import (
    "gno.land/p/demo/json"
)

func ApiGetMarketInfo(marketId string) string {
    marketInfo := GetMarketInfo(marketId)
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

// Helper function to marshal JSON
func marshal(node *json.Node) string {
    b, err := json.Marshal(node)
    if err != nil {
        panic(err.Error())
    }
    return string(b)
}