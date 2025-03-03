package vm

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/version"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var (
	tracer = otel.Tracer("vm_handler")
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
	ctx, span := traces.StartSpan(ctx, tracer, "vm_handler.Process")
	defer span.End()

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
	ctx, span := traces.StartSpan(ctx, tracer, "handleMsgAddPackage")
	span.SetAttributes(attribute.String("pkgPath", msg.Package.Path))
	defer span.End()

	err := vh.vm.AddPackage(ctx, msg)
	if err != nil {
		span.RecordError(err)
		return abciResult(err)
	}
	return sdk.Result{}
}

// Handle MsgCall.
func (vh vmHandler) handleMsgCall(ctx sdk.Context, msg MsgCall) (res sdk.Result) {
	ctx, span := traces.StartSpan(ctx, tracer, "handleMsgCall")
	span.SetAttributes(attribute.String("pkgPath", msg.PkgPath), attribute.String("func", msg.Func))
	defer span.End()

	resstr, err := vh.vm.Call(ctx, msg)
	if err != nil {
		span.RecordError(err)
		return abciResult(err)
	}
	res.Data = []byte(resstr)
	return
}

// Handle MsgRun.
func (vh vmHandler) handleMsgRun(ctx sdk.Context, msg MsgRun) (res sdk.Result) {
	ctx, span := traces.StartSpan(ctx, tracer, "handleMsgRun")
	span.SetAttributes(attribute.String("pkgPath", msg.Package.Path))
	defer span.End()

	resstr, err := vh.vm.Run(ctx, msg)
	if err != nil {
		span.RecordError(err)
		return abciResult(err)
	}
	res.Data = []byte(resstr)
	return
}

// ----------------------------------------
// Query

// query paths
const (
	QueryRender = "qrender"
	QueryFuncs  = "qfuncs"
	QueryEval   = "qeval"
	QueryFile   = "qfile"
	QueryDoc    = "qdoc"
)

func (vh vmHandler) Query(ctx sdk.Context, req abci.RequestQuery) abci.ResponseQuery {
	var (
		res  abci.ResponseQuery
		path = secondPart(req.Path)
	)

	switch path {
	case QueryRender:
		res = vh.queryRender(ctx, req)
	case QueryFuncs:
		res = vh.queryFuncs(ctx, req)
	case QueryEval:
		res = vh.queryEval(ctx, req)
	case QueryFile:
		res = vh.queryFile(ctx, req)
	case QueryDoc:
		res = vh.queryDoc(ctx, req)
	default:
		return sdk.ABCIResponseQueryFromError(
			std.ErrUnknownRequest(fmt.Sprintf(
				"unknown vm query endpoint %s in %s",
				secondPart(req.Path), req.Path)))
	}

	return res
}

// queryRender calls .Render(<path>) in readonly mode.
func (vh vmHandler) queryRender(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	reqData := string(req.Data)
	dot := strings.IndexByte(reqData, ':')
	if dot < 0 {
		panic("expected <pkgpath>:<path> syntax in query input data")
	}

	pkgPath, path := reqData[:dot], reqData[dot+1:]
	expr := fmt.Sprintf("Render(%q)", path)
	result, err := vh.vm.QueryEvalString(ctx, pkgPath, expr)
	if err != nil {
		if strings.Contains(err.Error(), "Render not declared") {
			err = NoRenderDeclError{}
		}
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

// queryEval evaluates any expression in readonly mode and returns the results.
func (vh vmHandler) queryEval(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	pkgPath, expr := parseQueryEvalData(string(req.Data))
	result, err := vh.vm.QueryEval(ctx, pkgPath, expr)
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

// queryDoc returns the JSON of the doc for a given pkgpath, suitable for printing
func (vh vmHandler) queryDoc(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	filepath := string(req.Data)
	jsonDoc, err := vh.vm.QueryDoc(ctx, filepath)
	if err != nil {
		res = sdk.ABCIResponseQueryFromError(err)
		return
	}
	res.Data = []byte(jsonDoc.JSON())
	return
}

// ----------------------------------------
// misc

func abciResult(err error) sdk.Result {
	res := sdk.ABCIResultFromError(err)
	res.Info += "vm.version=" + version.Version
	return res
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
