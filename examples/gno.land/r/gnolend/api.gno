package gnolend

import (
    "gno.land/p/demo/json"
)

// Market API Response
func ApiGetMarket(marketId string) string {
    market, _ := GetMarket(marketId)
    builder := json.Builder()
    builder.WriteObject("market", func(b *json.NodeBuilder) {
        rpcMarket := market.ToRpc()
        b.WriteString("totalSupplyAssets", rpcMarket.TotalSupplyAssets)
        b.WriteString("totalSupplyShares", rpcMarket.TotalSupplyShares)
        b.WriteString("totalBorrowAssets", rpcMarket.TotalBorrowAssets)
        b.WriteString("totalBorrowShares", rpcMarket.TotalBorrowShares)
        b.WriteNumber("lastUpdate", float64(rpcMarket.LastUpdate))
        b.WriteString("fee", rpcMarket.Fee)
    })
    
    return marshal(builder.Node())
}

func ApiGetMarketParams(marketId string) string {
    _, params := GetMarket(marketId)
	builder := json.Builder()
	builder.WriteObject("params", func(b *json.NodeBuilder) {
		rpcParams := params.ToRpc()
		b.WriteString("poolPath", rpcParams.PoolPath)
		b.WriteString("irm", rpcParams.IRM)
		b.WriteString("lltv", rpcParams.LLTV)
		b.WriteBool("isToken0Loan", rpcParams.IsToken0Loan)
	})
	
	return marshal(builder.Node())
}

// Position API Response
func ApiGetPosition(marketId, userAddr string) string {
    position := GetPosition(marketId, userAddr)
    
    builder := json.Builder()
    builder.WriteObject("position", func(b *json.NodeBuilder) {
        rpcPos := position.ToRpc()
        b.WriteString("supplyShares", rpcPos.SupplyShares)
        b.WriteString("borrowShares", rpcPos.BorrowShares)
        b.WriteString("collateral", rpcPos.Collateral)
    })
    
    return marshal(builder.Node())
}

// List Markets
func ApiListMarkets() string {
    marketList := GetMarketList()
    
    builder := json.Builder()
    builder.WriteArray("markets", func(ab *json.ArrayBuilder) {
        for _, marketId := range marketList {
            market, _ := GetMarket(marketId)
            ab.WriteObject(func(b *json.NodeBuilder) {
                rpcMarket := market.ToRpc()
                b.WriteString("totalSupplyAssets", rpcMarket.TotalSupplyAssets)
                b.WriteString("totalSupplyShares", rpcMarket.TotalSupplyShares)
                b.WriteString("totalBorrowAssets", rpcMarket.TotalBorrowAssets)
                b.WriteString("totalBorrowShares", rpcMarket.TotalBorrowShares)
                b.WriteNumber("lastUpdate", float64(rpcMarket.LastUpdate))
                b.WriteString("fee", rpcMarket.Fee)
            })
        }
    })
    
    return marshal(builder.Node())
}

// Helper function to marshal JSON
func marshal(node *json.Node) string {
    b, err := json.Marshal(node)
    if err != nil {
        panic(err.Error())
    }
    return string(b)
}