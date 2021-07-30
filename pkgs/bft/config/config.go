package config

import (
	"fmt"
	"os"
	"path/filepath"

	cns "github.com/gnolang/gno/pkgs/bft/consensus/config"
	mem "github.com/gnolang/gno/pkgs/bft/mempool/config"
	rpc "github.com/gnolang/gno/pkgs/bft/rpc/config"
	"github.com/gnolang/gno/pkgs/errors"
	p2p "github.com/gnolang/gno/pkgs/p2p/config"
)

// Config defines the top level configuration for a Tendermint node
type Config struct {
	// Top level options use an anonymous struct
	BaseConfig `mapstructure:",squash"`

	// Options for services
	RPC       *rpc.RPCConfig       `mapstructure:"rpc"`
	P2P       *p2p.P2PConfig       `mapstructure:"p2p"`
	Mempool   *mem.MempoolConfig   `mapstructure:"mempool"`
	Consensus *cns.ConsensusConfig `mapstructure:"consensus"`
}

// DefaultConfig returns a default configuration for a Tendermint node
func DefaultConfig() *Config {
	return &Config{
		BaseConfig: DefaultBaseConfig(),
		RPC:        rpc.DefaultRPCConfig(),
		P2P:        p2p.DefaultP2PConfig(),
		Mempool:    mem.DefaultMempoolConfig(),
		Consensus:  cns.DefaultConsensusConfig(),
	}
}

// TestConfig returns a configuration that can be used for testing
func TestConfig() *Config {
	return &Config{
		BaseConfig: TestBaseConfig(),
		RPC:        rpc.TestRPCConfig(),
		P2P:        p2p.TestP2PConfig(),
		Mempool:    mem.TestMempoolConfig(),
		Consensus:  cns.TestConsensusConfig(),
	}
}

// SetRoot sets the RootDir for all Config structs
func (cfg *Config) SetRoot(root string) *Config {
	cfg.BaseConfig.RootDir = root
	cfg.RPC.RootDir = root
	cfg.P2P.RootDir = root
	cfg.Mempool.RootDir = root
	cfg.Consensus.RootDir = root
	return cfg
}

// ValidateBasic performs basic validation (checking param bounds, etc.) and
// returns an error if any check fails.
func (cfg *Config) ValidateBasic() error {
	if err := cfg.BaseConfig.ValidateBasic(); err != nil {
		return err
	}
	if err := cfg.RPC.ValidateBasic(); err != nil {
		return errors.Wrap(err, "Error in [rpc] section")
	}
	if err := cfg.P2P.ValidateBasic(); err != nil {
		return errors.Wrap(err, "Error in [p2p] section")
	}
	if err := cfg.Mempool.ValidateBasic(); err != nil {
		return errors.Wrap(err, "Error in [mempool] section")
	}
	if err := cfg.Consensus.ValidateBasic(); err != nil {
		return errors.Wrap(err, "Error in [consensus] section")
	}
	return nil
}

//-----------------------------------------------------------------------------
// BaseConfig

const (
	// LogFormatPlain is a format for colored text
	LogFormatPlain = "plain"
	// LogFormatJSON is a format for json output
	LogFormatJSON = "json"
)

var (
	defaultConfigDir = "config"
	defaultDataDir   = "data"

	defaultConfigFileName   = "config.toml"
	defaultGenesisJSONName  = "genesis.json"
	defaultNodeKeyName      = "node_key.json"
	defaultPrivValKeyName   = "priv_validator_key.json"
	defaultPrivValStateName = "priv_validator_state.json"

	defaultConfigFilePath   = filepath.Join(defaultConfigDir, defaultConfigFileName)
	defaultGenesisJSONPath  = filepath.Join(defaultConfigDir, defaultGenesisJSONName)
	defaultPrivValKeyPath   = filepath.Join(defaultConfigDir, defaultPrivValKeyName)
	defaultPrivValStatePath = filepath.Join(defaultDataDir, defaultPrivValStateName)
	defaultNodeKeyPath      = filepath.Join(defaultConfigDir, defaultNodeKeyName)
)

// BaseConfig defines the base configuration for a Tendermint node
type BaseConfig struct {
	// chainID is unexposed and immutable but here for convenience
	chainID string

	// The root directory for all data.
	// This should be set in viper so it can unmarshal into this struct
	RootDir string `mapstructure:"home"`

	// TCP or UNIX socket address of the ABCI application,
	// or the name of an ABCI application compiled in with the Tendermint binary
	ProxyApp string `mapstructure:"proxy_app"`

	// A custom human readable name for this node
	Moniker string `mapstructure:"moniker"`

	// If this node is many blocks behind the tip of the chain, FastSync
	// allows them to catchup quickly by downloading blocks in parallel
	// and verifying their commits
	FastSyncMode bool `mapstructure:"fast_sync"`

	// Database backend: goleveldb | cleveldb | boltdb
	// * goleveldb (github.com/syndtr/goleveldb - most popular implementation)
	//   - pure go
	//   - stable
	// * cleveldb (uses levigo wrapper)
	//   - fast
	//   - requires gcc
	//   - use cleveldb build tag (go build -tags cleveldb)
	// * boltdb (uses etcd's fork of bolt - go.etcd.io/bbolt)
	//   - EXPERIMENTAL
	//   - may be faster is some use-cases (random reads - indexer)
	//   - use boltdb build tag (go build -tags boltdb)
	DBBackend string `mapstructure:"db_backend"`

	// Database directory
	DBPath string `mapstructure:"db_dir"`

	// Output level for logging
	LogLevel string `mapstructure:"log_level"`

	// Output format: 'plain' (colored text) or 'json'
	LogFormat string `mapstructure:"log_format"`

	// Path to the JSON file containing the initial validator set and other meta data
	Genesis string `mapstructure:"genesis_file"`

	// Path to the JSON file containing the private key to use as a validator in the consensus protocol
	PrivValidatorKey string `mapstructure:"priv_validator_key_file"`

	// Path to the JSON file containing the last sign state of a validator
	PrivValidatorState string `mapstructure:"priv_validator_state_file"`

	// TCP or UNIX socket address for Tendermint to listen on for
	// connections from an external PrivValidator process
	PrivValidatorListenAddr string `mapstructure:"priv_validator_laddr"`

	// A JSON file containing the private key to use for p2p authenticated encryption
	NodeKey string `mapstructure:"node_key_file"`

	// Mechanism to connect to the ABCI application: socket | grpc
	ABCI string `mapstructure:"abci"`

	// TCP or UNIX socket address for the profiling server to listen on
	ProfListenAddress string `mapstructure:"prof_laddr"`

	// If true, query the ABCI app on connecting to a new peer
	// so the app can decide if we should keep the connection or not
	FilterPeers bool `mapstructure:"filter_peers"` // false
}

// DefaultBaseConfig returns a default base configuration for a Tendermint node
func DefaultBaseConfig() BaseConfig {
	return BaseConfig{
		Genesis:            defaultGenesisJSONPath,
		PrivValidatorKey:   defaultPrivValKeyPath,
		PrivValidatorState: defaultPrivValStatePath,
		NodeKey:            defaultNodeKeyPath,
		Moniker:            defaultMoniker,
		ProxyApp:           "tcp://127.0.0.1:26658",
		ABCI:               "socket",
		LogLevel:           DefaultPackageLogLevels(),
		LogFormat:          LogFormatPlain,
		ProfListenAddress:  "",
		FastSyncMode:       true,
		FilterPeers:        false,
		DBBackend:          "goleveldb",
		DBPath:             "data",
	}
}

// TestBaseConfig returns a base configuration for testing a Tendermint node
func TestBaseConfig() BaseConfig {
	cfg := DefaultBaseConfig()
	cfg.chainID = "tendermint_test"
	cfg.ProxyApp = "kvstore"
	cfg.FastSyncMode = false
	cfg.DBBackend = "memdb"
	return cfg
}

func (cfg BaseConfig) ChainID() string {
	return cfg.chainID
}

// GenesisFile returns the full path to the genesis.json file
func (cfg BaseConfig) GenesisFile() string {
	return rootify(cfg.Genesis, cfg.RootDir)
}

// PrivValidatorKeyFile returns the full path to the priv_validator_key.json file
func (cfg BaseConfig) PrivValidatorKeyFile() string {
	return rootify(cfg.PrivValidatorKey, cfg.RootDir)
}

// PrivValidatorFile returns the full path to the priv_validator_state.json file
func (cfg BaseConfig) PrivValidatorStateFile() string {
	return rootify(cfg.PrivValidatorState, cfg.RootDir)
}

// NodeKeyFile returns the full path to the node_key.json file
func (cfg BaseConfig) NodeKeyFile() string {
	return rootify(cfg.NodeKey, cfg.RootDir)
}

// DBDir returns the full path to the database directory
func (cfg BaseConfig) DBDir() string {
	return rootify(cfg.DBPath, cfg.RootDir)
}

// ValidateBasic performs basic validation (checking param bounds, etc.) and
// returns an error if any check fails.
func (cfg BaseConfig) ValidateBasic() error {
	switch cfg.LogFormat {
	case LogFormatPlain, LogFormatJSON:
	default:
		return errors.New("unknown log_format (must be 'plain' or 'json')")
	}
	return nil
}

// DefaultLogLevel returns a default log level of "error"
func DefaultLogLevel() string {
	return "error"
}

// DefaultPackageLogLevels returns a default log level setting so all packages
// log at "error", while the `state` and `main` packages log at "info"
func DefaultPackageLogLevels() string {
	return fmt.Sprintf("main:info,state:info,*:%s", DefaultLogLevel())
}

var defaultMoniker = getDefaultMoniker()

// getDefaultMoniker returns a default moniker, which is the host name. If runtime
// fails to get the host name, "anonymous" will be returned.
func getDefaultMoniker() string {
	moniker, err := os.Hostname()
	if err != nil {
		moniker = "anonymous"
	}
	return moniker
}
