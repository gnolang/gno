package config

import (
	"fmt"
	"os"
	"path/filepath"

	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/pelletier/go-toml"
)

// DefaultDirPerm is the default permissions used when creating directories.
const DefaultDirPerm = 0o700

// LoadConfigFile loads the TOML node configuration from the specified path
func LoadConfigFile(path string) (*Config, error) {
	// Read the config file
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		return nil, readErr
	}

	// Parse the node config
	var nodeConfig Config

	if unmarshalErr := toml.Unmarshal(content, &nodeConfig); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	// Validate the config
	if validateErr := nodeConfig.ValidateBasic(); validateErr != nil {
		return nil, fmt.Errorf("unable to validate config, %w", validateErr)
	}

	return &nodeConfig, nil
}

/****** these are for production settings ***********/

// WriteConfigFile renders config using the template and writes it to configFilePath.
func WriteConfigFile(configFilePath string, config *Config) error {
	// Marshal the config
	configRaw, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("unable to TOML marshal config, %w", err)
	}

	if err := osm.WriteFile(configFilePath, configRaw, 0o644); err != nil {
		return fmt.Errorf("unable to write config file, %w", err)
	}

	return nil
}

/****** these are for test settings ***********/

func ResetTestRoot(testName string) (*Config, string) {
	chainID := "test-chain"

	// create a unique, concurrency-safe test directory under os.TempDir()
	testDir, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}

	rootDir, err := os.MkdirTemp(testDir, fmt.Sprintf("%s-%s_", chainID, testName))
	if err != nil {
		panic(err)
	}

	// ensure config and data subdirs are created
	if err := osm.EnsureDir(filepath.Join(rootDir, defaultConfigDir), DefaultDirPerm); err != nil {
		panic(err)
	}
	if err := osm.EnsureDir(filepath.Join(rootDir, defaultSecretsDir), DefaultDirPerm); err != nil {
		panic(err)
	}
	if err := osm.EnsureDir(filepath.Join(rootDir, DefaultDBDir), DefaultDirPerm); err != nil {
		panic(err)
	}

	baseConfig := DefaultBaseConfig()
	configFilePath := filepath.Join(rootDir, defaultConfigPath)
	// NOTE: this does not match the behaviour of the Gno.land node.
	// However, many tests rely on the fact that they can cleanup the directory
	// by doing RemoveAll on the rootDir; so to keep compatibility with that
	// behaviour, we place genesis.json in the rootDir.
	genesisFilePath := filepath.Join(rootDir, "genesis.json")
	privKeyFilePath := filepath.Join(rootDir, baseConfig.PrivValidatorKey)
	privStateFilePath := filepath.Join(rootDir, baseConfig.PrivValidatorState)

	// Write default config file if missing.
	if !osm.FileExists(configFilePath) {
		WriteConfigFile(configFilePath, DefaultConfig())
	}
	if !osm.FileExists(genesisFilePath) {
		if chainID == "" {
			chainID = "tendermint_test"
		}
		testGenesis := fmt.Sprintf(testGenesisFmt, chainID)
		osm.MustWriteFile(genesisFilePath, []byte(testGenesis), 0o644)
	}
	// we always overwrite the priv val
	osm.MustWriteFile(privKeyFilePath, []byte(testPrivValidatorKey), 0o644)
	osm.MustWriteFile(privStateFilePath, []byte(testPrivValidatorState), 0o644)

	config := TestConfig().SetRootDir(rootDir)

	return config, genesisFilePath
}

var testGenesisFmt = `{
  "genesis_time": "2018-10-10T08:20:13.695936996Z",
  "chain_id": "%s",
  "validators": [
    {
      "pub_key": {
        "@type": "/tm.PubKeyEd25519",
        "value": "cVt6w3C1DWYwwkAirnbsL49CoOe8T8ZR2BCB8MeOGRg="
      },
      "power": "10",
      "name": ""
    }
  ],
  "app_hash": ""
}`

var testPrivValidatorKey = `{
  "address": "g1uvwz22t0l2fv9az93wutmlusrjv5zdwx2n32d5",
  "pub_key": {
    "@type": "/tm.PubKeyEd25519",
    "value": "cVt6w3C1DWYwwkAirnbsL49CoOe8T8ZR2BCB8MeOGRg="
  },
  "priv_key": {
    "@type": "/tm.PrivKeyEd25519",
    "value": "Qq4Q9QH2flPSIJShbXPIocbrQtQ4S7Kdn31uI3sKZoJxW3rDcLUNZjDCQCKuduwvj0Kg57xPxlHYEIHwx44ZGA=="
  }
}`

var testPrivValidatorState = `{
  "height": "0",
  "round": "0",
  "step": 0
}`
