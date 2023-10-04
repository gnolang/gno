package restore

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/tx-archive/log"
	"github.com/gnolang/tx-archive/restore/client"
	"github.com/gnolang/tx-archive/restore/source"
)

// ExecuteRestore executes the node restore process
func ExecuteRestore(
	ctx context.Context,
	client client.Client,
	source source.Source,
	logger log.Logger,
) error {
	// Set up the teardown
	teardown := func() {
		if closeErr := source.Close(); closeErr != nil {
			logger.Error(
				"unable to gracefully close source",
				"err",
				closeErr.Error(),
			)
		}
	}

	defer teardown()

	var (
		tx      *std.Tx
		nextErr error

		totalTxs uint64
	)

	// Fetch next transactions
	for nextErr == nil {
		tx, nextErr = source.Next(ctx)
		if nextErr != nil {
			break
		}

		// Send the transaction
		if sendErr := client.SendTransaction(tx); sendErr != nil {
			// Invalid transaction sends are only logged,
			// and do not stop the restore process
			logger.Error(
				"unable to send transaction",
				"err",
				sendErr.Error(),
			)

			continue
		}

		totalTxs++
	}

	// Check if this is the end of the road
	if !errors.Is(nextErr, io.EOF) {
		return fmt.Errorf("unable to get next transaction, %w", nextErr)
	}

	// No more transactions to apply
	logger.Info(
		"restore process finished",
		"total",
		totalTxs,
	)

	return nil
}
