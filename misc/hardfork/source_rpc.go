package main

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	bftypes "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// rpcSource fetches chain state from a live (or recently-halted) node via RPC.
type rpcSource struct {
	rpcURL string
	client *rpcclient.RPCClient
}

func newRPCSource(rpcURL string) (*rpcSource, error) {
	client, err := rpcclient.NewHTTPClient(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("creating RPC client for %s: %w", rpcURL, err)
	}
	return &rpcSource{rpcURL: rpcURL, client: client}, nil
}

func (s *rpcSource) Description() string { return "RPC" }
func (s *rpcSource) Close() error        { return s.client.Close() }

func (s *rpcSource) FetchGenesis(ctx context.Context) (*bftypes.GenesisDoc, error) {
	res, err := s.client.Genesis(ctx)
	if err != nil {
		return nil, fmt.Errorf("RPC genesis call: %w", err)
	}
	return res.Genesis, nil
}

func (s *rpcSource) LatestHeight(ctx context.Context) (int64, error) {
	res, err := s.client.Status(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("RPC status call: %w", err)
	}
	return res.SyncInfo.LatestBlockHeight, nil
}

// FetchTxs iterates every block from fromHeight to toHeight, extracts
// successful transactions, and wraps them in TxWithMetadata.
//
// This is intentionally block-by-block rather than using TxSearch because:
// - TxSearch is not guaranteed to be available on all nodes
// - Block iteration gives us the exact block height and timestamp
// - We need metadata (BlockHeight, Timestamp, ChainID) for the hardfork replay
//
// For large chains this is slow (one RPC call per block). For production use
// the local dir source (which reads the block store directly) is faster.
func (s *rpcSource) FetchTxs(ctx context.Context, fromHeight, toHeight int64, io commands.IO) ([]gnoland.TxWithMetadata, error) {
	var txs []gnoland.TxWithMetadata

	// Get chain ID from genesis (needed for metadata)
	genesis, err := s.FetchGenesis(ctx)
	if err != nil {
		return nil, err
	}
	chainID := genesis.ChainID

	total := toHeight - fromHeight + 1
	var processed, txCount int64

	for h := fromHeight; h <= toHeight; h++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		processed++
		if processed%1000 == 0 || processed == total {
			io.Printf("\r  Blocks: %d/%d  Txs: %d", processed, total, txCount)
		}

		// Fetch block
		block, err := s.client.Block(ctx, &h)
		if err != nil {
			return nil, fmt.Errorf("fetching block %d: %w", h, err)
		}

		if len(block.Block.Data.Txs) == 0 {
			continue
		}

		// Fetch block results to filter out failed txs
		results, err := s.client.BlockResults(ctx, &h)
		if err != nil {
			return nil, fmt.Errorf("fetching block results %d: %w", h, err)
		}

		timestamp := block.Block.Header.Time.Unix()

		for i, rawTx := range block.Block.Data.Txs {
			// Skip failed transactions
			if i < len(results.Results.DeliverTxs) && results.Results.DeliverTxs[i].IsErr() {
				continue
			}

			// Decode the raw transaction bytes
			var stdTx std.Tx
			if err := amino.Unmarshal(rawTx, &stdTx); err != nil {
				// Skip malformed txs with a warning
				io.Printf("\n  WARNING: could not decode tx at height %d index %d: %v\n", h, i, err)
				continue
			}

			txs = append(txs, gnoland.TxWithMetadata{
				Tx: stdTx,
				Metadata: &gnoland.GnoTxMetadata{
					Timestamp:   timestamp,
					BlockHeight: h,
					ChainID:     chainID,
				},
			})
			txCount++
		}
	}

	io.Printf("\r  Blocks: %d/%d  Txs: %d\n", processed, total, txCount)
	return txs, nil
}
