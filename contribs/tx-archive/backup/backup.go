package backup

//nolint:revive // See https://github.com/gnolang/gno/issues/1197
import (
	"context"
	"fmt"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	_ "github.com/gnolang/gno/gno.land/pkg/sdk/vm"

	"github.com/gnolang/gno/contribs/tx-archive/backup/client"
	"github.com/gnolang/gno/contribs/tx-archive/backup/writer"
	"github.com/gnolang/gno/contribs/tx-archive/log"
	"github.com/gnolang/gno/contribs/tx-archive/log/noop"
)

const DefaultBatchSize = 1000

// Service is the chain backup service
type Service struct {
	client client.Client
	writer writer.Writer
	logger log.Logger

	batchSize     uint
	watchInterval time.Duration // interval for the watch routine
	skipFailedTxs bool
}

// NewService creates a new backup service
func NewService(client client.Client, writer writer.Writer, opts ...Option) *Service {
	s := &Service{
		client:        client,
		writer:        writer,
		logger:        noop.New(),
		watchInterval: 1 * time.Second,
	}

	for _, opt := range opts {
		opt(s)
	}

	// Batch size needs to be at least 1
	if s.batchSize == 0 {
		s.batchSize = DefaultBatchSize
	}

	return s
}

// ExecuteBackup executes the node backup process
func (s *Service) ExecuteBackup(ctx context.Context, cfg Config) error {
	// Verify the config
	if cfgErr := ValidateConfig(cfg); cfgErr != nil {
		return fmt.Errorf("invalid config, %w", cfgErr)
	}

	// Determine the right bound
	toBlock, boundErr := determineRightBound(s.client, cfg.ToBlock)
	if boundErr != nil {
		return fmt.Errorf("unable to determine right bound, %w", boundErr)
	}

	// Log info about what will be backed up
	s.logger.Info(
		"Existing blocks to backup",
		"from block", cfg.FromBlock,
		"to block", toBlock,
		"total", toBlock-cfg.FromBlock+1,
	)

	// Keep track of what has been backed up
	var results struct {
		blocksFetched uint64
		blocksWithTxs uint64
		txsBackedUp   uint64
	}

	// Log results on exit
	defer func() {
		s.logger.Info(
			"Total data backed up",
			"blocks fetched", results.blocksFetched,
			"blocks with transactions", results.blocksWithTxs,
			"transactions written", results.txsBackedUp,
		)
	}()

	// Internal function that fetches and writes a range of blocks
	fetchAndWrite := func(fromBlock, toBlock uint64) error {
		// Fetch by batches
		for batchStart := fromBlock; batchStart <= toBlock; {
			// Determine batch stop block
			batchStop := batchStart + uint64(s.batchSize) - 1
			if batchStop > toBlock {
				batchStop = toBlock
			}

			batchSize := batchStop - batchStart + 1

			// Verbose log for blocks to be fetched
			s.logger.Debug(
				"Fetching batch of blocks",
				"from", batchStart,
				"to", batchStop,
				"size", batchSize,
			)

			// Fetch current batch
			blocks, err := s.client.GetBlocks(ctx, batchStart, batchStop)
			if err != nil {
				return fmt.Errorf("unable to fetch blocks, %w", err)
			}

			// Keep track of the number of fetched blocks & those containing transactions
			results.blocksFetched += batchSize
			results.blocksWithTxs += uint64(len(blocks))

			// Verbose log for blocks containing transactions
			s.logger.Debug(
				"Batch fetched successfully",
				"blocks with transactions", fmt.Sprintf("%d/%d", len(blocks), batchSize),
			)

			// Iterate over the list of blocks containing transactions
			for _, block := range blocks {
				// Fetch current batch tx results, if any
				txResults, err := s.client.GetTxResults(block.Height)
				if err != nil {
					return fmt.Errorf("unable to fetch tx results, %w", err)
				}

				// Sanity check
				if len(txResults) != len(block.Txs) {
					return fmt.Errorf(
						"invalid txs results fetched %d, expected %d",
						len(txResults),
						len(block.Txs),
					)
				}

				for i, tx := range block.Txs {
					txResult := txResults[i]

					if !txResult.IsOK() && s.skipFailedTxs {
						// Skip saving failed transaction
						s.logger.Debug(
							"Skipping failed tx",
							"height", block.Height,
							"index", i,
						)

						continue
					}

					// Write the tx data to the file
					txData := &gnoland.TxWithMetadata{
						Tx: tx,
						Metadata: &gnoland.GnoTxMetadata{
							Timestamp: block.Timestamp,
						},
					}

					if writeErr := s.writer.WriteTxData(txData); writeErr != nil {
						return fmt.Errorf("unable to write tx data, %w", writeErr)
					}

					// Keep track of the number of backed up transactions
					results.txsBackedUp++

					// Verbose log for each transaction written
					s.logger.Debug(
						"Transaction backed up",
						"blockNum", block.Height,
						"tx count (block)", i+1,
						"tx count (total)", results.txsBackedUp,
					)
				}
			}

			batchStart = batchStop + 1
		}

		return nil
	}

	// Backup the existing transactions
	if fetchErr := fetchAndWrite(cfg.FromBlock, toBlock); fetchErr != nil {
		return fetchErr
	}

	// Check if there needs to be a watcher setup
	if cfg.Watch {
		s.logger.Info(
			"Existing blocks backup complete",
			"blocks fetched", results.blocksFetched,
			"blocks with transactions", results.blocksWithTxs,
			"transactions written", results.txsBackedUp,
		)

		s.logger.Info("Watch for new blocks to backup")

		ticker := time.NewTicker(s.watchInterval)
		defer ticker.Stop()

		lastBlock := toBlock

		for {
			select {
			case <-ctx.Done():
				s.logger.Info("Stop watching for new blocks to backup")

				return nil
			case <-ticker.C:
				// Fetch the latest block from the chain
				latest, latestErr := s.client.GetLatestBlockNumber()
				if latestErr != nil {
					return fmt.Errorf("unable to fetch latest block number, %w", latestErr)
				}

				// Check if there have been blocks in the meantime
				if lastBlock == latest {
					continue
				}

				// Catch up to the latest block
				if fetchErr := fetchAndWrite(lastBlock+1, latest); fetchErr != nil {
					return fetchErr
				}

				// Update the last exported block
				lastBlock = latest
			}
		}
	}

	return nil
}

// determineRightBound determines the
// right bound for the chain backup (block height)
func determineRightBound(
	client client.Client,
	potentialTo *uint64,
) (uint64, error) {
	// Get the latest block height from the chain
	latestBlockNumber, err := client.GetLatestBlockNumber()
	if err != nil {
		return 0, fmt.Errorf("unable to fetch latest block number, %w", err)
	}

	// Check if the chain has the block
	if potentialTo != nil && *potentialTo < latestBlockNumber {
		// Requested right bound is valid, use it
		return *potentialTo, nil
	}

	// Requested right bound is not valid, use the latest block number
	return latestBlockNumber, nil
}
