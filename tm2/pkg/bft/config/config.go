package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

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
)

var (
	errInvalidMoniker                    = errors.New("moniker not set")
	errInvalidDBBackend                  = errors.New("invalid DB backend")
	errInvalidDBPath                     = errors.New("invalid DB path")
	errInvalidGenesisPath                = errors.New("invalid genesis path")
	errInvalidPrivValidatorKeyPath       = errors.New("invalid private validator key path")
	errInvalidPrivValidatorStatePath     = errors.New("invalid private validator state file path")
	errInvalidABCIMechanism              = errors.New("invalid ABCI mechanism")
	errInvalidPrivValidatorListenAddress = errors.New("invalid PrivValidator listen address")
	errInvalidProfListenAddress          = errors.New("invalid profiling server listen address")
	errInvalidNodeKeyPath                = errors.New("invalid p2p node key path")
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
	RPC          *rpc.RPCConfig       `toml:"rpc" comment:"##### rpc server configuration options #####"`
	P2P          *p2p.P2PConfig       `toml:"p2p" comment:"##### peer to peer configuration options #####"`
	Mempool      *mem.MempoolConfig   `toml:"mempool" comment:"##### mempool configuration options #####"`
	Consensus    *cns.ConsensusConfig `toml:"consensus" comment:"##### consensus configuration options #####"`
	TxEventStore *eventstore.Config   `toml:"tx_event_store" comment:"##### event store #####"`
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
	}
}

type Option func(cfg *Config)

// LoadOrMakeConfigWithOptions loads the configuration located in the given
// root directory, at [defaultConfigFilePath].
//
// If the config does not exist, it is created, starting from the values in
// `DefaultConfig` and applying the defaults in opts.
func LoadOrMakeConfigWithOptions(root string, opts ...Option) (*Config, error) {
	// Initialize the config as default
	var (
		cfg        = DefaultConfig()
		configPath = join(root, defaultConfigFilePath)
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

// TestConfig returns a configuration that can be used for testing
func TestConfig() *Config {
	return &Config{
		BaseConfig:   testBaseConfig(),
		RPC:          rpc.TestRPCConfig(),
		P2P:          p2p.TestP2PConfig(),
		Mempool:      mem.TestMempoolConfig(),
		Consensus:    cns.TestConsensusConfig(),
		TxEventStore: eventstore.DefaultEventStoreConfig(),
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
func (cfg *Config) EnsureDirs() error {
	rootDir := cfg.BaseConfig.RootDir

	if err := osm.EnsureDir(rootDir, DefaultDirPerm); err != nil {
		return fmt.Errorf("no root directory, %w", err)
	}

	if err := osm.EnsureDir(filepath.Join(rootDir, defaultConfigDir), DefaultDirPerm); err != nil {
		return fmt.Errorf("no config directory, %w", err)
	}

	if err := osm.EnsureDir(filepath.Join(rootDir, defaultDataDir), DefaultDirPerm); err != nil {
		return fmt.Errorf("no data directory, %w", err)
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
	return nil
}

// -----------------------------------------------------------------------------
// BaseConfig

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
	ProxyApp string `toml:"proxy_app" comment:"TCP or UNIX socket address of the ABCI application, \n or the name of an ABCI application compiled in with the Tendermint binary"`

	// Local application instance in lieu of remote app.
	LocalApp abci.Application `toml:"-"`

	// A custom human readable name for this node
	Moniker string `toml:"moniker" comment:"A custom human readable name for this node"`

	// If this node is many blocks behind the tip of the chain, FastSync
	// allows them to catchup quickly by downloading blocks in parallel
	// and verifying their commits
	FastSyncMode bool `toml:"fast_sync" comment:"If this node is many blocks behind the tip of the chain, FastSync\n allows them to catchup quickly by downloading blocks in parallel\n and verifying their commits"`

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
	DBBackend string `toml:"db_backend" comment:"Database backend: goleveldb | cleveldb | boltdb\n * goleveldb (github.com/syndtr/goleveldb - most popular implementation)\n  - pure go\n  - stable\n * cleveldb (uses levigo wrapper)\n  - fast\n  - requires gcc\n  - use cleveldb build tag (go build -tags cleveldb)\n * boltdb (uses etcd's fork of bolt - go.etcd.io/bbolt)\n  - EXPERIMENTAL\n  - may be faster is some use-cases (random reads - indexer)\n  - use boltdb build tag (go build -tags boltdb)"`

	// Database directory
	DBPath string `toml:"db_dir" comment:"Database directory"`

	// Path to the JSON file containing the initial validator set and other meta data
	Genesis string `toml:"genesis_file" comment:"Path to the JSON file containing the initial validator set and other meta data"`

	// Path to the JSON file containing the private key to use as a validator in the consensus protocol
	PrivValidatorKey string `toml:"priv_validator_key_file" comment:"Path to the JSON file containing the private key to use as a validator in the consensus protocol"`

	// Path to the JSON file containing the last sign state of a validator
	PrivValidatorState string `toml:"priv_validator_state_file" comment:"Path to the JSON file containing the last sign state of a validator"`

	// TCP or UNIX socket address for Tendermint to listen on for
	// connections from an external PrivValidator process
	PrivValidatorListenAddr string `toml:"priv_validator_laddr" comment:"TCP or UNIX socket address for Tendermint to listen on for\n connections from an external PrivValidator process"`

	// A JSON file containing the private key to use for p2p authenticated encryption
	NodeKey string `toml:"node_key_file" comment:"Path to the JSON file containing the private key to use for node authentication in the p2p protocol"`

	// Mechanism to connect to the ABCI application: local | socket
	ABCI string `toml:"abci" comment:"Mechanism to connect to the ABCI application: socket | grpc"`

	// TCP or UNIX socket address for the profiling server to listen on
	ProfListenAddress string `toml:"prof_laddr" comment:"TCP or UNIX socket address for the profiling server to listen on"`

	// If true, query the ABCI app on connecting to a new peer
	// so the app can decide if we should keep the connection or not
	FilterPeers bool `toml:"filter_peers" comment:"If true, query the ABCI app on connecting to a new peer\n so the app can decide if we should keep the connection or not"` // false
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
		ABCI:               SocketABCI,
		ProfListenAddress:  "",
		FastSyncMode:       true,
		FilterPeers:        false,
		DBBackend:          db.GoLevelDBBackend.String(),
		DBPath:             "data",
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
	if cfg.DBBackend != db.GoLevelDBBackend.String() &&
		cfg.DBBackend != db.CLevelDBBackend.String() &&
		cfg.DBBackend != db.BoltDBBackend.String() {
		return errInvalidDBBackend
	}

	// Verify the DB path is set
	if cfg.DBPath == "" {
		return errInvalidDBPath
	}

	// Verify the genesis path is set
	if cfg.Genesis == "" {
		return errInvalidGenesisPath
	}

	// Verify the validator private key path is set
	if cfg.PrivValidatorKey == "" {
		return errInvalidPrivValidatorKeyPath
	}

	// Verify the validator state file path is set
	if cfg.PrivValidatorState == "" {
		return errInvalidPrivValidatorStatePath
	}

	// Verify the PrivValidator listen address
	if cfg.PrivValidatorListenAddr != "" &&
		!tcpUnixAddressRegex.MatchString(cfg.PrivValidatorListenAddr) {
		return errInvalidPrivValidatorListenAddress
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
