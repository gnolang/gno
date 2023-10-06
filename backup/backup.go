package backup

//nolint:revive // See https://github.com/gnolang/gno/issues/1197
import (
	"context"
	"fmt"

	"github.com/gnolang/tx-archive/backup/client"
	"github.com/gnolang/tx-archive/backup/writer"
	"github.com/gnolang/tx-archive/log"
	"github.com/gnolang/tx-archive/log/noop"
	"github.com/gnolang/tx-archive/types"

	_ "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
)

// Service is the chain backup service
type Service struct {
	client client.Client
	writer writer.Writer
	logger log.Logger
}

// NewService creates a new backup service
func NewService(client client.Client, writer writer.Writer, opts ...Option) *Service {
	s := &Service{
		client: client,
		writer: writer,
		logger: noop.New(),
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

	// Gather the chain data from the node
	for block := cfg.FromBlock; block <= toBlock; block++ {
		select {
		case <-ctx.Done():
			s.logger.Info("export procedure stopped")

			return nil
		default:
			txs, txErr := s.client.GetBlockTransactions(block)
			if txErr != nil {
				return fmt.Errorf("unable to fetch block transactions, %w", txErr)
			}

			// Skip empty blocks
			if len(txs) == 0 {
				continue
			}

			// Save the block transaction data, if any
			for _, tx := range txs {
				data := &types.TxData{
					Tx:       tx,
					BlockNum: block,
				}

				// Write the tx data to the file
				if writeErr := s.writer.WriteTxData(data); writeErr != nil {
					return fmt.Errorf("unable to write tx data, %w", writeErr)
				}
			}

			// Log the progress
			logProgress(s.logger, cfg.FromBlock, toBlock, block)
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

// logProgress logs the backup progress
func logProgress(logger log.Logger, from, to, current uint64) {
	total := to - from
	status := (float64(current) - float64(from)) / float64(total) * 100

	logger.Info(
		fmt.Sprintf("Total of %d blocks backed up", current-from+1),
		"total", total+1,
		"from", from,
		"to", to,
		"status", fmt.Sprintf("%.2f%%", status),
	)
}
