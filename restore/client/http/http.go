package http

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	rpcClient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Client is the TM2 HTTP client
type Client struct {
	client rpcClient.Client
}

// NewClient creates a new TM2 HTTP client
func NewClient(remote string) *Client {
	return &Client{
		client: rpcClient.NewHTTP(remote, ""),
	}
}

func (c *Client) SendTransaction(tx *std.Tx) error {
	aminoTx, err := amino.Marshal(tx)
	if err != nil {
		return fmt.Errorf(
			"unable to marshal transaction to amino binary, %w",
			err,
		)
	}

	// Broadcast sync
	_, err = c.client.BroadcastTxSync(aminoTx)
	if err != nil {
		return fmt.Errorf(
			"unable to broadcast sync transaction, %w",
			err,
		)
	}

	return nil
}
