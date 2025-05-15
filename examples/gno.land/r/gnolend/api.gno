package gnolend

import (
    "gno.land/p/demo/json"
)

func ApiGetMarketInfo(marketId string) string {
    marketInfo := GetMarketInfo(marketId)
    node := json.ObjectNode("", map[string]*json.Node{
        "marketInfo": marketInfo.JSON(),
    })
    
    return marshal(node)
}

func ApiGetMarket(marketId string) string {
    market, _ := GetMarket(marketId)
    node := json.ObjectNode("", map[string]*json.Node{
        "market": market.ToRpc().JSON(),
    })
    
    return marshal(node)
}

func ApiGetMarketParams(marketId string) string {
    _, params := GetMarket(marketId)
    node := json.ObjectNode("", map[string]*json.Node{
        "params": params.ToRpc().JSON(),
    })
    
    return marshal(node)
}

func ApiGetPosition(marketId, userAddr string) string {
    position := GetPosition(marketId, userAddr)
    node := json.ObjectNode("", map[string]*json.Node{
        "position": position.ToRpc().JSON(),
    })
    
    return marshal(node)
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
    
    node := json.ObjectNode("", map[string]*json.Node{
        "markets": markets,
    })
    
    return marshal(node)
}

// Helper function to marshal JSON
func marshal(node *json.Node) string {
    b, err := json.Marshal(node)
    if err != nil {
        panic(err.Error())
    }
    return string(b)
}