package bank

import (
	"github.com/gnolang/gno/tm2/pkg/std"
)

// TransferEvent records a successful bank transfer.
type TransferEvent struct {
	// From is empty for multisend credits; To is empty for multisend debits.
	From   string    `json:"from"`
	To     string    `json:"to"`
	Amount std.Coins `json:"amount"`
}

func (TransferEvent) AssertABCIEvent() {}
