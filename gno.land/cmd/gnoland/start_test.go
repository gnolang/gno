package main

import (
	"bytes"
	"context"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

// isNodeUp verifies the node at the given JSON-RPC URL is serving requests
func isNodeUp(ctx context.Context, url string) error {
	var (
		client = &http.Client{
			Timeout: 5 * time.Second,
		}

		statusURL = url + "/status"
	)

	cb := func() bool {
		resp, err := client.Get(statusURL)
		if err != nil {
			return true
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return false
		}

		return true
	}

	return retryUntilTimeout(ctx, cb)
}

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

func TestStart_Lazy(t *testing.T) {
	var (
		nodeDir     = t.TempDir()
		genesisFile = filepath.Join(nodeDir, "test_genesis.json")

		args = []string{
			"start",
			"--lazy",
			"--skip-failing-genesis-txs",

			// These two flags are tested together as they would otherwise
			// pollute this directory (cmd/gnoland) if not set.
			"--data-dir",
			nodeDir,
			"--genesis",
			genesisFile,
		}
	)

	// Prepare the IO
	mockOut := new(bytes.Buffer)
	mockErr := new(bytes.Buffer)
	io := commands.NewTestIO()
	io.SetOut(commands.WriteNopCloser(mockOut))
	io.SetErr(commands.WriteNopCloser(mockErr))

	// Create and run the command
	ctx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFn()

	g, gCtx := errgroup.WithContext(ctx)

	// Start the node
	g.Go(func() error {
		cmd := newRootCmd(io)

		return cmd.ParseAndRun(gCtx, args)
	})

	// Start the JSON-RPC poll service
	pollCtx, pollCancelFn := context.WithTimeout(ctx, 5*time.Second)
	defer pollCancelFn()

	nodeUpErr := isNodeUp(pollCtx, "http://127.0.0.1:26657")
	cancelFn() // stop the node

	require.NoError(t, g.Wait())
	require.NoError(t, nodeUpErr)

	// Make sure the genesis is generated
	assert.FileExists(t, genesisFile)

	// Make sure the config is generated (default)
	assert.FileExists(t, constructConfigPath(nodeDir))

	// Make sure the secrets are generated
	var (
		secretsPath        = constructSecretsPath(nodeDir)
		validatorKeyPath   = filepath.Join(secretsPath, defaultValidatorKeyName)
		validatorStatePath = filepath.Join(secretsPath, defaultValidatorStateName)
		nodeKeyPath        = filepath.Join(secretsPath, defaultNodeKeyName)
	)

	assert.DirExists(t, secretsPath)
	assert.FileExists(t, validatorKeyPath)
	assert.FileExists(t, validatorStatePath)
	assert.FileExists(t, nodeKeyPath)
}
