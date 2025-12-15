package spec

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			"zero ID",
			JSONRPCNumberID(0),
			"0",
		},
		{
			"non-zero ID",
			JSONRPCNumberID(100),
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
					NewJSONResponse(
						testCase.id,
						struct {
							Value string
						}{
							Value: "hello",
						},
						nil,
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

				data, err = json.Marshal(
					NewJSONResponse(
						testCase.id,
						nil,
						NewJSONError("Invalid JSON", ParseErrorCode),
					),
				)
				require.NoError(t, err)

				assert.Equal(
					t,
					fmt.Sprintf(
						`{"jsonrpc":"2.0","id":%v,"error":{"code":-32700,"message":"Invalid JSON"}}`,
						testCase.expectedID,
					),
					string(data),
				)

				data, err = json.Marshal(
					NewJSONResponse(
						testCase.id,
						nil,
						NewJSONError("Method not found", MethodNotFoundErrorCode),
					),
				)
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

				var expectedResponse *BaseJSONResponse

				assert.NoError(
					t,
					json.Unmarshal(
						fmt.Appendf(nil, `{"jsonrpc":"2.0","id":%v,"result":{"Value":"hello"}}`, testCase.expectedID),
						&expectedResponse,
					),
				)

				successResponse := NewJSONResponse(
					testCase.id,
					struct {
						Value string
					}{
						Value: "hello",
					},
					nil,
				)

				assert.Equal(t, expectedResponse, successResponse)
			})
		})
	}
}
