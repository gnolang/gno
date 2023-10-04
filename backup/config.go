package backup

import (
	"errors"
)

var errInvalidRange = errors.New("invalid backup block range")

// Config is the base chain backup configuration
type Config struct {
	ToBlock   *uint64 // the right bound for the block range; latest if not specified
	FromBlock uint64  // the left bound for the block range
}

// DefaultConfig returns the default backup configuration
func DefaultConfig() Config {
	return Config{
		ToBlock:   nil, // to latest block by default
		FromBlock: 0,   // from genesis by default
	}
}

// ValidateConfig validates the base backup configuration
func ValidateConfig(cfg Config) error {
	// Make sure the backup limits are correct
	if cfg.ToBlock != nil && *cfg.ToBlock < cfg.FromBlock {
		return errInvalidRange
	}

	return nil
}
