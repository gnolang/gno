package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

// retryUntilTimeout runs the callback until the timeout is exceeded, or
// the callback returns a flag indicating completion
func retryUntilTimeout(ctx context.Context, cb func() bool) error {
	ch := make(chan error, 1)

	go func() {
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				ch <- ctx.Err()

				return
			default:
				retry := cb()
				if !retry {
					ch <- nil
					return
				}
			}

			time.Sleep(500 * time.Millisecond)
		}
	}()

	return <-ch
}

// prepareNodeRPC sets the RPC listen address for the node to be an arbitrary
// free address. Setting the listen port to a free port on the machine avoids
// node collisions between different testing suites
func prepareNodeRPC(t *testing.T, nodeDir string) {
	t.Helper()

	args := []string{
		"config",
		"init",
		"--home",
		nodeDir,
	}

	// Prepare the IO
	mockOut := new(bytes.Buffer)
	mockErr := new(bytes.Buffer)
	io := commands.NewTestIO()
	io.SetOut(commands.WriteNopCloser(mockOut))
	io.SetErr(commands.WriteNopCloser(mockErr))

	// Prepare the cmd context
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	// Run config init
	require.NoError(t, newRootCmd(io).ParseAndRun(ctx, args))

	args = []string{
		"config",
		"set",
		"--home",
		nodeDir,
		"rpc.laddr",
		"tcp://0.0.0.0:0",
	}

	// Run config set
	require.NoError(t, newRootCmd(io).ParseAndRun(ctx, args))
}

func TestStart_Lazy(t *testing.T) {
	t.Parallel()

	homeDir := newTestHomeDirectory(t, t.TempDir())
	homeDir.genesisFile = homeDir.Path() + "test_genesis.json"

	// Prepare the config
	prepareNodeRPC(t, homeDir.Path())

	args := []string{
		"start",
		"--lazy",
		"--skip-failing-genesis-txs",

		// These two flags are tested together as they would otherwise
		// pollute this directory (cmd/gnoland) if not set.
		"--home",
		homeDir.Path(),
		"--genesis",
		homeDir.GenesisFilePath(),
	}

	// Prepare the IO
	mockOut := new(bytes.Buffer)
	mockErr := new(bytes.Buffer)
	io := commands.NewTestIO()
	io.SetOut(commands.WriteNopCloser(mockOut))
	io.SetErr(commands.WriteNopCloser(mockErr))

	// Create and run the command
	ctx, cancelFn := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelFn()

	// Set up the command ctx
	g, gCtx := errgroup.WithContext(ctx)

	// Start the node
	g.Go(func() error {
		return newRootCmd(io).ParseAndRun(gCtx, args)
	})

	// Set up the retry ctx
	retryCtx, retryCtxCancelFn := context.WithTimeout(ctx, 5*time.Second)
	defer retryCtxCancelFn()

	// This is a very janky way to verify the node has started.
	// The alternative is to poll the node's RPC endpoints, but for some reason
	// this introduces a lot of flakyness to the testing suite -- shocking!
	// In an effort to keep this simple, and avoid randomly failing tests,
	// we query the CLI output of the command
	require.NoError(t, retryUntilTimeout(retryCtx, func() bool {
		return !strings.Contains(mockOut.String(), startGraphic)
	}))

	cancelFn() // stop the node
	require.NoError(t, g.Wait())

	// Make sure the genesis is generated
	assert.FileExists(t, homeDir.GenesisFilePath())

	// Make sure the config is generated (default)
	assert.FileExists(t, homeDir.ConfigFile())

	// Make sure the secrets are generated
	assert.DirExists(t, homeDir.SecretsDir())
	assert.FileExists(t, homeDir.SecretsValidatorKey())
	assert.FileExists(t, homeDir.SecretsValidatorState())
	assert.FileExists(t, homeDir.SecretsValidatorState())
}
