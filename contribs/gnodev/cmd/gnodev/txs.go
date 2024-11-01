package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
)

func parseTxs(ctx context.Context, txFile string) ([]gnoland.TxWithMetadata, error) {
	if txFile == "" {
		return nil, nil
	}

	file, loadErr := os.Open(txFile)
	if loadErr != nil {
		return nil, fmt.Errorf("unable to open tx file %s: %w", txFile, loadErr)
	}
	defer file.Close()

	var (
		txs []gnoland.TxWithMetadata

		scanner = bufio.NewScanner(file)
	)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Parse the amino JSON
			var tx gnoland.TxWithMetadata
			if err := amino.UnmarshalJSON(scanner.Bytes(), &tx); err != nil {
				return nil, fmt.Errorf(
					"unable to unmarshal amino JSON, %w",
					err,
				)
			}

			txs = append(txs, tx)
		}
	}

	// Check for scanning errors
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf(
			"error encountered while reading file, %w",
			err,
		)
	}

	return txs, nil
}
