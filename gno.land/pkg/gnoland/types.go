package gnoland

import (
	"fmt"

	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/std"
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

type Balance struct {
	Address bft.Address
	Value   std.Coin
}

func (b Balance) String() string {
	return fmt.Sprintf("%s=%s", b.Address.String(), b.Value.String())
}

type Balances []Balance

func (bs Balances) Strings() []string {
	bss := make([]string, len(bs))
	for i, balance := range bs {
		bss[i] = balance.String()
	}
	return bss
}
