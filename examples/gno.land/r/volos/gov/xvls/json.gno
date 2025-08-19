package xvls

import (
	"std"

	"gno.land/p/demo/json"
)

// RpcTokenInfo represents xVLS token information in RPC format
type RpcTokenInfo struct {
	Name         string `json:"name"`
	Symbol       string `json:"symbol"`
	Decimals     int    `json:"decimals"`
	TotalSupply  int64  `json:"totalSupply"`
	VotingSupply int64  `json:"votingSupply"`
}

func TokenInfoToRpc() RpcTokenInfo {
	return RpcTokenInfo{
		Name:         Name(),
		Symbol:       Symbol(),
		Decimals:     Decimals(),
		TotalSupply:  TotalSupply(),
		VotingSupply: VotingSupply(),
	}
}

func (r RpcTokenInfo) JSON() *json.Node {
	return json.ObjectNode("tokenInfo", map[string]*json.Node{
		"name":         json.StringNode("name", r.Name),
		"symbol":       json.StringNode("symbol", r.Symbol),
		"decimals":     json.NumberNode("decimals", float64(r.Decimals)),
		"totalSupply":  json.NumberNode("totalSupply", float64(r.TotalSupply)),
		"votingSupply": json.NumberNode("votingSupply", float64(r.VotingSupply)),
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
