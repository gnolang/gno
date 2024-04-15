package rpctypes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

// JSONRPCID is a wrapper type for JSON-RPC request IDs,
// which can be a string value | number value | not set (nil)
type JSONRPCID interface {
	String() string
}

// JSONRPCStringID a wrapper for JSON-RPC string IDs
type JSONRPCStringID string

func (id JSONRPCStringID) String() string {
	return string(id)
}

// JSONRPCIntID a wrapper for JSON-RPC integer IDs
type JSONRPCIntID int

func (id JSONRPCIntID) String() string {
	return fmt.Sprintf("%d", id)
}

// parseID parses the given ID value
func parseID(idValue any) (JSONRPCID, error) {
	switch id := idValue.(type) {
	case string:
		return JSONRPCStringID(id), nil
	case float64:
		// json.Unmarshal uses float64 for all numbers
		// (https://golang.org/pkg/encoding/json/#Unmarshal),
		// but the JSONRPC2.0 spec says the id SHOULD NOT contain
		// decimals - so we truncate the decimals here.
		return JSONRPCIntID(int(id)), nil
	default:
		typ := reflect.TypeOf(id)
		return nil, fmt.Errorf("JSON-RPC ID (%v) is of unknown type (%v)", id, typ)
	}
}

// ----------------------------------------
// REQUEST

type RPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      JSONRPCID       `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"` // must be map[string]interface{} or []interface{}
}

// UnmarshalJSON custom JSON unmarshalling due to JSONRPCID being string or int
func (request *RPCRequest) UnmarshalJSON(data []byte) error {
	unsafeReq := &struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      any             `json:"id"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"` // must be map[string]any or []any
	}{}

	if err := json.Unmarshal(data, &unsafeReq); err != nil {
		return fmt.Errorf("unable to JSON-parse the RPC request, %w", err)
	}

	request.JSONRPC = unsafeReq.JSONRPC
	request.Method = unsafeReq.Method
	request.Params = unsafeReq.Params

	// Check if the ID is set
	if unsafeReq.ID == nil {
		return nil
	}

	// Parse the ID
	id, err := parseID(unsafeReq.ID)
	if err != nil {
		return fmt.Errorf("unable to parse request ID, %w", err)
	}

	request.ID = id

	return nil
}

func NewRPCRequest(id JSONRPCID, method string, params json.RawMessage) RPCRequest {
	return RPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
}

func (request RPCRequest) String() string {
	return fmt.Sprintf("[%s %s]", request.ID, request.Method)
}

// MapToRequest generates an RPC request with the given ID and method.
// The params are encoded as a JSON map
func MapToRequest(id JSONRPCID, method string, params map[string]any) (RPCRequest, error) {
	params_ := make(map[string]json.RawMessage, len(params))
	for name, value := range params {
		valueJSON, err := amino.MarshalJSON(value)
		if err != nil {
			return RPCRequest{}, fmt.Errorf("unable to parse param, %w", err)
		}

		params_[name] = valueJSON
	}

	payload, err := json.Marshal(params_) // NOTE: Amino doesn't handle maps yet.
	if err != nil {
		return RPCRequest{}, fmt.Errorf("unable to JSON marshal params, %w", err)
	}

	return NewRPCRequest(id, method, payload), nil
}

// ----------------------------------------
// RESPONSE

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

func (err RPCError) Error() string {
	const baseFormat = "RPC error %d - %s"
	if err.Data != "" {
		return fmt.Sprintf(baseFormat+": %s", err.Code, err.Message, err.Data)
	}

	return fmt.Sprintf(baseFormat, err.Code, err.Message)
}

type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      JSONRPCID       `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

type (
	RPCRequests  []RPCRequest
	RPCResponses []RPCResponse
)

// UnmarshalJSON custom JSON unmarshalling due to JSONRPCID being string or int
func (response *RPCResponse) UnmarshalJSON(data []byte) error {
	unsafeResp := &struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      interface{}     `json:"id"`
		Result  json.RawMessage `json:"result,omitempty"`
		Error   *RPCError       `json:"error,omitempty"`
	}{}

	// Parse the response
	if err := json.Unmarshal(data, &unsafeResp); err != nil {
		return fmt.Errorf("unable to JSON-parse the RPC response, %w", err)
	}

	response.JSONRPC = unsafeResp.JSONRPC
	response.Error = unsafeResp.Error
	response.Result = unsafeResp.Result

	// Check if any response ID is set
	if unsafeResp.ID == nil {
		return nil
	}

	// Parse the ID
	id, err := parseID(unsafeResp.ID)
	if err != nil {
		return fmt.Errorf("unable to parse response ID, %w", err)
	}

	response.ID = id

	return nil
}

func NewRPCSuccessResponse(id JSONRPCID, res any) RPCResponse {
	var rawMsg json.RawMessage

	if res != nil {
		var js []byte
		js, err := amino.MarshalJSON(res)
		if err != nil {
			return RPCInternalError(id, errors.Wrap(err, "Error marshalling response"))
		}
		rawMsg = js
	}

	return RPCResponse{JSONRPC: "2.0", ID: id, Result: rawMsg}
}

func NewRPCErrorResponse(id JSONRPCID, code int, msg string, data string) RPCResponse {
	return RPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &RPCError{Code: code, Message: msg, Data: data},
	}
}

func (response RPCResponse) String() string {
	if response.Error == nil {
		return fmt.Sprintf("[%s %v]", response.ID, response.Result)
	}
	return fmt.Sprintf("[%s %s]", response.ID, response.Error)
}

func RPCParseError(id JSONRPCID, err error) RPCResponse {
	return NewRPCErrorResponse(id, -32700, "Parse error. Invalid JSON", err.Error())
}

func RPCInvalidRequestError(id JSONRPCID, err error) RPCResponse {
	return NewRPCErrorResponse(id, -32600, "Invalid Request", err.Error())
}

func RPCMethodNotFoundError(id JSONRPCID) RPCResponse {
	return NewRPCErrorResponse(id, -32601, "Method not found", "")
}

func RPCInvalidParamsError(id JSONRPCID, err error) RPCResponse {
	return NewRPCErrorResponse(id, -32602, "Invalid params", err.Error())
}

func RPCInternalError(id JSONRPCID, err error) RPCResponse {
	return NewRPCErrorResponse(id, -32603, "Internal error", err.Error())
}

// ----------------------------------------

// WSRPCConnection represents a websocket connection.
type WSRPCConnection interface {
	// GetRemoteAddr returns a remote address of the connection.
	GetRemoteAddr() string
	// WriteRPCResponses writes the resp onto connection (BLOCKING).
	WriteRPCResponses(resp RPCResponses)
	// TryWriteRPCResponses tries to write the resp onto connection (NON-BLOCKING).
	TryWriteRPCResponses(resp RPCResponses) bool
	// Context returns the connection's context.
	Context() context.Context
}

// Context is the first parameter for all functions. It carries a json-rpc
// request, http request and websocket connection.
//
// - JSONReq is non-nil when JSONRPC is called over websocket or HTTP.
// - WSConn is non-nil when we're connected via a websocket.
// - HTTPReq is non-nil when URI or JSONRPC is called over HTTP.
type Context struct {
	// json-rpc request
	JSONReq *RPCRequest
	// websocket connection
	WSConn WSRPCConnection
	// http request
	HTTPReq *http.Request
}

// RemoteAddr returns the remote address (usually a string "IP:port").
// If neither HTTPReq nor WSConn is set, an empty string is returned.
// HTTP:
//
//	http.Request#RemoteAddr
//
// WS:
//
//	result of GetRemoteAddr
func (ctx *Context) RemoteAddr() string {
	if ctx.HTTPReq != nil {
		return ctx.HTTPReq.RemoteAddr
	} else if ctx.WSConn != nil {
		return ctx.WSConn.GetRemoteAddr()
	}
	return ""
}

// Context returns the request's context.
// The returned context is always non-nil; it defaults to the background context.
// HTTP:
//
//	The context is canceled when the client's connection closes, the request
//	is canceled (with HTTP/2), or when the ServeHTTP method returns.
//
// WS:
//
//	The context is canceled when the client's connections closes.
func (ctx *Context) Context() context.Context {
	if ctx.HTTPReq != nil {
		return ctx.HTTPReq.Context()
	} else if ctx.WSConn != nil {
		return ctx.WSConn.Context()
	}
	return context.Background()
}
