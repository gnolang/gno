package gnoclient

import (
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
)

// Client provides an interface for interacting with the blockchain.
type Client struct {
	Signer    Signer           // Signer for transaction authentication
	RPCClient rpcclient.Client // RPC client for blockchain communication
}

// validateSigner checks that the signer is correctly configured.
func (c *Client) validateSigner() error {
	if c.Signer == nil {
		return ErrMissingSigner
	}
	return nil
}

// validateRPCClient checks that the RPCClient is correctly configured.
func (c *Client) validateRPCClient() error {
	if c.RPCClient == nil {
		return ErrMissingRPCClient
	}
	return nil
}
