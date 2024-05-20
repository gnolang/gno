package vm

import (
	"fmt"
	"strings"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type vmHandler struct {
	vm *VMKeeper
}

// NewHandler returns a handler for "vm" type messages.
func NewHandler(vm *VMKeeper) vmHandler {
	return vmHandler{
		vm: vm,
	}
}

func (vh vmHandler) Process(ctx sdk.Context, msg std.Msg) sdk.Result {
	switch msg := msg.(type) {
	case MsgAddPackage:
		return vh.handleMsgAddPackage(ctx, msg)
	case MsgCall:
		return vh.handleMsgCall(ctx, msg)
	case MsgRun:
		return vh.handleMsgRun(ctx, msg)
	default:
		errMsg := fmt.Sprintf("unrecognized vm message type: %T", msg)
		return abciResult(std.ErrUnknownRequest(errMsg))
	}
}

// Handle MsgAddPackage.
func (vh vmHandler) handleMsgAddPackage(ctx sdk.Context, msg MsgAddPackage) sdk.Result {
	err := vh.vm.AddPackage(ctx, msg)
	if err != nil {
		return abciResult(err)
	}
	return sdk.Result{}
}

// Handle MsgCall.
func (vh vmHandler) handleMsgCall(ctx sdk.Context, msg MsgCall) (res sdk.Result) {
	resstr, err := vh.vm.Call(ctx, msg)
	if err != nil {
		return abciResult(err)
	}
	res.Data = []byte(resstr)
	return
	/* TODO handle events.
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyXXX, types.AttributeValueXXX),
		),
	)
	*/
}

// Handle MsgRun.
func (vh vmHandler) handleMsgRun(ctx sdk.Context, msg MsgRun) (res sdk.Result) {
	resstr, err := vh.vm.Run(ctx, msg)
	if err != nil {
		return abciResult(err)
	}
	res.Data = []byte(resstr)
	return
}

//----------------------------------------
// Query

// query paths
const (
	QueryPackage = "package"
	QueryStore   = "store"
	QueryRender  = "qrender"
	QueryFuncs   = "qfuncs"
	QueryEval    = "qeval"
	QueryFile    = "qfile"
)

func (vh vmHandler) Query(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	switch secondPart(req.Path) {
	case QueryPackage:
		return vh.queryPackage(ctx, req)
	case QueryStore:
		return vh.queryStore(ctx, req)
	case QueryRender:
		return vh.queryRender(ctx, req)
	case QueryFuncs:
		return vh.queryFuncs(ctx, req)
	case QueryEval:
		return vh.queryEval(ctx, req)
	case QueryFile:
		return vh.queryFile(ctx, req)
	default:
		res = sdk.ABCIResponseQueryFromError(
			std.ErrUnknownRequest(fmt.Sprintf(
				"unknown vm query endpoint %s in %s",
				secondPart(req.Path), req.Path)))
		return
	}
}

// queryPackage fetch a package's files.
func (vh vmHandler) queryPackage(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	res.Data = []byte(fmt.Sprintf("TODO: parse parts get or make fileset..."))
	return
}

// queryPackage fetch items from the store.
func (vh vmHandler) queryStore(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	res.Data = []byte(fmt.Sprintf("TODO: fetch from store"))
	return
}

// queryRender calls .Render(<path>) in readonly mode.
func (vh vmHandler) queryRender(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	reqData := string(req.Data)
	reqParts := strings.Split(reqData, "\n")
	if len(reqParts) != 2 {
		panic("expected two lines in query input data")
	}
	pkgPath := reqParts[0]
	path := reqParts[1]
	expr := fmt.Sprintf("Render(%q)", path)
	result, err := vh.vm.QueryEvalString(ctx, pkgPath, expr)
	if err != nil {
		res = sdk.ABCIResponseQueryFromError(err)
		return
	}
	res.Data = []byte(result)
	return
}

// queryFuncs returns public facing function signatures as JSON.
func (vh vmHandler) queryFuncs(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	reqData := string(req.Data)
	reqParts := strings.Split(reqData, "\n")
	if len(reqParts) != 1 {
		panic("expected one line in query input data")
	}
	pkgPath := reqParts[0]
	fsigs, err := vh.vm.QueryFuncs(ctx, pkgPath)
	if err != nil {
		res = sdk.ABCIResponseQueryFromError(err)
		return
	}
	res.Data = []byte(fsigs.JSON())
	return
}

// queryEval evaluates any expression in readonly mode and returns the results.
func (vh vmHandler) queryEval(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
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

// queryFile returns the file bytes, or list of files if directory.
// if file, res.Value is []byte("file").
// if dir, res.Value is []byte("dir").
func (vh vmHandler) queryFile(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	filepath := string(req.Data)
	result, err := vh.vm.QueryFile(ctx, filepath)
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
		if parts[0] != "vm" {
			panic("should not happen")
		}
		return parts[1]
	}
}
