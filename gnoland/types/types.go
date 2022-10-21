package types

import (
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/std"
)

type GnoAccount struct {
	std.BaseAccount
	*PackageAccount `json:"PackageAccount,omitempty",yaml:"package_account,omitempty"`
}

type PackageAccount struct {
	Owner crypto.Address `json:"owner",yaml:"owner"`
}

func ProtoGnoAccount() std.Account {
	return &GnoAccount{}
}

type GnoGenesisState struct {
	Balances []string `json:"balances"`
	Txs      []std.Tx `json:"txs"`
}
