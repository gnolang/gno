package boards

import (
	"errors"
	"strings"
)

// Option configures boards.
type Option func(*Board) error

// WithName assigns a name to the board.
func WithName(name string) Option {
	return func(b *Board) error {
		b.name = strings.TrimSpace(name)
		return nil
	}
}

// WithCreator assigns a creator address to the board.
func WithCreator(addr address) Option {
	return func(b *Board) error {
		if !addr.IsValid() {
			return errors.New("invalid board creator address: " + addr.String())
		}

		b.creator = addr
		return nil
	}
}

// EnableAliases enables support for tracking of previous board names.
// This can be useful in cases where the name of a board must always
// reference the same board over time.
func EnableAliases() Option {
	return func(b *Board) error {
		b.aliasesEnabled = true
		return nil
	}
}
