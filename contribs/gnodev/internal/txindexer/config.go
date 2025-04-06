package txindexer

// Config holds the configuration for the tx-indexer service
type Config struct {
	// Enabled indicates whether the tx-indexer is enabled
	Enabled bool
	// DBPath is the path to the database file
	DBPath string
	// HTTPRateLimit is the rate limit for HTTP requests
	HTTPRateLimit *int
	// ListenAddress is the address to listen on for incoming connections
	ListenAddress string
	// LogLevel is the log level for the tx-indexer
	LogLevel *string
	// MaxChunkSize is the maximum chunk size for fetching data
	MaxChunkSize *int
	// MaxSlots is the maximum number of slots (workers) for fetching data
	MaxSlots *int
	// Remote is the remote JSON-RPC URL of the Gno chain
	Remote *string
}

func (c *Config) validate() error {
	if c == nil {
		return nil
	}

	if c.Enabled && c.DBPath == "" {
		return ErrInvalidConfigMissingDBPath
	}

	return nil
}
