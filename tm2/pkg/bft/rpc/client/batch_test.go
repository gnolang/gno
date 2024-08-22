package client

import (
	"context"
	"testing"

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

// generateMockBatchClient generates a common
// mock batch handling client
func generateMockBatchClient(
	t *testing.T,
	method string,
	expectedRequests int,
	commonResult any,
) *mockClient {
	t.Helper()

	return &mockClient{
		sendBatchFn: func(_ context.Context, requests types.RPCRequests) (types.RPCResponses, error) {
			require.Len(t, requests, expectedRequests)

			responses := make(types.RPCResponses, len(requests))

			for index, request := range requests {
				require.Equal(t, "2.0", request.JSONRPC)
				require.NotEmpty(t, request.ID)
				require.Equal(t, method, request.Method)

				result, err := amino.MarshalJSON(commonResult)
				require.NoError(t, err)

				response := types.RPCResponse{
					JSONRPC: "2.0",
					ID:      request.ID,
					Result:  result,
					Error:   nil,
				}

				responses[index] = response
			}

			return responses, nil
		},
	}
}

func TestRPCBatch_Count(t *testing.T) {
	t.Parallel()

	var (
		c     = NewRPCClient(&mockClient{})
		batch = c.NewBatch()
	)

	// Make sure the batch is initially empty
	assert.Equal(t, 0, batch.Count())

	// Add a dummy request
	require.NoError(t, batch.Status())

	// Make sure the request is enqueued
	assert.Equal(t, 1, batch.Count())
}

func TestRPCBatch_Clear(t *testing.T) {
	t.Parallel()

	var (
		c     = NewRPCClient(&mockClient{})
		batch = c.NewBatch()
	)

	// Add a dummy request
	require.NoError(t, batch.Status())

	// Make sure the request is enqueued
	assert.Equal(t, 1, batch.Count())

	// Clear the batch
	assert.Equal(t, 1, batch.Clear())

	// Make sure no request is enqueued
	assert.Equal(t, 0, batch.Count())
}

func TestRPCBatch_Send(t *testing.T) {
	t.Parallel()

	t.Run("empty batch", func(t *testing.T) {
		t.Parallel()

		var (
			c     = NewRPCClient(&mockClient{})
			batch = c.NewBatch()
		)

		res, err := batch.Send(context.Background())

		assert.ErrorIs(t, err, errEmptyBatch)
		assert.Nil(t, res)
	})

	t.Run("valid batch", func(t *testing.T) {
		t.Parallel()

		var (
			numRequests    = 10
			expectedStatus = &ctypes.ResultStatus{
				NodeInfo: p2p.NodeInfo{
					Moniker: "dummy",
				},
			}

			mockClient = generateMockBatchClient(t, statusMethod, 10, expectedStatus)

			c     = NewRPCClient(mockClient)
			batch = c.NewBatch()
		)

		// Enqueue the requests
		for i := 0; i < numRequests; i++ {
			require.NoError(t, batch.Status())
		}

		// Send the batch
		results, err := batch.Send(context.Background())
		require.NoError(t, err)

		// Validate the results
		assert.Len(t, results, numRequests)

		for _, result := range results {
			castResult, ok := result.(*ctypes.ResultStatus)
			require.True(t, ok)

			assert.Equal(t, expectedStatus, castResult)
		}
	})
}

func TestRPCBatch_Endpoints(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		method          string
		expectedResult  any
		batchCallback   func(*RPCBatch)
		extractCallback func(any) any
	}{
		{
			statusMethod,
			&ctypes.ResultStatus{
				NodeInfo: p2p.NodeInfo{
					Moniker: "dummy",
				},
			},
			func(batch *RPCBatch) {
				require.NoError(t, batch.Status())
			},
			func(result any) any {
				castResult, ok := result.(*ctypes.ResultStatus)
				require.True(t, ok)

				return castResult
			},
		},
		{
			abciInfoMethod,
			&ctypes.ResultABCIInfo{
				Response: abci.ResponseInfo{
					LastBlockAppHash: []byte("dummy"),
				},
			},
			func(batch *RPCBatch) {
				require.NoError(t, batch.ABCIInfo())
			},
			func(result any) any {
				castResult, ok := result.(*ctypes.ResultABCIInfo)
				require.True(t, ok)

				return castResult
			},
		},
		{
			abciQueryMethod,
			&ctypes.ResultABCIQuery{
				Response: abci.ResponseQuery{
					Value: []byte("dummy"),
				},
			},
			func(batch *RPCBatch) {
				require.NoError(t, batch.ABCIQuery("path", []byte("dummy")))
			},
			func(result any) any {
				castResult, ok := result.(*ctypes.ResultABCIQuery)
				require.True(t, ok)

				return castResult
			},
		},
		{
			broadcastTxCommitMethod,
			&ctypes.ResultBroadcastTxCommit{
				Hash: []byte("dummy"),
			},
			func(batch *RPCBatch) {
				require.NoError(t, batch.BroadcastTxCommit([]byte("dummy")))
			},
			func(result any) any {
				castResult, ok := result.(*ctypes.ResultBroadcastTxCommit)
				require.True(t, ok)

				return castResult
			},
		},
		{
			broadcastTxAsyncMethod,
			&ctypes.ResultBroadcastTx{
				Hash: []byte("dummy"),
			},
			func(batch *RPCBatch) {
				require.NoError(t, batch.BroadcastTxAsync([]byte("dummy")))
			},
			func(result any) any {
				castResult, ok := result.(*ctypes.ResultBroadcastTx)
				require.True(t, ok)

				return castResult
			},
		},
		{
			broadcastTxSyncMethod,
			&ctypes.ResultBroadcastTx{
				Hash: []byte("dummy"),
			},
			func(batch *RPCBatch) {
				require.NoError(t, batch.BroadcastTxSync([]byte("dummy")))
			},
			func(result any) any {
				castResult, ok := result.(*ctypes.ResultBroadcastTx)
				require.True(t, ok)

				return castResult
			},
		},
		{
			unconfirmedTxsMethod,
			&ctypes.ResultUnconfirmedTxs{
				Count: 10,
			},
			func(batch *RPCBatch) {
				require.NoError(t, batch.UnconfirmedTxs(0))
			},
			func(result any) any {
				castResult, ok := result.(*ctypes.ResultUnconfirmedTxs)
				require.True(t, ok)

				return castResult
			},
		},
		{
			numUnconfirmedTxsMethod,
			&ctypes.ResultUnconfirmedTxs{
				Count: 10,
			},
			func(batch *RPCBatch) {
				require.NoError(t, batch.NumUnconfirmedTxs())
			},
			func(result any) any {
				castResult, ok := result.(*ctypes.ResultUnconfirmedTxs)
				require.True(t, ok)

				return castResult
			},
		},
		{
			netInfoMethod,
			&ctypes.ResultNetInfo{
				NPeers: 10,
			},
			func(batch *RPCBatch) {
				require.NoError(t, batch.NetInfo())
			},
			func(result any) any {
				castResult, ok := result.(*ctypes.ResultNetInfo)
				require.True(t, ok)

				return castResult
			},
		},
		{
			dumpConsensusStateMethod,
			&ctypes.ResultDumpConsensusState{
				RoundState: &cstypes.RoundState{
					Round: 10,
				},
			},
			func(batch *RPCBatch) {
				require.NoError(t, batch.DumpConsensusState())
			},
			func(result any) any {
				castResult, ok := result.(*ctypes.ResultDumpConsensusState)
				require.True(t, ok)

				return castResult
			},
		},
		{
			consensusStateMethod,
			&ctypes.ResultConsensusState{
				RoundState: cstypes.RoundStateSimple{
					ProposalBlockHash: []byte("dummy"),
				},
			},
			func(batch *RPCBatch) {
				require.NoError(t, batch.ConsensusState())
			},
			func(result any) any {
				castResult, ok := result.(*ctypes.ResultConsensusState)
				require.True(t, ok)

				return castResult
			},
		},
		{
			consensusParamsMethod,
			&ctypes.ResultConsensusParams{
				BlockHeight: 10,
			},
			func(batch *RPCBatch) {
				require.NoError(t, batch.ConsensusParams(nil))
			},
			func(result any) any {
				castResult, ok := result.(*ctypes.ResultConsensusParams)
				require.True(t, ok)

				return castResult
			},
		},
		{
			healthMethod,
			&ctypes.ResultHealth{},
			func(batch *RPCBatch) {
				require.NoError(t, batch.Health())
			},
			func(result any) any {
				castResult, ok := result.(*ctypes.ResultHealth)
				require.True(t, ok)

				return castResult
			},
		},
		{
			blockchainMethod,
			&ctypes.ResultBlockchainInfo{
				LastHeight: 100,
			},
			func(batch *RPCBatch) {
				require.NoError(t, batch.BlockchainInfo(0, 0))
			},
			func(result any) any {
				castResult, ok := result.(*ctypes.ResultBlockchainInfo)
				require.True(t, ok)

				return castResult
			},
		},
		{
			genesisMethod,
			&ctypes.ResultGenesis{
				Genesis: &bfttypes.GenesisDoc{
					ChainID: "dummy",
				},
			},
			func(batch *RPCBatch) {
				require.NoError(t, batch.Genesis())
			},
			func(result any) any {
				castResult, ok := result.(*ctypes.ResultGenesis)
				require.True(t, ok)

				return castResult
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
			func(batch *RPCBatch) {
				require.NoError(t, batch.Block(nil))
			},
			func(result any) any {
				castResult, ok := result.(*ctypes.ResultBlock)
				require.True(t, ok)

				return castResult
			},
		},
		{
			blockResultsMethod,
			&ctypes.ResultBlockResults{
				Height: 10,
			},
			func(batch *RPCBatch) {
				require.NoError(t, batch.BlockResults(nil))
			},
			func(result any) any {
				castResult, ok := result.(*ctypes.ResultBlockResults)
				require.True(t, ok)

				return castResult
			},
		},
		{
			commitMethod,
			&ctypes.ResultCommit{
				CanonicalCommit: true,
			},
			func(batch *RPCBatch) {
				require.NoError(t, batch.Commit(nil))
			},
			func(result any) any {
				castResult, ok := result.(*ctypes.ResultCommit)
				require.True(t, ok)

				return castResult
			},
		},
		{
			txMethod,
			&ctypes.ResultTx{
				Hash:   []byte("tx hash"),
				Height: 10,
			},
			func(batch *RPCBatch) {
				require.NoError(t, batch.Tx([]byte("tx hash")))
			},
			func(result any) any {
				castResult, ok := result.(*ctypes.ResultTx)
				require.True(t, ok)

				return castResult
			},
		},
		{
			validatorsMethod,
			&ctypes.ResultValidators{
				BlockHeight: 10,
			},
			func(batch *RPCBatch) {
				require.NoError(t, batch.Validators(nil))
			},
			func(result any) any {
				castResult, ok := result.(*ctypes.ResultValidators)
				require.True(t, ok)

				return castResult
			},
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.method, func(t *testing.T) {
			t.Parallel()

			var (
				numRequests = 10
				mockClient  = generateMockBatchClient(
					t,
					testCase.method,
					numRequests,
					testCase.expectedResult,
				)

				c     = NewRPCClient(mockClient)
				batch = c.NewBatch()
			)

			// Enqueue the requests
			for i := 0; i < numRequests; i++ {
				testCase.batchCallback(batch)
			}

			// Send the batch
			results, err := batch.Send(context.Background())
			require.NoError(t, err)

			// Validate the results
			assert.Len(t, results, numRequests)

			for _, result := range results {
				castResult := testCase.extractCallback(result)

				assert.Equal(t, testCase.expectedResult, castResult)
			}
		})
	}
}
