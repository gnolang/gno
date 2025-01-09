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
	case MsgEval:
		return vh.handleMsgEval(ctx, msg)
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

func (vh vmHandler) handleMsgCall(ctx sdk.Context, msg MsgCall) (res sdk.Result) {
	resstr, err := vh.vm.Call(ctx, msg)
	if err != nil {
		return abciResult(err)
	}

	res.Data = []byte(resstr)
	return res
}

// Handle MsgEval.
func (vh vmHandler) handleMsgEval(ctx sdk.Context, msg MsgEval) (res sdk.Result) {
	resstr, err := vh.vm.Eval(ctx, msg)
	if err != nil {
		return abciResult(err)
	}

	res.Data = []byte(resstr)
	return res
}

// Handle MsgRun.
func (vh vmHandler) handleMsgRun(ctx sdk.Context, msg MsgRun) (res sdk.Result) {
	resstr, err := vh.vm.Run(ctx, msg)
	if err != nil {
		return abciResult(err)
	}

	res.Data = []byte(resstr)
	return res
}

// ----------------------------------------
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

func (vh vmHandler) Query(ctx sdk.Context, req abci.RequestQuery) abci.ResponseQuery {
	var (
		res  abci.ResponseQuery
		path = secondPart(req.Path)
	)

	switch path {
	case QueryPackage:
		res = vh.queryPackage(ctx, req)
	case QueryStore:
		res = vh.queryStore(ctx, req)
	case QueryRender:
		res = vh.queryRender(ctx, req)
	case QueryFuncs:
		res = vh.queryFuncs(ctx, req)
	case QueryFile:
		res = vh.queryFile(ctx, req)
	case QueryEval:
		res = vh.queryEval(ctx, req)
	default:
		return sdk.ABCIResponseQueryFromError(
			std.ErrUnknownRequest(fmt.Sprintf(
				"unknown vm query endpoint %s in %s",
				secondPart(req.Path), req.Path)))
	}

	return res
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
	dot := strings.IndexByte(reqData, ':')
	if dot < 0 {
		panic("expected <pkgpath>:<path> syntax in query input data")
	}

	pkgPath, path := reqData[:dot], reqData[dot+1:]

	// Generate msg eval request
	expr := fmt.Sprintf("Render(%q)", path)
	msgEval := NewMsgEval(FormatString, pkgPath, expr)

	// Try evaluate `Render` function
	result, err := vh.vm.Eval(ctx, msgEval)
	if err != nil {
		res = sdk.ABCIResponseQueryFromError(err)
		return
	}
	res.Data = []byte(result)
	return
}

// queryFuncs returns public facing function signatures as JSON.
func (vh vmHandler) queryFuncs(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	pkgPath := string(req.Data)
	fsigs, err := vh.vm.QueryFuncs(ctx, pkgPath)
	if err != nil {
		res = sdk.ABCIResponseQueryFromError(err)
		return
	}
	res.Data = []byte(fsigs.JSON())
	return
}

// queryEval evaluates any expression in readonly mode and returns the results based on the given format.
func (vh vmHandler) queryEval(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	var format Format
	switch ss := strings.Split(req.Path, "/"); len(ss) {
	case 2:
		format = FormatDefault
	case 3:
		format = Format(ss[2])
	default:
		res = sdk.ABCIResponseQueryFromError(fmt.Errorf("invalid query"))
		return

	}

	switch format {
	case FormatMachine, FormatJSON, FormatString:
	default:
		err := fmt.Errorf("invalid query result format %q", format)
		res = sdk.ABCIResponseQueryFromError(err)
		return
	}

	pkgpath, expr := parseQueryEvalData(string(req.Data))
	fmt.Printf("-> %q %q %q\r\n", format, pkgpath, expr)
	msgEval := NewMsgEval(format, pkgpath, expr)
	if expr == "" {
		res = sdk.ABCIResponseQueryFromError(fmt.Errorf("expr cannot be empty"))
		return
	}

	result, err := vh.vm.Eval(ctx, msgEval)
	if err != nil {
		res = sdk.ABCIResponseQueryFromError(err)
		return
	}

	res.Data = []byte(result)
	return
}

// queryEvalJSON evaluates any expression in readonly mode and returns the results in JSON format.
func (vh vmHandler) queryEvalJSON(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	pkgath, expr := parseQueryEvalData(string(req.Data))
	msgEval := NewMsgEval(FormatJSON, pkgath, expr)
	result, err := vh.vm.Eval(ctx, msgEval)
	if err != nil {
		res = sdk.ABCIResponseQueryFromError(err)
		return
	}
	res.Data = []byte(result)
	return
}

// parseQueryEval parses the input string of vm/qeval. It takes the first dot
// after the first slash (if any) to separe the pkgPath and the expr.
// For instance, in gno.land/r/realm.MyFunction(), gno.land/r/realm is the
// pkgPath,and MyFunction() is the expr.
func parseQueryEvalData(data string) (pkgPath, expr string) {
	slash := strings.IndexByte(data, '/')
	if slash >= 0 {
		pkgPath += data[:slash]
		data = data[slash:]
	}
	dot := strings.IndexByte(data, '.')
	if dot < 0 {
		panic(panicInvalidQueryEvalData)
	}
	pkgPath += data[:dot]
	expr = data[dot+1:]
	return
}

const (
	panicInvalidQueryEvalData = "expected <pkgpath>.<expression> syntax in query input data"
)

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

// ----------------------------------------
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
