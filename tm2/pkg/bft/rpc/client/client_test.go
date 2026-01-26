package client

import (
	"context"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	abciTypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/abci"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/blocks"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/consensus"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/health"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/mempool"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/net"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/status"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/tx"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	bfttypes "github.com/gnolang/gno/tm2/pkg/bft/types"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateMockRequestClient generates a single RPC request mock client
func generateMockRequestClient(
	t *testing.T,
	method string,
	verifyParamsFn func(*testing.T, []any),
	responseData any,
) *mockClient {
	t.Helper()

	return &mockClient{
		sendRequestFn: func(
			_ context.Context,
			request *spec.BaseJSONRequest,
		) (*spec.BaseJSONResponse, error) {
			// Validate the request
			require.Equal(t, "2.0", request.JSONRPC)
			require.NotNil(t, request.ID)
			require.Equal(t, request.Method, method)

			// Validate the params
			verifyParamsFn(t, request.Params)

			// Prepare the result
			result, err := amino.MarshalJSON(responseData)
			require.NoError(t, err)

			// Prepare the response
			response := &spec.BaseJSONResponse{
				Result: result, // direct
				Error:  nil,
				BaseJSON: spec.BaseJSON{
					JSONRPC: spec.JSONRPCVersion,
					ID:      request.ID,
				},
			}

			return response, nil
		},
	}
}

// generateMockRequestsClient generates a batch RPC request mock client
func generateMockRequestsClient(
	t *testing.T,
	method string,
	verifyParamsFn func(*testing.T, []any),
	responseData []any,
) *mockClient {
	t.Helper()

	return &mockClient{
		sendBatchFn: func(
			_ context.Context,
			requests spec.BaseJSONRequests,
		) (spec.BaseJSONResponses, error) {
			responses := make(spec.BaseJSONResponses, 0, len(requests))

			// Validate the requests
			for index, r := range requests {
				require.Equal(t, "2.0", r.JSONRPC)
				require.NotNil(t, r.ID)
				require.Equal(t, r.Method, method)

				// Validate the params
				verifyParamsFn(t, r.Params)

				// Prepare the result
				result, err := amino.MarshalJSON(responseData[index])
				require.NoError(t, err)

				// Prepare the response
				response := &spec.BaseJSONResponse{
					Result: result, // direct
					Error:  nil,
					BaseJSON: spec.BaseJSON{
						JSONRPC: spec.JSONRPCVersion,
						ID:      r.ID,
					},
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
		expectedStatus = &status.ResultStatus{
			NodeInfo: p2pTypes.NodeInfo{
				Moniker: "dummy",
			},
		}

		verifyFn = func(t *testing.T, params []any) {
			t.Helper()

			assert.Len(t, params, 1)
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
	status, err := c.Status(context.Background(), nil)
	require.NoError(t, err)

	assert.Equal(t, expectedStatus, status)
}

func TestRPCClient_ABCIInfo(t *testing.T) {
	t.Parallel()

	var (
		expectedInfo = &abciTypes.ResultABCIInfo{
			Response: abci.ResponseInfo{
				LastBlockAppHash: []byte("dummy"),
			},
		}

		verifyFn = func(t *testing.T, params []any) {
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
	info, err := c.ABCIInfo(context.Background())
	require.NoError(t, err)

	assert.Equal(t, expectedInfo, info)
}

func TestRPCClient_ABCIQuery(t *testing.T) {
	t.Parallel()

	var (
		path = "path"
		data = []byte("data")
		opts = DefaultABCIQueryOptions

		expectedQuery = &abciTypes.ResultABCIQuery{
			Response: abci.ResponseQuery{
				Value: []byte("dummy"),
			},
		}

		verifyFn = func(t *testing.T, params []any) {
			t.Helper()

			assert.Equal(t, path, params[0])
			assert.Equal(t, data, params[1])
			assert.Equal(t, opts.Height, params[2])
			assert.Equal(t, opts.Prove, params[3])
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
	query, err := c.ABCIQuery(context.Background(), path, data)
	require.NoError(t, err)

	assert.Equal(t, expectedQuery, query)
}

func TestRPCClient_BroadcastTxCommit(t *testing.T) {
	t.Parallel()

	var (
		tx = []byte("tx")

		expectedTxCommit = &mempool.ResultBroadcastTxCommit{
			Hash: []byte("dummy"),
		}

		verifyFn = func(t *testing.T, params []any) {
			t.Helper()

			assert.Equal(t, bfttypes.Tx(tx), params[0])
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
	txCommit, err := c.BroadcastTxCommit(context.Background(), tx)
	require.NoError(t, err)

	assert.Equal(t, expectedTxCommit, txCommit)
}

func TestRPCClient_BroadcastTxAsync(t *testing.T) {
	t.Parallel()

	var (
		tx = []byte("tx")

		expectedTxBroadcast = &mempool.ResultBroadcastTx{
			Hash: []byte("dummy"),
		}

		verifyFn = func(t *testing.T, params []any) {
			t.Helper()

			assert.Equal(t, bfttypes.Tx(tx), params[0])
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
	txAsync, err := c.BroadcastTxAsync(context.Background(), tx)
	require.NoError(t, err)

	assert.Equal(t, expectedTxBroadcast, txAsync)
}

func TestRPCClient_BroadcastTxSync(t *testing.T) {
	t.Parallel()

	var (
		tx = []byte("tx")

		expectedTxBroadcast = &mempool.ResultBroadcastTx{
			Hash: []byte("dummy"),
		}

		verifyFn = func(t *testing.T, params []any) {
			t.Helper()

			assert.Equal(t, bfttypes.Tx(tx), params[0])
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
	txSync, err := c.BroadcastTxSync(context.Background(), tx)
	require.NoError(t, err)

	assert.Equal(t, expectedTxBroadcast, txSync)
}

func TestRPCClient_UnconfirmedTxs(t *testing.T) {
	t.Parallel()

	var (
		limit = 10

		expectedResult = &mempool.ResultUnconfirmedTxs{
			Count: 10,
		}

		verifyFn = func(t *testing.T, params []any) {
			t.Helper()

			assert.Equal(t, limit, params[0])
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
	result, err := c.UnconfirmedTxs(context.Background(), limit)
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_NumUnconfirmedTxs(t *testing.T) {
	t.Parallel()

	var (
		expectedResult = &mempool.ResultUnconfirmedTxs{
			Count: 10,
		}

		verifyFn = func(t *testing.T, params []any) {
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
	result, err := c.NumUnconfirmedTxs(context.Background())
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_NetInfo(t *testing.T) {
	t.Parallel()

	var (
		expectedResult = &net.ResultNetInfo{
			NPeers: 10,
		}

		verifyFn = func(t *testing.T, params []any) {
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
	result, err := c.NetInfo(context.Background())
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_DumpConsensusState(t *testing.T) {
	t.Parallel()

	var (
		expectedResult = &consensus.ResultDumpConsensusState{
			RoundState: &cstypes.RoundState{
				Round: 10,
			},
		}

		verifyFn = func(t *testing.T, params []any) {
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
	result, err := c.DumpConsensusState(context.Background())
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_ConsensusState(t *testing.T) {
	t.Parallel()

	var (
		expectedResult = &consensus.ResultConsensusState{
			RoundState: cstypes.RoundStateSimple{
				ProposalBlockHash: []byte("dummy"),
			},
		}

		verifyFn = func(t *testing.T, params []any) {
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
	result, err := c.ConsensusState(context.Background())
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_ConsensusParams(t *testing.T) {
	t.Parallel()

	var (
		blockHeight = int64(10)

		expectedResult = &consensus.ResultConsensusParams{
			BlockHeight: blockHeight,
		}

		verifyFn = func(t *testing.T, params []any) {
			t.Helper()

			assert.Equal(t, blockHeight, params[0])
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
	result, err := c.ConsensusParams(context.Background(), &blockHeight)
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_Health(t *testing.T) {
	t.Parallel()

	var (
		expectedResult = &health.ResultHealth{}

		verifyFn = func(t *testing.T, params []any) {
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
	result, err := c.Health(context.Background())
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_BlockchainInfo(t *testing.T) {
	t.Parallel()

	var (
		minHeight = int64(5)
		maxHeight = int64(10)

		expectedResult = &blocks.ResultBlockchainInfo{
			LastHeight: 100,
		}

		verifyFn = func(t *testing.T, params []any) {
			t.Helper()

			assert.Equal(t, minHeight, params[0])
			assert.Equal(t, maxHeight, params[1])
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
	result, err := c.BlockchainInfo(context.Background(), minHeight, maxHeight)
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_Genesis(t *testing.T) {
	t.Parallel()

	var (
		expectedResult = &net.ResultGenesis{
			Genesis: &bfttypes.GenesisDoc{
				ChainID: "dummy",
			},
		}

		verifyFn = func(t *testing.T, params []any) {
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
	result, err := c.Genesis(context.Background())
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_Block(t *testing.T) {
	t.Parallel()

	var (
		height = int64(10)

		expectedResult = &blocks.ResultBlock{
			BlockMeta: &bfttypes.BlockMeta{
				Header: bfttypes.Header{
					Height: height,
				},
			},
		}

		verifyFn = func(t *testing.T, params []any) {
			t.Helper()

			assert.Equal(t, height, params[0])
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
	result, err := c.Block(context.Background(), &height)
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_BlockResults(t *testing.T) {
	t.Parallel()

	var (
		height = int64(10)

		expectedResult = &blocks.ResultBlockResults{
			Height: height,
		}

		verifyFn = func(t *testing.T, params []any) {
			t.Helper()

			assert.Equal(t, height, params[0])
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
	result, err := c.BlockResults(context.Background(), &height)
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_Commit(t *testing.T) {
	t.Parallel()

	var (
		height = int64(10)

		expectedResult = &blocks.ResultCommit{
			CanonicalCommit: true,
		}

		verifyFn = func(t *testing.T, params []any) {
			t.Helper()

			assert.Equal(t, height, params[0])
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
	result, err := c.Commit(context.Background(), &height)
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_Tx(t *testing.T) {
	t.Parallel()

	var (
		hash = []byte("tx hash")

		expectedResult = &tx.ResultTx{
			Hash:   hash,
			Height: 10,
		}

		verifyFn = func(t *testing.T, params []any) {
			t.Helper()

			assert.Equal(t, hash, params[0])
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
	result, err := c.Tx(context.Background(), hash)
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_Validators(t *testing.T) {
	t.Parallel()

	var (
		height = int64(10)

		expectedResult = &consensus.ResultValidators{
			BlockHeight: height,
		}

		verifyFn = func(t *testing.T, params []any) {
			t.Helper()

			assert.Equal(t, height, params[0])
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
	result, err := c.Validators(context.Background(), &height)
	require.NoError(t, err)

	assert.Equal(t, expectedResult, result)
}

func TestRPCClient_Batch(t *testing.T) {
	t.Parallel()

	convertResults := func(results []*status.ResultStatus) []any {
		res := make([]any, len(results))

		for index, item := range results {
			res[index] = item
		}

		return res
	}

	var (
		expectedStatuses = []*status.ResultStatus{
			{
				NodeInfo: p2pTypes.NodeInfo{
					Moniker: "dummy",
				},
			},
			{
				NodeInfo: p2pTypes.NodeInfo{
					Moniker: "dummy",
				},
			},
			{
				NodeInfo: p2pTypes.NodeInfo{
					Moniker: "dummy",
				},
			},
		}

		verifyFn = func(t *testing.T, params []any) {
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

	batch.Status()
	batch.Status()
	batch.Status()

	require.EqualValues(t, 3, batch.Count())

	// Send the batch
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	results, err := batch.Send(ctx)
	require.NoError(t, err)

	require.Len(t, results, len(expectedStatuses))

	for index, result := range results {
		castResult, ok := result.(*status.ResultStatus)
		require.True(t, ok)

		assert.Equal(t, expectedStatuses[index], castResult)
	}
}
