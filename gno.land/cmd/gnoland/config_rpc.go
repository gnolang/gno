package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type configRPCCfg struct {
	commonEditCfg

	rootDir                  string
	listenAddress            string
	corsAllowedOrigins       commands.StringArr
	corsAllowedMethods       commands.StringArr
	corsAllowedHeaders       commands.StringArr
	grpcListenAddress        string
	grpcMaxOpenConnections   int
	unsafe                   string
	maxOpenConnections       int
	timeoutBroadcastTxCommit time.Duration
	maxBodyBytes             int64
	maxHeaderBytes           int
	tlsCertFile              string
	tlsKeyFile               string
}

// newConfigRPCCmd creates the new config rpc command
func newConfigRPCCmd(io commands.IO) *commands.Command {
	cfg := &configRPCCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "rpc",
			ShortUsage: "config rpc [flags]",
			ShortHelp:  "Edits the Gno node's RPC configuration",
			LongHelp:   "Edits the Gno node's RPC configuration locally",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execConfigRPC(cfg, io)
		},
	)

	return cmd
}

func (c *configRPCCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonEditCfg.RegisterFlags(fs)

	fs.StringVar(
		&c.rootDir,
		"root-dir",
		"",
		"the root directory for all rpc data",
	)

	fs.StringVar(
		&c.listenAddress,
		"listen-address",
		"",
		"the TCP or UNIX socket address for the RPC server to listen on",
	)

	fs.Var(
		&c.corsAllowedOrigins,
		"cors-allowed-origins",
		"a list of origins a cross-domain request can be executed from",
	)

	fs.Var(
		&c.corsAllowedMethods,
		"cors-allowed-methods",
		"a list of methods the client is allowed to use with cross-domain requests",
	)

	fs.Var(
		&c.corsAllowedHeaders,
		"cors-allowed-headers",
		"a list of non simple headers the client is allowed to use with cross-domain requests",
	)

	fs.StringVar(
		&c.grpcListenAddress,
		"grpc-listen-address",
		"",
		"the TCP or UNIX socket address for the gRPC server to listen on",
	)

	fs.IntVar(
		&c.grpcMaxOpenConnections,
		"grpc-max-open-connections",
		-1,
		"the maxiumum number of simultaneous connections",
	)

	fs.StringVar(
		&c.unsafe,
		"unsafe-rpc",
		offValue,
		"activate unsafe RPC commands like /dial_persistent_peers and /unsafe_flush_mempool",
	)

	fs.IntVar(
		&c.maxOpenConnections,
		"rpc-max-open-connections",
		-1,
		"the maximum number of simultaneous RPC connections (including WebSocket)",
	)

	fs.DurationVar(
		&c.timeoutBroadcastTxCommit,
		"timeout-broadcast-commit",
		time.Second*0,
		"how long to wait for a tx to be committed during /broadcast_tx_commit",
	)

	fs.Int64Var(
		&c.maxBodyBytes,
		"max-body-bytes",
		-1,
		"the maximum size of request body, in bytes",
	)

	fs.IntVar(
		&c.maxHeaderBytes,
		"max-header-bytes",
		-1,
		"the maximum size of request header, in bytes",
	)

	fs.StringVar(
		&c.tlsCertFile,
		"tls-cert-file",
		"",
		"the path to a file containing certificate that is used to create the HTTPS server",
	)

	fs.StringVar(
		&c.tlsKeyFile,
		"tls-key-file",
		"",
		"the path to a file containing matching private key that is used to create the HTTPS server",
	)
}

func execConfigRPC(cfg *configRPCCfg, io commands.IO) error {
	// Load the config
	loadedCfg, err := config.LoadConfigFile(cfg.configPath)
	if err != nil {
		return fmt.Errorf("unable to load config, %w", err)
	}

	// Set the root dir, if any
	if cfg.rootDir != "" {
		loadedCfg.RPC.RootDir = cfg.rootDir
	}

	// Set the listen address, if any
	if cfg.listenAddress != "" {
		loadedCfg.RPC.ListenAddress = cfg.listenAddress
	}

	// Set the CORS Allowed Origins, if any
	if len(cfg.corsAllowedOrigins) != 0 {
		loadedCfg.RPC.CORSAllowedOrigins = cfg.corsAllowedOrigins
	}

	// Set the CORS Allowed Methods, if any
	if len(cfg.corsAllowedMethods) != 0 {
		loadedCfg.RPC.CORSAllowedMethods = cfg.corsAllowedMethods
	}

	// Set the CORS Allowed Headers, if any
	if len(cfg.corsAllowedHeaders) != 0 {
		loadedCfg.RPC.CORSAllowedHeaders = cfg.corsAllowedHeaders
	}

	// Set the GRPC listen address, if any
	if cfg.grpcListenAddress != "" {
		loadedCfg.RPC.GRPCListenAddress = cfg.grpcListenAddress
	}

	// Set the max open GRPC connections, if any
	if cfg.grpcMaxOpenConnections >= 0 {
		loadedCfg.RPC.GRPCMaxOpenConnections = cfg.grpcMaxOpenConnections
	}

	// Set the unsafe flag, if any
	unsafeVal, err := parseToggleValue(cfg.unsafe)
	if err != nil {
		return err
	}

	if unsafeVal != loadedCfg.RPC.Unsafe {
		loadedCfg.RPC.Unsafe = unsafeVal
	}

	// Set the max open RPC connections, if any
	if cfg.maxOpenConnections >= 0 {
		loadedCfg.RPC.MaxOpenConnections = cfg.maxOpenConnections
	}

	// Set the broadcast commit timeout, if any
	if cfg.timeoutBroadcastTxCommit > 0 {
		loadedCfg.RPC.TimeoutBroadcastTxCommit = cfg.timeoutBroadcastTxCommit
	}

	// Set the max RPC body bytes, if any
	if cfg.maxBodyBytes >= 0 {
		loadedCfg.RPC.MaxBodyBytes = cfg.maxBodyBytes
	}

	// Set the max RPC header bytes, if any
	if cfg.maxHeaderBytes >= 0 {
		loadedCfg.RPC.MaxHeaderBytes = cfg.maxHeaderBytes
	}

	// Set the TLS certificate file, if any
	if cfg.tlsCertFile != "" {
		loadedCfg.RPC.TLSCertFile = cfg.tlsCertFile
	}

	// Set the TLS key file, if any
	if cfg.tlsKeyFile != "" {
		loadedCfg.RPC.TLSKeyFile = cfg.tlsKeyFile
	}

	// Make sure the config is now valid
	if err := loadedCfg.ValidateBasic(); err != nil {
		return fmt.Errorf("unable to validate config, %w", err)
	}

	// Save the config
	if err := config.WriteConfigFile(cfg.configPath, loadedCfg); err != nil {
		return fmt.Errorf("unable to save updated config, %w", err)
	}

	io.Printfln("Updated rpc configuration saved at %s", cfg.configPath)

	return nil
}
