package vm

import (
	"sync"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// vm.VMKeeperI defines a module interface that supports Gno
// smart contracts programming (scripting).
// TODO: tune this, remove unnecessary
type VMKeeperI interface {
	AddPackage(ctx sdk.Context, msg MsgAddPackage) error
	Call(ctx sdk.Context, msg MsgCall) (res string, err error)
	DispatchInternalMsg(GnoMsg)
	EventLoop(*sync.WaitGroup)
	SubmitTxFee(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error
	QueryEvalString(ctx sdk.Context, pkgPath string, expr string) (res string, err error)
	QueryFuncs(ctx sdk.Context, pkgPath string) (fsigs FunctionSignatures, err error)
	QueryEval(ctx sdk.Context, pkgPath string, expr string) (res string, err error)
	QueryFile(ctx sdk.Context, filepath string) (res string, err error)
}
