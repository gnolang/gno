package vm

import (
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/sdk"
)

type ExecContext struct {
	ChainID string
	Height  int64
	Msg     MsgCall
	PkgAddr crypto.Address

	sdkCtx sdk.Context // TODO: ensure hidden or refactor.
}
