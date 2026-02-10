package client

import (
	"context"

	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
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
	ABCIInfo(ctx context.Context) (*ctypes.ResultABCIInfo, error)
	ABCIQuery(ctx context.Context, path string, data []byte) (*ctypes.ResultABCIQuery, error)
	ABCIQueryWithOptions(ctx context.Context, path string, data []byte,
		opts ABCIQueryOptions) (*ctypes.ResultABCIQuery, error)

	// Writing to abci app
	BroadcastTxCommit(ctx context.Context, tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error)
	BroadcastTxAsync(ctx context.Context, tx types.Tx) (*ctypes.ResultBroadcastTx, error)
	BroadcastTxSync(ctx context.Context, tx types.Tx) (*ctypes.ResultBroadcastTx, error)
}

// SignClient groups together the functionality needed to get valid signatures
// and prove anything about the chain.
type SignClient interface {
	Block(ctx context.Context, height *int64) (*ctypes.ResultBlock, error)
	BlockResults(ctx context.Context, height *int64) (*ctypes.ResultBlockResults, error)
	Commit(ctx context.Context, height *int64) (*ctypes.ResultCommit, error)
	Validators(ctx context.Context, height *int64) (*ctypes.ResultValidators, error)
}

// HistoryClient provides access to data from genesis to now in large chunks.
type HistoryClient interface {
	Genesis(ctx context.Context) (*ctypes.ResultGenesis, error)
	BlockchainInfo(ctx context.Context, minHeight, maxHeight int64) (*ctypes.ResultBlockchainInfo, error)
}

// StatusClient provides access to general chain info.
type StatusClient interface {
	Status(ctx context.Context, heightGte *int64) (*ctypes.ResultStatus, error)
}

// NetworkClient is general info about the network state. May not be needed
// usually.
type NetworkClient interface {
	NetInfo(ctx context.Context) (*ctypes.ResultNetInfo, error)
	DumpConsensusState(ctx context.Context) (*ctypes.ResultDumpConsensusState, error)
	ConsensusState(ctx context.Context) (*ctypes.ResultConsensusState, error)
	ConsensusParams(ctx context.Context, height *int64) (*ctypes.ResultConsensusParams, error)
	Health(ctx context.Context) (*ctypes.ResultHealth, error)
}

// MempoolClient shows us data about current mempool state.
type MempoolClient interface {
	UnconfirmedTxs(ctx context.Context, limit int) (*ctypes.ResultUnconfirmedTxs, error)
	NumUnconfirmedTxs(ctx context.Context) (*ctypes.ResultUnconfirmedTxs, error)
}

type TxClient interface {
	Tx(ctx context.Context, hash []byte) (*ctypes.ResultTx, error)
}
