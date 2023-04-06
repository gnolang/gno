package client

/*
The client package provides a general purpose interface (Client) for connecting
to a tendermint node, as well as higher-level functionality.

The main implementation for production code is client.HTTP, which
connects via http to the jsonrpc interface of the tendermint node.

For connecting to a node running in the same process (eg. when
compiling the abci app in the same process), you can use the client.Local
implementation.

For mocking out server responses during testing to see behavior for
arbitrary return values, use the mock package.

In addition to the Client interface, which should be used externally
for maximum flexibility and testability, and two implementations,
this package also provides helper functions that work on any Client
implementation.
*/

import (
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

// Client wraps most important rpc calls a client would make.
//
// NOTE: Events cannot be subscribed to from the RPC APIs. For events
// subscriptions and filters and queries, an external API must be used that
// first synchronously consumes the events from the node's synchronous event
// switch, or reads logged events from the filesystem.
type Client interface {
	// service.Service
	ABCIClient
	HistoryClient
	NetworkClient
	SignClient
	StatusClient
	MempoolClient
}

// ABCIClient groups together the functionality that principally affects the
// ABCI app.
//
// In many cases this will be all we want, so we can accept an interface which
// is easier to mock.
type ABCIClient interface {
	// Reading from abci app
	ABCIInfo() (*ctypes.ResultABCIInfo, error)
	ABCIQuery(path string, data []byte) (*ctypes.ResultABCIQuery, error)
	ABCIQueryWithOptions(path string, data []byte,
		opts ABCIQueryOptions) (*ctypes.ResultABCIQuery, error)

	// Writing to abci app
	BroadcastTxCommit(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error)
	BroadcastTxAsync(tx types.Tx) (*ctypes.ResultBroadcastTx, error)
	BroadcastTxSync(tx types.Tx) (*ctypes.ResultBroadcastTx, error)
}

// SignClient groups together the functionality needed to get valid signatures
// and prove anything about the chain.
type SignClient interface {
	Block(height *int64) (*ctypes.ResultBlock, error)
	BlockResults(height *int64) (*ctypes.ResultBlockResults, error)
	Commit(height *int64) (*ctypes.ResultCommit, error)
	Validators(height *int64) (*ctypes.ResultValidators, error)
}

// HistoryClient provides access to data from genesis to now in large chunks.
type HistoryClient interface {
	Genesis() (*ctypes.ResultGenesis, error)
	BlockchainInfo(minHeight, maxHeight int64) (*ctypes.ResultBlockchainInfo, error)
}

// StatusClient provides access to general chain info.
type StatusClient interface {
	Status() (*ctypes.ResultStatus, error)
}

// NetworkClient is general info about the network state. May not be needed
// usually.
type NetworkClient interface {
	NetInfo() (*ctypes.ResultNetInfo, error)
	DumpConsensusState() (*ctypes.ResultDumpConsensusState, error)
	ConsensusState() (*ctypes.ResultConsensusState, error)
	ConsensusParams(height *int64) (*ctypes.ResultConsensusParams, error)
	Health() (*ctypes.ResultHealth, error)
}

// MempoolClient shows us data about current mempool state.
type MempoolClient interface {
	UnconfirmedTxs(limit int) (*ctypes.ResultUnconfirmedTxs, error)
	NumUnconfirmedTxs() (*ctypes.ResultUnconfirmedTxs, error)
}
