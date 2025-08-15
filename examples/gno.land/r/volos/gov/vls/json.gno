package vls

import (
	"gno.land/p/demo/json"
	"std"
)

// RpcTokenInfo represents VLS token information in RPC format
type RpcTokenInfo struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Decimals    int    `json:"decimals"`
	TotalSupply int64  `json:"totalSupply"`
}

func TokenInfoToRpc() RpcTokenInfo {
	return RpcTokenInfo{
		Name:        Name(),
		Symbol:      Symbol(),
		Decimals:    Decimals(),
		TotalSupply: TotalSupply(),
	}
}

func (r RpcTokenInfo) JSON() *json.Node {
	return json.ObjectNode("tokenInfo", map[string]*json.Node{
		"name":        json.StringNode("name", r.Name),
		"symbol":      json.StringNode("symbol", r.Symbol),
		"decimals":    json.NumberNode("decimals", float64(r.Decimals)),
		"totalSupply": json.NumberNode("totalSupply", float64(r.TotalSupply)),
	})
}

// RpcBalance represents a balance in RPC format
type RpcBalance struct {
	Address string `json:"address"`
	Balance int64  `json:"balance"`
}

func BalanceToRpc(addr std.Address) RpcBalance {
	return RpcBalance{
		Address: addr.String(),
		Balance: BalanceOf(addr),
	}
}

func (r RpcBalance) JSON() *json.Node {
	return json.ObjectNode("balance", map[string]*json.Node{
		"address": json.StringNode("address", r.Address),
		"balance": json.NumberNode("balance", float64(r.Balance)),
	})
}

// RpcAllowance represents an allowance in RPC format
type RpcAllowance struct {
	Owner    string `json:"owner"`
	Spender  string `json:"spender"`
	Allowance int64 `json:"allowance"`
}

func AllowanceToRpc(owner, spender std.Address) RpcAllowance {
	return RpcAllowance{
		Owner:     owner.String(),
		Spender:   spender.String(),
		Allowance: Allowance(owner, spender),
	}
}

func (r RpcAllowance) JSON() *json.Node {
	return json.ObjectNode("allowance", map[string]*json.Node{
		"owner":     json.StringNode("owner", r.Owner),
		"spender":   json.StringNode("spender", r.Spender),
		"allowance": json.NumberNode("allowance", float64(r.Allowance)),
	})
}
