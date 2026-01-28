package server

import "log/slog"

type Option func(s *JSONRPC)

// WithLogger sets the logger to be used
// with the JSON-RPC server
func WithLogger(logger *slog.Logger) Option {
	return func(s *JSONRPC) {
		s.logger = logger
	}
}
