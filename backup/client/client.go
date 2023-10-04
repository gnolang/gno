package client

import "github.com/gnolang/gno/tm2/pkg/std"

// Client defines the client interface for fetching chain data
type Client interface {
	// GetLatestBlockNumber returns the latest block height from the chain
	GetLatestBlockNumber() (uint64, error)

	// GetBlockTransactions returns the transactions contained
	// within the specified block, if any
	GetBlockTransactions(uint64) ([]std.Tx, error)
}
