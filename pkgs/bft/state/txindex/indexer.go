package txindex

import "github.com/gnolang/gno/pkgs/bft/types"

// TxIndexer indexes transactions for later processing
type TxIndexer interface {
	// Start starts the transaction indexer
	Start() error

	// Close stops the transaction indexer
	Close() error

	// Index analyzes, indexes and stores a single transaction
	Index(result *types.TxResult) error
}
