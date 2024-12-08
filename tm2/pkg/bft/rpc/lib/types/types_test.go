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
						[]byte(fmt.Sprintf(`{"jsonrpc":"2.0","id":%v,"result":{"Value":"hello"}}`, testCase.expectedID)),
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
