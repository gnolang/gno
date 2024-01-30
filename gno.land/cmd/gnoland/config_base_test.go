package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// initializeTestConfig initializes a default configuration
// at a temporary path
func initializeTestConfig(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.toml")
	cfg := config.DefaultConfig()

	require.NoError(t, config.WriteConfigFile(path, cfg))

	return path
}

func TestConfig_Base_Invalid(t *testing.T) {
	t.Parallel()

	t.Run("invalid config path", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"config",
			"base",
			"--config-path",
			"",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, "unable to load config")
	})

	t.Run("invalid config change", func(t *testing.T) {
		t.Parallel()

		// Setup the test config
		path := initializeTestConfig(t)

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"config",
			"base",
			"--config-path",
			path,
			"--db-backend",
			"random db backend",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, "unable to validate config")
	})

	t.Run("invalid sync toggle value", func(t *testing.T) {
		t.Parallel()

		// Setup the test config
		path := initializeTestConfig(t)

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"config",
			"base",
			"--config-path",
			path,
			"--fast-sync",
			"random toggle",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, errInvalidToggleValue)
	})

	t.Run("invalid filter peers toggle value", func(t *testing.T) {
		t.Parallel()

		// Setup the test config
		path := initializeTestConfig(t)

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"config",
			"base",
			"--config-path",
			path,
			"--filter-peers",
			"random toggle",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, errInvalidToggleValue)
	})
}

func TestConfig_Base_Valid(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name     string
		flags    []string
		verifyFn func(loadedCfg *config.Config, value string)
	}{
		{
			"root dir updated",
			[]string{
				"--root-dir",
				"example root dir",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RootDir)
			},
		},
		{
			"proxy app updated",
			[]string{
				"--proxy-app",
				"example proxy app",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.ProxyApp)
			},
		},
		{
			"moniker updated",
			[]string{
				"--moniker",
				"example moniker",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Moniker)
			},
		},
		{
			"fast sync mode updated",
			[]string{
				"--fast-sync",
				"off",
			},
			func(loadedCfg *config.Config, value string) {
				boolVal := false
				if value == onValue {
					boolVal = true
				}

				assert.Equal(t, boolVal, loadedCfg.FastSyncMode)
			},
		},
		{
			"db backend updated",
			[]string{
				"--db-backend",
				config.LevelDBName,
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.DBBackend)
			},
		},
		{
			"db path updated",
			[]string{
				"--db-path",
				"example path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.DBPath)
			},
		},
		{
			"genesis path updated",
			[]string{
				"--genesis-file",
				"example path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Genesis)
			},
		},
		{
			"validator key updated",
			[]string{
				"--validator-key-file",
				"example path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.PrivValidatorKey)
			},
		},
		{
			"validator state file updated",
			[]string{
				"--validator-state-file",
				"example path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.PrivValidatorState)
			},
		},
		{
			"validator listen addr updated",
			[]string{
				"--validator-laddr",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.PrivValidatorListenAddr)
			},
		},
		{
			"node key path updated",
			[]string{
				"--node-key",
				"example path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.NodeKey)
			},
		},
		{
			"abci updated",
			[]string{
				"--abci",
				config.LocalABCI,
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.ABCI)
			},
		},
		{
			"profiling listen address updated",
			[]string{
				"--prof-laddr",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.ProfListenAddress)
			},
		},
		{
			"filter peers flag updated",
			[]string{
				"--filter-peers",
				onValue,
			},
			func(loadedCfg *config.Config, value string) {
				boolVal := false
				if value == onValue {
					boolVal = true
				}

				assert.Equal(t, boolVal, loadedCfg.FilterPeers)
			},
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Setup the test config
			path := initializeTestConfig(t)
			args := []string{
				"config",
				"base",
				"--config-path",
				path,
			}

			// Create the command
			cmd := newRootCmd(commands.NewTestIO())
			args = append(args, testCase.flags...)

			// Run the command
			cmdErr := cmd.ParseAndRun(context.Background(), args)
			require.NoError(t, cmdErr)

			// Make sure the config was updated
			loadedCfg, err := config.LoadConfigFile(path)
			require.NoError(t, err)

			testCase.verifyFn(loadedCfg, testCase.flags[len(testCase.flags)-1])
		})
	}
}
