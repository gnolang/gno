package client

import (
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Client defines the client interface for fetching chain data
type Client interface {
	// GetLatestBlockNumber returns the latest block height from the chain
	GetLatestBlockNumber() (uint64, error)

	// GetBlock returns the transactions contained
	// within the specified block, if any, apart from the block height and
	// its timestamp in milliseconds.
	GetBlock(uint64) (*Block, error)
}

type Block struct {
	Txs       []std.Tx
	Height    uint64
	Timestamp int64
}
