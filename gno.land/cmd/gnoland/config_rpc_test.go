package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_RPC_Invalid(t *testing.T) {
	t.Parallel()

	t.Run("invalid config path", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"config",
			"rpc",
			"--config-path",
			"",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, "unable to load config")
	})

	t.Run("invalid unsafe toggle value", func(t *testing.T) {
		t.Parallel()

		// Setup the test config
		path := initializeTestConfig(t)

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"config",
			"rpc",
			"--config-path",
			path,
			"--unsafe-rpc",
			"random toggle",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, errInvalidToggleValue)
	})
}

func TestConfig_RPC_Valid(t *testing.T) {
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
				assert.Equal(t, value, loadedCfg.RPC.RootDir)
			},
		},
		{
			"listen address updated",
			[]string{
				"--listen-address",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.ListenAddress)
			},
		},
		{
			"CORS Allowed Origins updated",
			[]string{
				"--cors-allowed-origins",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, []string{value}, loadedCfg.RPC.CORSAllowedOrigins)
			},
		},
		{
			"CORS Allowed Methods updated",
			[]string{
				"--cors-allowed-methods",
				"POST",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, []string{value}, loadedCfg.RPC.CORSAllowedMethods)
			},
		},
		{
			"CORS Allowed Headers updated",
			[]string{
				"--cors-allowed-headers",
				"*",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, []string{value}, loadedCfg.RPC.CORSAllowedHeaders)
			},
		},
		{
			"GRPC listen address updated",
			[]string{
				"--grpc-listen-address",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.GRPCListenAddress)
			},
		},
		{
			"GRPC max open connections updated",
			[]string{
				"--grpc-max-open-connections",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.RPC.GRPCMaxOpenConnections))
			},
		},
		{
			"unsafe value updated",
			[]string{
				"--unsafe-rpc",
				onValue,
			},
			func(loadedCfg *config.Config, value string) {
				boolVal := false
				if value == onValue {
					boolVal = true
				}

				assert.Equal(t, boolVal, loadedCfg.RPC.Unsafe)
			},
		},
		{
			"RPC max open connections updated",
			[]string{
				"--rpc-max-open-connections",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.RPC.MaxOpenConnections))
			},
		},
		{
			"tx commit broadcast timeout updated",
			[]string{
				"--timeout-broadcast-commit",
				(time.Second * 10).String(),
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.TimeoutBroadcastTxCommit.String())
			},
		},
		{
			"max body bytes updated",
			[]string{
				"--max-body-bytes",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.RPC.MaxBodyBytes))
			},
		},
		{
			"max header bytes updated",
			[]string{
				"--max-header-bytes",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.RPC.MaxHeaderBytes))
			},
		},
		{
			"TLS cert file updated",
			[]string{
				"--tls-cert-file",
				"example path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.TLSCertFile)
			},
		},
		{
			"TLS key file updated",
			[]string{
				"--tls-key-file",
				"example path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.TLSKeyFile)
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
				"rpc",
				"--config-path",
				path,
			}

			// Create the command
			cmd := newRootCmd(commands.NewTestIO())
			args = append(args, testCase.flags...)

			// Run the command
			cmdErr := cmd.ParseAndRun(context.Background(), args)
			require.NoError(t, cmdErr)

			// Make sure config was updated
			loadedCfg, err := config.LoadConfigFile(path)
			require.NoError(t, err)

			testCase.verifyFn(loadedCfg, testCase.flags[len(testCase.flags)-1])
		})
	}
}
