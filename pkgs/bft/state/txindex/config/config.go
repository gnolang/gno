package config

import "github.com/gnolang/gno/pkgs/bft/state/txindex/null"

// IndexerParams defines the arbitrary indexer config params
type IndexerParams map[string]any

// Config defines the specific transaction
// indexer configuration
type Config struct {
	IndexerType string
	Params      IndexerParams
}

// GetParam fetches the specific config param, if any.
// Returns nil if the param is not present
func (c *Config) GetParam(name string) any {
	if c.Params != nil {
		return c.Params[name]
	}

	return nil
}

// DefaultIndexerConfig returns the default indexer config
func DefaultIndexerConfig() *Config {
	return &Config{
		IndexerType: null.IndexerType,
	}
}
