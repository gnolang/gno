package stdlibs

import (
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/sdk"
	"github.com/gnolang/gno/pkgs/std"
)

type ExecContext struct {
	ChainID     string
	Height      int64
	Msg         sdk.Msg
	Caller      crypto.Address
	TxSend      std.Coins
	TxSendSpent *std.Coins // mutable
	PkgAddr     crypto.Address
	Banker      Banker
}
