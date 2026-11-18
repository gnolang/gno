package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// serveNodeResult starts a JSON-RPC endpoint that answers method with result
// the way a node does. It fails the test on any other method, and on a request
// whose path param is not wantPath.
func serveNodeResult(t *testing.T, method, wantPath, result string) *httptest.Server {
	t.Helper()

	node := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request rpctypes.RPCRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decoding JSON-RPC request: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if request.Method != method {
			t.Errorf("JSON-RPC method = %q, want %q", request.Method, method)
			http.Error(w, "unexpected method", http.StatusBadRequest)
			return
		}

		var params struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			t.Errorf("decoding JSON-RPC params: %v", err)
		}
		if params.Path != wantPath {
			t.Errorf("JSON-RPC path param = %q, want %q", params.Path, wantPath)
		}

		w.Header().Set("Content-Type", "application/json")
		response := rpctypes.RPCResponse{JSONRPC: "2.0", ID: request.ID, Result: json.RawMessage(result)}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("writing %s response: %v", method, err)
		}
	}))
	t.Cleanup(node.Close)

	return node
}

// getAccountFromNode reads the account at address through a node client whose
// node answers the account query with result.
func getAccountFromNode(t *testing.T, address, result string) (std.Account, error) {
	t.Helper()

	node := serveNodeResult(t, "abci_query", "auth/accounts/"+address, result)

	cli, err := newNodeClient(node.URL)
	require.NoError(t, err)

	return cli.GetAccount(crypto.MustAddressFromString(address))
}

// accountQueryResult is the abci_query result a node returns with data as its
// response data. A nil data is sent as null.
func accountQueryResult(data []byte) string {
	return string(amino.MustMarshalJSON(&ctypes.ResultABCIQuery{
		Response: abci.ResponseQuery{ResponseBase: abci.ResponseBase{Data: data}},
	}))
}

// transferCommitResult is the broadcast_tx_commit result of a 100 GNOT faucet
// drip on onyx-1 (height 22949). deliver_tx is the result the node recorded for
// that transaction. check_tx is omitted: CheckTx runs only the ante handler, so
// it carries no events.
const transferCommitResult = `{
	"deliver_tx": {
		"ResponseBase": {
			"Error": null,
			"Data": null,
			"Events": [
				{
					"@type": "/bank.TransferEvent",
					"from": "g1jf3xq3yur9tp9ts0h9qkl9t96kptvzq8cwshyw",
					"to": "g1aeddlftlfk27ret5rf750d7w5dume3kcsm8r8m",
					"coins": "100000000ugnot"
				}
			],
			"Log": "msg:0,success:true,log:,events:[]",
			"Info": ""
		},
		"GasWanted": "5000000",
		"GasUsed": "933276"
	},
	"hash": "vRDuCfcBACb7Zjk/SBQ8BTv2fTqxa9yBBawGdrdG1Wo=",
	"height": "22949"
}`

// TestNodeClient_DecodesBankTransferEvent checks that the node client decodes a
// broadcast_tx_commit result carrying a bank.TransferEvent. The client decodes
// with the amino types of the tm2 that gnofaucet links, so a drip whose result
// holds an event type missing from that tm2 is reported as failed although the
// transfer is committed.
func TestNodeClient_DecodesBankTransferEvent(t *testing.T) {
	t.Parallel()

	// A broadcast carries only the transaction, so the request has no path.
	node := serveNodeResult(t, "broadcast_tx_commit", "", transferCommitResult)

	cli, err := newNodeClient(node.URL)
	require.NoError(t, err)

	res, err := cli.SendTransactionCommit(&std.Tx{})
	require.NoError(t, err)
	assert.Len(t, res.DeliverTx.Events, 1)
}

// faucetAccountAddress is the onyx-1 faucet account that
// faucetAccountQueryResult describes.
const faucetAccountAddress = "g1jf3xq3yur9tp9ts0h9qkl9t96kptvzq8cwshyw"

// faucetAccountQueryResult is the abci_query result for auth/accounts of the
// onyx-1 faucet account. Its data is a gno.land account: std.BaseAccount plus
// the attributes field.
const faucetAccountQueryResult = `{
	"response": {
		"ResponseBase": {
			"Error": null,
			"Data": "ewogICJCYXNlQWNjb3VudCI6IHsKICAgICJhZGRyZXNzIjogImcxamYzeHEzeXVyOXRwOXRzMGg5cWtsOXQ5NmtwdHZ6cThjd3NoeXciLAogICAgImNvaW5zIjogIjk5OTk5OTk5NTM3MzAwMDAwMHVnbm90IiwKICAgICJwdWJsaWNfa2V5IjogewogICAgICAiQHR5cGUiOiAiL3RtLlB1YktleVNlY3AyNTZrMSIsCiAgICAgICJ2YWx1ZSI6ICJBL0F0eURkYkpEVy80WHh3S3RBSGhQeFBCZ0I2WDZMeWxqSC9KVTJRMjJYMyIKICAgIH0sCiAgICAiYWNjb3VudF9udW1iZXIiOiAiNiIsCiAgICAic2VxdWVuY2UiOiAiMjciCiAgfSwKICAiYXR0cmlidXRlcyI6ICIwIgp9",
			"Events": null,
			"Log": "",
			"Info": ""
		},
		"Key": null,
		"Value": null,
		"Proof": null,
		"Height": "0"
	}
}`

// TestNodeClient_DecodesGnoAccount checks that the node client decodes the
// account a gno.land node returns. The faucet reads each of its accounts this
// way before a drip and skips any it cannot decode, so an account decoding
// failure turns every drip into "no funded account found".
func TestNodeClient_DecodesGnoAccount(t *testing.T) {
	t.Parallel()

	account, err := getAccountFromNode(t, faucetAccountAddress, faucetAccountQueryResult)
	require.NoError(t, err)

	assert.Equal(t, crypto.MustAddressFromString(faucetAccountAddress), account.GetAddress())
	assert.Equal(t, std.MustParseCoins("999999995373000000ugnot"), account.GetCoins())
	assert.Equal(t, uint64(6), account.GetAccountNumber())
	assert.Equal(t, uint64(27), account.GetSequence())
}

// missingAccountQueryResult is the abci_query result for auth/accounts of an
// onyx-1 address that has no account: the node returns null data.
const missingAccountQueryResult = `{
	"response": {
		"ResponseBase": {
			"Error": null,
			"Data": "bnVsbA==",
			"Events": null,
			"Log": "",
			"Info": ""
		},
		"Key": null,
		"Value": null,
		"Proof": null,
		"Height": "0"
	}
}`

// TestNodeClient_ReportsMissingAccount checks that the node client reports an
// address the node holds no account for as std.UnknownAddressError.
func TestNodeClient_ReportsMissingAccount(t *testing.T) {
	t.Parallel()

	const address = "g1k8mmsvg89ctlucrf8ar9ae0ndvna0333agen50"

	testCases := []struct {
		name   string
		result string
	}{
		{name: "null data", result: missingAccountQueryResult},
		{name: "empty data", result: accountQueryResult(nil)},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := getAccountFromNode(t, address, testCase.result)

			assert.ErrorIs(t, err, std.UnknownAddressError{})
		})
	}
}

// invalidAddressQueryResult is the abci_query result onyx-1 returns for
// auth/accounts of a malformed address: the node rejects the query.
const invalidAddressQueryResult = `{
	"response": {
		"ResponseBase": {
			"Error": {
				"@type": "/std.InvalidAddressError"
			},
			"Data": null,
			"Events": null,
			"Log": "--= Error =--\nData: std.InvalidAddressError{abciError:std.abciError{}}\nMsg Traces:\n    0  /gnoroot/tm2/pkg/std/errors.go:96 - invalid query address notbech32\nStack Trace:\n    0  /gnoroot/tm2/pkg/errors/errors.go:93\n    1  /gnoroot/tm2/pkg/std/errors.go:96\n    2  /gnoroot/tm2/pkg/sdk/auth/handler.go:189\n    3  /gnoroot/tm2/pkg/sdk/auth/handler.go:166\n    4  /gnoroot/tm2/pkg/sdk/baseapp.go:574\n    5  /gnoroot/tm2/pkg/sdk/baseapp.go:455\n    6  /gnoroot/tm2/pkg/bft/abci/client/local_client.go:193\n    7  /gnoroot/tm2/pkg/bft/appconn/app_conn.go:144\n    8  /gnoroot/tm2/pkg/bft/rpc/core/abci.go:14\n    9  /usr/local/go/src/reflect/value.go:581\n   10  /usr/local/go/src/reflect/value.go:365\n... 13 more lines elided",
			"Info": ""
		},
		"Key": null,
		"Value": null,
		"Proof": null,
		"Height": "0"
	}
}`

// TestNodeClient_GetAccountErrors checks that the node client returns an
// error, and no account, for each node answer it does not accept. The node's
// own answer is recorded from onyx-1; the malformed data cases are constructed,
// as a correct node never returns them. BaseAccount is decoded strictly, so a
// BaseAccount field that std.BaseAccount lacks is rejected on purpose.
func TestNodeClient_GetAccountErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		address   string
		result    string
		wantErrAs any // when set, a pointer the error must match with errors.As
	}{
		{
			name:      "node rejects the query",
			address:   faucetAccountAddress,
			result:    invalidAddressQueryResult,
			wantErrAs: new(std.InvalidAddressError),
		},
		{
			name:      "data is not JSON",
			address:   faucetAccountAddress,
			result:    accountQueryResult([]byte("not json")),
			wantErrAs: new(*json.SyntaxError),
		},
		{
			name:    "data has no BaseAccount field",
			address: faucetAccountAddress,
			result:  accountQueryResult([]byte(`{"attributes":"0"}`)),
		},
		{
			name:    "BaseAccount has a field std.BaseAccount lacks",
			address: faucetAccountAddress,
			result:  accountQueryResult([]byte(`{"BaseAccount":{"address":"` + faucetAccountAddress + `","unknown":"0"},"attributes":"0"}`)),
		},
		{
			name:    "BaseAccount is null",
			address: faucetAccountAddress,
			result:  accountQueryResult([]byte(`{"BaseAccount":null,"attributes":"0"}`)),
		},
		{
			name:    "account belongs to another address",
			address: "g1aeddlftlfk27ret5rf750d7w5dume3kcsm8r8m",
			result:  faucetAccountQueryResult,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			account, err := getAccountFromNode(t, testCase.address, testCase.result)

			assert.Nil(t, account)
			require.Error(t, err)
			assert.NotErrorIs(t, err, std.UnknownAddressError{})
			if testCase.wantErrAs != nil {
				assert.ErrorAs(t, err, testCase.wantErrAs)
			}
		})
	}
}

// TestNodeClient_ReportsUnreachableNode checks that the node client returns an
// error when the node does not answer.
func TestNodeClient_ReportsUnreachableNode(t *testing.T) {
	t.Parallel()

	// Nothing serves port 1 (tcpmux) on the loopback address, whereas the port
	// of a closed test server may be reused by a parallel test.
	cli, err := newNodeClient("http://127.0.0.1:1")
	require.NoError(t, err)

	account, err := cli.GetAccount(crypto.MustAddressFromString(faucetAccountAddress))

	assert.Nil(t, account)
	assert.Error(t, err)
}
