package ws

import (
	"log/slog"
)

type Option func(*Client)

// WithLogger sets the WS client logger
func WithLogger(logger *slog.Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}
