package restore

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gnolang/gno/contribs/tx-archive/log"
	"github.com/gnolang/gno/contribs/tx-archive/log/noop"
	"github.com/gnolang/gno/contribs/tx-archive/restore/client"
	"github.com/gnolang/gno/contribs/tx-archive/restore/source"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Service is the chain restore service
type Service struct {
	client client.Client
	source source.Source
	logger log.Logger

	watchInterval time.Duration // interval for the watch routine
}

// NewService creates a new restore service
func NewService(client client.Client, source source.Source, opts ...Option) *Service {
	s := &Service{
		client:        client,
		source:        source,
		logger:        noop.New(),
		watchInterval: 1 * time.Second,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// ExecuteRestore executes the node restore process
func (s *Service) ExecuteRestore(ctx context.Context, watch bool) error {
	fetchTxAndSend := func() error {
		var (
			tx      *std.Tx
			nextErr error

			totalTxs uint64
		)

		// Fetch next transactions
		for nextErr == nil {
			tx, nextErr = s.source.Next(ctx)
			if nextErr != nil {
				break
			}

			// Send the transaction
			if sendErr := s.client.SendTransaction(ctx, tx); sendErr != nil {
				// Invalid transaction sends are only logged,
				// and do not stop the restore process
				s.logger.Error(
					"unable to send transaction",
					"err",
					sendErr.Error(),
				)

				continue
			}

			totalTxs++

			s.logger.Info(
				"sent transaction",
				"total",
				totalTxs,
			)
		}

		// Check if this is the end of the road
		if !errors.Is(nextErr, io.EOF) {
			return fmt.Errorf("unable to get next transaction, %w", nextErr)
		}

		return nil
	}

	// Execute the initial restore
	if fetchErr := fetchTxAndSend(); fetchErr != nil {
		return fetchErr
	}

	// Check if there needs to be a watcher setup
	if watch {
		ticker := time.NewTicker(s.watchInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				s.logger.Info("restore procedure stopped")

				return nil
			case <-ticker.C:
				if fetchErr := fetchTxAndSend(); fetchErr != nil {
					return fetchErr
				}
			}
		}
	}

	// No more transactions to apply
	s.logger.Info("restore process finished")

	return nil
}
