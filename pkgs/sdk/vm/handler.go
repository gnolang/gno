package vm

import (
	"fmt"
	"strings"

	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
	"github.com/gnolang/gno/pkgs/sdk"
	"github.com/gnolang/gno/pkgs/sdk/auth"
	"github.com/gnolang/gno/pkgs/std"
)

type vmHandler struct {
	vm VMKeeper
}

// NewHandler returns a handler for "vm" type messages.
func NewHandler(vm VMKeeper) vmHandler {
	return vmHandler{
		vm: vm,
	}
}

func (vh vmHandler) Process(ctx sdk.Context, msg std.Msg) sdk.Result {
	switch msg := msg.(type) {
	case MsgAddPackage:
		return vh.handleMsgAddPackage(ctx, msg)
	case MsgEval:
		return vh.handleMsgEval(ctx, msg)
	default:
		errMsg := fmt.Sprintf("unrecognized vm message type: %T", msg)
		return abciResult(std.ErrUnknownRequest(errMsg))
	}
}

// Handle MsgAddPackage.
func (vh vmHandler) handleMsgAddPackage(ctx sdk.Context, msg MsgAddPackage) sdk.Result {
	// TODO record write new vm to disk fs
	// TODO if already exists, vhoud fail. (one vhot write for now)
	// TODO deduct coins from user for payment.
	// TODO
	err := vh.vm.AddPackage(ctx, msg.Creator, msg.PkgPath, msg.Files)
	if err != nil {
		return abciResult(err)
	}
	amount, err := std.ParseCoins("1gnot") // XXX calculate
	if err != nil {
		return abciResult(err)
	}
	err = vh.vm.bank.SendCoins(ctx, msg.Creator, auth.FeeCollectorAddress(), amount)
	if err != nil {
		return abciResult(err)
	}
	return sdk.Result{}
}

// Handle MsgEval.
func (vh vmHandler) handleMsgEval(ctx sdk.Context, msg MsgEval) sdk.Result {
	// TODO create new machine
	// TODO with app common store
	// TODO
	amount, err := std.ParseCoins("1gnot") // XXX calculate
	if err != nil {
		return abciResult(err)
	}
	err = vh.vm.bank.SendCoins(ctx, msg.Caller, auth.FeeCollectorAddress(), amount)
	if err != nil {
		return abciResult(err)
	}
	/*
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute(sdk.AttributeKeyXXX, types.AttributeValueXXX),
			),
		)
	*/
	return sdk.Result{}
}

//----------------------------------------
// Query

// query package path
const QueryPackage = "package"

// query store path
const QueryStore = "store"

func (vh vmHandler) Query(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	switch queryPath(req.Path) {
	case QueryPackage:
		return vh.queryPackage(ctx, req)
	case QueryStore:
		return vh.queryStore(ctx, req)
	default:
		res.Error = sdk.ABCIError(
			std.ErrUnknownRequest("unknown vm query endpoint"))
		return
	}
}

// queryPackage fetch a package's files.
func (vh vmHandler) queryPackage(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	parts := strings.Split(req.Path, "/")
	if parts[0] != "vm" {
		panic("vhould not happen")
	}
	res.Data = []byte(fmt.Sprintf("TODO: parse parts get or make fileset..."))
	return
}

// queryPackage fetch items from the store.
func (vh vmHandler) queryStore(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	parts := strings.Split(req.Path, "/")
	if parts[0] != "vm" {
		panic("vhould not happen")
	}
	res.Data = []byte(fmt.Sprintf("TODO: parse parts get or make fileset..."))
	return
}

//----------------------------------------
// misc

func abciResult(err error) sdk.Result {
	return sdk.ABCIResultFromError(err)
}

// returns the second component of a query path.
func queryPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 2 {
		return ""
	} else {
		return parts[1]
	}
}
