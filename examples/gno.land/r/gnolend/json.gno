package gnolend

import (
    "gno.land/p/demo/json"
)

// RpcMarket matches the pool example's Rpc-prefixed pattern
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
    return json.ObjectNode("", map[string]*json.Node{
        "totalSupplyAssets": json.StringNode("totalSupplyAssets", r.TotalSupplyAssets),
        "totalSupplyShares": json.StringNode("totalSupplyShares", r.TotalSupplyShares),
        "totalBorrowAssets": json.StringNode("totalBorrowAssets", r.TotalBorrowAssets),
        "totalBorrowShares": json.StringNode("totalBorrowShares", r.TotalBorrowShares),
        "lastUpdate":        json.NumberNode("lastUpdate", float64(r.LastUpdate)),
        "fee":               json.StringNode("fee", r.Fee),
    })
}

// RpcPosition with proper JSON tags
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
    return json.ObjectNode("", map[string]*json.Node{
        "supplyShares": json.StringNode("supplyShares", r.SupplyShares),
        "borrowShares": json.StringNode("borrowShares", r.BorrowShares),
        "collateral":   json.StringNode("collateral", r.Collateral),
    })
}

// RpcMarketParams with JSON tags
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
    return json.ObjectNode("", map[string]*json.Node{
        "poolPath":     json.StringNode("poolPath", r.PoolPath),
        "irm":          json.StringNode("irm", r.IRM),
        "lltv":         json.StringNode("lltv", r.LLTV),
        "isToken0Loan": json.BoolNode("isToken0Loan", r.IsToken0Loan),
    })
}

// RpcIRM implementation
type RpcIRM struct {
    Name       string `json:"name"`
    Type       string `json:"type"`
    Parameters string `json:"parameters"`
}

func IRMToRpc(irm IRM) RpcIRM {
    return RpcIRM{
        Name:       irm.Name(),
        Type:       "IRM",
        Parameters: "{}", // Implement parameter serialization if needed
    }
}

func (r RpcIRM) JSON() *json.Node {
    return json.ObjectNode("", map[string]*json.Node{
        "name":       json.StringNode("name", r.Name),
        "type":       json.StringNode("type", r.Type),
        "parameters": json.StringNode("parameters", r.Parameters),
    })
} 