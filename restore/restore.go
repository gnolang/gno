package restore

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/tx-archive/log"
	"github.com/gnolang/tx-archive/log/noop"
	"github.com/gnolang/tx-archive/restore/client"
	"github.com/gnolang/tx-archive/restore/source"
)

// Service is the chain restore service
type Service struct {
	client client.Client
	source source.Source
	logger log.Logger
}

// NewService creates a new restore service
func NewService(client client.Client, source source.Source, opts ...Option) *Service {
	s := &Service{
		client: client,
		source: source,
		logger: noop.New(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// ExecuteRestore executes the node restore process
func (s *Service) ExecuteRestore(ctx context.Context) error {
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
		if sendErr := s.client.SendTransaction(tx); sendErr != nil {
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

	// No more transactions to apply
	s.logger.Info(
		"restore process finished",
		"total",
		totalTxs,
	)

	return nil
}
