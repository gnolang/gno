package restore

import "errors"

const (
	DefaultRemote = "http://127.0.0.1:26657"
)

var errInvalidRemote = errors.New("invalid remote address")

// Config is the base chain restore config
type Config struct {
	Remote string // the remote JSON-RPC URL of the chain
}

// DefaultConfig returns the default restore configuration
func DefaultConfig() Config {
	return Config{
		Remote: DefaultRemote,
	}
}

// ValidateConfig validates the base restore configuration
func ValidateConfig(cfg Config) error {
	if cfg.Remote == "" {
		return errInvalidRemote
	}

	return nil
}
