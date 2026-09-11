package bank

import (
	"github.com/gnolang/gno/tm2/pkg/std"
)

// TransferEvent records a successful bank transfer.
type TransferEvent struct {
	From  string    `json:"from"`
	To    string    `json:"to"`
	Coins std.Coins `json:"coins"`
}

func (TransferEvent) AssertABCIEvent() {}
