package rpc

//nolint:revive // See https://github.com/gnolang/gno/issues/1197
import (
	"context"
	"errors"
	"fmt"

	_ "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	rpcClient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/std"

	"github.com/gnolang/tx-archive/backup/client"
)

var _ client.Client = &Client{}

// Client is the TM2 RPC client
type Client struct {
	client *rpcClient.RPCClient
}

// NewHTTPClient creates a new TM2 HTTP RPC client
func NewHTTPClient(remote string) (*Client, error) {
	c, err := rpcClient.NewHTTPClient(remote)
	if err != nil {
		return nil, fmt.Errorf("unable to create HTTP client, %w", err)
	}

	return &Client{
		client: c,
	}, nil
}

// NewWSClient creates a new TM2 WebSocket RPC client
func NewWSClient(remote string) (*Client, error) {
	c, err := rpcClient.NewWSClient(remote)
	if err != nil {
		return nil, fmt.Errorf("unable to create WebSocket client, %w", err)
	}

	return &Client{
		client: c,
	}, nil
}

func (c *Client) GetLatestBlockNumber() (uint64, error) {
	status, err := c.client.Status()
	if err != nil {
		return 0, fmt.Errorf(
			"unable to fetch latest block number, %w",
			err,
		)
	}

	return uint64(status.SyncInfo.LatestBlockHeight), nil
}

func (c *Client) GetBlocks(ctx context.Context, from, to uint64) ([]*client.Block, error) {
	// Check if the block range is valid
	if from > to {
		return nil, fmt.Errorf(
			"invalid block range, from (%d) bigger than to (%d)",
			from,
			to,
		)
	}

	// Prepare batch of requests
	batch := c.client.NewBatch()

	for currBlock := from; currBlock <= to; currBlock++ {
		// Add current block to the batch
		currBlockInt64 := int64(currBlock)

		if err := batch.Block(&currBlockInt64); err != nil {
			return nil, fmt.Errorf(
				"unable to batch block request, %w",
				err,
			)
		}
	}

	// Send batch of requests
	results, err := batch.Send(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to send batch of request, %w",
			err,
		)
	}

	// Gather blocks containing transaction in RPC results
	var blocks []*client.Block

	for _, result := range results {
		blockRes, ok := result.(*ctypes.ResultBlock)
		if !ok {
			return nil, errors.New("unable to cast request result to TxData")
		}

		// If block contain transaction, gather them
		if len(blockRes.Block.Data.Txs) > 0 {
			txs := make([]std.Tx, 0, len(blockRes.Block.Data.Txs))

			// Decode amino transactions
			for _, encodedTx := range blockRes.Block.Data.Txs {
				var tx std.Tx

				if unmarshalErr := amino.Unmarshal(encodedTx, &tx); unmarshalErr != nil {
					return nil, fmt.Errorf(
						"unable to unmarshal amino tx, %w",
						err,
					)
				}

				txs = append(txs, tx)
			}

			// Add block including transactions, timestamp and block height to slice
			blocks = append(blocks, &client.Block{
				Timestamp: blockRes.Block.Time.UnixMilli(),
				Height:    uint64(blockRes.Block.Height),
				Txs:       txs,
			})
		}
	}

	return blocks, nil
}
