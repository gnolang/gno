package source

import (
	"context"

	"github.com/gnolang/gno/tm2/pkg/std"
)

// Source defines the interface for any
// source of transactions that need to be restored
type Source interface {
	// Next fetches the next transaction to be restored.
	// This call can be BLOCKING
	Next(context.Context) (*std.Tx, error)

	// Close shuts down the transaction source
	Close() error
}
