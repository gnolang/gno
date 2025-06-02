package client

import (
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/std"
	"golang.org/x/net/context"
)

// Client defines the client interface for fetching chain data
type Client interface {
	// GetLatestBlockNumber returns the latest block height from the chain
	GetLatestBlockNumber() (uint64, error)

	// GetBlocks returns a slice of Block - including the block height and its
	// timestamp in milliseconds - in the requested range only if they contain
	// transactions
	GetBlocks(ctx context.Context, from, to uint64) ([]*Block, error)

	// GetTxResults returns the block transaction results (if any)
	GetTxResults(block uint64) ([]*abci.ResponseDeliverTx, error)
}

type Block struct {
	Txs       []std.Tx
	Height    uint64
	Timestamp int64
}
