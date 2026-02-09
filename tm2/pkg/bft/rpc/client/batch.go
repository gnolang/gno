package client

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abciTypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/abci"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/blocks"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/consensus"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/health"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/mempool"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/net"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/status"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/tx"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/client"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

var errEmptyBatch = errors.New("RPC batch is empty")

type RPCBatch struct {
	batch rpcclient.Batch

	// resultMap maps the request ID -> result Amino type
	// Why?
	// There is a weird quirk in this RPC system where request results
	// are marshaled into Amino JSON, before being handed off to the client.
	// The client, of course, needs to unmarshal the Amino JSON-encoded response result
	// back into a concrete type.
	// Since working with an RPC batch is asynchronous
	// (requests are added at any time, but results are retrieved when the batch is sent)
	// there needs to be a record of what specific type the result needs to be Amino unmarshalled to
	resultMap map[string]any

	mux sync.RWMutex
}

func (b *RPCBatch) Count() int {
	b.mux.RLock()
	defer b.mux.RUnlock()

	return b.batch.Count()
}

func (b *RPCBatch) Clear() int {
	b.mux.Lock()
	defer b.mux.Unlock()

	return b.batch.Clear()
}

func (b *RPCBatch) Send(ctx context.Context) ([]any, error) {
	b.mux.Lock()
	defer b.mux.Unlock()

	// Save the initial batch size
	batchSize := b.batch.Count()

	// Sanity check for not sending empty batches
	if batchSize == 0 {
		return nil, errEmptyBatch
	}

	// Send the batch
	responses, err := b.batch.Send(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to send RPC batch, %w", err)
	}

	var (
		results = make([]any, 0, batchSize)
		errs    = make([]error, 0, batchSize)
	)

	// Parse the response results
	for _, response := range responses {
		// Check the error
		if response.Error != nil {
			errs = append(errs, response.Error)
			results = append(results, nil)

			continue
		}

		// Get the result type from the result map
		result, exists := b.resultMap[response.ID.String()]
		if !exists {
			return nil, fmt.Errorf("unexpected response with ID %s", response.ID)
		}

		// Amino JSON-unmarshal the response result
		if err := amino.UnmarshalJSON(response.Result, result); err != nil {
			return nil, fmt.Errorf("unable to parse response result, %w", err)
		}

		results = append(results, result)
	}

	return results, errors.Join(errs...)
}

func (b *RPCBatch) addRequest(request *spec.BaseJSONRequest, result any) {
	b.mux.Lock()
	defer b.mux.Unlock()

	// Save the result type
	b.resultMap[request.ID.String()] = result

	// Add the request to the batch
	b.batch.AddRequest(request)
}

func (b *RPCBatch) Status() {
	// Prepare the RPC request
	request := newRequest(
		statusMethod,
		nil,
	)

	b.addRequest(request, &status.ResultStatus{})
}

func (b *RPCBatch) ABCIInfo() {
	// Prepare the RPC request
	request := newRequest(
		abciInfoMethod,
		nil,
	)

	b.addRequest(request, &abciTypes.ResultABCIInfo{})
}

func (b *RPCBatch) ABCIQuery(path string, data []byte) {
	b.ABCIQueryWithOptions(path, data, DefaultABCIQueryOptions)
}

func (b *RPCBatch) ABCIQueryWithOptions(path string, data []byte, opts ABCIQueryOptions) {
	// Prepare the RPC request
	request := newRequest(
		abciQueryMethod,
		[]any{
			path,
			data,
			opts.Height,
			opts.Prove,
		},
	)

	b.addRequest(request, &abciTypes.ResultABCIQuery{})
}

func (b *RPCBatch) BroadcastTxCommit(tx types.Tx) {
	// Prepare the RPC request
	request := newRequest(
		broadcastTxCommitMethod,
		[]any{
			tx,
		},
	)

	b.addRequest(request, &mempool.ResultBroadcastTxCommit{})
}

func (b *RPCBatch) BroadcastTxAsync(tx types.Tx) {
	b.broadcastTX(broadcastTxAsyncMethod, tx)
}

func (b *RPCBatch) BroadcastTxSync(tx types.Tx) {
	b.broadcastTX(broadcastTxSyncMethod, tx)
}

func (b *RPCBatch) broadcastTX(route string, tx types.Tx) {
	// Prepare the RPC request
	request := newRequest(
		route,
		[]any{
			tx,
		},
	)

	b.addRequest(request, &mempool.ResultBroadcastTx{})
}

func (b *RPCBatch) UnconfirmedTxs(limit int) {
	// Prepare the RPC request
	request := newRequest(
		unconfirmedTxsMethod,
		[]any{
			limit,
		},
	)

	b.addRequest(request, &mempool.ResultUnconfirmedTxs{})
}

func (b *RPCBatch) NumUnconfirmedTxs() {
	// Prepare the RPC request
	request := newRequest(
		numUnconfirmedTxsMethod,
		nil,
	)

	b.addRequest(request, &mempool.ResultUnconfirmedTxs{})
}

func (b *RPCBatch) NetInfo() {
	// Prepare the RPC request
	request := newRequest(
		netInfoMethod,
		nil,
	)

	b.addRequest(request, &net.ResultNetInfo{})
}

func (b *RPCBatch) DumpConsensusState() {
	// Prepare the RPC request
	request := newRequest(
		dumpConsensusStateMethod,
		nil,
	)

	b.addRequest(request, &consensus.ResultDumpConsensusState{})
}

func (b *RPCBatch) ConsensusState() {
	// Prepare the RPC request
	request := newRequest(
		consensusStateMethod,
		nil,
	)

	b.addRequest(request, &consensus.ResultConsensusState{})
}

func (b *RPCBatch) ConsensusParams(height *int64) {
	var v int64
	if height != nil {
		v = *height
	}

	// Prepare the RPC request
	request := newRequest(
		consensusParamsMethod,
		[]any{
			v,
		},
	)

	b.addRequest(request, &consensus.ResultConsensusParams{})
}

func (b *RPCBatch) Health() {
	// Prepare the RPC request
	request := newRequest(
		healthMethod,
		nil,
	)

	b.addRequest(request, &health.ResultHealth{})
}

func (b *RPCBatch) BlockchainInfo(minHeight, maxHeight int64) {
	// Prepare the RPC request
	request := newRequest(
		blockchainMethod,
		[]any{
			minHeight,
			maxHeight,
		},
	)

	b.addRequest(request, &blocks.ResultBlockchainInfo{})
}

func (b *RPCBatch) Genesis() {
	// Prepare the RPC request
	request := newRequest(genesisMethod, nil)

	b.addRequest(request, &net.ResultGenesis{})
}

func (b *RPCBatch) Block(height *int64) {
	var v int64
	if height != nil {
		v = *height
	}

	// Prepare the RPC request
	request := newRequest(
		blockMethod,
		[]any{
			v,
		},
	)

	b.addRequest(request, &blocks.ResultBlock{})
}

func (b *RPCBatch) BlockResults(height *int64) {
	var v int64
	if height != nil {
		v = *height
	}

	// Prepare the RPC request
	request := newRequest(
		blockResultsMethod,
		[]any{
			v,
		},
	)

	b.addRequest(request, &blocks.ResultBlockResults{})
}

func (b *RPCBatch) Commit(height *int64) {
	var v int64
	if height != nil {
		v = *height
	}

	// Prepare the RPC request
	request := newRequest(
		commitMethod,
		[]any{
			v,
		},
	)

	b.addRequest(request, &blocks.ResultCommit{})
}

func (b *RPCBatch) Tx(hash []byte) {
	// Prepare the RPC request
	request := newRequest(
		txMethod,
		[]any{
			hash,
		},
	)

	b.addRequest(request, &tx.ResultTx{})
}

func (b *RPCBatch) Validators(height *int64) {
	var v int64
	if height != nil {
		v = *height
	}

	// Prepare the RPC request
	request := newRequest(
		validatorsMethod,
		[]any{
			v,
		},
	)

	b.addRequest(request, &consensus.ResultValidators{})
}
