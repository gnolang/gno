package gnoland

import (
	"github.com/gnolang/gno/pkgs/std"
)

type GnoAccount struct {
	std.BaseAccount
}

func ProtoGnoAccount() std.Account {
	return &GnoAccount{}
}

type GnoGenesisState struct {
	Balances []string `json:"balances"`
	Txs      []std.Tx `json:"txs"`
}
