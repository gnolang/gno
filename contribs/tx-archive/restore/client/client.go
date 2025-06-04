package client

import "github.com/gnolang/gno/tm2/pkg/std"

// Client defines the client interface for sending tx data
type Client interface {
	// SendTransaction executes a broadcast sync send
	// of the specified transaction to the chain
	SendTransaction(*std.Tx) error
}
