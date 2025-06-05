package restore

import "github.com/gnolang/gno/contribs/tx-archive/log"

type Option func(s *Service)

// WithLogger specifies the logger for the restore service
func WithLogger(l log.Logger) Option {
	return func(s *Service) {
		s.logger = l
	}
}
