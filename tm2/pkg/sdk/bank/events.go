package bank

import (
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// TransferEvent is emitted when coins are transferred between accounts.
type TransferEvent struct {
	From   crypto.Address `json:"from"`
	To     crypto.Address `json:"to"`
	Amount std.Coins      `json:"amount"`
}

func (TransferEvent) AssertABCIEvent() {}
