package backup

import "github.com/gnolang/tx-archive/log"

type Option func(s *Service)

// WithLogger specifies the logger for the backup service
func WithLogger(l log.Logger) Option {
	return func(s *Service) {
		s.logger = l
	}
}
