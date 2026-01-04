package integration

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Define a flag to indicate whether to run the embedded command
var runCommand = flag.Bool("run-node-process", false, "execute the embedded command")

func TestMain(m *testing.M) {
	flag.Parse()

	// Check if the embedded command should be executed
	if !*runCommand {
		os.Exit(m.Run())
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if err := RunMain(ctx, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

// TestGnolandIntegration tests the forking of a Gnoland node.
func TestNodeProcess(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	gnoRootDir := gnoenv.RootDir()

	// Define paths for the build directory and the gnoland binary
	gnolandDBDir := filepath.Join(t.TempDir(), "db")

	// Prepare a minimal node configuration for testing
	cfg := TestingMinimalNodeConfig(gnoRootDir)

	var stdio bytes.Buffer
	defer func() {
		t.Log("node output:")
		t.Log(stdio.String())
	}()

	start := time.Now()
	node := runTestingNodeProcess(t, ctx, ProcessConfig{
		Stderr: &stdio, Stdout: &stdio,
		Node: &ProcessNodeConfig{
			Verbose:      true,
			ValidatorKey: ed25519.GenPrivKey(),
			DBDir:        gnolandDBDir,
			RootDir:      gnoRootDir,
			TMConfig:     cfg.TMConfig,
			Genesis:      NewMarshalableGenesisDoc(cfg.Genesis),
		},
	})
	t.Logf("time to start the node: %v", time.Since(start).String())

	// Create a new HTTP client to interact with the integration node
	cli, err := client.NewHTTPClient(node.Address())
	require.NoError(t, err)

	// Retrieve node info
	info, err := cli.ABCIInfo(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, info.Response.Data)

	// Attempt to stop the node
	err = node.Stop()
	require.NoError(t, err)

	// Attempt to stop the node a second time, should not fail
	err = node.Stop()
	require.NoError(t, err)
}

// TestGnolandIntegration tests the forking of a Gnoland node.
func TestInMemoryNodeProcess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	gnoRootDir := gnoenv.RootDir()

	// Define paths for the build directory and the gnoland binary
	gnolandDBDir := filepath.Join(t.TempDir(), "db")

	// Prepare a minimal node configuration for testing
	cfg := TestingMinimalNodeConfig(gnoRootDir)

	var stdio bytes.Buffer
	defer func() {
		t.Log("node output:")
		t.Log(stdio.String())
	}()

	start := time.Now()
	node, err := RunInMemoryProcess(ctx, ProcessConfig{
		Stderr: &stdio, Stdout: &stdio,
		Node: &ProcessNodeConfig{
			Verbose:      true,
			ValidatorKey: ed25519.GenPrivKey(),
			DBDir:        gnolandDBDir,
			RootDir:      gnoRootDir,
			TMConfig:     cfg.TMConfig,
			Genesis:      NewMarshalableGenesisDoc(cfg.Genesis),
		},
	})
	require.NoError(t, err)
	t.Logf("time to start the node: %v", time.Since(start).String())

	// Create a new HTTP client to interact with the integration node
	cli, err := client.NewHTTPClient(node.Address())
	require.NoError(t, err)

	// Retrieve node info
	info, err := cli.ABCIInfo(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, info.Response.Data)

	// Attempt to stop the node
	err = node.Stop()
	require.NoError(t, err)

	// Attempt to stop the node a second time, should not fail
	err = node.Stop()
	require.NoError(t, err)
}
