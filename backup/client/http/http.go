package http

//nolint:revive // See https://github.com/gnolang/gno/issues/1197
import (
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

func (c *Client) GetBlockTransactions(blockNum uint64) ([]std.Tx, error) {
	// Fetch the block
	blockNumInt64 := int64(blockNum)

	block, err := c.client.Block(&blockNumInt64)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to fetch block, %w",
			err,
		)
	}

	// Decode amino transactions
	txs := make([]std.Tx, 0, len(block.Block.Data.Txs))

	for _, encodedTx := range block.Block.Data.Txs {
		var tx std.Tx

		if unmarshalErr := amino.Unmarshal(encodedTx, &tx); unmarshalErr != nil {
			return nil, fmt.Errorf(
				"unable to unmarshal amino tx, %w",
				err,
			)
		}

		txs = append(txs, tx)
	}

	return txs, nil
}
