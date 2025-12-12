package config

// RPCConfig defines the configuration options for the Tendermint RPC server
type RPCConfig struct {
	// TCP or UNIX socket address for the RPC server to listen on
	ListenAddress string `json:"laddr" toml:"laddr" comment:"TCP or UNIX socket address for the RPC server to listen on"`
}

// DefaultRPCConfig returns a default configuration for the RPC server
func DefaultRPCConfig() *RPCConfig {
	return &RPCConfig{
		ListenAddress: "tcp://0.0.0.0:26657",
	}
}
