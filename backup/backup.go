package backup

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/gnolang/tx-archive/backup/client"
	"github.com/gnolang/tx-archive/log"
	"github.com/gnolang/tx-archive/types"
)

// ExecuteBackup executes the node backup process
func ExecuteBackup(
	client client.Client,
	logger log.Logger,
	cfg Config,
) error {
	// Verify the config
	if cfgErr := ValidateConfig(cfg); cfgErr != nil {
		return fmt.Errorf("invalid config, %w", cfgErr)
	}

	// Open the file for writing
	outputFile, openErr := os.OpenFile(
		cfg.OutputFile,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0o755,
	)
	if openErr != nil {
		return fmt.Errorf("unable to open file %s, %w", cfg.OutputFile, openErr)
	}

	closeFile := func() error {
		if err := outputFile.Close(); err != nil {
			logger.Error("unable to close output file", "err", err.Error())

			return err
		}

		return nil
	}

	teardown := func() {
		if err := closeFile(); err != nil {
			if removeErr := os.Remove(outputFile.Name()); removeErr != nil {
				logger.Error("unable to remove file", "err", err.Error())
			}
		}
	}

	// Set up the teardown
	defer teardown()

	// Determine the right bound
	toBlock, boundErr := determineRightBound(client, cfg.ToBlock)
	if boundErr != nil {
		return fmt.Errorf("unable to determine right bound, %w", boundErr)
	}

	// Gather the chain data from the node
	blockData, blockDataErr := getBlockData(client, logger, cfg.FromBlock, toBlock)
	if blockDataErr != nil {
		return fmt.Errorf("unable to fetch block data, %w", blockDataErr)
	}

	// Prepare the archive
	metadata, metadataErr := generateMetadata(blockData)
	if metadataErr != nil {
		return fmt.Errorf("unable to generate metadata, %w", metadataErr)
	}

	archive := &types.Archive{
		BlockData: blockData,
		Metadata:  metadata,
	}

	// Marshal the archive data
	archiveRaw, marshalErr := json.Marshal(archive)
	if marshalErr != nil {
		return fmt.Errorf("unable to marshal archive JSON, %w", marshalErr)
	}

	// Write the archive data to a file
	_, writeErr := outputFile.Write(archiveRaw)
	if writeErr != nil {
		return fmt.Errorf("unable to write archive JSON, %w", writeErr)
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

// getBlockData fetches the block data from the chain
func getBlockData(
	client client.Client,
	logger log.Logger,
	from,
	to uint64,
) ([]*types.BlockData, error) {
	blockData := make([]*types.BlockData, 0, to-from+1)

	for block := from; block <= to; block++ {
		txs, txErr := client.GetBlockTransactions(block)
		if txErr != nil {
			return nil, fmt.Errorf("unable to fetch block transactions, %w", txErr)
		}

		// Save the block transaction data
		data := &types.BlockData{
			Txs:      txs,
			BlockNum: block,
		}
		blockData = append(blockData, data)

		// Log the progress
		logProgress(logger, from, to, block)
	}

	return blockData, nil
}

// logProgress logs the backup progress
func logProgress(logger log.Logger, from, to, current uint64) {
	total := to - from
	status := (float64(current) - float64(from)) / float64(total) * 100

	logger.Info(
		fmt.Sprintf("Total of %d blocks backed up", current-from+1),
		"total", total,
		"from", from,
		"to", true,
		"status", fmt.Sprintf("%.2f%%", status),
	)
}

func generateMetadata(blockData []*types.BlockData) (*types.Metadata, error) {
	if len(blockData) == 0 {
		return nil, errors.New("unable to generate metadata, no block data")
	}

	return &types.Metadata{
		EarliestBlockHeight: blockData[0].BlockNum,
		LatestBlockHeight:   blockData[len(blockData)-1].BlockNum,
	}, nil
}
