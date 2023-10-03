package backup

import (
	"errors"
	"os"
)

const (
	DefaultOutputFileLocation = "./backup.json"
)

var (
	errInvalidOutputLocation = errors.New("invalid output file location")
	errOutputFileExists      = errors.New("output file already exists")
	errInvalidRange          = errors.New("invalid backup block range")
)

// Config is the base chain backup configuration
type Config struct {
	ToBlock    *uint64 // the right bound for the block range; latest if not specified
	OutputFile string  // the output file path
	FromBlock  uint64  // the left bound for the block range
	Overwrite  bool    // flag indicating if the output file should be overwritten
}

// DefaultConfig returns the default backup configuration
func DefaultConfig() Config {
	return Config{
		OutputFile: DefaultOutputFileLocation,
		Overwrite:  false, // no overwrites by default
		ToBlock:    nil,   // to latest block by default
		FromBlock:  0,     // from genesis by default
	}
}

// ValidateConfig validates the base backup configuration
func ValidateConfig(cfg Config) error {
	// Make sure the output file path is valid
	if cfg.OutputFile == "" {
		return errInvalidOutputLocation
	}

	// Make sure the output file can be overwritten, if it exists
	if _, err := os.Stat(cfg.OutputFile); err == nil && !cfg.Overwrite {
		// File already exists, and the overwrite flag is not set
		return errOutputFileExists
	}

	// Make sure the backup limits are correct
	if cfg.ToBlock != nil && *cfg.ToBlock < cfg.FromBlock {
		return errInvalidRange
	}

	return nil
}
