package rpctypes

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/errors"
)

func TestJSONRPCID_Marshal_Unmarshal(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name       string
		id         JSONRPCID
		expectedID string
	}{
		{
			"short string",
			JSONRPCStringID("1"),
			`"1"`,
		},
		{
			"long string",
			JSONRPCStringID("alphabet"),
			`"alphabet"`,
		},
		{
			"empty string",
			JSONRPCStringID(""),
			`""`,
		},
		{
			"unicode string",
			JSONRPCStringID("àáâ"),
			`"àáâ"`,
		},
		{
			"negative number",
			JSONRPCIntID(-1),
			"-1",
		},
		{
			"zero ID",
			JSONRPCIntID(0),
			"0",
		},
		{
			"non-zero ID",
			JSONRPCIntID(100),
			"100",
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			t.Run("marshal", func(t *testing.T) {
				t.Parallel()

				data, err := json.Marshal(
					NewRPCSuccessResponse(testCase.id, struct {
						Value string
					}{
						Value: "hello",
					},
					),
				)
				require.NoError(t, err)

				assert.Equal(
					t,
					fmt.Sprintf(
						`{"jsonrpc":"2.0","id":%v,"result":{"Value":"hello"}}`,
						testCase.expectedID,
					),
					string(data),
				)

				data, err = json.Marshal(RPCParseError(testCase.id, errors.New("Hello world")))
				require.NoError(t, err)

				assert.Equal(
					t,
					fmt.Sprintf(
						`{"jsonrpc":"2.0","id":%v,"error":{"code":-32700,"message":"Parse error. Invalid JSON","data":"Hello world"}}`,
						testCase.expectedID,
					),
					string(data),
				)

				data, err = json.Marshal(RPCMethodNotFoundError(testCase.id))
				require.NoError(t, err)

				assert.Equal(
					t,
					fmt.Sprintf(
						`{"jsonrpc":"2.0","id":%v,"error":{"code":-32601,"message":"Method not found"}}`,
						testCase.expectedID,
					),
					string(data),
				)
			})

			t.Run("unmarshal", func(t *testing.T) {
				t.Parallel()

				var expectedResponse RPCResponse

				assert.NoError(
					t,
					json.Unmarshal(
						fmt.Appendf(nil, `{"jsonrpc":"2.0","id":%v,"result":{"Value":"hello"}}`, testCase.expectedID),
						&expectedResponse,
					),
				)

				successResponse := NewRPCSuccessResponse(
					testCase.id,
					struct {
						Value string
					}{
						Value: "hello",
					},
				)

				assert.Equal(t, expectedResponse, successResponse)
			})
		})
	}
}

func TestRPCResponse_UnmarshalJSON_NilID(t *testing.T) {
	t.Parallel()

	t.Run("error response with null ID should return error", func(t *testing.T) {
		t.Parallel()

		// Per JSON-RPC spec, error responses can have null ID (e.g., parse errors)
		jsonData := []byte(`{"jsonrpc":"2.0","id":null,"error":{"code":-32700,"message":"Parse error"}}`)

		var response RPCResponse
		err := json.Unmarshal(jsonData, &response)

		// Should return the RPCError
		require.Error(t, err)
		assert.IsType(t, &RPCError{}, err)
		assert.Equal(t, -32700, err.(*RPCError).Code)
		assert.Equal(t, "Parse error", err.(*RPCError).Message)
	})

	t.Run("success response with null ID should fail", func(t *testing.T) {
		t.Parallel()

		// Success responses must have a valid ID
		jsonData := []byte(`{"jsonrpc":"2.0","id":null,"result":"something"}`)

		var response RPCResponse
		err := json.Unmarshal(jsonData, &response)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "request ID cannot be nil")
	})

	t.Run("error response with valid ID should parse successfully", func(t *testing.T) {
		t.Parallel()

		jsonData := []byte(`{"jsonrpc":"2.0","id":"123","error":{"code":-32600,"message":"Invalid Request"}}`)

		var response RPCResponse
		err := json.Unmarshal(jsonData, &response)

		// Should parse successfully - error is in the response, not unmarshaling
		require.NoError(t, err)
		assert.Equal(t, JSONRPCStringID("123"), response.ID)
		require.NotNil(t, response.Error)
		assert.Equal(t, -32600, response.Error.Code)
		assert.Equal(t, "Invalid Request", response.Error.Message)
	})
}
