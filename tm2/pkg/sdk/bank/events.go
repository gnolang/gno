package bank

import (
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// TransferEvent is emitted for 1:1 transfers.
type TransferEvent struct {
	From   crypto.Address `json:"from"`
	To     crypto.Address `json:"to"`
	Amount std.Coins      `json:"amount"`
}

func (TransferEvent) AssertABCIEvent() {}

// CoinSpentEvent is emitted when coins leave an account.
type CoinSpentEvent struct {
	Spender crypto.Address `json:"spender"`
	Amount  std.Coins      `json:"amount"`
}

func (CoinSpentEvent) AssertABCIEvent() {}

// CoinReceivedEvent is emitted when coins enter an account.
type CoinReceivedEvent struct {
	Receiver crypto.Address `json:"receiver"`
	Amount   std.Coins      `json:"amount"`
}

func (CoinReceivedEvent) AssertABCIEvent() {}
