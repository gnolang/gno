package core

import (
	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/abci"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/blocks"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/consensus"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/health"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/mempool"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/net"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/status"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/tx"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server"
	"github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/events"
)

// SetupABCI sets up the following endpoints:
//   - abci_info
//   - abci_query
func SetupABCI(server *server.JSONRPC, proxyAppConn appconn.Query) {
	h := abci.NewHandler(proxyAppConn)

	server.RegisterHandler(
		"abci_info",
		h.InfoHandler,
	)

	server.RegisterHandler(
		"abci_query",
		h.QueryHandler,
		"path", "data", "height", "prove",
	)
}

// SetupBlocks sets up the following endpoints:
//   - blockchain
//   - block
//   - commit
//   - block_results
func SetupBlocks(server *server.JSONRPC, store state.BlockStore, stateDB dbm.DB) {
	h := blocks.NewHandler(store, stateDB)

	server.RegisterHandler(
		"blockchain",
		h.BlockchainInfoHandler,
		"minHeight", "maxHeight",
	)

	server.RegisterHandler(
		"block",
		h.BlockHandler,
		"height",
	)

	server.RegisterHandler(
		"commit",
		h.CommitHandler,
		"height",
	)

	server.RegisterHandler(
		"block_results",
		h.BlockResultsHandler,
		"height",
	)
}

// SetupConsensus sets up the following endpoints:
//   - validators
//   - dump_consensus_state
//   - consensus_state
//   - consensus_params
func SetupConsensus(
	server *server.JSONRPC,
	consensusState consensus.Consensus,
	stateDB dbm.DB,
	peers ctypes.Peers,
) {
	h := consensus.NewHandler(consensusState, stateDB, peers)

	server.RegisterHandler(
		"validators",
		h.ValidatorsHandler,
		"height",
	)

	server.RegisterHandler(
		"dump_consensus_state",
		h.DumpConsensusStateHandler,
	)

	server.RegisterHandler(
		"consensus_state",
		h.ConsensusStateHandler,
	)

	server.RegisterHandler(
		"consensus_params",
		h.ConsensusParamsHandler,
		"height",
	)
}

// SetupHealth sets up the following endpoints:
//   - health
func SetupHealth(server *server.JSONRPC) {
	server.RegisterHandler(
		"health",
		health.HealthHandler,
	)
}

// SetupMempool sets up the following endpoints:
//   - broadcast_tx_async
//   - broadcast_tx_sync
//   - broadcast_tx_commit
//   - unconfirmed_txs
//   - num_unconfirmed_txs
func SetupMempool(
	server *server.JSONRPC,
	mp mempool.Mempool,
	evsw events.EventSwitch,
) {
	h := mempool.NewHandler(mp, evsw)

	server.RegisterHandler(
		"broadcast_tx_async",
		h.BroadcastTxAsyncHandler,
		"tx",
	)

	server.RegisterHandler(
		"broadcast_tx_sync",
		h.BroadcastTxSyncHandler,
		"tx",
	)

	server.RegisterHandler(
		"broadcast_tx_commit",
		h.BroadcastTxCommitHandler,
		"tx",
	)

	server.RegisterHandler(
		"unconfirmed_txs",
		h.UnconfirmedTxsHandler,
		"limit",
	)

	server.RegisterHandler(
		"num_unconfirmed_txs",
		h.NumUnconfirmedTxsHandler,
	)
}

// SetupNet sets up the following endpoints:
//   - net_info
//   - genesis
func SetupNet(
	server *server.JSONRPC,
	peers ctypes.Peers,
	transport ctypes.Transport,
	genesisDoc *types.GenesisDoc,
) {
	h := net.NewHandler(peers, transport, genesisDoc)

	server.RegisterHandler(
		"net_info",
		h.NetInfoHandler,
	)

	server.RegisterHandler(
		"genesis",
		h.GenesisHandler,
	)
}

// SetupTx sets up the following endpoints:
//   - tx
func SetupTx(
	server *server.JSONRPC,
	blockStore state.BlockStore,
	stateDB dbm.DB,
) {
	h := tx.NewHandler(blockStore, stateDB)

	server.RegisterHandler(
		"tx",
		h.TxHandler,
		"hash",
	)
}

// SetupStatus sets up the following endpoints:
//   - status
func SetupStatus(server *server.JSONRPC, buildFn status.BuildStatusFn) {
	h := status.NewHandler(buildFn)

	server.RegisterHandler(
		"status",
		h.StatusHandler,
		"heightGte",
	)
}
