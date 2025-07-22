package backup

import (
	"errors"
)

var (
	errInvalidRange     = errors.New("invalid backup block range")
	errInvalidFromBlock = errors.New("from block must be after genesis (0)")
)

// Config is the base chain backup configuration
type Config struct {
	ToBlock   *uint64 // the right bound for the block range; latest if not specified
	FromBlock uint64  // the left bound for the block range

	Watch        bool // flag indicating if incoming tx data should be backed up
	SkipFailedTx bool // flag indicating if failed txs should be ignored
}

// DefaultConfig returns the default backup configuration
func DefaultConfig() Config {
	return Config{
		ToBlock:      nil,   // to latest block by default
		FromBlock:    1,     // from genesis + 1 by default
		Watch:        false, // no tracking by default
		SkipFailedTx: false, // include all txs
	}
}

// ValidateConfig validates the base backup configuration
func ValidateConfig(cfg Config) error {
	// Make sure the backup limits are correct
	if cfg.ToBlock != nil && *cfg.ToBlock < cfg.FromBlock {
		return errInvalidRange
	}

	// Make sure the from-block is after genesis
	if cfg.FromBlock == 0 {
		return errInvalidFromBlock
	}

	return nil
}
