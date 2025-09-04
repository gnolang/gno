package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

// prepareNodeRPC sets the RPC listen address for the node to be an arbitrary
// free address. Setting the listen port to a free port on the machine avoids
// node collisions between different testing suites
func prepareNodeRPC(t *testing.T, nodeDir string, addr string) {
	t.Helper()

	path := constructConfigPath(nodeDir)
	args := []string{
		"config",
		"init",
		"--config-path",
		path,
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

	args = []string{"config", "set",
		"--config-path", path,
		"rpc.laddr", addr,
	}

	// Run config set
	require.NoError(t, newRootCmd(io).ParseAndRun(ctx, args))
}

func TestStart_Lazy(t *testing.T) {
	// Running a full node is cpu consuming
	// Do run this one in parallel
	// t.Parallel()

	// We allow one minute by node lifespan
	const maxTestDeadline = time.Minute

	shortTempDir := func(t *testing.T) string {
		t.Helper()

		dir, err := os.MkdirTemp("/tmp", "socktest-*")
		require.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(dir) })
		return dir
	}

	tests := []struct {
		name           string
		additionalArgs []string
	}{
		{
			name:           "with skip-failing-genesis-txs",
			additionalArgs: []string{"--skip-failing-genesis-txs"},
		},
		{
			name:           "with skip-genesis-sig-verification",
			additionalArgs: []string{"--skip-genesis-sig-verification"},
		},
		{
			name:           "with 2 skips",
			additionalArgs: []string{"--skip-genesis-sig-verification", "--skip-failing-genesis-txs"},
		},
		// XXX: {name: "no args", additionalArgs: []string{}}, // not compatible with current genesis.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Generate temp socket filepath for listening.
			// (Use short path to avoid > 120 char socket path).
			sockFile := filepath.Join(shortTempDir(t), "rpc.sock")
			sockAddr := fmt.Sprintf("unix://%s", sockFile)

			var (
				nodeDir     = t.TempDir()
				genesisFile = filepath.Join(nodeDir, "test_genesis.json")
			)

			// Prepare the config
			prepareNodeRPC(t, nodeDir, sockAddr)

			args := []string{
				"start",
				"--lazy",
			}

			// Add additional args
			args = append(args, tc.additionalArgs...)

			args = append(args,
				// These two flags are tested together as they would otherwise
				// pollute this directory (cmd/gnoland) if not set.
				"--data-dir",
				nodeDir,
				"--genesis",
				genesisFile,
			)

			// Prepare the IO
			mockOut := new(bytes.Buffer)
			mockErr := new(bytes.Buffer)
			io := commands.NewTestIO()
			io.SetOut(commands.WriteNopCloser(mockOut))
			io.SetErr(commands.WriteNopCloser(mockErr))

			// Create and run the command
			deadline := time.Now().Add(maxTestDeadline)
			ctx, cancelFn := context.WithDeadline(context.Background(), deadline)
			defer cancelFn()

			// Set up the command ctx
			g, gCtx := errgroup.WithContext(ctx)

			// Start the node
			g.Go(func() error {
				defer cancelFn()
				return newRootCmd(io).ParseAndRun(gCtx, args)
			})

			t.Logf("node: check for ascii graphic to show up - time left %s", time.Until(deadline))

			// Check that starting ascii graphic display
			require.Eventuallyf(t, func() bool {
				return strings.Contains(mockOut.String(), startGraphic)
			}, time.Until(deadline), time.Millisecond*500, "node: ascii graphic never show up")

			cli, err := client.NewHTTPClient(sockAddr)
			require.NoError(t, err)

			t.Logf("rpc: get node infos - time left %s", time.Until(deadline))

			// Check that rpc endpoint is correctly listening on our socket
			require.EventuallyWithT(t, func(c *assert.CollectT) {
				info, qerr := cli.ABCIInfo(gCtx)
				require.NoError(c, qerr)
				require.NoError(c, info.Response.Error)
			}, time.Until(deadline), time.Millisecond*500, "rpc: unable get node infos")

			t.Logf("rpc: query vm/qpaths - time left %s", time.Until(deadline))

			// Check the node as fully loaded by checking rpc qpaths endpoint
			require.EventuallyWithT(t, func(c *assert.CollectT) {
				qres, qerr := cli.ABCIQuery(gCtx, "vm/qpaths", []byte("gno.land"))
				require.NoError(c, qerr)
				require.NoError(c, qres.Response.Error)
				paths := strings.Split(string(qres.Response.Data), "\n")
				require.Greater(c, len(paths), 1, "query qpaths: no package has been loaded")
			}, time.Until(deadline), time.Millisecond*500, "rpc: unable to call rpc vm/qpaths")

			t.Logf("node: stopping - time left %s", time.Until(deadline))

			cancelFn() // stop the node
			require.NoError(t, g.Wait())

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
		})
	}
}
