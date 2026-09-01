package bank

import (
	"github.com/gnolang/gno/tm2/pkg/std"
)

// TransferEvent records a successful bank transfer.
type TransferEvent struct {
	// From is empty for multisends, whose inputs cannot be mapped one-to-one
	// to their outputs.
	From   string    `json:"from"`
	To     string    `json:"to"`
	Amount std.Coins `json:"amount"`
}

func (TransferEvent) AssertABCIEvent() {}
