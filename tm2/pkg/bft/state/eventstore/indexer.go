package eventstore

import "github.com/gnolang/gno/tm2/pkg/bft/types"

// TxEventStore stores transaction events for later processing
type TxEventStore interface {
	// Start starts the transaction event store
	Start() error

	// Stop stops the transaction event store
	Stop() error

	// GetType returns the event store type
	GetType() string

	// Index analyzes, indexes and stores a single transaction
	Index(result types.TxResult) error
}
