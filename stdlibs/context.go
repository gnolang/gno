package stdlibs

import (
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/sdk"
	"github.com/gnolang/gno/pkgs/std"
)

type ExecContext struct {
	ChainID       string
	Height        int64
	Timestamp     int64
	Msg           sdk.Msg
	OrigCaller    crypto.Bech32Address
	OrigPkgAddr   crypto.Bech32Address
	OrigSend      std.Coins
	OrigSendSpent *std.Coins // mutable
	Banker        Banker
}
