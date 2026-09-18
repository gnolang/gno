package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	// register every DB backend so validation accepts whatever the default is
	_ "github.com/gnolang/gno/tm2/pkg/db/_all"
)

// writeConfigBytes writes raw TOML content at the default config path under root
func writeConfigBytes(t *testing.T, root string, content []byte) {
	t.Helper()

	cfgPath := filepath.Join(root, defaultConfigPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(cfgPath), 0o755))
	require.NoError(t, os.WriteFile(cfgPath, content, 0o644))
}

// writeConfig saves the given config at the default config path under root
func writeConfig(t *testing.T, root string, cfg *Config) {
	t.Helper()

	cfgPath := filepath.Join(root, defaultConfigPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(cfgPath), 0o755))
	require.NoError(t, WriteConfigFile(cfgPath, cfg))
}

func TestConfig_LoadConfig(t *testing.T) {
	t.Parallel()

	t.Run("explicit zero values are honored", func(t *testing.T) {
		t.Parallel()

		cfgDir := t.TempDir()

		// Write a full config whose values differ from the defaults
		// only by being explicitly set to Go zero values
		cfg := DefaultConfig()
		cfg.Mempool.Recheck = false
		cfg.Consensus.CreateEmptyBlocks = false
		writeConfig(t, cfgDir, cfg)

		loadedCfg, loadErr := LoadConfig(cfgDir)
		require.NoError(t, loadErr)

		assert.False(t, loadedCfg.Mempool.Recheck)
		assert.False(t, loadedCfg.Consensus.CreateEmptyBlocks)
	})

	t.Run("explicit empty slice is honored", func(t *testing.T) {
		t.Parallel()

		cfgDir := t.TempDir()

		// The default for cors_allowed_methods is a non-empty slice;
		// an explicit empty array in the file must disable it
		cfg := DefaultConfig()
		cfg.RPC.CORSAllowedMethods = []string{}
		writeConfig(t, cfgDir, cfg)

		loadedCfg, loadErr := LoadConfig(cfgDir)
		require.NoError(t, loadErr)

		assert.Empty(t, loadedCfg.RPC.CORSAllowedMethods)
	})

	t.Run("keys absent from the file keep defaults", func(t *testing.T) {
		t.Parallel()

		cfgDir := t.TempDir()

		// Write a partial config, with entire sections and
		// individual keys within present sections omitted
		writeConfigBytes(t, cfgDir, []byte(
			"moniker = \"from-file\"\n[p2p]\npersistent_peers = \"node0@127.0.0.1:26656\"\n",
		))

		loadedCfg, loadErr := LoadConfig(cfgDir)
		require.NoError(t, loadErr)

		defaultCfg := DefaultConfig()

		// Present keys come from the file
		assert.Equal(t, "from-file", loadedCfg.Moniker)
		assert.Equal(t, "node0@127.0.0.1:26656", loadedCfg.P2P.PersistentPeers)

		// Keys absent from a present section keep their defaults
		assert.Equal(t, defaultCfg.P2P.MaxNumOutboundPeers, loadedCfg.P2P.MaxNumOutboundPeers)

		// Absent sections keep their defaults
		assert.Equal(t, defaultCfg.Mempool.Recheck, loadedCfg.Mempool.Recheck)
		assert.Equal(t, defaultCfg.Consensus.CreateEmptyBlocks, loadedCfg.Consensus.CreateEmptyBlocks)
		assert.Equal(t, defaultCfg.Mempool.Size, loadedCfg.Mempool.Size)
	})

	t.Run("non-zero values load unchanged", func(t *testing.T) {
		t.Parallel()

		cfgDir := t.TempDir()

		cfg := DefaultConfig()
		cfg.P2P.PersistentPeers = "node0@127.0.0.1:26656"
		cfg.P2P.SendRate = 1024000
		writeConfig(t, cfgDir, cfg)

		loadedCfg, loadErr := LoadConfig(cfgDir)
		require.NoError(t, loadErr)

		assert.Equal(t, cfg.P2P.PersistentPeers, loadedCfg.P2P.PersistentPeers)
		assert.Equal(t, cfg.P2P.SendRate, loadedCfg.P2P.SendRate)

		// A slice present in the file replaces the default slice
		// rather than appending to it
		assert.Equal(t, cfg.RPC.CORSAllowedMethods, loadedCfg.RPC.CORSAllowedMethods)
	})
}

func TestConfig_LoadOrMakeConfigWithOptions(t *testing.T) {
	t.Parallel()

	t.Run("existing configuration", func(t *testing.T) {
		t.Parallel()

		// Provide an empty directory
		cfgDir := t.TempDir()
		cfgPath := filepath.Join(cfgDir, defaultConfigPath)

		// Create a default config
		cfg := DefaultConfig()
		cfg.SetRootDir(cfgDir)

		// Make an incremental changes
		cfg.Moniker = "custom moniker"

		// Make sure the cfg paths are initialized
		require.NoError(t, cfg.EnsureDirs())

		// Write the config
		require.NoError(t, WriteConfigFile(cfgPath, cfg))

		// Load the config
		loadedCfg, loadErr := LoadOrMakeConfigWithOptions(cfgDir)
		require.NoError(t, loadErr)

		assert.Equal(t, cfg, loadedCfg)
	})

	t.Run("no existing config", func(t *testing.T) {
		t.Parallel()

		// Provide an empty directory
		cfgDir := t.TempDir()
		cfgPath := filepath.Join(cfgDir, defaultConfigPath)

		cfg, err := LoadOrMakeConfigWithOptions(cfgDir)
		require.NoError(t, err)

		// Make sure the returned cfg is the default one
		expectedCfg := DefaultConfig()
		expectedCfg.SetRootDir(cfgDir)

		assert.Equal(t, expectedCfg, cfg)

		// Make sure the returned config was saved
		loadedCfg, loadErr := LoadConfigFile(cfgPath)
		require.NoError(t, loadErr)

		loadedCfg.SetRootDir(cfgDir)

		assert.Equal(t, cfg, loadedCfg)
	})

	t.Run("no existing config, with options", func(t *testing.T) {
		t.Parallel()

		moniker := "dummy moniker"

		// Provide an empty directory
		cfgDir := t.TempDir()
		cfgPath := filepath.Join(cfgDir, defaultConfigPath)

		cfg, err := LoadOrMakeConfigWithOptions(
			cfgDir,
			func(cfg *Config) {
				cfg.BaseConfig.Moniker = moniker
			},
		)
		require.NoError(t, err)

		// Make sure the returned config was saved
		loadedCfg, loadErr := LoadConfigFile(cfgPath)
		require.NoError(t, loadErr)

		loadedCfg.SetRootDir(cfgDir)

		assert.Equal(t, cfg, loadedCfg)
	})

	t.Run("file values take precedence over options", func(t *testing.T) {
		t.Parallel()

		cfgDir := t.TempDir()

		// Write a full config with an explicit zero value
		// and a custom moniker
		cfg := DefaultConfig()
		cfg.Moniker = "from-file"
		cfg.Mempool.Recheck = false
		writeConfig(t, cfgDir, cfg)

		loadedCfg, loadErr := LoadOrMakeConfigWithOptions(
			cfgDir,
			func(cfg *Config) {
				cfg.Moniker = "from-opt"
				cfg.Mempool.Recheck = true
			},
		)
		require.NoError(t, loadErr)

		assert.Equal(t, "from-file", loadedCfg.Moniker)
		assert.False(t, loadedCfg.Mempool.Recheck)
	})

	t.Run("options are kept for keys absent from the file", func(t *testing.T) {
		t.Parallel()

		cfgDir := t.TempDir()

		// Write a partial config that does not set the moniker
		writeConfigBytes(t, cfgDir, []byte("[mempool]\nrecheck = false\n"))

		loadedCfg, loadErr := LoadOrMakeConfigWithOptions(
			cfgDir,
			func(cfg *Config) {
				cfg.Moniker = "from-opt"
			},
		)
		require.NoError(t, loadErr)

		assert.Equal(t, "from-opt", loadedCfg.Moniker)
		assert.False(t, loadedCfg.Mempool.Recheck)
	})
}

func TestConfig_ValidateBaseConfig(t *testing.T) {
	t.Parallel()

	t.Run("valid default config", func(t *testing.T) {
		t.Parallel()

		c := DefaultConfig()

		assert.NoError(t, c.BaseConfig.ValidateBasic())
	})

	t.Run("invalid moniker", func(t *testing.T) {
		t.Parallel()

		c := DefaultConfig()
		c.Moniker = ""

		assert.ErrorIs(t, c.BaseConfig.ValidateBasic(), errInvalidMoniker)
	})

	t.Run("invalid DB backend", func(t *testing.T) {
		t.Parallel()

		c := DefaultConfig()
		c.DBBackend = "totally valid backend"

		assert.ErrorIs(t, c.BaseConfig.ValidateBasic(), errInvalidDBBackend)
	})

	t.Run("DB path not set", func(t *testing.T) {
		t.Parallel()

		c := DefaultConfig()
		c.DBPath = ""

		assert.ErrorIs(t, c.BaseConfig.ValidateBasic(), errInvalidDBPath)
	})

	t.Run("node key path not set", func(t *testing.T) {
		t.Parallel()

		c := DefaultConfig()
		c.NodeKey = ""

		assert.ErrorIs(t, c.BaseConfig.ValidateBasic(), errInvalidNodeKeyPath)
	})

	t.Run("invalid ABCI mechanism", func(t *testing.T) {
		t.Parallel()

		c := DefaultConfig()
		c.ABCI = "hopes and dreams"

		assert.ErrorIs(t, c.BaseConfig.ValidateBasic(), errInvalidABCIMechanism)
	})

	t.Run("invalid prof listen address", func(t *testing.T) {
		t.Parallel()

		c := DefaultConfig()
		c.ProfListenAddress = "beep.boop"

		assert.ErrorIs(t, c.BaseConfig.ValidateBasic(), errInvalidProfListenAddress)
	})
}

func TestConfig_DBDir(t *testing.T) {
	t.Parallel()

	t.Run("DB path is absolute", func(t *testing.T) {
		t.Parallel()

		c := DefaultConfig()
		c.RootDir = "/root"
		c.DBPath = "/abs/path"

		assert.Equal(t, c.DBPath, c.DBDir())
		assert.NotEqual(t, filepath.Join(c.RootDir, c.DBPath), c.DBDir())
	})

	t.Run("DB path is relative", func(t *testing.T) {
		t.Parallel()

		c := DefaultConfig()
		c.RootDir = "/root"
		c.DBPath = "relative/path"

		assert.Equal(t, filepath.Join(c.RootDir, c.DBPath), c.DBDir())
	})
}
