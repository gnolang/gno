package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	bfttypes "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateMockRequestClient generates a single RPC request mock client
func generateMockRequestClient(
	t *testing.T,
	method string,
	verifyParamsFn func(*testing.T, map[string]any),
	responseData any,
) *mockClient {
	t.Helper()

	return &mockClient{
		sendRequestFn: func(
			_ context.Context,
			request types.RPCRequest,
		) (*types.RPCResponse, error) {
			// Validate the request
			require.Equal(t, "2.0", request.JSONRPC)
			require.NotNil(t, request.ID)
			require.Equal(t, request.Method, method)

			// Validate the params
			var params map[string]any
			require.NoError(t, json.Unmarshal(request.Params, &params))

			verifyParamsFn(t, params)

			// Prepare the result
			result, err := amino.MarshalJSON(responseData)
			require.NoError(t, err)

			// Prepare the response
			response := &types.RPCResponse{
				JSONRPC: "2.0",
				ID:      request.ID,
				Result:  result,
				Error:   nil,
			}

			return response, nil
		},
	}
}

// generateMockRequestsClient generates a batch RPC request mock client
func generateMockRequestsClient(
	t *testing.T,
	method string,
	verifyParamsFn func(*testing.T, map[string]any),
	responseData []any,
) *mockClient {
	t.Helper()

	return &mockClient{
		sendBatchFn: func(
			_ context.Context,
			requests types.RPCRequests,
		) (types.RPCResponses, error) {
			responses := make(types.RPCResponses, 0, len(requests))

			// Validate the requests
			for index, r := range requests {
				require.Equal(t, "2.0", r.JSONRPC)
				require.NotNil(t, r.ID)
				require.Equal(t, r.Method, method)

				// Validate the params
				var params map[string]any
				require.NoError(t, json.Unmarshal(r.Params, &params))

				verifyParamsFn(t, params)

				// Prepare the result
				result, err := amino.MarshalJSON(responseData[index])
				require.NoError(t, err)

				// Prepare the response
				response := types.RPCResponse{
					JSONRPC: "2.0",
					ID:      r.ID,
					Result:  result,
					Error:   nil,
				}

				responses = append(responses, response)
			}

			return responses, nil
		},
	}
}

func TestRPCClient_Status(t *testing.T) {
	t.Parallel()

	var (
		expectedStatus = &ctypes.ResultStatus{
			NodeInfo: p2p.NodeInfo{
				Moniker: "dummy",
			},
		}

		verifyFn = func(t *testing.T, params map[string]any) {
			t.Helper()

			assert.Len(t, params, 0)
		}

		mockClient = generateMockRequestClient(
			t,
			statusMethod,
			verifyFn,
			expectedStatus,
		)
	)

	// Create the client
	c := NewRPCClient(mockClient)

	// Get the status
	status, err := c.Status()
	require.NoError(t, err)

	assert.Equal(t, expectedStatus, status)
}

func TestRPCClient_ABCIInfo(t *testing.T) {
	t.Parallel()

	var (
		expectedInfo = &ctypes.ResultABCIInfo{
			Response: abci.ResponseInfo{
				LastBlockAppHash: []byte("dummy"),
			},
		}

		verifyFn = func(t *testing.T, params map[string]any) {
			t.Helper()

			assert.Len(t, params, 0)
		}

		mockClient = generateMockRequestClient(
			t,
			abciInfoMethod,
			verifyFn,
			expectedInfo,
		)
	)

	// Create the client
	c := NewRPCClient(mockClient)

	// Get the info
	info, err := c.ABCIInfo()
	require.NoError(t, err)

	assert.Equal(t, expectedInfo, info)
}

func TestRPCClient_ABCIQuery(t *testing.T) {
	t.Parallel()

	var (
		path = "path"
		data = []byte("data")
		opts = DefaultABCIQueryOptions

		expectedQuery = &ctypes.ResultABCIQuery{
			Response: abci.ResponseQuery{
				Value: []byte("dummy"),
			},
		}

		verifyFn = func(t *testing.T, params map[string]any) {
			t.Helper()

			assert.Equal(t, path, params["path"])
			assert.Equal(t, base64.StdEncoding.EncodeToString(data), params["data"])
			assert.Equal(t, fmt.Sprintf("%d", opts.Height), params["height"])
			assert.Equal(t, opts.Prove, params["prove"])
		}

		mockClient = generateMockRequestClient(
			t,
			abciQueryMethod,
			verifyFn,
			expectedQuery,
		)
	)

	// Create the client
	c := NewRPCClient(mockClient)

	// Get the query
	query, err := c.ABCIQuery(path, data)
	require.NoError(t, err)

	assert.Equal(t, expectedQuery, query)
}

func TestRPCClient_BroadcastTxCommit(t *testing.T) {
	t.Parallel()

	var (
		tx = []byte("tx")

		expectedTxCommit = &ctypes.ResultBroadcastTxCommit{
			Hash: []byte("dummy"),
		}

		verifyFn = func(t *testing.T, params map[string]any) {
			t.Helper()

			assert.Equal(t, base64.StdEncoding.EncodeToString(tx), params["tx"])
		}

		mockClient = generateMockRequestClient(
			t,
			broadcastTxCommitMethod,
			verifyFn,
			expectedTxCommit,
		)
	)

	// Create the client
	c := NewRPCClient(mockClient)

	// Get the broadcast
	txCommit, err := c.BroadcastTxCommit(tx)
	require.NoError(t, err)

	assert.Equal(t, expectedTxCommit, txCommit)
}

func TestRPCClient_BroadcastTxAsync(t *testing.T) {
	t.Parallel()

	var (
		tx = []byte("tx")

		expectedTxBroadcast = &ctypes.ResultBroadcastTx{
			Hash: []byte("dummy"),
		}

		verifyFn = func(t *testing.T, params map[string]any) {
			t.Helper()

			assert.Equal(t, base64.StdEncoding.EncodeToString(tx), params["tx"])
		}

		mockClient = generateMockRequestClient(
			t,
			broadcastTxAsyncMethod,
			verifyFn,
			expectedTxBroadcast,
		)
	)

	// Create the client
	c := NewRPCClient(mockClient)

	// Get the broadcast
	txAsync, err := c.BroadcastTxAsync(tx)
	require.NoError(t, err)

	assert.Equal(t, expectedTxBroadcast, txAsync)
}

func TestRPCClient_BroadcastTxSync(t *testing.T) {
	t.Parallel()

	var (
		tx = []byte("tx")

		expectedTxBroadcast = &ctypes.ResultBroadcastTx{
			Hash: []byte("dummy"),
		}

		verifyFn = func(t *testing.T, params map[string]any) {
			t.Helper()

			assert.Equal(t, base64.StdEncoding.EncodeToString(tx), params["tx"])
		}

		mockClient = generateMockRequestClient(
			t,
			broadcastTxSyncMethod,
			verifyFn,
			expectedTxBroadcast,
		)
	)

	// Create the client
	c := NewRPCClient(mockClient)

	// Get the broadcast
	txSync, err := c.BroadcastTxSync(tx)
	require.NoError(t, err)

	assert.Equal(t, expectedTxBroadcast, txSync)
}

func TestRPCClient_UnconfirmedTxs(t *testing.T) {
	t.Parallel()

	var (
		limit = 10

		expectedResult = &ctypes.ResultUnconfirmedTxs{
			Count: 10,
		}

		verifyFn = func(t *testing.T, params map[string]any) {
			t.Helper()

			assert.Equal(t, fmt.Sprintf("%d", limit), params["limit"])
		}

		mockClient = generateMockRequestClient(
			t,
			unconfirmedTxsMethod,
			verifyFn,
			expectedResult,
		)
	)

	// Create the client
	c := NewRPCClient(mockClient)

	// Get the result
	result, err := c.UnconfirmedTxs(limit)
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_NumUnconfirmedTxs(t *testing.T) {
	t.Parallel()

	var (
		expectedResult = &ctypes.ResultUnconfirmedTxs{
			Count: 10,
		}

		verifyFn = func(t *testing.T, params map[string]any) {
			t.Helper()

			assert.Len(t, params, 0)
		}

		mockClient = generateMockRequestClient(
			t,
			numUnconfirmedTxsMethod,
			verifyFn,
			expectedResult,
		)
	)

	// Create the client
	c := NewRPCClient(mockClient)

	// Get the result
	result, err := c.NumUnconfirmedTxs()
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_NetInfo(t *testing.T) {
	t.Parallel()

	var (
		expectedResult = &ctypes.ResultNetInfo{
			NPeers: 10,
		}

		verifyFn = func(t *testing.T, params map[string]any) {
			t.Helper()

			assert.Len(t, params, 0)
		}

		mockClient = generateMockRequestClient(
			t,
			netInfoMethod,
			verifyFn,
			expectedResult,
		)
	)

	// Create the client
	c := NewRPCClient(mockClient)

	// Get the result
	result, err := c.NetInfo()
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_DumpConsensusState(t *testing.T) {
	t.Parallel()

	var (
		expectedResult = &ctypes.ResultDumpConsensusState{
			RoundState: &cstypes.RoundState{
				Round: 10,
			},
		}

		verifyFn = func(t *testing.T, params map[string]any) {
			t.Helper()

			assert.Len(t, params, 0)
		}

		mockClient = generateMockRequestClient(
			t,
			dumpConsensusStateMethod,
			verifyFn,
			expectedResult,
		)
	)

	// Create the client
	c := NewRPCClient(mockClient)

	// Get the result
	result, err := c.DumpConsensusState()
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_ConsensusState(t *testing.T) {
	t.Parallel()

	var (
		expectedResult = &ctypes.ResultConsensusState{
			RoundState: cstypes.RoundStateSimple{
				ProposalBlockHash: []byte("dummy"),
			},
		}

		verifyFn = func(t *testing.T, params map[string]any) {
			t.Helper()

			assert.Len(t, params, 0)
		}

		mockClient = generateMockRequestClient(
			t,
			consensusStateMethod,
			verifyFn,
			expectedResult,
		)
	)

	// Create the client
	c := NewRPCClient(mockClient)

	// Get the result
	result, err := c.ConsensusState()
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_ConsensusParams(t *testing.T) {
	t.Parallel()

	var (
		blockHeight = int64(10)

		expectedResult = &ctypes.ResultConsensusParams{
			BlockHeight: blockHeight,
		}

		verifyFn = func(t *testing.T, params map[string]any) {
			t.Helper()

			assert.Equal(t, fmt.Sprintf("%d", blockHeight), params["height"])
		}

		mockClient = generateMockRequestClient(
			t,
			consensusParamsMethod,
			verifyFn,
			expectedResult,
		)
	)

	// Create the client
	c := NewRPCClient(mockClient)

	// Get the result
	result, err := c.ConsensusParams(&blockHeight)
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_Health(t *testing.T) {
	t.Parallel()

	var (
		expectedResult = &ctypes.ResultHealth{}

		verifyFn = func(t *testing.T, params map[string]any) {
			t.Helper()

			assert.Len(t, params, 0)
		}

		mockClient = generateMockRequestClient(
			t,
			healthMethod,
			verifyFn,
			expectedResult,
		)
	)

	// Create the client
	c := NewRPCClient(mockClient)

	// Get the result
	result, err := c.Health()
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_BlockchainInfo(t *testing.T) {
	t.Parallel()

	var (
		minHeight = int64(5)
		maxHeight = int64(10)

		expectedResult = &ctypes.ResultBlockchainInfo{
			LastHeight: 100,
		}

		verifyFn = func(t *testing.T, params map[string]any) {
			t.Helper()

			assert.Equal(t, fmt.Sprintf("%d", minHeight), params["minHeight"])
			assert.Equal(t, fmt.Sprintf("%d", maxHeight), params["maxHeight"])
		}

		mockClient = generateMockRequestClient(
			t,
			blockchainMethod,
			verifyFn,
			expectedResult,
		)
	)

	// Create the client
	c := NewRPCClient(mockClient)

	// Get the result
	result, err := c.BlockchainInfo(minHeight, maxHeight)
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_Genesis(t *testing.T) {
	t.Parallel()

	var (
		expectedResult = &ctypes.ResultGenesis{
			Genesis: &bfttypes.GenesisDoc{
				ChainID: "dummy",
			},
		}

		verifyFn = func(t *testing.T, params map[string]any) {
			t.Helper()

			assert.Len(t, params, 0)
		}

		mockClient = generateMockRequestClient(
			t,
			genesisMethod,
			verifyFn,
			expectedResult,
		)
	)

	// Create the client
	c := NewRPCClient(mockClient)

	// Get the result
	result, err := c.Genesis()
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_Block(t *testing.T) {
	t.Parallel()

	var (
		height = int64(10)

		expectedResult = &ctypes.ResultBlock{
			BlockMeta: &bfttypes.BlockMeta{
				Header: bfttypes.Header{
					Height: height,
				},
			},
		}

		verifyFn = func(t *testing.T, params map[string]any) {
			t.Helper()

			assert.Equal(t, fmt.Sprintf("%d", height), params["height"])
		}

		mockClient = generateMockRequestClient(
			t,
			blockMethod,
			verifyFn,
			expectedResult,
		)
	)

	// Create the client
	c := NewRPCClient(mockClient)

	// Get the result
	result, err := c.Block(&height)
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_BlockResults(t *testing.T) {
	t.Parallel()

	var (
		height = int64(10)

		expectedResult = &ctypes.ResultBlockResults{
			Height: height,
		}

		verifyFn = func(t *testing.T, params map[string]any) {
			t.Helper()

			assert.Equal(t, fmt.Sprintf("%d", height), params["height"])
		}

		mockClient = generateMockRequestClient(
			t,
			blockResultsMethod,
			verifyFn,
			expectedResult,
		)
	)

	// Create the client
	c := NewRPCClient(mockClient)

	// Get the result
	result, err := c.BlockResults(&height)
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_Commit(t *testing.T) {
	t.Parallel()

	var (
		height = int64(10)

		expectedResult = &ctypes.ResultCommit{
			CanonicalCommit: true,
		}

		verifyFn = func(t *testing.T, params map[string]any) {
			t.Helper()

			assert.Equal(t, fmt.Sprintf("%d", height), params["height"])
		}

		mockClient = generateMockRequestClient(
			t,
			commitMethod,
			verifyFn,
			expectedResult,
		)
	)

	// Create the client
	c := NewRPCClient(mockClient)

	// Get the result
	result, err := c.Commit(&height)
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_Tx(t *testing.T) {
	t.Parallel()

	var (
		hash = []byte("tx hash")

		expectedResult = &ctypes.ResultTx{
			Hash:   hash,
			Height: 10,
		}

		verifyFn = func(t *testing.T, params map[string]any) {
			t.Helper()

			assert.Equal(t, base64.StdEncoding.EncodeToString(hash), params["hash"])
		}

		mockClient = generateMockRequestClient(
			t,
			txMethod,
			verifyFn,
			expectedResult,
		)
	)

	// Create the client
	c := NewRPCClient(mockClient)

	// Get the result
	result, err := c.Tx(hash)
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_Validators(t *testing.T) {
	t.Parallel()

	var (
		height = int64(10)

		expectedResult = &ctypes.ResultValidators{
			BlockHeight: height,
		}

		verifyFn = func(t *testing.T, params map[string]any) {
			t.Helper()

			assert.Equal(t, fmt.Sprintf("%d", height), params["height"])
		}

		mockClient = generateMockRequestClient(
			t,
			validatorsMethod,
			verifyFn,
			expectedResult,
		)
	)

	// Create the client
	c := NewRPCClient(mockClient)

	// Get the result
	result, err := c.Validators(&height)
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_Batch(t *testing.T) {
	t.Parallel()

	convertResults := func(results []*ctypes.ResultStatus) []any {
		res := make([]any, len(results))

		for index, item := range results {
			res[index] = item
		}

		return res
	}

	var (
		expectedStatuses = []*ctypes.ResultStatus{
			{
				NodeInfo: p2p.NodeInfo{
					Moniker: "dummy",
				},
			},
			{
				NodeInfo: p2p.NodeInfo{
					Moniker: "dummy",
				},
			},
			{
				NodeInfo: p2p.NodeInfo{
					Moniker: "dummy",
				},
			},
		}

		verifyFn = func(t *testing.T, params map[string]any) {
			t.Helper()

			assert.Len(t, params, 0)
		}

		mockClient = generateMockRequestsClient(
			t,
			statusMethod,
			verifyFn,
			convertResults(expectedStatuses),
		)
	)

	// Create the client
	c := NewRPCClient(mockClient)

	// Create the batch
	batch := c.NewBatch()

	require.NoError(t, batch.Status())
	require.NoError(t, batch.Status())
	require.NoError(t, batch.Status())

	require.EqualValues(t, 3, batch.Count())

	// Send the batch
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	results, err := batch.Send(ctx)
	require.NoError(t, err)

	require.Len(t, results, len(expectedStatuses))

	for index, result := range results {
		castResult, ok := result.(*ctypes.ResultStatus)
		require.True(t, ok)

		assert.Equal(t, expectedStatuses[index], castResult)
	}
}
