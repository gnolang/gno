package client

import (
	"context"

	"github.com/gnolang/gno/tm2/pkg/std"
)

// Client defines the client interface for sending tx data
type Client interface {
	// SendTransaction executes a broadcast sync send
	// of the specified transaction to the chain
	SendTransaction(context.Context, *std.Tx) error
}
