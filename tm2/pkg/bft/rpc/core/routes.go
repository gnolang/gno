package core

import (
	rpc "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server"
)

// NOTE: Amino is registered in rpc/core/types/codec.go.

// Routes builds the RPC route map for this Environment. Each route binds a
// method value on env, so requests dispatch to this specific Environment's
// state without going through package globals.
//
// If unsafe is true, the "unsafe_*" routes (mempool flush, CPU/heap
// profiler) are included.
func (env *Environment) Routes(unsafe bool) map[string]*rpc.RPCFunc {
	routes := map[string]*rpc.RPCFunc{
		// info API
		"health":               rpc.NewRPCFunc(env.Health, ""),
		"status":               rpc.NewRPCFunc(env.Status, "heightGte"),
		"net_info":             rpc.NewRPCFunc(env.NetInfo, ""),
		"blockchain":           rpc.NewRPCFunc(env.BlockchainInfo, "minHeight,maxHeight"),
		"genesis":              rpc.NewRPCFunc(env.Genesis, ""),
		"block":                rpc.NewRPCFunc(env.Block, "height"),
		"block_results":        rpc.NewRPCFunc(env.BlockResults, "height"),
		"commit":               rpc.NewRPCFunc(env.Commit, "height"),
		"tx":                   rpc.NewRPCFunc(env.Tx, "hash"),
		"validators":           rpc.NewRPCFunc(env.Validators, "height"),
		"dump_consensus_state": rpc.NewRPCFunc(env.DumpConsensusState, ""),
		"consensus_state":      rpc.NewRPCFunc(env.ConsensusState, ""),
		"consensus_params":     rpc.NewRPCFunc(env.ConsensusParams, "height"),
		"unconfirmed_txs":      rpc.NewRPCFunc(env.UnconfirmedTxs, "limit"),
		"num_unconfirmed_txs":  rpc.NewRPCFunc(env.NumUnconfirmedTxs, ""),

		// tx broadcast API
		"broadcast_tx_commit": rpc.NewRPCFunc(env.BroadcastTxCommit, "tx"),
		"broadcast_tx_sync":   rpc.NewRPCFunc(env.BroadcastTxSync, "tx"),
		"broadcast_tx_async":  rpc.NewRPCFunc(env.BroadcastTxAsync, "tx"),

		// abci API
		"abci_query": rpc.NewRPCFunc(env.ABCIQuery, "path,data,height,prove"),
		"abci_info":  rpc.NewRPCFunc(env.ABCIInfo, ""),
	}

	if unsafe {
		// control API
		routes["unsafe_flush_mempool"] = rpc.NewRPCFunc(env.UnsafeFlushMempool, "")
		// profiler API
		routes["unsafe_start_cpu_profiler"] = rpc.NewRPCFunc(env.UnsafeStartCPUProfiler, "filename")
		routes["unsafe_stop_cpu_profiler"] = rpc.NewRPCFunc(env.UnsafeStopCPUProfiler, "")
		routes["unsafe_write_heap_profile"] = rpc.NewRPCFunc(env.UnsafeWriteHeapProfile, "filename")
	}

	return routes
}
