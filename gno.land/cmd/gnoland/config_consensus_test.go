package main

import (
	"context"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Consensus_Invalid(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name  string
		flags []string
	}{
		{
			"invalid skip commit timeout toggle value",
			[]string{
				"--skip-commit-timeout",
				"random toggle",
			},
		},
		{
			"invalid create empty blocks toggle value",
			[]string{
				"--create-empty-blocks",
				"random toggle",
			},
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Setup the test config
			path := initializeTestConfig(t)

			// Create the command
			cmd := newRootCmd(commands.NewTestIO())
			args := []string{
				"config",
				"consensus",
				"--config-path",
				path,
			}

			args = append(args, testCase.flags...)

			// Run the command
			cmdErr := cmd.ParseAndRun(context.Background(), args)
			assert.ErrorIs(t, cmdErr, errInvalidToggleValue)
		})
	}

	t.Run("invalid config path", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"config",
			"consensus",
			"--config-path",
			"",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, "unable to load config")
	})
}

func TestConfig_Consensus_Valid(t *testing.T) {
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
				assert.Equal(t, value, loadedCfg.Consensus.RootDir)
			},
		},
		{
			"WAL path updated",
			[]string{
				"--wal-path",
				"example WAL path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.WalPath)
			},
		},
		{
			"propose timeout updated",
			[]string{
				"--timeout-propose",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPropose.String())
			},
		},
		{
			"propose timeout delta updated",
			[]string{
				"--timeout-propose-delta",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutProposeDelta.String())
			},
		},
		{
			"prevote timeout updated",
			[]string{
				"--timeout-prevote",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPrevote.String())
			},
		},
		{
			"prevote timeout delta updated",
			[]string{
				"--timeout-prevote-delta",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPrevoteDelta.String())
			},
		},
		{
			"precommit timeout updated",
			[]string{
				"--timeout-precommit",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPrecommit.String())
			},
		},
		{
			"precommit timeout delta updated",
			[]string{
				"--timeout-precommit-delta",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPrecommitDelta.String())
			},
		},
		{
			"commit timeout updated",
			[]string{
				"--timeout-commit",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutCommit.String())
			},
		},
		{
			"skip commit timeout toggle updated",
			[]string{
				"--skip-commit-timeout",
				onValue,
			},
			func(loadedCfg *config.Config, value string) {
				boolVal := false
				if value == onValue {
					boolVal = true
				}

				assert.Equal(t, boolVal, loadedCfg.Consensus.SkipTimeoutCommit)
			},
		},
		{
			"create empty blocks toggle updated",
			[]string{
				"--create-empty-blocks",
				offValue,
			},
			func(loadedCfg *config.Config, value string) {
				boolVal := false
				if value == onValue {
					boolVal = true
				}

				assert.Equal(t, boolVal, loadedCfg.Consensus.CreateEmptyBlocks)
			},
		},
		{
			"create empty blocks interval updated",
			[]string{
				"--empty-blocks-interval",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.CreateEmptyBlocksInterval.String())
			},
		},
		{
			"peer gossip sleep duration updated",
			[]string{
				"--gossip-sleep-duration",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.PeerGossipSleepDuration.String())
			},
		},
		{
			"peer query majority sleep duration updated",
			[]string{
				"--query-sleep-duration",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.PeerQueryMaj23SleepDuration.String())
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
				"consensus",
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
