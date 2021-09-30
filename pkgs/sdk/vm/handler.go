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
	amount, err := std.ParseCoins("1gnot") // XXX calculate
	if err != nil {
		return abciResult(err)
	}
	err = vh.vm.bank.SendCoins(ctx, msg.Creator, auth.FeeCollectorAddress(), amount)
	if err != nil {
		return abciResult(err)
	}
	err = vh.vm.AddPackage(ctx, msg)
	if err != nil {
		return abciResult(err)
	}
	return sdk.Result{}
}

// Handle MsgEval.
func (vh vmHandler) handleMsgEval(ctx sdk.Context, msg MsgEval) (res sdk.Result) {
	amount, err := std.ParseCoins("1gnot") // XXX calculate
	if err != nil {
		return abciResult(err)
	}
	err = vh.vm.bank.SendCoins(ctx, msg.Caller, auth.FeeCollectorAddress(), amount)
	if err != nil {
		return abciResult(err)
	}
	out, err := vh.vm.Eval(ctx, msg)
	if err != nil {
		return abciResult(err)
	}
	res.Data = []byte(out)
	return
	/*
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute(sdk.AttributeKeyXXX, types.AttributeValueXXX),
			),
		)
	*/
}

//----------------------------------------
// Query

// query paths
const QueryPackage = "package"
const QueryStore = "store"
const QueryEval = "queryeval"

func (vh vmHandler) Query(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	switch secondPart(req.Path) {
	case QueryPackage:
		return vh.queryPackage(ctx, req)
	case QueryStore:
		return vh.queryStore(ctx, req)
	case QueryEval:
		return vh.queryEval(ctx, req)
	default:
		res = sdk.ABCIResponseQueryFromError(
			std.ErrUnknownRequest("unknown vm query endpoint"))
		return
	}
}

// queryPackage fetch a package's files.
func (vh vmHandler) queryPackage(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	parts := strings.Split(req.Path, "/")
	if parts[0] != "vm" {
		panic("should not happen")
	}
	res.Data = []byte(fmt.Sprintf("TODO: parse parts get or make fileset..."))
	return
}

// queryPackage fetch items from the store.
func (vh vmHandler) queryStore(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	parts := strings.Split(req.Path, "/")
	if parts[0] != "vm" {
		panic("should not happen")
	}
	res.Data = []byte(fmt.Sprintf("TODO: fetch from store"))
	return
}

// queryEval evaluates a pure readonly expression.
func (vh vmHandler) queryEval(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	parts := strings.Split(req.Path, "/")
	if parts[0] != "vm" {
		panic("should not happen")
	}
	reqData := string(req.Data)
	reqParts := strings.Split(reqData, "\n")
	if len(reqParts) != 2 {
		panic("expected two lines in query input data")
	}
	pkgPath := reqParts[0]
	expr := reqParts[1]
	result, err := vh.vm.QueryEval(ctx, pkgPath, expr)
	if err != nil {
		res = sdk.ABCIResponseQueryFromError(err)
		return
	}
	res.Data = []byte(result)
	return
}

//----------------------------------------
// misc

func abciResult(err error) sdk.Result {
	return sdk.ABCIResultFromError(err)
}

// returns the second component of a path.
func secondPart(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return ""
	} else {
		return parts[1]
	}
}
