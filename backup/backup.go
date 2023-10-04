package backup

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/gnolang/tx-archive/backup/client"
	"github.com/gnolang/tx-archive/log"
	"github.com/gnolang/tx-archive/types"
)

// ExecuteBackup executes the node backup process
func ExecuteBackup(
	client client.Client,
	writer io.Writer,
	logger log.Logger,
	cfg Config,
) error {
	// Verify the config
	if cfgErr := ValidateConfig(cfg); cfgErr != nil {
		return fmt.Errorf("invalid config, %w", cfgErr)
	}

	// Determine the right bound
	toBlock, boundErr := determineRightBound(client, cfg.ToBlock)
	if boundErr != nil {
		return fmt.Errorf("unable to determine right bound, %w", boundErr)
	}

	// Gather the chain data from the node
	for block := cfg.FromBlock; block <= toBlock; block++ {
		txs, txErr := client.GetBlockTransactions(block)
		if txErr != nil {
			return fmt.Errorf("unable to fetch block transactions, %w", txErr)
		}

		// Save the block transaction data, if any
		for _, tx := range txs {
			data := &types.TxData{
				Tx:       tx,
				BlockNum: block,
			}

			// Write the tx data to the file
			if writeErr := writeTxData(writer, data); writeErr != nil {
				return fmt.Errorf("unable to write tx data, %w", writeErr)
			}
		}

		// Log the progress
		logProgress(logger, cfg.FromBlock, toBlock, block)
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

// writeTxData outputs the tx data to the writer
func writeTxData(writer io.Writer, txData *types.TxData) error {
	// Marshal tx data into JSON
	jsonData, err := json.Marshal(txData)
	if err != nil {
		return fmt.Errorf("unable to marshal JSON data, %w", err)
	}

	// Write the JSON data as a line to the file
	_, err = writer.Write(jsonData)
	if err != nil {
		return fmt.Errorf("unable to write to output, %w", err)
	}

	// Write a newline character to separate JSON objects
	_, err = writer.Write([]byte("\n"))
	if err != nil {
		return fmt.Errorf("unable to write newline output, %w", err)
	}

	return nil
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
