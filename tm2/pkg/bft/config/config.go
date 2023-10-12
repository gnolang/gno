package config

import (
	"fmt"
	"os"
	"path/filepath"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	cns "github.com/gnolang/gno/tm2/pkg/bft/consensus/config"
	mem "github.com/gnolang/gno/tm2/pkg/bft/mempool/config"
	rpc "github.com/gnolang/gno/tm2/pkg/bft/rpc/config"
	"github.com/gnolang/gno/tm2/pkg/errors"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	p2p "github.com/gnolang/gno/tm2/pkg/p2p/config"
)

// Config defines the top level configuration for a Tendermint node
type Config struct {
	// Top level options use an anonymous struct
	BaseConfig `toml:",squash"`

	// Options for services
	RPC       *rpc.RPCConfig       `toml:"rpc"`
	P2P       *p2p.P2PConfig       `toml:"p2p"`
	Mempool   *mem.MempoolConfig   `toml:"mempool"`
	Consensus *cns.ConsensusConfig `toml:"consensus"`
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

type ConfigOptions func(cfg *Config)

// LoadOrMakeConfigWithOptions loads configuration or saves one
// made by modifying the default config with override options
func LoadOrMakeConfigWithOptions(root string, options ConfigOptions) (*Config, error) {
	var cfg *Config

	configPath := join(root, defaultConfigFilePath)
	if osm.FileExists(configPath) {
		var loadErr error

		// Load the configuration
		if cfg, loadErr = LoadConfigFile(configPath); loadErr != nil {
			return nil, loadErr
		}

		cfg.SetRootDir(root)
		cfg.EnsureDirs()
	} else {
		cfg = DefaultConfig()
		options(cfg)
		cfg.SetRootDir(root)
		cfg.EnsureDirs()
		WriteConfigFile(configPath, cfg)

		// Validate the configuration
		if validateErr := cfg.ValidateBasic(); validateErr != nil {
			return nil, fmt.Errorf("unable to validate config, %w", validateErr)
		}
	}

	return cfg, nil
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

// SetRootDir sets the RootDir for all Config structs
func (cfg *Config) SetRootDir(root string) *Config {
	cfg.BaseConfig.RootDir = root
	cfg.RPC.RootDir = root
	cfg.P2P.RootDir = root
	cfg.Mempool.RootDir = root
	cfg.Consensus.RootDir = root
	return cfg
}

// EnsureDirs ensures default directories in root dir (and root dir).
func (cfg *Config) EnsureDirs() {
	rootDir := cfg.BaseConfig.RootDir
	if err := osm.EnsureDir(rootDir, DefaultDirPerm); err != nil {
		panic(err.Error())
	}
	if err := osm.EnsureDir(filepath.Join(rootDir, defaultConfigDir), DefaultDirPerm); err != nil {
		panic(err.Error())
	}
	if err := osm.EnsureDir(filepath.Join(rootDir, defaultDataDir), DefaultDirPerm); err != nil {
		panic(err.Error())
	}
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

// -----------------------------------------------------------------------------
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
	RootDir string `toml:"home"`

	// TCP or UNIX socket address of the ABCI application,
	// or the name of an ABCI application compiled in with the Tendermint binary,
	// or empty if local application instance.
	ProxyApp string `toml:"proxy_app"`

	// Local application instance in lieu of remote app.
	LocalApp abci.Application

	// A custom human readable name for this node
	Moniker string `toml:"moniker"`

	// If this node is many blocks behind the tip of the chain, FastSync
	// allows them to catchup quickly by downloading blocks in parallel
	// and verifying their commits
	FastSyncMode bool `toml:"fast_sync"`

	// Database backend: goleveldb | cleveldb | boltdb
	// * goleveldb (github.com/gnolang/goleveldb - most popular implementation)
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
	DBBackend string `toml:"db_backend"`

	// Database directory
	DBPath string `toml:"db_dir"`

	// Output level for logging
	LogLevel string `toml:"log_level"`

	// Output format: 'plain' (colored text) or 'json'
	LogFormat string `toml:"log_format"`

	// Path to the JSON file containing the initial validator set and other meta data
	Genesis string `toml:"genesis_file"`

	// Path to the JSON file containing the private key to use as a validator in the consensus protocol
	PrivValidatorKey string `toml:"priv_validator_key_file"`

	// Path to the JSON file containing the last sign state of a validator
	PrivValidatorState string `toml:"priv_validator_state_file"`

	// TCP or UNIX socket address for Tendermint to listen on for
	// connections from an external PrivValidator process
	PrivValidatorListenAddr string `toml:"priv_validator_laddr"`

	// A JSON file containing the private key to use for p2p authenticated encryption
	NodeKey string `toml:"node_key_file"`

	// Mechanism to connect to the ABCI application: local | socket
	ABCI string `toml:"abci"`

	// TCP or UNIX socket address for the profiling server to listen on
	ProfListenAddress string `toml:"prof_laddr"`

	// If true, query the ABCI app on connecting to a new peer
	// so the app can decide if we should keep the connection or not
	FilterPeers bool `toml:"filter_peers"` // false
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
	cfg.ProxyApp = "mock://kvstore"
	cfg.FastSyncMode = false
	cfg.DBBackend = "memdb"
	return cfg
}

func (cfg BaseConfig) ChainID() string {
	return cfg.chainID
}

// GenesisFile returns the full path to the genesis.json file
func (cfg BaseConfig) GenesisFile() string {
	return join(cfg.RootDir, cfg.Genesis)
}

// PrivValidatorKeyFile returns the full path to the priv_validator_key.json file
func (cfg BaseConfig) PrivValidatorKeyFile() string {
	return join(cfg.RootDir, cfg.PrivValidatorKey)
}

// PrivValidatorFile returns the full path to the priv_validator_state.json file
func (cfg BaseConfig) PrivValidatorStateFile() string {
	return join(cfg.RootDir, cfg.PrivValidatorState)
}

// NodeKeyFile returns the full path to the node_key.json file
func (cfg BaseConfig) NodeKeyFile() string {
	return join(cfg.RootDir, cfg.NodeKey)
}

// DBDir returns the full path to the database directory
func (cfg BaseConfig) DBDir() string {
	return join(cfg.RootDir, cfg.DBPath)
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
