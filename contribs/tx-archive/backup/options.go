package backup

import "github.com/gnolang/gno/contribs/tx-archive/log"

type Option func(s *Service)

// WithLogger specifies the logger for the backup service
func WithLogger(l log.Logger) Option {
	return func(s *Service) {
		s.logger = l
	}
}

// WithBatchSize specifies the number of requests to batch
func WithBatchSize(size uint) Option {
	return func(s *Service) {
		s.batchSize = size
	}
}

// WithSkipFailedTxs specifies if failed txs should be backed up
func WithSkipFailedTxs(skip bool) Option {
	return func(s *Service) {
		s.skipFailedTxs = skip
	}
}
