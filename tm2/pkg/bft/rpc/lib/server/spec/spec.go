package spec

import (
	"encoding/json"
	"fmt"
)

const JSONRPCVersion = "2.0"

// BaseJSON defines the base JSON fields
// all JSON-RPC requests and responses need to have
type BaseJSON struct {
	JSONRPC string `json:"jsonrpc"`
	ID      uint   `json:"id,omitempty"` // TODO support string IDs
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
	Result any            `json:"result"`
	Error  *BaseJSONError `json:"error,omitempty"`
	BaseJSON
}

// BaseJSONError defines the base JSON response error format
type BaseJSONError struct {
	Data    any    `json:"data,omitempty"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// NewJSONRequest creates a new JSON-RPC request
func NewJSONRequest(
	id uint,
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
	id uint,
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

// GenerateInvalidParamCountError generates the JSON-RPC invalid param count error
func GenerateInvalidParamCountError() *BaseJSONError {
	return NewJSONError(
		"Invalid number of parameters",
		InvalidParamsErrorCode,
	)
}

func ParseObjectParameter[T any](param any, data *T) error {
	marshaled, err := json.Marshal(param)
	if err != nil {
		return err
	}

	err = json.Unmarshal(marshaled, data)
	if err != nil {
		return err
	}

	return nil
}
