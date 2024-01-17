package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_LoadOrMakeConfigWithOptions(t *testing.T) {
	t.Parallel()

	t.Run("existing configuration", func(t *testing.T) {
		t.Parallel()

		// Provide an empty directory
		cfgDir := t.TempDir()
		cfgPath := join(cfgDir, defaultConfigFilePath)

		// Create a default config
		cfg := DefaultConfig()
		cfg.SetRootDir(cfgDir)

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
		cfgPath := join(cfgDir, defaultConfigFilePath)

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

		monkier := "dummy monkier"

		// Provide an empty directory
		cfgDir := t.TempDir()
		cfgPath := join(cfgDir, defaultConfigFilePath)

		cfg, err := LoadOrMakeConfigWithOptions(
			cfgDir,
			func(cfg *Config) {
				cfg.BaseConfig.Moniker = monkier
			},
		)
		require.NoError(t, err)

		// Make sure the returned config was saved
		loadedCfg, loadErr := LoadConfigFile(cfgPath)
		require.NoError(t, loadErr)

		loadedCfg.SetRootDir(cfgDir)

		assert.Equal(t, cfg, loadedCfg)
	})
}
