package http

//nolint:revive // See https://github.com/gnolang/gno/issues/1197
import (
	"context"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	rpcClient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/std"

	_ "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
)

// Client is the TM2 HTTP client
type Client struct {
	client rpcClient.Client
}

// NewClient creates a new TM2 HTTP client
func NewClient(remote string) (*Client, error) {
	c, err := rpcClient.NewHTTPClient(remote)
	if err != nil {
		return nil, fmt.Errorf("unable to create HTTP client, %w", err)
	}

	return &Client{
		client: c,
	}, nil
}

func (c *Client) SendTransaction(ctx context.Context, tx *std.Tx) error {
	aminoTx, err := amino.Marshal(tx)
	if err != nil {
		return fmt.Errorf(
			"unable to marshal transaction to amino binary, %w",
			err,
		)
	}

	// Broadcast sync
	_, err = c.client.BroadcastTxSync(ctx, aminoTx)
	if err != nil {
		return fmt.Errorf(
			"unable to broadcast sync transaction, %w",
			err,
		)
	}

	return nil
}
