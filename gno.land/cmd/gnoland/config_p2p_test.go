package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_P2P_Invalid(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name  string
		flags []string
	}{
		{
			"invalid upnp toggle value",
			[]string{
				"--upnp",
				"random toggle",
			},
		},
		{
			"invalid pex toggle value",
			[]string{
				"--pex-reactor",
				"random toggle",
			},
		},
		{
			"invalid seed mode toggle value",
			[]string{
				"--seed-mode",
				"random toggle",
			},
		},
		{
			"invalid allow duplicate IPs toggle value",
			[]string{
				"--allow-duplicate-ip",
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
				"p2p",
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
			"p2p",
			"--config-path",
			"",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, "unable to load config")
	})
}

func TestConfig_P2P_Valid(t *testing.T) {
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
				assert.Equal(t, value, loadedCfg.P2P.RootDir)
			},
		},
		{
			"listen address updated",
			[]string{
				"--listen-address",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.ListenAddress)
			},
		},
		{
			"external address updated",
			[]string{
				"--external-address",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.ExternalAddress)
			},
		},
		{
			"seeds updated",
			[]string{
				"--seeds",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.Seeds)
			},
		},
		{
			"persistent peers updated",
			[]string{
				"--persistent-peers",
				"nodeID@0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.PersistentPeers)
			},
		},
		{
			"upnp toggle updated",
			[]string{
				"--upnp",
				offValue,
			},
			func(loadedCfg *config.Config, value string) {
				boolVal := false
				if value == onValue {
					boolVal = true
				}

				assert.Equal(t, boolVal, loadedCfg.P2P.UPNP)
			},
		},
		{
			"max inbound peers updated",
			[]string{
				"--max-inbound-peers",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.MaxNumInboundPeers))
			},
		},
		{
			"max outbound peers updated",
			[]string{
				"--max-outbound-peers",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.MaxNumOutboundPeers))
			},
		},
		{
			"flush throttle timeout updated",
			[]string{
				"--flush-throttle-timeout",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.FlushThrottleTimeout.String())
			},
		},
		{
			"max package payload size updated",
			[]string{
				"--max-message-payload",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.MaxPacketMsgPayloadSize))
			},
		},
		{
			"send rate updated",
			[]string{
				"--send-rate",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.SendRate))
			},
		},
		{
			"receive rate updated",
			[]string{
				"--receive-rate",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.RecvRate))
			},
		},
		{
			"pex reactor toggle updated",
			[]string{
				"--pex-reactor",
				offValue,
			},
			func(loadedCfg *config.Config, value string) {
				boolVal := false
				if value == onValue {
					boolVal = true
				}

				assert.Equal(t, boolVal, loadedCfg.P2P.PexReactor)
			},
		},
		{
			"seed mode updated",
			[]string{
				"--seed-mode",
				offValue,
			},
			func(loadedCfg *config.Config, value string) {
				boolVal := false
				if value == onValue {
					boolVal = true
				}

				assert.Equal(t, boolVal, loadedCfg.P2P.SeedMode)
			},
		},
		{
			"private peer IDs updated",
			[]string{
				"--private-peer-ids",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.PrivatePeerIDs)
			},
		},
		{
			"allow duplicate IPs updated",
			[]string{
				"--allow-duplicate-ip",
				offValue,
			},
			func(loadedCfg *config.Config, value string) {
				boolVal := false
				if value == onValue {
					boolVal = true
				}

				assert.Equal(t, boolVal, loadedCfg.P2P.AllowDuplicateIP)
			},
		},
		{
			"handshake timeout updated",
			[]string{
				"--handshake-timeout",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.HandshakeTimeout.String())
			},
		},
		{
			"dial timeout updated",
			[]string{
				"--dial-timeout",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.DialTimeout.String())
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
				"p2p",
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
