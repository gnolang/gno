package vm

import (
	"context"
	"fmt"
	"strings"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/telemetry"
	"github.com/gnolang/gno/tm2/pkg/telemetry/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
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

func (vh vmHandler) Process(msg std.Msg) sdk.Result {
	switch msg := msg.(type) {
	case MsgAddPackage:
		return vh.handleMsgAddPackage(msg)
	case MsgCall:
		return vh.handleMsgCall(msg)
	case MsgRun:
		return vh.handleMsgRun(msg)
	default:
		errMsg := fmt.Sprintf("unrecognized vm message type: %T", msg)
		return abciResult(std.ErrUnknownRequest(errMsg))
	}
}

// Handle MsgAddPackage.
func (vh vmHandler) handleMsgAddPackage(msg MsgAddPackage) sdk.Result {
	err := vh.vm.AddPackage(msg)
	if err != nil {
		return abciResult(err)
	}
	return sdk.Result{}
}

// Handle MsgCall.
func (vh vmHandler) handleMsgCall(msg MsgCall) (res sdk.Result) {
	resstr, err := vh.vm.Call(msg)
	if err != nil {
		return abciResult(err)
	}
	res.Data = []byte(resstr)
	return
}

// Handle MsgRun.
func (vh vmHandler) handleMsgRun(msg MsgRun) (res sdk.Result) {
	resstr, err := vh.vm.Run(msg)
	if err != nil {
		return abciResult(err)
	}
	res.Data = []byte(resstr)
	return
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

func (vh vmHandler) Query(req abci.RequestQuery) abci.ResponseQuery {
	var (
		res  abci.ResponseQuery
		path = secondPart(req.Path)
	)

	switch path {
	case QueryPackage:
		res = vh.queryPackage(req)
	case QueryStore:
		res = vh.queryStore(req)
	case QueryRender:
		res = vh.queryRender(req)
	case QueryFuncs:
		res = vh.queryFuncs(req)
	case QueryEval:
		res = vh.queryEval(req)
	case QueryFile:
		res = vh.queryFile(req)
	default:
		return sdk.ABCIResponseQueryFromError(
			std.ErrUnknownRequest(fmt.Sprintf(
				"unknown vm query endpoint %s in %s",
				secondPart(req.Path), req.Path)))
	}

	logQueryTelemetry(path, res.IsErr())

	return res
}

// logQueryTelemetry logs the relevant VM query telemetry
func logQueryTelemetry(path string, isErr bool) {
	if !telemetry.MetricsEnabled() {
		return
	}

	metrics.VMQueryCalls.Add(
		context.Background(),
		1,
		metric.WithAttributes(
			attribute.KeyValue{
				Key:   "path",
				Value: attribute.StringValue(path),
			},
		),
	)

	if isErr {
		metrics.VMQueryErrors.Add(context.Background(), 1)
	}
}

// queryPackage fetch a package's files.
func (vh vmHandler) queryPackage(req abci.RequestQuery) (res abci.ResponseQuery) {
	res.Data = []byte(fmt.Sprintf("TODO: parse parts get or make fileset..."))
	return
}

// queryPackage fetch items from the store.
func (vh vmHandler) queryStore(req abci.RequestQuery) (res abci.ResponseQuery) {
	res.Data = []byte(fmt.Sprintf("TODO: fetch from store"))
	return
}

// queryRender calls .Render(<path>) in readonly mode.
func (vh vmHandler) queryRender(req abci.RequestQuery) (res abci.ResponseQuery) {
	reqData := string(req.Data)
	reqParts := strings.Split(reqData, "\n")
	if len(reqParts) != 2 {
		panic("expected two lines in query input data")
	}
	pkgPath := reqParts[0]
	path := reqParts[1]
	expr := fmt.Sprintf("Render(%q)", path)
	result, err := vh.vm.QueryEvalString(pkgPath, expr)
	if err != nil {
		res = sdk.ABCIResponseQueryFromError(err)
		return
	}
	res.Data = []byte(result)
	return
}

// queryFuncs returns public facing function signatures as JSON.
func (vh vmHandler) queryFuncs(req abci.RequestQuery) (res abci.ResponseQuery) {
	reqData := string(req.Data)
	reqParts := strings.Split(reqData, "\n")
	if len(reqParts) != 1 {
		panic("expected one line in query input data")
	}
	pkgPath := reqParts[0]
	fsigs, err := vh.vm.QueryFuncs(pkgPath)
	if err != nil {
		res = sdk.ABCIResponseQueryFromError(err)
		return
	}
	res.Data = []byte(fsigs.JSON())
	return
}

// queryEval evaluates any expression in readonly mode and returns the results.
func (vh vmHandler) queryEval(req abci.RequestQuery) (res abci.ResponseQuery) {
	reqData := string(req.Data)
	reqParts := strings.Split(reqData, "\n")
	if len(reqParts) != 2 {
		panic("expected two lines in query input data")
	}
	pkgPath := reqParts[0]
	expr := reqParts[1]
	result, err := vh.vm.QueryEval(pkgPath, expr)
	if err != nil {
		res = sdk.ABCIResponseQueryFromError(err)
		return
	}
	res.Data = []byte(result)
	return
}

// queryFile returns the file bytes, or list of files if directory.
func (vh vmHandler) queryFile(req abci.RequestQuery) (res abci.ResponseQuery) {
	filepath := string(req.Data)
	result, err := vh.vm.QueryFile(filepath)
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
