package config

// Config defines the configuration options for the Tendermint RPC server
type Config struct {
	// TCP or UNIX socket address for the RPC server to listen on
	ListenAddress string `json:"laddr" toml:"laddr" comment:"TCP or UNIX socket address for the RPC server to listen on"`
}

// DefaultConfig returns a default configuration for the RPC server
func DefaultConfig() *Config {
	return &Config{
		ListenAddress: "tcp://0.0.0.0:26657",
	}
}
