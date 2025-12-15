package spec

import (
	"encoding/json"
	"fmt"
)

const JSONRPCVersion = "2.0"

// JSONRPCID is a wrapper type for JSON-RPC request IDs,
// which can be a string or number (or omitted)
type JSONRPCID interface {
	String() string
}

// JSONRPCStringID is a wrapper for JSON-RPC string IDs
type JSONRPCStringID string

func (id JSONRPCStringID) String() string {
	return string(id)
}

// JSONRPCIntID is a wrapper for JSON-RPC number IDs
type JSONRPCIntID uint

func (id JSONRPCIntID) String() string {
	return fmt.Sprintf("%d", id)
}

// parseID parses the generic JSON value into a JSON-RPC ID (string / number)
func parseID(idValue any) (JSONRPCID, error) {
	switch v := idValue.(type) {
	case string:
		return JSONRPCStringID(v), nil
	case float64:
		// encoding/json uses float64 for numbers
		return JSONRPCIntID(uint(v)), nil
	case nil:
		// omitted
		return nil, nil
	default:
		return nil, fmt.Errorf("JSON-RPC ID (%v) is of unknown type (%T)", v, v)
	}
}

// BaseJSON defines the base JSON fields
// all JSON-RPC requests and responses need to have
type BaseJSON struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      JSONRPCID `json:"id,omitempty"`
}

// BaseJSONRequest defines the base JSON request format
type BaseJSONRequest struct {
	BaseJSON

	Method string `json:"method"`
	Params []any  `json:"params"`
}

// BaseJSONRequests represents a batch of JSON-RPC requests
type BaseJSONRequests []*BaseJSONRequest

// BaseJSONResponses represents a batch of JSON-RPC responses
type BaseJSONResponses []*BaseJSONResponse

// BaseJSONResponse defines the base JSON response format
type BaseJSONResponse struct {
	Result any            `json:"result,omitempty"`
	Error  *BaseJSONError `json:"error,omitempty"`
	BaseJSON
}

// BaseJSONError defines the base JSON response error format
type BaseJSONError struct {
	Data    any    `json:"data,omitempty"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewJSONRequest creates a new JSON-RPC request
func NewJSONRequest(
	id JSONRPCID,
	method string,
	params []any,
) *BaseJSONRequest {
	return &BaseJSONRequest{
		BaseJSON: BaseJSON{
			ID:      id,
			JSONRPC: JSONRPCVersion,
		},
		Method: method,
		Params: params,
	}
}

// NewJSONResponse creates a new JSON-RPC response
func NewJSONResponse(
	id JSONRPCID,
	result any,
	err *BaseJSONError,
) *BaseJSONResponse {
	return &BaseJSONResponse{
		BaseJSON: BaseJSON{
			ID:      id,
			JSONRPC: JSONRPCVersion,
		},
		Result: result,
		Error:  err,
	}
}

// NewJSONError creates a new JSON-RPC error
func NewJSONError(message string, code int) *BaseJSONError {
	return &BaseJSONError{
		Code:    code,
		Message: message,
	}
}

// GenerateResponseError generates the JSON-RPC server error response
func GenerateResponseError(err error) *BaseJSONError {
	return NewJSONError(err.Error(), ServerErrorCode)
}

// GenerateInvalidParamError generates the JSON-RPC invalid param error response
func GenerateInvalidParamError(index int) *BaseJSONError {
	return NewJSONError(
		fmt.Sprintf(
			"Invalid %s parameter",
			getOrdinalSuffix(index),
		),
		InvalidParamsErrorCode,
	)
}

func getOrdinalSuffix(num int) string {
	switch num % 10 {
	case 1:
		if num%100 != 11 {
			return fmt.Sprintf("%d%s", num, "st")
		}
	case 2:
		if num%100 != 12 {
			return fmt.Sprintf("%d%s", num, "nd")
		}
	case 3:
		if num%100 != 13 {
			return fmt.Sprintf("%d%s", num, "rd")
		}
	}

	return fmt.Sprintf("%d%s", num, "th")
}

func (r BaseJSONRequest) MarshalJSON() ([]byte, error) {
	var id any
	switch v := r.ID.(type) {
	case nil:
		// omitted
	case JSONRPCStringID:
		id = string(v)
	case JSONRPCIntID:
		id = uint(v)
	default:
		if v != nil {
			return nil, fmt.Errorf("unsupported JSON-RPC ID type %T", v)
		}
	}

	var raw struct {
		JSONRPC string `json:"jsonrpc"`
		ID      any    `json:"id,omitempty"`
		Method  string `json:"method"`
		Params  []any  `json:"params"`
	}

	raw.JSONRPC = r.JSONRPC
	raw.ID = id
	raw.Method = r.Method
	raw.Params = r.Params

	return json.Marshal(raw)
}

func (r *BaseJSONRequest) UnmarshalJSON(data []byte) error {
	var raw struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      any             `json:"id"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("unable to JSON-parse request: %w", err)
	}

	r.JSONRPC = raw.JSONRPC
	r.Method = raw.Method

	// Parse ID
	id, err := parseID(raw.ID)
	if err != nil {
		return fmt.Errorf("unable to parse request ID: %w", err)
	}

	r.ID = id

	// Parse params as []any
	if len(raw.Params) == 0 || string(raw.Params) == "null" {
		r.Params = nil

		return nil
	}

	var params []any
	if err := json.Unmarshal(raw.Params, &params); err != nil {
		return fmt.Errorf("unable to parse request params: %w", err)
	}

	r.Params = params

	return nil
}

func (r BaseJSONResponse) MarshalJSON() ([]byte, error) {
	var id any
	switch v := r.ID.(type) {
	case nil:
	case JSONRPCStringID:
		id = string(v)
	case JSONRPCIntID:
		id = uint(v)
	default:
		if v != nil {
			return nil, fmt.Errorf("unsupported JSON-RPC ID type %T", v)
		}
	}

	var raw struct {
		JSONRPC string         `json:"jsonrpc"`
		ID      any            `json:"id,omitempty"`
		Result  any            `json:"result,omitempty"`
		Error   *BaseJSONError `json:"error,omitempty"`
	}

	raw.JSONRPC = r.JSONRPC
	raw.ID = id
	raw.Result = r.Result
	raw.Error = r.Error

	return json.Marshal(raw)
}

func (r *BaseJSONResponse) UnmarshalJSON(data []byte) error {
	var raw struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      any             `json:"id"`
		Result  json.RawMessage `json:"result,omitempty"`
		Error   *BaseJSONError  `json:"error,omitempty"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("unable to JSON-parse response: %w", err)
	}

	r.JSONRPC = raw.JSONRPC
	r.Error = raw.Error

	id, err := parseID(raw.ID)
	if err != nil {
		return fmt.Errorf("unable to parse response ID: %w", err)
	}

	r.ID = id

	var result any
	if len(raw.Result) > 0 && string(raw.Result) != "null" {
		if err := json.Unmarshal(raw.Result, &result); err != nil {
			return fmt.Errorf("unable to parse response result: %w", err)
		}
	}
	r.Result = result

	return nil
}
