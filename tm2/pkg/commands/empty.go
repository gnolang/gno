package commands

import "flag"

// EmptyConfig is an empty command configuration
// that should be substituted in commands that require one
type EmptyConfig struct{}

// NewEmptyConfig creates a new instance of the empty configuration
func NewEmptyConfig() *EmptyConfig {
	return &EmptyConfig{}
}

// RegisterFlags ignores flag set registration
func (ec *EmptyConfig) RegisterFlags(_ *flag.FlagSet) {}
