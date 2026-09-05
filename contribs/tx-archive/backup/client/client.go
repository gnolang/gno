package client

import (
	"context"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Client defines the client interface for fetching chain data
type Client interface {
	// GetLatestBlockNumber returns the latest block height from the chain
	GetLatestBlockNumber() (uint64, error)

	// GetChainID returns the chain ID of the source chain
	GetChainID() (string, error)

	// GetBlocks returns a slice of Block - including the block height and its
	// timestamp in milliseconds - in the requested range only if they contain
	// transactions
	GetBlocks(ctx context.Context, from, to uint64) ([]*Block, error)

	// GetTxResults returns the block transaction results (if any)
	GetTxResults(block uint64) ([]*abci.ResponseDeliverTx, error)

	// GetAccountAtHeight returns the (account_number, sequence) pair for
	// the given address at the given block height. Used by the hardfork-
	// metadata signer-info resolver to anchor brute-force sequence search.
	// Returns (0, 0, nil) when the account doesn't exist yet at that height.
	GetAccountAtHeight(addr crypto.Address, height uint64) (accNum, sequence uint64, err error)
}

type Block struct {
	Txs       []std.Tx
	Height    uint64
	Timestamp int64
}
