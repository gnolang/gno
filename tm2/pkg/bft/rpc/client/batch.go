package client

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/gnolang/gno/tm2/pkg/amino"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/client"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

var errEmptyBatch = errors.New("RPC batch is empty")

type RPCBatch struct {
	batch rpcclient.Batch

	// resultMap maps the request ID -> result Amino type
	// Why?
	// There is a weird quirk in this RPC system where request results
	// are marshalled into Amino JSON, before being handed off to the client.
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

func (b *RPCBatch) addRequest(request rpctypes.RPCRequest, result any) {
	b.mux.Lock()
	defer b.mux.Unlock()

	// Save the result type
	b.resultMap[request.ID.String()] = result

	// Add the request to the batch
	b.batch.AddRequest(request)
}

func (b *RPCBatch) Status() error {
	// Prepare the RPC request
	request, err := newRequest(
		statusMethod,
		map[string]any{},
	)
	if err != nil {
		return fmt.Errorf("unable to create request, %w", err)
	}

	b.addRequest(request, &ctypes.ResultStatus{})

	return nil
}

func (b *RPCBatch) ABCIInfo() error {
	// Prepare the RPC request
	request, err := newRequest(
		abciInfoMethod,
		map[string]any{},
	)
	if err != nil {
		return fmt.Errorf("unable to create request, %w", err)
	}

	b.addRequest(request, &ctypes.ResultABCIInfo{})

	return nil
}

func (b *RPCBatch) ABCIQuery(path string, data []byte) error {
	return b.ABCIQueryWithOptions(path, data, DefaultABCIQueryOptions)
}

func (b *RPCBatch) ABCIQueryWithOptions(path string, data []byte, opts ABCIQueryOptions) error {
	// Prepare the RPC request
	request, err := newRequest(
		abciQueryMethod,
		map[string]any{
			"path":   path,
			"data":   data,
			"height": opts.Height,
			"prove":  opts.Prove,
		},
	)
	if err != nil {
		return fmt.Errorf("unable to create request, %w", err)
	}

	b.addRequest(request, &ctypes.ResultABCIQuery{})

	return nil
}

func (b *RPCBatch) BroadcastTxCommit(tx types.Tx) error {
	// Prepare the RPC request
	request, err := newRequest(
		broadcastTxCommitMethod,
		map[string]any{"tx": tx},
	)
	if err != nil {
		return fmt.Errorf("unable to create request, %w", err)
	}

	b.addRequest(request, &ctypes.ResultBroadcastTxCommit{})

	return nil
}

func (b *RPCBatch) BroadcastTxAsync(tx types.Tx) error {
	return b.broadcastTX(broadcastTxAsyncMethod, tx)
}

func (b *RPCBatch) BroadcastTxSync(tx types.Tx) error {
	return b.broadcastTX(broadcastTxSyncMethod, tx)
}

func (b *RPCBatch) broadcastTX(route string, tx types.Tx) error {
	// Prepare the RPC request
	request, err := newRequest(
		route,
		map[string]any{"tx": tx},
	)
	if err != nil {
		return fmt.Errorf("unable to create request, %w", err)
	}

	b.addRequest(request, &ctypes.ResultBroadcastTx{})

	return nil
}

func (b *RPCBatch) UnconfirmedTxs(limit int) error {
	// Prepare the RPC request
	request, err := newRequest(
		unconfirmedTxsMethod,
		map[string]any{"limit": limit},
	)
	if err != nil {
		return fmt.Errorf("unable to create request, %w", err)
	}

	b.addRequest(request, &ctypes.ResultUnconfirmedTxs{})

	return nil
}

func (b *RPCBatch) NumUnconfirmedTxs() error {
	// Prepare the RPC request
	request, err := newRequest(
		numUnconfirmedTxsMethod,
		map[string]any{},
	)
	if err != nil {
		return fmt.Errorf("unable to create request, %w", err)
	}

	b.addRequest(request, &ctypes.ResultUnconfirmedTxs{})

	return nil
}

func (b *RPCBatch) NetInfo() error {
	// Prepare the RPC request
	request, err := newRequest(
		netInfoMethod,
		map[string]any{},
	)
	if err != nil {
		return fmt.Errorf("unable to create request, %w", err)
	}

	b.addRequest(request, &ctypes.ResultNetInfo{})

	return nil
}

func (b *RPCBatch) DumpConsensusState() error {
	// Prepare the RPC request
	request, err := newRequest(
		dumpConsensusStateMethod,
		map[string]any{},
	)
	if err != nil {
		return fmt.Errorf("unable to create request, %w", err)
	}

	b.addRequest(request, &ctypes.ResultDumpConsensusState{})

	return nil
}

func (b *RPCBatch) ConsensusState() error {
	// Prepare the RPC request
	request, err := newRequest(
		consensusStateMethod,
		map[string]any{},
	)
	if err != nil {
		return fmt.Errorf("unable to create request, %w", err)
	}

	b.addRequest(request, &ctypes.ResultConsensusState{})

	return nil
}

func (b *RPCBatch) ConsensusParams(height *int64) error {
	params := map[string]any{}
	if height != nil {
		params["height"] = height
	}

	// Prepare the RPC request
	request, err := newRequest(
		consensusParamsMethod,
		params,
	)
	if err != nil {
		return fmt.Errorf("unable to create request, %w", err)
	}

	b.addRequest(request, &ctypes.ResultConsensusParams{})

	return nil
}

func (b *RPCBatch) Health() error {
	// Prepare the RPC request
	request, err := newRequest(
		healthMethod,
		map[string]any{},
	)
	if err != nil {
		return fmt.Errorf("unable to create request, %w", err)
	}

	b.addRequest(request, &ctypes.ResultHealth{})

	return nil
}

func (b *RPCBatch) BlockchainInfo(minHeight, maxHeight int64) error {
	// Prepare the RPC request
	request, err := newRequest(
		blockchainMethod,
		map[string]any{
			"minHeight": minHeight,
			"maxHeight": maxHeight,
		},
	)
	if err != nil {
		return fmt.Errorf("unable to create request, %w", err)
	}

	b.addRequest(request, &ctypes.ResultBlockchainInfo{})

	return nil
}

func (b *RPCBatch) Genesis() error {
	// Prepare the RPC request
	request, err := newRequest(genesisMethod, map[string]any{})
	if err != nil {
		return fmt.Errorf("unable to create request, %w", err)
	}

	b.addRequest(request, &ctypes.ResultGenesis{})

	return nil
}

func (b *RPCBatch) Block(height *int64) error {
	params := map[string]any{}
	if height != nil {
		params["height"] = height
	}

	// Prepare the RPC request
	request, err := newRequest(blockMethod, params)
	if err != nil {
		return fmt.Errorf("unable to create request, %w", err)
	}

	b.addRequest(request, &ctypes.ResultBlock{})

	return nil
}

func (b *RPCBatch) BlockResults(height *int64) error {
	params := map[string]any{}
	if height != nil {
		params["height"] = height
	}

	// Prepare the RPC request
	request, err := newRequest(blockResultsMethod, params)
	if err != nil {
		return fmt.Errorf("unable to create request, %w", err)
	}

	b.addRequest(request, &ctypes.ResultBlockResults{})

	return nil
}

func (b *RPCBatch) Commit(height *int64) error {
	params := map[string]any{}
	if height != nil {
		params["height"] = height
	}

	// Prepare the RPC request
	request, err := newRequest(commitMethod, params)
	if err != nil {
		return fmt.Errorf("unable to create request, %w", err)
	}

	b.addRequest(request, &ctypes.ResultCommit{})

	return nil
}

func (b *RPCBatch) Tx(hash []byte) error {
	// Prepare the RPC request
	request, err := newRequest(
		txMethod,
		map[string]interface{}{
			"hash": hash,
		},
	)
	if err != nil {
		return fmt.Errorf("unable to create request, %w", err)
	}

	b.addRequest(request, &ctypes.ResultTx{})

	return nil
}

func (b *RPCBatch) Validators(height *int64) error {
	params := map[string]any{}
	if height != nil {
		params["height"] = height
	}

	// Prepare the RPC request
	request, err := newRequest(validatorsMethod, params)
	if err != nil {
		return fmt.Errorf("unable to create request, %w", err)
	}

	b.addRequest(request, &ctypes.ResultValidators{})

	return nil
}
