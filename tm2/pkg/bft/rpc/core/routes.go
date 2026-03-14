package core

import (
	rpc "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server"
)

// TODO: better system than "unsafe" prefix
// NOTE: Amino is registered in rpc/core/types/codec.go.
var Routes = map[string]*rpc.RPCFunc{
	// info API
	"health":               rpc.NewRPCFunc(Health, ""),
	"status":               rpc.NewRPCFunc(Status, "heightGte"),
	"net_info":             rpc.NewRPCFunc(NetInfo, ""),
	"blockchain":           rpc.NewRPCFunc(BlockchainInfo, "minHeight,maxHeight"),
	"genesis":              rpc.NewRPCFunc(Genesis, ""),
	"block":                rpc.NewRPCFunc(Block, "height"),
	"block_results":        rpc.NewRPCFunc(BlockResults, "height"),
	"commit":               rpc.NewRPCFunc(Commit, "height"),
	"tx":                   rpc.NewRPCFunc(Tx, "hash"),
	"validators":           rpc.NewRPCFunc(Validators, "height"),
	"dump_consensus_state": rpc.NewRPCFunc(DumpConsensusState, ""),
	"consensus_state":      rpc.NewRPCFunc(ConsensusState, ""),
	"consensus_params":     rpc.NewRPCFunc(ConsensusParams, "height"),
	"unconfirmed_txs":      rpc.NewRPCFunc(UnconfirmedTxs, "limit"),
	"num_unconfirmed_txs":  rpc.NewRPCFunc(NumUnconfirmedTxs, ""),

	// tx broadcast API
	"broadcast_tx_commit": rpc.NewRPCFunc(BroadcastTxCommit, "tx"),
	"broadcast_tx_sync":   rpc.NewRPCFunc(BroadcastTxSync, "tx"),
	"broadcast_tx_async":  rpc.NewRPCFunc(BroadcastTxAsync, "tx"),

	// abci API
	"abci_query": rpc.NewRPCFunc(ABCIQuery, "path,data,height,prove"),
	"abci_info":  rpc.NewRPCFunc(ABCIInfo, ""),

	// blocks backup API
	"backup": rpc.NewWSRPCFunc(BackupBlocks, "start,end"),
}

func AddUnsafeRoutes() {
	// control API
	Routes["unsafe_flush_mempool"] = rpc.NewRPCFunc(UnsafeFlushMempool, "")

	// profiler API
	Routes["unsafe_start_cpu_profiler"] = rpc.NewRPCFunc(UnsafeStartCPUProfiler, "filename")
	Routes["unsafe_stop_cpu_profiler"] = rpc.NewRPCFunc(UnsafeStopCPUProfiler, "")
	Routes["unsafe_write_heap_profile"] = rpc.NewRPCFunc(UnsafeWriteHeapProfile, "filename")
}
