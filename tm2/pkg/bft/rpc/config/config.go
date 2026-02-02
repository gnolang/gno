package config

import (
	"errors"
	"net/http"
	"path/filepath"
	"time"
)

// -----------------------------------------------------------------------------
// RPCConfig

const (
	defaultConfigDir = "config"
)

// RPCConfig defines the configuration options for the Tendermint RPC server
type RPCConfig struct {
	RootDir string `json:"home" toml:"home"`

	// TCP or UNIX socket address for the RPC server to listen on
	ListenAddress string `json:"laddr" toml:"laddr" comment:"TCP or UNIX socket address for the RPC server to listen on"`

	// A list of origins a cross-domain request can be executed from.
	// If the special '*' value is present in the list, all origins will be allowed.
	// An origin may contain a wildcard (*) to replace 0 or more characters (i.e.: http://*.domain.com).
	// Only one wildcard can be used per origin.
	CORSAllowedOrigins []string `json:"cors_allowed_origins" toml:"cors_allowed_origins" comment:"A list of origins a cross-domain request can be executed from\n Default value '[]' disables cors support\n Use '[\"*\"]' to allow any origin"`

	// A list of methods the client is allowed to use with cross-domain requests.
	CORSAllowedMethods []string `json:"cors_allowed_methods" toml:"cors_allowed_methods" comment:"A list of methods the client is allowed to use with cross-domain requests"`

	// A list of non simple headers the client is allowed to use with cross-domain requests.
	CORSAllowedHeaders []string `json:"cors_allowed_headers" toml:"cors_allowed_headers" comment:"A list of non simple headers the client is allowed to use with cross-domain requests"`

	// TCP or UNIX socket address for the gRPC server to listen on
	// NOTE: This server only supports /broadcast_tx_commit
	GRPCListenAddress string `json:"grpc_laddr" toml:"grpc_laddr" comment:"TCP or UNIX socket address for the gRPC server to listen on\n NOTE: This server only supports /broadcast_tx_commit"`

	// Maximum number of simultaneous connections.
	// Does not include RPC (HTTP&WebSocket) connections. See max_open_connections
	// If you want to accept a larger number than the default, make sure
	// you increase your OS limits.
	// 0 - unlimited.
	GRPCMaxOpenConnections int `json:"grpc_max_open_connections" toml:"grpc_max_open_connections" comment:"Maximum number of simultaneous connections.\n Does not include RPC (HTTP&WebSocket) connections. See max_open_connections\n If you want to accept a larger number than the default, make sure\n you increase your OS limits.\n 0 - unlimited.\n Should be < {ulimit -Sn} - {MaxNumInboundPeers} - {MaxNumOutboundPeers} - {N of wal, db and other open files}\n 1024 - 40 - 10 - 50 = 924 = ~900"`

	// Activate unsafe RPC commands like /dial_persistent_peers and /unsafe_flush_mempool
	Unsafe bool `json:"unsafe" toml:"unsafe" comment:"Activate unsafe RPC commands like /dial_seeds and /unsafe_flush_mempool"`

	// Maximum number of simultaneous connections (including WebSocket).
	// Does not include gRPC connections. See grpc_max_open_connections
	// If you want to accept a larger number than the default, make sure
	// you increase your OS limits.
	// 0 - unlimited.
	// Should be < {ulimit -Sn} - {MaxNumInboundPeers} - {MaxNumOutboundPeers} - {N of wal, db and other open files}
	// 1024 - 40 - 10 - 50 = 924 = ~900
	MaxOpenConnections int `json:"max_open_connections" toml:"max_open_connections" comment:"Maximum number of simultaneous connections (including WebSocket).\n Does not include gRPC connections. See grpc_max_open_connections\n If you want to accept a larger number than the default, make sure\n you increase your OS limits.\n 0 - unlimited.\n Should be < {ulimit -Sn} - {MaxNumInboundPeers} - {MaxNumOutboundPeers} - {N of wal, db and other open files}\n 1024 - 40 - 10 - 50 = 924 = ~900"`

	// How long to wait for a tx to be committed during /broadcast_tx_commit
	// WARNING: Using a value larger than 10s will result in increasing the
	// global HTTP write timeout, which applies to all connections and endpoints.
	// See https://github.com/tendermint/tendermint/issues/3435
	TimeoutBroadcastTxCommit time.Duration `json:"timeout_broadcast_tx_commit" toml:"timeout_broadcast_tx_commit" comment:"How long to wait for a tx to be committed during /broadcast_tx_commit.\n WARNING: Using a value larger than 10s will result in increasing the\n global HTTP write timeout, which applies to all connections and endpoints.\n See https://github.com/tendermint/tendermint/issues/3435"`

	// Maximum size of request body, in bytes
	MaxBodyBytes int64 `json:"max_body_bytes" toml:"max_body_bytes" comment:"Maximum size of request body, in bytes"`

	// Maximum size of request header, in bytes
	MaxHeaderBytes int `json:"max_header_bytes" toml:"max_header_bytes" comment:"Maximum size of request header, in bytes"`

	// The path to a file containing certificate that is used to create the HTTPS server.
	// Might be either absolute path or path related to tendermint's config directory.
	//
	// If the certificate is signed by a certificate authority,
	// the certFile should be the concatenation of the server's certificate, any intermediates,
	// and the CA's certificate.
	//
	// NOTE: both tls_cert_file and tls_key_file must be present for Tendermint to create HTTPS server. Otherwise, HTTP server is run.
	TLSCertFile string `json:"tls_cert_file" toml:"tls_cert_file" comment:"The path to a file containing certificate that is used to create the HTTPS server.\n Might be either absolute path or path related to tendermint's config directory.\n If the certificate is signed by a certificate authority,\n the certFile should be the concatenation of the server's certificate, any intermediates,\n and the CA's certificate.\n NOTE: both tls_cert_file and tls_key_file must be present for Tendermint to create HTTPS server. Otherwise, HTTP server is run."`

	// The path to a file containing matching private key that is used to create the HTTPS server.
	// Might be either absolute path or path related to tendermint's config directory.
	//
	// NOTE: both tls_cert_file and tls_key_file must be present for Tendermint to create HTTPS server. Otherwise, HTTP server is run.
	TLSKeyFile string `json:"tls_key_file" toml:"tls_key_file" comment:"The path to a file containing matching private key that is used to create the HTTPS server.\n Might be either absolute path or path related to tendermint's config directory.\n NOTE: both tls_cert_file and tls_key_file must be present for Tendermint to create HTTPS server. Otherwise, HTTP server is run."`
}

// DefaultRPCConfig returns a default configuration for the RPC server
func DefaultRPCConfig() *RPCConfig {
	return &RPCConfig{
		ListenAddress:          "tcp://127.0.0.1:26657",
		CORSAllowedOrigins:     []string{"*"},
		CORSAllowedMethods:     []string{http.MethodHead, http.MethodGet, http.MethodPost, http.MethodOptions},
		CORSAllowedHeaders:     []string{"Origin", "Accept", "Content-Type", "X-Requested-With", "X-Server-Time"},
		GRPCListenAddress:      "",
		GRPCMaxOpenConnections: 900,

		Unsafe:             false,
		MaxOpenConnections: 900,

		TimeoutBroadcastTxCommit: 10 * time.Second,

		MaxBodyBytes:   int64(1000000), // 1MB
		MaxHeaderBytes: 1 << 20,        // same as the net/http default

		TLSCertFile: "",
		TLSKeyFile:  "",
	}
}

// TestRPCConfig returns a configuration for testing the RPC server
func TestRPCConfig() *RPCConfig {
	cfg := DefaultRPCConfig()
	cfg.ListenAddress = "tcp://0.0.0.0:26657"
	cfg.GRPCListenAddress = "tcp://0.0.0.0:26658"
	cfg.Unsafe = true
	return cfg
}

// ValidateBasic performs basic validation (checking param bounds, etc.) and
// returns an error if any check fails.
func (cfg *RPCConfig) ValidateBasic() error {
	if cfg.GRPCMaxOpenConnections < 0 {
		return errors.New("grpc_max_open_connections can't be negative")
	}
	if cfg.MaxOpenConnections < 0 {
		return errors.New("max_open_connections can't be negative")
	}
	if cfg.TimeoutBroadcastTxCommit < 0 {
		return errors.New("timeout_broadcast_tx_commit can't be negative")
	}
	if cfg.MaxBodyBytes < 0 {
		return errors.New("max_body_bytes can't be negative")
	}
	if cfg.MaxHeaderBytes < 0 {
		return errors.New("max_header_bytes can't be negative")
	}
	return nil
}

// IsCorsEnabled returns true if cross-origin resource sharing is enabled.
// XXX review.
func (cfg *RPCConfig) IsCorsEnabled() bool {
	return len(cfg.CORSAllowedOrigins) != 0
}

func (cfg RPCConfig) KeyFile() string {
	path := cfg.TLSKeyFile
	if filepath.IsAbs(path) {
		return path
	}
	return join(cfg.RootDir, filepath.Join(defaultConfigDir, path))
}

func (cfg RPCConfig) CertFile() string {
	path := cfg.TLSCertFile
	if filepath.IsAbs(path) {
		return path
	}
	return join(cfg.RootDir, filepath.Join(defaultConfigDir, path))
}

func (cfg RPCConfig) IsTLSEnabled() bool {
	return cfg.TLSCertFile != "" && cfg.TLSKeyFile != ""
}

// helper function to make config creation independent of root dir
func join(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(root, path)
}
