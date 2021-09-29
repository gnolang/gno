package vm

import (
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/sdk"
)

type EvalContext struct {
	ChainID string
	Height  int64
	Msg     MsgEval
	PkgAddr crypto.Address

	sdkCtx sdk.Context // TODO: ensure hidden or refactor.
}
