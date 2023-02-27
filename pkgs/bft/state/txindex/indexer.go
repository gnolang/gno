package txindex

import "github.com/gnolang/gno/pkgs/bft/types"

// TxIndexer indexes transactions for later processing
type TxIndexer interface {
	// Start starts the transaction indexer
	Start() error

	// Stop stops the transaction indexer
	Stop() error

	// GetType returns the indexer type
	GetType() string

	// Index analyzes, indexes and stores a single transaction
	Index(result *types.TxResult) error
}
