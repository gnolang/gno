package xvls

import (
	"std"

	"gno.land/p/demo/json"
)

// ApiGetTokenInfo returns xVLS token information as JSON object
func ApiGetTokenInfo() string {
	tokenInfo := TokenInfoToRpc()
	return marshal(tokenInfo.JSON())
}

// ApiGetBalance returns a balance as JSON object
func ApiGetBalance(address string) string {
	addr := std.Address(address)
	if !addr.IsValid() {
		return marshalError("invalid address")
	}

	balance := BalanceToRpc(addr)
	return marshal(balance.JSON())
}

func marshal(node *json.Node) string {
	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}
	return string(b)
}

func marshalError(message string) string {
	errorNode := json.ObjectNode("", map[string]*json.Node{
		"error": json.StringNode("error", message),
	})
	return marshal(errorNode)
}
