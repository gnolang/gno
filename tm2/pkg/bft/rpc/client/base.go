package client

import (
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

var _ Client = (*baseRPCClient)(nil)

func (c *baseRPCClient) Status() (*ctypes.ResultStatus, error) {
	result := new(ctypes.ResultStatus)
	err := c.caller.Call("status", map[string]any{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "Status")
	}
	return result, nil
}

func (c *baseRPCClient) ABCIInfo() (*ctypes.ResultABCIInfo, error) {
	result := new(ctypes.ResultABCIInfo)
	err := c.caller.Call("abci_info", map[string]any{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "ABCIInfo")
	}
	return result, nil
}

func (c *baseRPCClient) ABCIQuery(path string, data []byte) (*ctypes.ResultABCIQuery, error) {
	return c.ABCIQueryWithOptions(path, data, DefaultABCIQueryOptions)
}

func (c *baseRPCClient) ABCIQueryWithOptions(path string, data []byte, opts ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
	result := new(ctypes.ResultABCIQuery)
	err := c.caller.Call("abci_query",
		map[string]any{"path": path, "data": data, "height": opts.Height, "prove": opts.Prove},
		result)
	if err != nil {
		return nil, errors.Wrap(err, "ABCIQuery")
	}
	return result, nil
}

func (c *baseRPCClient) BroadcastTxCommit(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	result := new(ctypes.ResultBroadcastTxCommit)
	err := c.caller.Call("broadcast_tx_commit", map[string]any{"tx": tx}, result)
	if err != nil {
		return nil, errors.Wrap(err, "broadcast_tx_commit")
	}
	return result, nil
}

func (c *baseRPCClient) BroadcastTxAsync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return c.broadcastTX("broadcast_tx_async", tx)
}

func (c *baseRPCClient) BroadcastTxSync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return c.broadcastTX("broadcast_tx_sync", tx)
}

func (c *baseRPCClient) broadcastTX(route string, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	result := new(ctypes.ResultBroadcastTx)
	err := c.caller.Call(route, map[string]any{"tx": tx}, result)
	if err != nil {
		return nil, errors.Wrap(err, route)
	}
	return result, nil
}

func (c *baseRPCClient) UnconfirmedTxs(limit int) (*ctypes.ResultUnconfirmedTxs, error) {
	result := new(ctypes.ResultUnconfirmedTxs)
	err := c.caller.Call("unconfirmed_txs", map[string]any{"limit": limit}, result)
	if err != nil {
		return nil, errors.Wrap(err, "unconfirmed_txs")
	}
	return result, nil
}

func (c *baseRPCClient) NumUnconfirmedTxs() (*ctypes.ResultUnconfirmedTxs, error) {
	result := new(ctypes.ResultUnconfirmedTxs)
	err := c.caller.Call("num_unconfirmed_txs", map[string]any{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "num_unconfirmed_txs")
	}
	return result, nil
}

func (c *baseRPCClient) NetInfo() (*ctypes.ResultNetInfo, error) {
	result := new(ctypes.ResultNetInfo)
	err := c.caller.Call("net_info", map[string]any{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "NetInfo")
	}
	return result, nil
}

func (c *baseRPCClient) DumpConsensusState() (*ctypes.ResultDumpConsensusState, error) {
	result := new(ctypes.ResultDumpConsensusState)
	err := c.caller.Call("dump_consensus_state", map[string]any{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "DumpConsensusState")
	}
	return result, nil
}

func (c *baseRPCClient) ConsensusState() (*ctypes.ResultConsensusState, error) {
	result := new(ctypes.ResultConsensusState)
	err := c.caller.Call("consensus_state", map[string]any{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "ConsensusState")
	}
	return result, nil
}

func (c *baseRPCClient) ConsensusParams(height *int64) (*ctypes.ResultConsensusParams, error) {
	result := new(ctypes.ResultConsensusParams)

	if err := c.caller.Call(
		"consensus_params",
		map[string]any{
			"height": height,
		},
		result,
	); err != nil {
		return nil, errors.Wrap(err, "ConsensusParams")
	}

	return result, nil
}

func (c *baseRPCClient) Health() (*ctypes.ResultHealth, error) {
	result := new(ctypes.ResultHealth)
	err := c.caller.Call("health", map[string]any{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "Health")
	}
	return result, nil
}

func (c *baseRPCClient) BlockchainInfo(minHeight, maxHeight int64) (*ctypes.ResultBlockchainInfo, error) {
	result := new(ctypes.ResultBlockchainInfo)
	err := c.caller.Call("blockchain",
		map[string]any{"minHeight": minHeight, "maxHeight": maxHeight},
		result)
	if err != nil {
		return nil, errors.Wrap(err, "BlockchainInfo")
	}
	return result, nil
}

func (c *baseRPCClient) Genesis() (*ctypes.ResultGenesis, error) {
	result := new(ctypes.ResultGenesis)
	err := c.caller.Call("genesis", map[string]any{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "Genesis")
	}
	return result, nil
}

func (c *baseRPCClient) Block(height *int64) (*ctypes.ResultBlock, error) {
	result := new(ctypes.ResultBlock)
	err := c.caller.Call("block", map[string]any{"height": height}, result)
	if err != nil {
		return nil, errors.Wrap(err, "Block")
	}
	return result, nil
}

func (c *baseRPCClient) BlockResults(height *int64) (*ctypes.ResultBlockResults, error) {
	result := new(ctypes.ResultBlockResults)
	err := c.caller.Call("block_results", map[string]any{"height": height}, result)
	if err != nil {
		return nil, errors.Wrap(err, "Block Result")
	}
	return result, nil
}

func (c *baseRPCClient) Commit(height *int64) (*ctypes.ResultCommit, error) {
	result := new(ctypes.ResultCommit)
	err := c.caller.Call("commit", map[string]any{"height": height}, result)
	if err != nil {
		return nil, errors.Wrap(err, "Commit")
	}
	return result, nil
}

func (c *baseRPCClient) Tx(hash []byte, prove bool) (*ctypes.ResultTx, error) {
	result := new(ctypes.ResultTx)
	params := map[string]any{
		"hash":  hash,
		"prove": prove,
	}
	err := c.caller.Call("tx", params, result)
	if err != nil {
		return nil, errors.Wrap(err, "Tx")
	}
	return result, nil
}

func (c *baseRPCClient) TxSearch(query string, prove bool, page, perPage int) (*ctypes.ResultTxSearch, error) {
	result := new(ctypes.ResultTxSearch)
	params := map[string]any{
		"query":    query,
		"prove":    prove,
		"page":     page,
		"per_page": perPage,
	}
	err := c.caller.Call("tx_search", params, result)
	if err != nil {
		return nil, errors.Wrap(err, "TxSearch")
	}
	return result, nil
}

func (c *baseRPCClient) Validators(height *int64) (*ctypes.ResultValidators, error) {
	result := new(ctypes.ResultValidators)
	params := map[string]any{}
	if height != nil {
		params["height"] = height
	}
	err := c.caller.Call("validators", params, result)
	if err != nil {
		return nil, errors.Wrap(err, "Validators")
	}
	return result, nil
}
