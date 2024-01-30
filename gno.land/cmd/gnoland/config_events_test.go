package main

import (
	"context"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/state/eventstore/file"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Events_Invalid(t *testing.T) {
	t.Parallel()

	// Create the command
	cmd := newRootCmd(commands.NewTestIO())
	args := []string{
		"config",
		"events",
		"--config-path",
		"",
	}

	// Run the command
	cmdErr := cmd.ParseAndRun(context.Background(), args)
	assert.ErrorContains(t, cmdErr, "unable to load config")
}

func TestConfig_Events_Valid(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name     string
		flags    []string
		verifyFn func(loadedCfg *config.Config, value string)
	}{
		{
			"event store type updated",
			[]string{
				"--event-store-type",
				file.EventStoreType,
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.TxEventStore.EventStoreType)
			},
		},
		{
			"event store params updated",
			[]string{
				"--event-store-params",
				"key1=value1",
				"--event-store-params",
				"key2=value2",
			},
			func(loadedCfg *config.Config, value string) {
				val, ok := loadedCfg.TxEventStore.Params["key1"]
				assert.True(t, ok)
				assert.Equal(t, "value1", val)

				val, ok = loadedCfg.TxEventStore.Params["key2"]
				assert.True(t, ok)
				assert.Equal(t, "value2", val)
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
				"events",
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
