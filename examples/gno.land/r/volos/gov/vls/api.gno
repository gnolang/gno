package vls

import (
	"gno.land/p/demo/json"
	"std"
)

// ApiGetTokenInfo returns VLS token information as JSON string
func ApiGetTokenInfo() string {
	tokenInfo := TokenInfoToRpc()
	return marshal(tokenInfo.JSON())
}

// ApiGetBalance returns a balance as JSON string
func ApiGetBalance(address string) string {
	addr := std.Address(address)
	if !addr.IsValid() {
		return marshalError("invalid address")
	}
	
	balance := BalanceToRpc(addr)
	return marshal(balance.JSON())
}

// ApiGetAllowance returns an allowance as JSON string
func ApiGetAllowance(owner, spender string) string {
	ownerAddr := std.Address(owner)
	spenderAddr := std.Address(spender)
	
	if !ownerAddr.IsValid() {
		return marshalError("invalid owner address")
	}
	if !spenderAddr.IsValid() {
		return marshalError("invalid spender address")
	}
	
	allowance := AllowanceToRpc(ownerAddr, spenderAddr)
	return marshal(allowance.JSON())
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
