package eventstore

import "github.com/gnolang/gno/tm2/pkg/bft/types"

const (
	StatusOn  = "on"
	StatusOff = "off"
)

// TxEventStore stores transaction events for later processing
type TxEventStore interface {
	// Start starts the transaction event store
	Start() error

	// Stop stops the transaction event store
	Stop() error

	// GetType returns the event store type
	GetType() string

	// Append analyzes and appends a single transaction
	// to the event store
	Append(result types.TxResult) error
}
