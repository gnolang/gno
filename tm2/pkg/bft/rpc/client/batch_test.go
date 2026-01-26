package client

import (
	"context"
	"testing"

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
		sendBatchFn: func(_ context.Context, requests spec.BaseJSONRequests) (spec.BaseJSONResponses, error) {
			require.Len(t, requests, expectedRequests)

			responses := make(spec.BaseJSONResponses, len(requests))

			for index, request := range requests {
				require.Equal(t, "2.0", request.JSONRPC)
				require.NotEmpty(t, request.ID)
				require.Equal(t, method, request.Method)

				result, err := amino.MarshalJSON(commonResult)
				require.NoError(t, err)

				response := &spec.BaseJSONResponse{
					Result: result,
					Error:  nil,
					BaseJSON: spec.BaseJSON{
						JSONRPC: spec.JSONRPCVersion,
						ID:      request.ID,
					},
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
	batch.Status()

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
	batch.Status()

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
			expectedStatus = &status.ResultStatus{
				NodeInfo: p2pTypes.NodeInfo{
					Moniker: "dummy",
				},
			}

			mockClient = generateMockBatchClient(t, statusMethod, 10, expectedStatus)

			c     = NewRPCClient(mockClient)
			batch = c.NewBatch()
		)

		// Enqueue the requests
		for range numRequests {
			batch.Status()
		}

		// Send the batch
		results, err := batch.Send(context.Background())
		require.NoError(t, err)

		// Validate the results
		assert.Len(t, results, numRequests)

		for _, result := range results {
			castResult, ok := result.(*status.ResultStatus)
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
			&status.ResultStatus{
				NodeInfo: p2pTypes.NodeInfo{
					Moniker: "dummy",
				},
			},
			func(batch *RPCBatch) {
				batch.Status()
			},
			func(result any) any {
				castResult, ok := result.(*status.ResultStatus)
				require.True(t, ok)

				return castResult
			},
		},
		{
			abciInfoMethod,
			&abciTypes.ResultABCIInfo{
				Response: abci.ResponseInfo{
					LastBlockAppHash: []byte("dummy"),
				},
			},
			func(batch *RPCBatch) {
				batch.ABCIInfo()
			},
			func(result any) any {
				castResult, ok := result.(*abciTypes.ResultABCIInfo)
				require.True(t, ok)

				return castResult
			},
		},
		{
			abciQueryMethod,
			&abciTypes.ResultABCIQuery{
				Response: abci.ResponseQuery{
					Value: []byte("dummy"),
				},
			},
			func(batch *RPCBatch) {
				batch.ABCIQuery("path", []byte("dummy"))
			},
			func(result any) any {
				castResult, ok := result.(*abciTypes.ResultABCIQuery)
				require.True(t, ok)

				return castResult
			},
		},
		{
			broadcastTxCommitMethod,
			&mempool.ResultBroadcastTxCommit{
				Hash: []byte("dummy"),
			},
			func(batch *RPCBatch) {
				batch.BroadcastTxCommit([]byte("dummy"))
			},
			func(result any) any {
				castResult, ok := result.(*mempool.ResultBroadcastTxCommit)
				require.True(t, ok)

				return castResult
			},
		},
		{
			broadcastTxAsyncMethod,
			&mempool.ResultBroadcastTx{
				Hash: []byte("dummy"),
			},
			func(batch *RPCBatch) {
				batch.BroadcastTxAsync([]byte("dummy"))
			},
			func(result any) any {
				castResult, ok := result.(*mempool.ResultBroadcastTx)
				require.True(t, ok)

				return castResult
			},
		},
		{
			broadcastTxSyncMethod,
			&mempool.ResultBroadcastTx{
				Hash: []byte("dummy"),
			},
			func(batch *RPCBatch) {
				batch.BroadcastTxSync([]byte("dummy"))
			},
			func(result any) any {
				castResult, ok := result.(*mempool.ResultBroadcastTx)
				require.True(t, ok)

				return castResult
			},
		},
		{
			unconfirmedTxsMethod,
			&mempool.ResultUnconfirmedTxs{
				Count: 10,
			},
			func(batch *RPCBatch) {
				batch.UnconfirmedTxs(0)
			},
			func(result any) any {
				castResult, ok := result.(*mempool.ResultUnconfirmedTxs)
				require.True(t, ok)

				return castResult
			},
		},
		{
			numUnconfirmedTxsMethod,
			&mempool.ResultUnconfirmedTxs{
				Count: 10,
			},
			func(batch *RPCBatch) {
				batch.NumUnconfirmedTxs()
			},
			func(result any) any {
				castResult, ok := result.(*mempool.ResultUnconfirmedTxs)
				require.True(t, ok)

				return castResult
			},
		},
		{
			netInfoMethod,
			&net.ResultNetInfo{
				NPeers: 10,
			},
			func(batch *RPCBatch) {
				batch.NetInfo()
			},
			func(result any) any {
				castResult, ok := result.(*net.ResultNetInfo)
				require.True(t, ok)

				return castResult
			},
		},
		{
			dumpConsensusStateMethod,
			&consensus.ResultDumpConsensusState{
				RoundState: &cstypes.RoundState{
					Round: 10,
				},
			},
			func(batch *RPCBatch) {
				batch.DumpConsensusState()
			},
			func(result any) any {
				castResult, ok := result.(*consensus.ResultDumpConsensusState)
				require.True(t, ok)

				return castResult
			},
		},
		{
			consensusStateMethod,
			&consensus.ResultConsensusState{
				RoundState: cstypes.RoundStateSimple{
					ProposalBlockHash: []byte("dummy"),
				},
			},
			func(batch *RPCBatch) {
				batch.ConsensusState()
			},
			func(result any) any {
				castResult, ok := result.(*consensus.ResultConsensusState)
				require.True(t, ok)

				return castResult
			},
		},
		{
			consensusParamsMethod,
			&consensus.ResultConsensusParams{
				BlockHeight: 10,
			},
			func(batch *RPCBatch) {
				batch.ConsensusParams(nil)
			},
			func(result any) any {
				castResult, ok := result.(*consensus.ResultConsensusParams)
				require.True(t, ok)

				return castResult
			},
		},
		{
			healthMethod,
			&health.ResultHealth{},
			func(batch *RPCBatch) {
				batch.Health()
			},
			func(result any) any {
				castResult, ok := result.(*health.ResultHealth)
				require.True(t, ok)

				return castResult
			},
		},
		{
			blockchainMethod,
			&blocks.ResultBlockchainInfo{
				LastHeight: 100,
			},
			func(batch *RPCBatch) {
				batch.BlockchainInfo(0, 0)
			},
			func(result any) any {
				castResult, ok := result.(*blocks.ResultBlockchainInfo)
				require.True(t, ok)

				return castResult
			},
		},
		{
			genesisMethod,
			&net.ResultGenesis{
				Genesis: &bfttypes.GenesisDoc{
					ChainID: "dummy",
				},
			},
			func(batch *RPCBatch) {
				batch.Genesis()
			},
			func(result any) any {
				castResult, ok := result.(*net.ResultGenesis)
				require.True(t, ok)

				return castResult
			},
		},
		{
			blockMethod,
			&blocks.ResultBlock{
				BlockMeta: &bfttypes.BlockMeta{
					Header: bfttypes.Header{
						Height: 10,
					},
				},
			},
			func(batch *RPCBatch) {
				batch.Block(nil)
			},
			func(result any) any {
				castResult, ok := result.(*blocks.ResultBlock)
				require.True(t, ok)

				return castResult
			},
		},
		{
			blockResultsMethod,
			&blocks.ResultBlockResults{
				Height: 10,
			},
			func(batch *RPCBatch) {
				batch.BlockResults(nil)
			},
			func(result any) any {
				castResult, ok := result.(*blocks.ResultBlockResults)
				require.True(t, ok)

				return castResult
			},
		},
		{
			commitMethod,
			&blocks.ResultCommit{
				CanonicalCommit: true,
			},
			func(batch *RPCBatch) {
				batch.Commit(nil)
			},
			func(result any) any {
				castResult, ok := result.(*blocks.ResultCommit)
				require.True(t, ok)

				return castResult
			},
		},
		{
			txMethod,
			&tx.ResultTx{
				Hash:   []byte("tx hash"),
				Height: 10,
			},
			func(batch *RPCBatch) {
				batch.Tx([]byte("tx hash"))
			},
			func(result any) any {
				castResult, ok := result.(*tx.ResultTx)
				require.True(t, ok)

				return castResult
			},
		},
		{
			validatorsMethod,
			&consensus.ResultValidators{
				BlockHeight: 10,
			},
			func(batch *RPCBatch) {
				batch.Validators(nil)
			},
			func(result any) any {
				castResult, ok := result.(*consensus.ResultValidators)
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
			for range numRequests {
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
