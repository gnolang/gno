package main

import (
	"context"
	"encoding/json"
	"fmt"

	tm2Client "github.com/gnolang/faucet/client/http"
	"github.com/gnolang/gno/tm2/pkg/amino"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// nodeClient is the client the faucet uses to talk to the node. It reads
// accounts itself and relies on the faucet library's client for everything
// else.
type nodeClient struct {
	*tm2Client.Client

	rpc rpcclient.Client
}

// newNodeClient creates the client the faucet uses to talk to the node at
// remote.
func newNodeClient(remote string) (*nodeClient, error) {
	cli, err := tm2Client.NewClient(remote)
	if err != nil {
		return nil, fmt.Errorf("unable to create TM2 client, %w", err)
	}

	rpc, err := rpcclient.NewHTTPClient(remote)
	if err != nil {
		return nil, fmt.Errorf("unable to create RPC client, %w", err)
	}

	return &nodeClient{
		Client: cli,
		rpc:    rpc,
	}, nil
}

// GetAccount fetches the account at address.
//
// A gno.land node returns its own account type: a std.BaseAccount next to
// fields the faucet does not use. Amino rejects JSON fields that its target
// type lacks, so the response is split with encoding/json, which ignores them,
// and only the BaseAccount field is decoded with amino. That decoding stays
// strict: a BaseAccount field that this tree's std.BaseAccount lacks is an
// error.
//
// An address with no account yields an error wrapping std.UnknownAddressError.
func (c *nodeClient) GetAccount(address crypto.Address) (std.Account, error) {
	path := fmt.Sprintf("auth/accounts/%s", address.String())

	res, err := c.rpc.ABCIQuery(context.Background(), path, []byte{})
	if err != nil {
		return nil, fmt.Errorf("unable to query account, %w", err)
	}
	if res.Response.Error != nil {
		return nil, fmt.Errorf("node rejected the account query, %w", res.Response.Error)
	}

	// The node answers null for an address that has no account.
	data := res.Response.Data
	if len(data) == 0 || string(data) == "null" {
		return nil, fmt.Errorf("unable to fetch account %s, %w", address, std.UnknownAddressError{})
	}

	var account struct {
		BaseAccount json.RawMessage `json:"BaseAccount"`
	}
	if err := json.Unmarshal(data, &account); err != nil {
		return nil, fmt.Errorf("unable to decode account, %w", err)
	}

	var base std.BaseAccount
	if err := amino.UnmarshalJSON(account.BaseAccount, &base); err != nil {
		return nil, fmt.Errorf("unable to decode account, %w", err)
	}

	// The faucet signs with the key of the returned account's address, so the
	// account must be the one requested.
	if base.Address != address {
		return nil, fmt.Errorf("node returned the account of %s for %s", base.Address, address)
	}

	return &base, nil
}
