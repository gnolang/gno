package gnoland

import (
	"errors"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	ErrBalanceEmptyAddress = errors.New("balance address is empty")
	ErrBalanceEmptyAmount  = errors.New("balance amount is empty")
)

type GnoAccount struct {
	std.BaseAccount
}

func ProtoGnoAccount() std.Account {
	return &GnoAccount{}
}

type GnoGenesisState struct {
	Balances []Balance `json:"balances"`
	Txs      []std.Tx  `json:"txs"`
	// Should match len(Txs), or be null
	TxContexts []vm.ExecContextCustom `json:"tx_contexts"`
}
