package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	bfttypes "github.com/gnolang/gno/tm2/pkg/bft/types"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestServer creates a test RPC server
func createTestServer(
	t *testing.T,
	handler http.Handler,
) *httptest.Server {
	t.Helper()

	s := httptest.NewServer(handler)
	t.Cleanup(s.Close)

	return s
}

// defaultHTTPHandler generates a default HTTP test handler
func defaultHTTPHandler(
	t *testing.T,
	method string,
	responseResult any,
) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/json", r.Header.Get("content-type"))

		// Parse the message
		var req types.RPCRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))

		// Basic request validation
		require.Equal(t, req.JSONRPC, "2.0")
		require.Equal(t, req.Method, method)

		// Marshal the result data to Amino JSON
		result, err := amino.MarshalJSON(responseResult)
		require.NoError(t, err)

		// Send a response back
		response := types.RPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  result,
		}

		// Marshal the response
		marshalledResponse, err := json.Marshal(response)
		require.NoError(t, err)

		_, err = w.Write(marshalledResponse)
		require.NoError(t, err)
	}
}

// defaultWSHandler generates a default WS test handler
func defaultWSHandler(
	t *testing.T,
	method string,
	responseResult any,
) http.HandlerFunc {
	t.Helper()

	upgrader := websocket.Upgrader{}

	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)

		defer c.Close()

		for {
			mt, message, err := c.ReadMessage()
			if websocket.IsUnexpectedCloseError(err) {
				return
			}

			require.NoError(t, err)

			// Parse the message
			var req types.RPCRequest
			require.NoError(t, json.Unmarshal(message, &req))

			// Basic request validation
			require.Equal(t, req.JSONRPC, "2.0")
			require.Equal(t, req.Method, method)

			// Marshal the result data to Amino JSON
			result, err := amino.MarshalJSON(responseResult)
			require.NoError(t, err)

			// Send a response back
			response := types.RPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  result,
			}

			// Marshal the response
			marshalledResponse, err := json.Marshal(response)
			require.NoError(t, err)

			require.NoError(t, c.WriteMessage(mt, marshalledResponse))
		}
	}
}

type e2eTestCase struct {
	name   string
	client *RPCClient
}

// generateE2ETestCases generates RPC client test cases (HTTP / WS)
func generateE2ETestCases(
	t *testing.T,
	method string,
	responseResult any,
) []e2eTestCase {
	t.Helper()

	// Create the http client
	httpServer := createTestServer(t, defaultHTTPHandler(t, method, responseResult))
	httpClient, err := NewHTTPClient(httpServer.URL)
	require.NoError(t, err)

	// Create the WS client
	wsServer := createTestServer(t, defaultWSHandler(t, method, responseResult))
	wsClient, err := NewWSClient("ws" + strings.TrimPrefix(wsServer.URL, "http"))
	require.NoError(t, err)

	return []e2eTestCase{
		{
			name:   "http",
			client: httpClient,
		},
		{
			name:   "ws",
			client: wsClient,
		},
	}
}

func TestRPCClient_E2E_Endpoints(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name           string
		expectedResult any
		verifyFn       func(*RPCClient, any)
	}{
		{
			statusMethod,
			&ctypes.ResultStatus{
				NodeInfo: p2pTypes.NodeInfo{
					Moniker: "dummy",
				},
			},
			func(client *RPCClient, expectedResult any) {
				status, err := client.Status(context.Background(), nil)
				require.NoError(t, err)

				assert.Equal(t, expectedResult, status)
			},
		},
		{
			abciInfoMethod,
			&ctypes.ResultABCIInfo{
				Response: abci.ResponseInfo{
					LastBlockAppHash: []byte("dummy"),
				},
			},
			func(client *RPCClient, expectedResult any) {
				result, err := client.ABCIInfo(context.Background())
				require.NoError(t, err)

				assert.Equal(t, expectedResult, result)
			},
		},
		{
			abciQueryMethod,
			&ctypes.ResultABCIQuery{
				Response: abci.ResponseQuery{
					Value: []byte("dummy"),
				},
			},
			func(client *RPCClient, expectedResult any) {
				result, err := client.ABCIQuery(context.Background(), "path", []byte("dummy"))
				require.NoError(t, err)

				assert.Equal(t, expectedResult, result)
			},
		},
		{
			broadcastTxCommitMethod,
			&ctypes.ResultBroadcastTxCommit{
				Hash: []byte("dummy"),
			},
			func(client *RPCClient, expectedResult any) {
				result, err := client.BroadcastTxCommit(context.Background(), []byte("dummy"))
				require.NoError(t, err)

				assert.Equal(t, expectedResult, result)
			},
		},
		{
			broadcastTxAsyncMethod,
			&ctypes.ResultBroadcastTx{
				Hash: []byte("dummy"),
			},
			func(client *RPCClient, expectedResult any) {
				result, err := client.BroadcastTxAsync(context.Background(), []byte("dummy"))
				require.NoError(t, err)

				assert.Equal(t, expectedResult, result)
			},
		},
		{
			broadcastTxSyncMethod,
			&ctypes.ResultBroadcastTx{
				Hash: []byte("dummy"),
			},
			func(client *RPCClient, expectedResult any) {
				result, err := client.BroadcastTxSync(context.Background(), []byte("dummy"))
				require.NoError(t, err)

				assert.Equal(t, expectedResult, result)
			},
		},
		{
			unconfirmedTxsMethod,
			&ctypes.ResultUnconfirmedTxs{
				Count: 10,
			},
			func(client *RPCClient, expectedResult any) {
				result, err := client.UnconfirmedTxs(context.Background(), 0)
				require.NoError(t, err)

				assert.Equal(t, expectedResult, result)
			},
		},
		{
			numUnconfirmedTxsMethod,
			&ctypes.ResultUnconfirmedTxs{
				Count: 10,
			},
			func(client *RPCClient, expectedResult any) {
				result, err := client.NumUnconfirmedTxs(context.Background())
				require.NoError(t, err)

				assert.Equal(t, expectedResult, result)
			},
		},
		{
			netInfoMethod,
			&ctypes.ResultNetInfo{
				NPeers: 10,
			},
			func(client *RPCClient, expectedResult any) {
				result, err := client.NetInfo(context.Background())
				require.NoError(t, err)

				assert.Equal(t, expectedResult, result)
			},
		},
		{
			dumpConsensusStateMethod,
			&ctypes.ResultDumpConsensusState{
				RoundState: &cstypes.RoundState{
					Round: 10,
				},
			},
			func(client *RPCClient, expectedResult any) {
				result, err := client.DumpConsensusState(context.Background())
				require.NoError(t, err)

				assert.Equal(t, expectedResult, result)
			},
		},
		{
			consensusStateMethod,
			&ctypes.ResultConsensusState{
				RoundState: cstypes.RoundStateSimple{
					ProposalBlockHash: []byte("dummy"),
				},
			},
			func(client *RPCClient, expectedResult any) {
				result, err := client.ConsensusState(context.Background())
				require.NoError(t, err)

				assert.Equal(t, expectedResult, result)
			},
		},
		{
			consensusParamsMethod,
			&ctypes.ResultConsensusParams{
				BlockHeight: 10,
			},
			func(client *RPCClient, expectedResult any) {
				result, err := client.ConsensusParams(context.Background(), nil)
				require.NoError(t, err)

				assert.Equal(t, expectedResult, result)
			},
		},
		{
			healthMethod,
			&ctypes.ResultHealth{},
			func(client *RPCClient, expectedResult any) {
				result, err := client.Health(context.Background())
				require.NoError(t, err)

				assert.Equal(t, expectedResult, result)
			},
		},
		{
			blockchainMethod,
			&ctypes.ResultBlockchainInfo{
				LastHeight: 100,
			},
			func(client *RPCClient, expectedResult any) {
				result, err := client.BlockchainInfo(context.Background(), 0, 0)
				require.NoError(t, err)

				assert.Equal(t, expectedResult, result)
			},
		},
		{
			genesisMethod,
			&ctypes.ResultGenesis{
				Genesis: &bfttypes.GenesisDoc{
					ChainID: "dummy",
				},
			},
			func(client *RPCClient, expectedResult any) {
				result, err := client.Genesis(context.Background())
				require.NoError(t, err)

				assert.Equal(t, expectedResult, result)
			},
		},
		{
			blockMethod,
			&ctypes.ResultBlock{
				BlockMeta: &bfttypes.BlockMeta{
					Header: bfttypes.Header{
						Height: 10,
					},
				},
			},
			func(client *RPCClient, expectedResult any) {
				result, err := client.Block(context.Background(), nil)
				require.NoError(t, err)

				assert.Equal(t, expectedResult, result)
			},
		},
		{
			blockResultsMethod,
			&ctypes.ResultBlockResults{
				Height: 10,
			},
			func(client *RPCClient, expectedResult any) {
				result, err := client.BlockResults(context.Background(), nil)
				require.NoError(t, err)

				assert.Equal(t, expectedResult, result)
			},
		},
		{
			commitMethod,
			&ctypes.ResultCommit{
				CanonicalCommit: true,
			},
			func(client *RPCClient, expectedResult any) {
				result, err := client.Commit(context.Background(), nil)
				require.NoError(t, err)

				assert.Equal(t, expectedResult, result)
			},
		},
		{
			txMethod,
			&ctypes.ResultTx{
				Hash:   []byte("tx hash"),
				Height: 10,
			},
			func(client *RPCClient, expectedResult any) {
				result, err := client.Tx(context.Background(), []byte("tx hash"))
				require.NoError(t, err)

				assert.Equal(t, expectedResult, result)
			},
		},
		{
			validatorsMethod,
			&ctypes.ResultValidators{
				BlockHeight: 10,
			},
			func(client *RPCClient, expectedResult any) {
				result, err := client.Validators(context.Background(), nil)
				require.NoError(t, err)

				assert.Equal(t, expectedResult, result)
			},
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			clientTable := generateE2ETestCases(
				t,
				testCase.name,
				testCase.expectedResult,
			)

			for _, clientCase := range clientTable {
				clientCase := clientCase

				t.Run(clientCase.name, func(t *testing.T) {
					t.Parallel()

					defer func() {
						require.NoError(t, clientCase.client.Close())
					}()

					testCase.verifyFn(clientCase.client, testCase.expectedResult)
				})
			}
		})
	}
}
