package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"time"

	"dario.cat/mergo"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	cns "github.com/gnolang/gno/tm2/pkg/bft/consensus/config"
	mem "github.com/gnolang/gno/tm2/pkg/bft/mempool/config"
	rpc "github.com/gnolang/gno/tm2/pkg/bft/rpc/config"
	eventstore "github.com/gnolang/gno/tm2/pkg/bft/state/eventstore/types"
	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/errors"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	p2p "github.com/gnolang/gno/tm2/pkg/p2p/config"
	sdk "github.com/gnolang/gno/tm2/pkg/sdk/config"
	telemetry "github.com/gnolang/gno/tm2/pkg/telemetry/config"
)

var (
	errInvalidMoniker           = errors.New("moniker not set")
	errInvalidDBBackend         = errors.New("invalid DB backend")
	errInvalidDBPath            = errors.New("invalid DB path")
	errInvalidABCIMechanism     = errors.New("invalid ABCI mechanism")
	errInvalidProfListenAddress = errors.New("invalid profiling server listen address")
	errInvalidNodeKeyPath       = errors.New("invalid p2p node key path")
)

const (
	LocalABCI  = "local"
	SocketABCI = "socket"
)

// Regular expression for TCP or UNIX socket address
// TCP address: host:port (IPv4 example)
// UNIX address: unix:// followed by the path
var tcpUnixAddressRegex = regexp.MustCompile(`^(?:[0-9]{1,3}(\.[0-9]{1,3}){3}:[0-9]+|unix://.+)`)

// Config defines the top level configuration for a Tendermint node
type Config struct {
	// Top level options use an anonymous struct
	BaseConfig `toml:",squash"`

	// Options for services
	RPC          *rpc.RPCConfig       `json:"rpc" toml:"rpc" comment:"##### rpc server configuration options #####"`
	P2P          *p2p.P2PConfig       `json:"p2p" toml:"p2p" comment:"##### peer to peer configuration options #####"`
	Mempool      *mem.MempoolConfig   `json:"mempool" toml:"mempool" comment:"##### mempool configuration options #####"`
	Consensus    *cns.ConsensusConfig `json:"consensus" toml:"consensus" comment:"##### consensus configuration options #####"`
	TxEventStore *eventstore.Config   `json:"tx_event_store" toml:"tx_event_store" comment:"##### event store #####"`
	Telemetry    *telemetry.Config    `json:"telemetry" toml:"telemetry" comment:"##### node telemetry #####"`
	Application  *sdk.AppConfig       `json:"application" toml:"application" comment:"##### app settings #####"`
}

// DefaultConfig returns a default configuration for a Tendermint node
func DefaultConfig() *Config {
	return &Config{
		BaseConfig:   DefaultBaseConfig(),
		RPC:          rpc.DefaultRPCConfig(),
		P2P:          p2p.DefaultP2PConfig(),
		Mempool:      mem.DefaultMempoolConfig(),
		Consensus:    cns.DefaultConsensusConfig(),
		TxEventStore: eventstore.DefaultEventStoreConfig(),
		Telemetry:    telemetry.DefaultTelemetryConfig(),
		Application:  sdk.DefaultAppConfig(),
	}
}

type Option func(cfg *Config)

// LoadConfig loads the node configuration from disk
func LoadConfig(root string) (*Config, error) {
	// Initialize the config as default
	var (
		cfg        = DefaultConfig()
		configPath = filepath.Join(root, defaultConfigPath)
	)

	if !osm.FileExists(configPath) {
		return nil, fmt.Errorf("config file at %q does not exist", configPath)
	}

	// Load the configuration
	loadedCfg, loadErr := LoadConfigFile(configPath)
	if loadErr != nil {
		return nil, loadErr
	}

	// Merge the loaded config with the default values.
	// This is done in case the loaded config is missing values
	if err := mergo.Merge(loadedCfg, cfg); err != nil {
		return nil, err
	}

	// Set the root directory
	loadedCfg.SetRootDir(root)

	return loadedCfg, nil
}

// LoadOrMakeConfigWithOptions loads the configuration located in the given
// root directory, at [defaultConfigFilePath].
//
// If the config does not exist, it is created, starting from the values in
// `DefaultConfig` and applying the defaults in opts.
func LoadOrMakeConfigWithOptions(root string, opts ...Option) (*Config, error) {
	// Initialize the config as default
	var (
		cfg        = DefaultConfig()
		configPath = filepath.Join(root, defaultConfigPath)
	)

	// Config doesn't exist, create it
	// from the default one
	for _, opt := range opts {
		opt(cfg)
	}

	// Check if the config exists
	if osm.FileExists(configPath) {
		// Load the configuration
		loadedCfg, loadErr := LoadConfigFile(configPath)
		if loadErr != nil {
			return nil, loadErr
		}

		// Merge the loaded config with the default values
		if err := mergo.Merge(loadedCfg, cfg); err != nil {
			return nil, err
		}

		// Set the root directory
		loadedCfg.SetRootDir(root)

		// Make sure the directories are initialized
		if err := loadedCfg.EnsureDirs(); err != nil {
			return nil, err
		}

		return loadedCfg, nil
	}

	cfg.SetRootDir(root)

	// Make sure the directories are initialized
	if err := cfg.EnsureDirs(); err != nil {
		return nil, err
	}

	// Validate the configuration
	if validateErr := cfg.ValidateBasic(); validateErr != nil {
		return nil, fmt.Errorf("unable to validate config, %w", validateErr)
	}

	// Save the config
	if err := WriteConfigFile(configPath, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// testP2PConfig returns a configuration for testing the peer-to-peer layer
func testP2PConfig() *p2p.P2PConfig {
	cfg := p2p.DefaultP2PConfig()
	cfg.ListenAddress = "tcp://0.0.0.0:26656"
	cfg.FlushThrottleTimeout = 10 * time.Millisecond

	return cfg
}

// TestConfig returns a configuration that can be used for testing
func TestConfig() *Config {
	return &Config{
		BaseConfig:   testBaseConfig(),
		RPC:          rpc.TestRPCConfig(),
		P2P:          testP2PConfig(),
		Mempool:      mem.TestMempoolConfig(),
		Consensus:    cns.TestConsensusConfig(),
		TxEventStore: eventstore.DefaultEventStoreConfig(),
		Telemetry:    telemetry.DefaultTelemetryConfig(),
		Application:  sdk.DefaultAppConfig(),
	}
}

// SetRootDir sets the RootDir for all Config structs
func (cfg *Config) SetRootDir(root string) *Config {
	cfg.BaseConfig.RootDir = root
	cfg.RPC.RootDir = root
	cfg.P2P.RootDir = root
	cfg.Mempool.RootDir = root
	cfg.Consensus.RootDir = root
	cfg.Consensus.PrivValidator.RootDir = (filepath.Join(root, DefaultSecretsDir))

	return cfg
}

// EnsureDirs ensures default directories in root dir (and root dir).
func (cfg *Config) EnsureDirs() error {
	rootDir := cfg.BaseConfig.RootDir

	if err := osm.EnsureDir(rootDir, DefaultDirPerm); err != nil {
		return fmt.Errorf("no root directory, %w", err)
	}

	if err := osm.EnsureDir(filepath.Join(rootDir, DefaultConfigDir), DefaultDirPerm); err != nil {
		return fmt.Errorf("no config directory, %w", err)
	}

	if err := osm.EnsureDir(filepath.Join(rootDir, DefaultSecretsDir), DefaultDirPerm); err != nil {
		return fmt.Errorf("no secrets directory, %w", err)
	}

	if err := osm.EnsureDir(filepath.Join(rootDir, DefaultDBDir), DefaultDirPerm); err != nil {
		return fmt.Errorf("no DB directory, %w", err)
	}

	return nil
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
	if err := cfg.Application.ValidateBasic(); err != nil {
		return errors.Wrap(err, "Error in [application] section")
	}
	return nil
}

// -----------------------------------------------------------------------------

var (
	DefaultDBDir      = "db"
	DefaultConfigDir  = "config"
	DefaultSecretsDir = "secrets"

	DefaultConfigFileName = "config.toml"
	defaultNodeKeyName    = "node_key.json"

	defaultConfigPath  = filepath.Join(DefaultConfigDir, DefaultConfigFileName)
	defaultNodeKeyPath = filepath.Join(DefaultSecretsDir, defaultNodeKeyName)
)

// BaseConfig defines the base configuration for a Tendermint node.
type BaseConfig struct {
	// chainID is unexposed and immutable but here for convenience
	chainID string

	// The root directory for all data.
	// The node directory contains:
	//
	//	┌── db/
	//	│   ├── blockstore.db (folder)
	//	│   ├── gnolang.db (folder)
	//	│   └── state.db (folder)
	//	├── wal/
	//	│   └── cs.wal (folder)
	//	├── secrets/
	//	│   ├── priv_validator_state.json
	//	│   ├── node_key.json
	//	│   └── priv_validator_key.json
	//	└── config/
	//	    └── config.toml (optional)
	RootDir string `toml:"home"`

	// TCP or UNIX socket address of the ABCI application,
	// or the name of an ABCI application compiled in with the Tendermint binary,
	// or empty if local application instance.
	ProxyApp string `toml:"proxy_app" comment:"TCP or UNIX socket address of the ABCI application, \n or the name of an ABCI application compiled in with the Tendermint binary"`

	// Local application instance in lieu of remote app.
	LocalApp abci.Application `toml:"-"`

	// A custom human readable name for this node
	Moniker string `toml:"moniker" comment:"A custom human readable name for this node"`

	// If this node is many blocks behind the tip of the chain, FastSync
	// allows them to catchup quickly by downloading blocks in parallel
	// and verifying their commits
	FastSyncMode bool `toml:"fast_sync" comment:"If this node is many blocks behind the tip of the chain, FastSync\n allows them to catchup quickly by downloading blocks in parallel\n and verifying their commits"`

	// Database backend: pebbledb | goleveldb | boltdb
	// * pebbledb (github.com/cockroachdb/pebble)
	//   - pure go
	//   - stable
	// * goleveldb (github.com/syndtr/goleveldb)
	//   - pure go
	//   - stable
	//   - use goleveldb build tag
	// * boltdb (uses etcd's fork of bolt - go.etcd.io/bbolt)
	//   - EXPERIMENTAL
	//   - may be faster is some use-cases (random reads - indexer)
	//   - use boltdb build tag (go build -tags boltdb)
	DBBackend string `toml:"db_backend" comment:"Database backend: pebbledb | goleveldb | boltdb\n* pebbledb (github.com/cockroachdb/pebble)\n  - pure go\n  - stable\n* goleveldb (github.com/syndtr/goleveldb)\n  - pure go\n  - stable\n  - use goleveldb build tag\n* boltdb (uses etcd's fork of bolt - go.etcd.io/bbolt)\n  - EXPERIMENTAL\n  - may be faster is some use-cases (random reads - indexer)\n  - use boltdb build tag (go build -tags boltdb)"`

	// Database directory
	DBPath string `toml:"db_dir" comment:"Database directory"`

	// A JSON file containing the private key to use for p2p authenticated encryption
	NodeKey string `toml:"node_key_file" comment:"Path to the JSON file containing the private key to use for node authentication in the p2p protocol"`

	// Mechanism to connect to the ABCI application: local | socket
	ABCI string `toml:"abci" comment:"Mechanism to connect to the ABCI application: socket | grpc"`

	// TCP or UNIX socket address for the profiling server to listen on
	ProfListenAddress string `toml:"prof_laddr" comment:"TCP or UNIX socket address for the profiling server to listen on"`
}

// DefaultBaseConfig returns a default base configuration for a Tendermint node
func DefaultBaseConfig() BaseConfig {
	return BaseConfig{
		NodeKey:           defaultNodeKeyPath,
		Moniker:           defaultMoniker,
		ProxyApp:          "tcp://127.0.0.1:26658",
		ABCI:              SocketABCI,
		ProfListenAddress: "",
		FastSyncMode:      true,
		DBBackend:         db.PebbleDBBackend.String(),
		DBPath:            DefaultDBDir,
	}
}

// testBaseConfig returns a base configuration for testing a Tendermint node
func testBaseConfig() BaseConfig {
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

// NodeKeyFile returns the full path to the node_key.json file
func (cfg BaseConfig) NodeKeyFile() string {
	return filepath.Join(cfg.RootDir, cfg.NodeKey)
}

// DBDir returns the full path to the database directory
func (cfg BaseConfig) DBDir() string {
	if filepath.IsAbs(cfg.DBPath) {
		return cfg.DBPath
	}

	return filepath.Join(cfg.RootDir, cfg.DBPath)
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

// ValidateBasic performs basic validation (checking param bounds, etc.) and
// returns an error if any check fails.
func (cfg BaseConfig) ValidateBasic() error {
	// Verify the moniker
	if cfg.Moniker == "" {
		return errInvalidMoniker
	}

	// Verify the DB backend
	// This will reject also any databases that haven't been added with build tags.
	// always reject memdb, as it shouldn't be used as a real-life database.
	if cfg.DBBackend == "memdb" ||
		!slices.Contains(db.BackendList(), db.BackendType(cfg.DBBackend)) {
		return errInvalidDBBackend
	}

	// Verify the DB path is set
	if cfg.DBPath == "" {
		return errInvalidDBPath
	}

	// Verify the p2p private key exists
	if cfg.NodeKey == "" {
		return errInvalidNodeKeyPath
	}

	// Verify the correct ABCI mechanism is set
	if cfg.ABCI != LocalABCI &&
		cfg.ABCI != SocketABCI {
		return errInvalidABCIMechanism
	}

	// Verify the profiling listen address
	if cfg.ProfListenAddress != "" && !tcpUnixAddressRegex.MatchString(cfg.ProfListenAddress) {
		return errInvalidProfListenAddress
	}

	return nil
}
