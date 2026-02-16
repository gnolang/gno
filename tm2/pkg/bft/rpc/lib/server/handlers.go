package rpcserver

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	goerrors "errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gnolang/gno/tm2/pkg/telemetry"
	"github.com/gnolang/gno/tm2/pkg/telemetry/metrics"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
	"github.com/gorilla/websocket"

	"github.com/gnolang/gno/tm2/pkg/amino"
	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/service"
)

// RegisterRPCFuncs adds a route for each function in the funcMap, as well as general jsonrpc and websocket handlers for all functions.
// "result" is the interface on which the result objects are registered, and is populated with every RPCResponse
func RegisterRPCFuncs(mux *http.ServeMux, funcMap map[string]*RPCFunc, logger *slog.Logger) {
	// Check if metrics are enabled
	if telemetry.MetricsEnabled() {
		// HTTP endpoints
		for funcName, rpcFunc := range funcMap {
			mux.HandleFunc(
				"/"+funcName,
				telemetryMiddleware(
					makeHTTPHandler(rpcFunc, logger),
				),
			)
		}

		// JSONRPC endpoints
		mux.HandleFunc(
			"/",
			telemetryMiddleware(
				handleInvalidJSONRPCPaths(makeJSONRPCHandler(funcMap, logger)),
			),
		)

		return
	}

	// HTTP endpoints
	for funcName, rpcFunc := range funcMap {
		mux.HandleFunc("/"+funcName, makeHTTPHandler(rpcFunc, logger))
	}

	// JSONRPC endpoints
	mux.HandleFunc("/", handleInvalidJSONRPCPaths(makeJSONRPCHandler(funcMap, logger)))
}

// -------------------------------------
// function introspection

// RPCFunc contains the introspected type information for a function
type RPCFunc struct {
	f        reflect.Value  // underlying rpc function
	args     []reflect.Type // type of each function arg
	returns  []reflect.Type // type of each return arg
	argNames []string       // name of each argument
	ws       bool           // websocket only
}

// NewRPCFunc wraps a function for introspection.
// f is the function, args are comma separated argument names
func NewRPCFunc(f any, args string) *RPCFunc {
	return newRPCFunc(f, args, false)
}

// NewWSRPCFunc wraps a function for introspection and use in the websockets.
func NewWSRPCFunc(f any, args string) *RPCFunc {
	return newRPCFunc(f, args, true)
}

func newRPCFunc(f any, args string, ws bool) *RPCFunc {
	var argNames []string
	if args != "" {
		argNames = strings.Split(args, ",")
	}
	return &RPCFunc{
		f:        reflect.ValueOf(f),
		args:     funcArgTypes(f),
		returns:  funcReturnTypes(f),
		argNames: argNames,
		ws:       ws,
	}
}

// return a function's argument types
func funcArgTypes(f any) []reflect.Type {
	t := reflect.TypeOf(f)
	n := t.NumIn()
	typez := make([]reflect.Type, n)
	for i := range n {
		typez[i] = t.In(i)
	}
	return typez
}

// return a function's return types
func funcReturnTypes(f any) []reflect.Type {
	t := reflect.TypeOf(f)
	n := t.NumOut()
	typez := make([]reflect.Type, n)
	for i := range n {
		typez[i] = t.Out(i)
	}
	return typez
}

// function introspection
// -----------------------------------------------------------------------------
// rpc.json

// jsonrpc calls grab the given method's function info and runs reflect.Call
func makeJSONRPCHandler(funcMap map[string]*RPCFunc, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, span := traces.Tracer().Start(r.Context(), traceMakeJSONRPCHandler)
		defer span.End()
		b, err := io.ReadAll(r.Body)
		if err != nil {
			WriteRPCResponseHTTP(w, types.RPCInvalidRequestError(types.JSONRPCStringID(""), errors.Wrap(err, "error reading request body")))
			return
		}
		// if its an empty request (like from a browser),
		// just display a list of functions
		if len(b) == 0 {
			writeListOfEndpoints(w, r, funcMap)
			return
		}

		// --- Branch 1: Attempt to Unmarshal as a Batch (Slice) of Requests ---
		var requests types.RPCRequests
		if err := json.Unmarshal(b, &requests); err == nil {
			var responses types.RPCResponses
			for _, req := range requests {
				if resp := processRequest(r, req, funcMap, logger); resp != nil {
					responses = append(responses, *resp)
				}
			}

			if len(responses) > 0 {
				WriteRPCResponseArrayHTTP(w, responses)
				return
			}
		}

		// --- Branch 2: Attempt to Unmarshal as a Single Request ---
		var request types.RPCRequest
		if err := json.Unmarshal(b, &request); err == nil {
			if resp := processRequest(r, request, funcMap, logger); resp != nil {
				WriteRPCResponseHTTP(w, *resp)
				return
			}
		} else {
			WriteRPCResponseHTTP(w, types.RPCParseError(types.JSONRPCStringID(""), errors.Wrap(err, "error unmarshalling request")))
			return
		}
	}
}

// processRequest checks and processes a single JSON-RPC request.
// If the request should produce a response, it returns a pointer to that response.
// Otherwise (e.g. if the request is a notification or fails validation), it returns nil.
func processRequest(r *http.Request, req types.RPCRequest, funcMap map[string]*RPCFunc, logger *slog.Logger) *types.RPCResponse {
	_, span := traces.Tracer().Start(r.Context(), "processRequest")
	defer span.End()
	// Skip notifications (an empty ID indicates no response should be sent)
	if req.ID == types.JSONRPCStringID("") {
		logger.Debug("Skipping notification (empty ID)")
		return nil
	}

	// Check that the URL path is valid (assume only "/" is acceptable)
	if len(r.URL.Path) > 1 {
		resp := types.RPCInvalidRequestError(req.ID, fmt.Errorf("invalid path: %s", r.URL.Path))
		return &resp
	}

	// Look up the requested method in the function map.
	rpcFunc, ok := funcMap[req.Method]
	if !ok || rpcFunc.ws {
		resp := types.RPCMethodNotFoundError(req.ID)
		return &resp
	}

	ctx := &types.Context{JSONReq: &req, HTTPReq: r}
	args := []reflect.Value{reflect.ValueOf(ctx)}
	if len(req.Params) > 0 {
		fnArgs, err := jsonParamsToArgs(rpcFunc, req.Params)
		if err != nil {
			resp := types.RPCInvalidParamsError(req.ID, errors.Wrap(err, "error converting json params to arguments"))
			return &resp
		}
		args = append(args, fnArgs...)
	}

	// Call the RPC function using reflection.
	returns := rpcFunc.f.Call(args)
	logger.Info("HTTPJSONRPC", "method", req.Method, "args", args, "returns", returns)

	// Convert the reflection return values into a result value for JSON serialization.
	result, err := unreflectResult(returns)
	if err != nil {
		resp := types.RPCInternalError(req.ID, err)
		return &resp
	}

	// Build and return a successful response.
	resp := types.NewRPCSuccessResponse(req.ID, result)
	return &resp
}

func handleInvalidJSONRPCPaths(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Since the pattern "/" matches all paths not matched by other registered patterns we check whether the path is indeed
		// "/", otherwise return a 404 error
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		next(w, r)
	}
}

// telemetryMiddleware is the telemetry middleware handler
func telemetryMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		// Log the response time
		metrics.HTTPRequestTime.Record(
			context.Background(),
			time.Since(start).Milliseconds(),
		)
	}
}

func mapParamsToArgs(rpcFunc *RPCFunc, params map[string]json.RawMessage, argsOffset int) ([]reflect.Value, error) {
	values := make([]reflect.Value, len(rpcFunc.argNames))
	for i, argName := range rpcFunc.argNames {
		argType := rpcFunc.args[i+argsOffset]

		if p, ok := params[argName]; ok && p != nil && len(p) > 0 {
			val := reflect.New(argType)
			err := amino.UnmarshalJSON(p, val.Interface())
			if err != nil {
				return nil, err
			}
			values[i] = val.Elem()
		} else { // use default for that type
			values[i] = reflect.Zero(argType)
		}
	}

	return values, nil
}

func arrayParamsToArgs(rpcFunc *RPCFunc, params []json.RawMessage, argsOffset int) ([]reflect.Value, error) {
	if len(rpcFunc.argNames) != len(params) {
		return nil, errors.New("expected %v parameters (%v), got %v (%v)",
			len(rpcFunc.argNames), rpcFunc.argNames, len(params), params)
	}

	values := make([]reflect.Value, len(params))
	for i, p := range params {
		argType := rpcFunc.args[i+argsOffset]
		val := reflect.New(argType)
		err := amino.UnmarshalJSON(p, val.Interface())
		if err != nil {
			return nil, err
		}
		values[i] = val.Elem()
	}
	return values, nil
}

// raw is unparsed json (from json.RawMessage) encoding either a map or an
// array.
//
// Example:
//
//	rpcFunc.args = [rpctypes.Context string]
//	rpcFunc.argNames = ["arg"]
func jsonParamsToArgs(rpcFunc *RPCFunc, raw []byte) ([]reflect.Value, error) {
	const argsOffset = 1

	// TODO: Make more efficient, perhaps by checking the first character for '{' or '['?
	// First, try to get the map.
	var m map[string]json.RawMessage
	err := json.Unmarshal(raw, &m)
	if err == nil {
		return mapParamsToArgs(rpcFunc, m, argsOffset)
	}

	// Otherwise, try an array.
	var a []json.RawMessage
	err = json.Unmarshal(raw, &a)
	if err == nil {
		return arrayParamsToArgs(rpcFunc, a, argsOffset)
	}

	// Otherwise, bad format, we cannot parse
	return nil, errors.New("unknown type for JSON params: %v. Expected map or array", err)
}

// rpc.json
// -----------------------------------------------------------------------------
// rpc.http

// convert from a function name to the http handler
func makeHTTPHandler(rpcFunc *RPCFunc, logger *slog.Logger) http.HandlerFunc {
	// Exception for websocket endpoints
	if rpcFunc.ws {
		return func(w http.ResponseWriter, r *http.Request) {
			WriteRPCResponseHTTP(w, types.RPCMethodNotFoundError(types.JSONRPCStringID("")))
		}
	}

	// All other endpoints
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("HTTP HANDLER", "req", r)
		_, span := traces.Tracer().Start(r.Context(), traceMakeHTTPHandler)
		defer span.End()

		ctx := &types.Context{HTTPReq: r}
		args := []reflect.Value{reflect.ValueOf(ctx)}

		fnArgs, err := httpParamsToArgs(rpcFunc, r)
		if err != nil {
			WriteRPCResponseHTTP(w, types.RPCInvalidParamsError(types.JSONRPCStringID(""), errors.Wrap(err, "error converting http params to arguments")))
			return
		}
		args = append(args, fnArgs...)

		returns := rpcFunc.f.Call(args)

		logger.Info("HTTPRestRPC", "method", r.URL.Path, "args", args, "returns", returns)
		result, err := unreflectResult(returns)
		if err != nil {
			var statusErr *types.HTTPStatusError
			if goerrors.As(err, &statusErr) {
				WriteRPCResponseHTTPError(w, statusErr.Code, types.RPCInternalError(types.JSONRPCStringID(""), err))
				return
			}

			WriteRPCResponseHTTP(w, types.RPCInternalError(types.JSONRPCStringID(""), err))
			return
		}
		WriteRPCResponseHTTP(w, types.NewRPCSuccessResponse(types.JSONRPCStringID(""), result))
	}
}

// Convert an http query to a list of properly typed values.
// To be properly decoded the arg must be a concrete type from tendermint (if its an interface).
func httpParamsToArgs(rpcFunc *RPCFunc, r *http.Request) ([]reflect.Value, error) {
	const argsOffset = 1

	paramsMap := make(map[string]json.RawMessage)
	for _, argName := range rpcFunc.argNames {
		arg := GetParam(r, argName)
		if arg == "" {
			// Empty param
			continue
		}

		// Handle hex string
		if strings.HasPrefix(arg, "0x") {
			decoded, err := hex.DecodeString(arg[2:])
			if err != nil {
				return nil, fmt.Errorf("unable to decode hex string: %w", err)
			}

			data, err := amino.MarshalJSON(decoded)
			if err != nil {
				return nil, fmt.Errorf("unable to marshal argument to JSON: %w", err)
			}

			paramsMap[argName] = data

			continue
		}

		// Handle integer string by adding quotes to ensure it is treated as a JSON string.
		// This is required by Amino JSON to unmarshal values into integers.
		if _, err := strconv.Atoi(arg); err == nil {
			// arg is a number, wrap it
			arg = "\"" + arg + "\""
		}

		// Handle invalid JSON: ensure it's wrapped as a JSON-encoded string
		if !json.Valid([]byte(arg)) {
			data, err := amino.MarshalJSON(arg)
			if err != nil {
				return nil, fmt.Errorf("unable to marshal argument to JSON: %w", err)
			}

			paramsMap[argName] = data

			continue
		}

		// Default: treat the argument as a JSON raw message
		paramsMap[argName] = json.RawMessage([]byte(arg))
	}

	return mapParamsToArgs(rpcFunc, paramsMap, argsOffset)
}

// rpc.http
// -----------------------------------------------------------------------------
// rpc.websocket

const (
	defaultWSWriteChanCapacity = 1000
	defaultWSWriteWait         = 10 * time.Second
	defaultWSReadWait          = 30 * time.Second
	defaultWSPingPeriod        = (defaultWSReadWait * 9) / 10
)

// A single websocket connection contains listener id, underlying ws
// connection.
//
// In case of an error, the connection is stopped.
type wsConnection struct {
	service.BaseService

	remoteAddr string
	baseConn   *websocket.Conn
	writeChan  chan types.RPCResponses

	funcMap map[string]*RPCFunc

	// write channel capacity
	writeChanCapacity int

	// each write times out after this.
	writeWait time.Duration

	// Connection times out if we haven't received *anything* in this long, not even pings.
	readWait time.Duration

	// Send pings to server with this period. Must be less than readWait, but greater than zero.
	pingPeriod time.Duration

	// Maximum message size.
	readLimit int64

	// callback which is called upon disconnect
	onDisconnect func(remoteAddr string)

	ctx    context.Context
	cancel context.CancelFunc
}

// NewWSConnection wraps websocket.Conn.
//
// See the commentary on the func(*wsConnection) functions for a detailed
// description of how to configure ping period and pong wait time. NOTE: if the
// write buffer is full, pongs may be dropped, which may cause clients to
// disconnect. see https://github.com/gorilla/websocket/issues/97
func NewWSConnection(
	baseConn *websocket.Conn,
	funcMap map[string]*RPCFunc,
	options ...func(*wsConnection),
) *wsConnection {
	wsc := &wsConnection{
		remoteAddr:        baseConn.RemoteAddr().String(),
		baseConn:          baseConn,
		funcMap:           funcMap,
		writeWait:         defaultWSWriteWait,
		writeChanCapacity: defaultWSWriteChanCapacity,
		readWait:          defaultWSReadWait,
		pingPeriod:        defaultWSPingPeriod,
	}
	for _, option := range options {
		option(wsc)
	}
	wsc.baseConn.SetReadLimit(wsc.readLimit)
	wsc.BaseService = *service.NewBaseService(nil, "wsConnection", wsc)
	return wsc
}

// OnDisconnect sets a callback which is used upon disconnect - not
// Goroutine-safe. Nop by default.
func OnDisconnect(onDisconnect func(remoteAddr string)) func(*wsConnection) {
	return func(wsc *wsConnection) {
		wsc.onDisconnect = onDisconnect
	}
}

// WriteWait sets the amount of time to wait before a websocket write times out.
// It should only be used in the constructor - not Goroutine-safe.
func WriteWait(writeWait time.Duration) func(*wsConnection) {
	return func(wsc *wsConnection) {
		wsc.writeWait = writeWait
	}
}

// WriteChanCapacity sets the capacity of the websocket write channel.
// It should only be used in the constructor - not Goroutine-safe.
func WriteChanCapacity(capacity int) func(*wsConnection) {
	return func(wsc *wsConnection) {
		wsc.writeChanCapacity = capacity
	}
}

// ReadWait sets the amount of time to wait before a websocket read times out.
// It should only be used in the constructor - not Goroutine-safe.
func ReadWait(readWait time.Duration) func(*wsConnection) {
	return func(wsc *wsConnection) {
		wsc.readWait = readWait
	}
}

// PingPeriod sets the duration for sending websocket pings.
// It should only be used in the constructor - not Goroutine-safe.
func PingPeriod(pingPeriod time.Duration) func(*wsConnection) {
	return func(wsc *wsConnection) {
		wsc.pingPeriod = pingPeriod
	}
}

// ReadLimit sets the maximum size for reading message.
// It should only be used in the constructor - not Goroutine-safe.
func ReadLimit(readLimit int64) func(*wsConnection) {
	return func(wsc *wsConnection) {
		wsc.readLimit = readLimit
	}
}

// OnStart implements service.Service by starting the read and write routines. It
// blocks until the connection closes.
func (wsc *wsConnection) OnStart() error {
	wsc.writeChan = make(chan types.RPCResponses, wsc.writeChanCapacity)

	// Read subscriptions/unsubscriptions to events
	go wsc.readRoutine()
	// Write responses, BLOCKING.
	wsc.writeRoutine()

	return nil
}

// OnStop implements service.Service by unsubscribing remoteAddr from all subscriptions.
func (wsc *wsConnection) OnStop() {
	// Both read and write loops close the websocket connection when they exit their loops.
	// The writeChan is never closed, to allow WriteRPCResponses() to fail.

	if wsc.onDisconnect != nil {
		wsc.onDisconnect(wsc.remoteAddr)
	}

	if wsc.ctx != nil {
		wsc.cancel()
	}
}

// GetRemoteAddr returns the remote address of the underlying connection.
// It implements WSRPCConnection
func (wsc *wsConnection) GetRemoteAddr() string {
	return wsc.remoteAddr
}

// WriteRPCResponse pushes a response to the writeChan, and blocks until it is accepted.
// It implements WSRPCConnection. It is Goroutine-safe.
func (wsc *wsConnection) WriteRPCResponses(resp types.RPCResponses) {
	select {
	case <-wsc.Quit():
		return
	case wsc.writeChan <- resp:
	}
}

// TryWriteRPCResponse attempts to push a response to the writeChan, but does not block.
// It implements WSRPCConnection. It is Goroutine-safe
func (wsc *wsConnection) TryWriteRPCResponses(resp types.RPCResponses) bool {
	select {
	case <-wsc.Quit():
		return false
	case wsc.writeChan <- resp:
		return true
	default:
		return false
	}
}

// Context returns the connection's context.
// The context is canceled when the client's connection closes.
func (wsc *wsConnection) Context() context.Context {
	if wsc.ctx != nil {
		return wsc.ctx
	}
	wsc.ctx, wsc.cancel = context.WithCancel(context.Background())
	return wsc.ctx
}

// Read from the socket and subscribe to or unsubscribe from events
func (wsc *wsConnection) readRoutine() {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("WSJSONRPC: %v", r)
			}
			wsc.Logger.Error("Panic in WSJSONRPC handler", "err", err, "stack", string(debug.Stack()))
			wsc.WriteRPCResponses(types.RPCResponses{types.RPCInternalError(types.JSONRPCStringID("unknown"), err)})
			go wsc.readRoutine()
		} else {
			wsc.baseConn.Close() //nolint: errcheck
		}
	}()

	wsc.baseConn.SetPongHandler(func(m string) error {
		return wsc.baseConn.SetReadDeadline(time.Now().Add(wsc.readWait))
	})

	telemetryEnabled := telemetry.MetricsEnabled()

	for {
		select {
		case <-wsc.Quit():
			return
		default:
			// reset deadline for every type of message (control or data)
			if err := wsc.baseConn.SetReadDeadline(time.Now().Add(wsc.readWait)); err != nil {
				wsc.Logger.Error("failed to set read deadline", "err", err)
			}
			var in []byte
			_, in, err := wsc.baseConn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					wsc.Logger.Info("Client closed the connection")
				} else {
					wsc.Logger.Error("Failed to read request", "err", err)
				}
				wsc.Stop()
				return
			}

			// Log the request response start time
			responseStart := time.Now()

			// first try to unmarshal the incoming request as an array of RPC requests
			var (
				requests  types.RPCRequests
				responses types.RPCResponses
			)

			// Try to unmarshal the requests as a batch
			if err := json.Unmarshal(in, &requests); err != nil {
				// Next, try to unmarshal as a single request
				var request types.RPCRequest
				if err := json.Unmarshal(in, &request); err != nil {
					wsc.WriteRPCResponses(
						types.RPCResponses{
							types.RPCParseError(
								types.JSONRPCStringID(""),
								errors.Wrap(err, "error unmarshalling request"),
							),
						},
					)

					return
				}

				requests = []types.RPCRequest{request}
			}

			for _, request := range requests {
				request := request

				// A Notification is a Request object without an "id" member.
				// The Server MUST NOT reply to a Notification, including those that are within a batch request.
				if request.ID == types.JSONRPCStringID("") {
					wsc.Logger.Debug("Skipping notification JSON-RPC request")

					continue
				}

				// Now, fetch the RPCFunc and execute it.
				rpcFunc := wsc.funcMap[request.Method]
				if rpcFunc == nil {
					responses = append(responses, types.RPCMethodNotFoundError(request.ID))

					continue
				}

				ctx := &types.Context{JSONReq: &request, WSConn: wsc}
				args := []reflect.Value{reflect.ValueOf(ctx)}
				if len(request.Params) > 0 {
					fnArgs, err := jsonParamsToArgs(rpcFunc, request.Params)
					if err != nil {
						responses = append(responses, types.RPCInternalError(request.ID, errors.Wrap(err, "error converting json params to arguments")))

						continue
					}
					args = append(args, fnArgs...)
				}

				returns := rpcFunc.f.Call(args)

				// TODO: Need to encode args/returns to string if we want to log them
				wsc.Logger.Info("WSJSONRPC", "method", request.Method)

				result, err := unreflectResult(returns)
				if err != nil {
					responses = append(responses, types.RPCInternalError(request.ID, err))

					continue
				}

				responses = append(responses, types.NewRPCSuccessResponse(request.ID, result))

				if len(responses) > 0 {
					wsc.WriteRPCResponses(responses)

					// Log telemetry
					if telemetryEnabled {
						metrics.WSRequestTime.Record(
							context.Background(),
							time.Since(responseStart).Milliseconds(),
						)
					}
				}
			}
		}
	}
}

// receives on a write channel and writes out on the socket
func (wsc *wsConnection) writeRoutine() {
	pingTicker := time.NewTicker(wsc.pingPeriod)
	defer func() {
		pingTicker.Stop()
		if err := wsc.baseConn.Close(); err != nil {
			wsc.Logger.Error("Error closing connection", "err", err)
		}
	}()

	// https://github.com/gorilla/websocket/issues/97
	pongs := make(chan string, 1)
	wsc.baseConn.SetPingHandler(func(m string) error {
		select {
		case pongs <- m:
		default:
		}
		return nil
	})

	for {
		select {
		case m := <-pongs:
			err := wsc.writeMessageWithDeadline(websocket.PongMessage, []byte(m))
			if err != nil {
				wsc.Logger.Info("Failed to write pong (client may disconnect)", "err", err)
			}
		case <-pingTicker.C:
			err := wsc.writeMessageWithDeadline(websocket.PingMessage, []byte{})
			if err != nil {
				wsc.Logger.Error("Failed to write ping", "err", err)
				wsc.Stop()
				return
			}
		case msgs := <-wsc.writeChan:
			var writeData any

			if len(msgs) == 1 {
				writeData = msgs[0]
			} else {
				writeData = msgs
			}

			jsonBytes, err := json.MarshalIndent(writeData, "", "  ")
			if err != nil {
				wsc.Logger.Error("Failed to marshal RPCResponse to JSON", "err", err)
			} else if err = wsc.writeMessageWithDeadline(websocket.TextMessage, jsonBytes); err != nil {
				wsc.Logger.Error("Failed to write response", "err", err)
				wsc.Stop()
				return
			}
		case <-wsc.Quit():
			return
		}
	}
}

// All writes to the websocket must (re)set the write deadline.
// If some writes don't set it while others do, they may timeout incorrectly (https://github.com/tendermint/tendermint/issues/553)
func (wsc *wsConnection) writeMessageWithDeadline(msgType int, msg []byte) error {
	if err := wsc.baseConn.SetWriteDeadline(time.Now().Add(wsc.writeWait)); err != nil {
		return err
	}
	return wsc.baseConn.WriteMessage(msgType, msg)
}

// ----------------------------------------

// WebsocketManager provides a WS handler for incoming connections and passes a
// map of functions along with any additional params to new connections.
// NOTE: The websocket path is defined externally, e.g. in node/node.go
type WebsocketManager struct {
	websocket.Upgrader

	funcMap       map[string]*RPCFunc
	logger        *slog.Logger
	wsConnOptions []func(*wsConnection)
}

// NewWebsocketManager returns a new WebsocketManager that passes a map of
// functions, connection options and logger to new WS connections.
func NewWebsocketManager(funcMap map[string]*RPCFunc, wsConnOptions ...func(*wsConnection)) *WebsocketManager {
	return &WebsocketManager{
		funcMap: funcMap,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// TODO ???
				return true
			},
		},
		logger:        log.NewNoopLogger(),
		wsConnOptions: wsConnOptions,
	}
}

// SetLogger sets the logger.
func (wm *WebsocketManager) SetLogger(l *slog.Logger) {
	wm.logger = l
}

// WebsocketHandler upgrades the request/response (via http.Hijack) and starts
// the wsConnection.
func (wm *WebsocketManager) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	wsConn, err := wm.Upgrade(w, r, nil)
	if err != nil {
		// TODO - return http error
		wm.logger.Error("Failed to upgrade to websocket connection", "err", err)
		return
	}

	// register connection
	con := NewWSConnection(wsConn, wm.funcMap, wm.wsConnOptions...)
	con.SetLogger(wm.logger.With("remote", wsConn.RemoteAddr()))
	wm.logger.Info("New websocket connection", "remote", con.remoteAddr)
	err = con.Start() // Blocking
	if err != nil {
		wm.logger.Error("Error starting connection", "err", err)
	}
}

// rpc.websocket
// -----------------------------------------------------------------------------

// NOTE: assume returns is result struct and error. If error is not nil, return it
func unreflectResult(returns []reflect.Value) (any, error) {
	errV := returns[1]
	if errVI := errV.Interface(); errVI != nil {
		return nil, errors.NewWithData(errVI)
	}
	rv := returns[0]
	// If the result is a registered interface, we need a pointer to it so
	// we can marshal with type info.
	if rv.Kind() == reflect.Interface {
		rvp := reflect.New(rv.Type())
		rvp.Elem().Set(rv)
		return rvp.Interface(), nil
	} else {
		return rv.Interface(), nil
	}
}

// writes a list of available rpc endpoints as an html page
func writeListOfEndpoints(w http.ResponseWriter, r *http.Request, funcMap map[string]*RPCFunc) {
	noArgNames := []string{}
	argNames := []string{}
	for name, funcData := range funcMap {
		if len(funcData.args) == 0 {
			noArgNames = append(noArgNames, name)
		} else {
			argNames = append(argNames, name)
		}
	}
	sort.Strings(noArgNames)
	sort.Strings(argNames)
	buf := new(bytes.Buffer)
	buf.WriteString("<html><body>")
	buf.WriteString("<br>Available endpoints:<br>")

	for _, name := range noArgNames {
		link := fmt.Sprintf("//%s/%s", r.Host, name)
		fmt.Fprintf(buf, "<a href=\"%s\">%s</a></br>", link, link)
	}

	buf.WriteString("<br>Endpoints that require arguments:<br>")
	for _, name := range argNames {
		link := fmt.Sprintf("//%s/%s?", r.Host, name)
		funcData := funcMap[name]
		for i, argName := range funcData.argNames {
			link += argName + "=_"
			if i < len(funcData.argNames)-1 {
				link += "&"
			}
		}
		fmt.Fprintf(buf, "<a href=\"%s\">%s</a></br>", link, link)
	}
	buf.WriteString("</body></html>")
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	w.Write(buf.Bytes()) //nolint: errcheck
}
