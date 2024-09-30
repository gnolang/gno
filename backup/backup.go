package backup

//nolint:revive // See https://github.com/gnolang/gno/issues/1197
import (
	"context"
	"fmt"
	"time"

	_ "github.com/gnolang/gno/gno.land/pkg/sdk/vm"

	"github.com/gnolang/tx-archive/backup/client"
	"github.com/gnolang/tx-archive/backup/writer"
	"github.com/gnolang/tx-archive/log"
	"github.com/gnolang/tx-archive/log/noop"
	"github.com/gnolang/tx-archive/types"
)

// Service is the chain backup service
type Service struct {
	client client.Client
	writer writer.Writer
	logger log.Logger

	watchInterval time.Duration // interval for the watch routine
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

	// Keep track of total txs backed up
	totalTxs := uint64(0)

	fetchAndWrite := func(height uint64) error {
		block, txErr := s.client.GetBlock(height)
		if txErr != nil {
			return fmt.Errorf("unable to fetch block transactions, %w", txErr)
		}

		// Skip empty blocks
		if len(block.Txs) == 0 {
			return nil
		}

		// Save the block transaction data, if any
		for _, tx := range block.Txs {
			data := &types.TxData{
				Tx:        tx,
				BlockNum:  block.Height,
				Timestamp: block.Timestamp,
			}

			// Write the tx data to the file
			if writeErr := s.writer.WriteTxData(data); writeErr != nil {
				return fmt.Errorf("unable to write tx data, %w", writeErr)
			}

			totalTxs++

			// Log the progress
			s.logger.Info(
				"Transaction backed up",
				"total", totalTxs,
			)
		}

		return nil
	}

	// Gather the chain data from the node
	for block := cfg.FromBlock; block <= toBlock; block++ {
		select {
		case <-ctx.Done():
			s.logger.Info("backup procedure stopped")

			return nil
		default:
			if fetchErr := fetchAndWrite(block); fetchErr != nil {
				return fetchErr
			}
		}
	}

	// Check if there needs to be a watcher setup
	if cfg.Watch {
		ticker := time.NewTicker(s.watchInterval)
		defer ticker.Stop()

		lastBlock := toBlock

		for {
			select {
			case <-ctx.Done():
				s.logger.Info("export procedure stopped")

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
				for block := lastBlock + 1; block <= latest; block++ {
					if fetchErr := fetchAndWrite(block); fetchErr != nil {
						return fetchErr
					}
				}

				// Update the last exported block
				lastBlock = latest
			}
		}
	}

	s.logger.Info("Backup complete")

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
