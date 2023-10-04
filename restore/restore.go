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
	client client.Client,
	source source.Source,
	logger log.Logger,
	cfg Config,
) error {
	// Verify the config
	if cfgErr := ValidateConfig(cfg); cfgErr != nil {
		return fmt.Errorf("invalid config, %w", cfgErr)
	}

	defer func() {
		if closeErr := source.Close(); closeErr != nil {
			logger.Error(
				"unable to gracefully close source",
				"err",
				closeErr.Error(),
			)
		}
	}()

	var (
		tx      *std.Tx
		nextErr error

		totalTxs uint64
	)

	// Fetch next transactions
	// TODO add ctx
	for tx, nextErr = source.Next(context.Background()); nextErr == nil; {
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
	if errors.Is(nextErr, io.EOF) {
		// No more transactions to apply
		logger.Info(
			"restore process finished",
			"total",
			totalTxs,
		)

		return nil
	}

	return fmt.Errorf("unable to get next transaction, %w", nextErr)
}
