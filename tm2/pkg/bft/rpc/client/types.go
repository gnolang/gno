package client

import (
	"context"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/abci"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/blocks"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/consensus"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/health"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/mempool"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/net"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/status"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/tx"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

// ABCIQueryOptions can be used to provide options for ABCIQuery call other
// than the DefaultABCIQueryOptions.
type ABCIQueryOptions struct {
	Height int64
	Prove  bool
}

// DefaultABCIQueryOptions are latest height (0) and prove false.
var DefaultABCIQueryOptions = ABCIQueryOptions{Height: 0, Prove: false}

// Client wraps most important rpc calls a client would make.
//
// NOTE: Events cannot be subscribed to from the RPC APIs. For events
// subscriptions and filters and queries, an external API must be used that
// first synchronously consumes the events from the node's synchronous event
// switch, or reads logged events from the filesystem.
type Client interface {
	ABCIClient
	HistoryClient
	NetworkClient
	SignClient
	StatusClient
	MempoolClient
	TxClient
}

// ABCIClient groups together the functionality that principally affects the
// ABCI app.
//
// In many cases this will be all we want, so we can accept an interface which
// is easier to mock.
type ABCIClient interface {
	// Reading from abci app
	ABCIInfo(ctx context.Context) (*abci.ResultABCIInfo, error)
	ABCIQuery(ctx context.Context, path string, data []byte) (*abci.ResultABCIQuery, error)
	ABCIQueryWithOptions(ctx context.Context, path string, data []byte,
		opts ABCIQueryOptions) (*abci.ResultABCIQuery, error)

	// Writing to abci app
	BroadcastTxCommit(ctx context.Context, tx types.Tx) (*mempool.ResultBroadcastTxCommit, error)
	BroadcastTxAsync(ctx context.Context, tx types.Tx) (*mempool.ResultBroadcastTx, error)
	BroadcastTxSync(ctx context.Context, tx types.Tx) (*mempool.ResultBroadcastTx, error)
}

// SignClient groups together the functionality needed to get valid signatures
// and prove anything about the chain.
type SignClient interface {
	Block(ctx context.Context, height *int64) (*blocks.ResultBlock, error)
	BlockResults(ctx context.Context, height *int64) (*blocks.ResultBlockResults, error)
	Commit(ctx context.Context, height *int64) (*blocks.ResultCommit, error)
	Validators(ctx context.Context, height *int64) (*consensus.ResultValidators, error)
}

// HistoryClient provides access to data from genesis to now in large chunks.
type HistoryClient interface {
	Genesis(ctx context.Context) (*net.ResultGenesis, error)
	BlockchainInfo(ctx context.Context, minHeight, maxHeight int64) (*blocks.ResultBlockchainInfo, error)
}

// StatusClient provides access to general chain info.
type StatusClient interface {
	Status(ctx context.Context, heightGte *int64) (*status.ResultStatus, error)
}

// NetworkClient is general info about the network state. May not be needed
// usually.
type NetworkClient interface {
	NetInfo(ctx context.Context) (*net.ResultNetInfo, error)
	DumpConsensusState(ctx context.Context) (*consensus.ResultDumpConsensusState, error)
	ConsensusState(ctx context.Context) (*consensus.ResultConsensusState, error)
	ConsensusParams(ctx context.Context, height *int64) (*consensus.ResultConsensusParams, error)
	Health(ctx context.Context) (*health.ResultHealth, error)
}

// MempoolClient shows us data about current mempool state.
type MempoolClient interface {
	UnconfirmedTxs(ctx context.Context, limit int) (*mempool.ResultUnconfirmedTxs, error)
	NumUnconfirmedTxs(ctx context.Context) (*mempool.ResultUnconfirmedTxs, error)
}

type TxClient interface {
	Tx(ctx context.Context, hash []byte) (*tx.ResultTx, error)
}
