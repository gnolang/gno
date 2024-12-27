package integration

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGnolandIntegration tests the forking of a Gnoland node.
func TestNodeProcess(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	gnoRootDir := gnoenv.RootDir()

	// Define paths for the build directory and the gnoland binary.
	gnolandDBDir := filepath.Join(t.TempDir(), "db")

	// Compile the gnoland binary.
	gnolandBin := buildGnoland(t, gnoRootDir)

	// Prepare a minimal node configuration for testing.
	cfg := TestingMinimalNodeConfig(gnoRootDir)

	var stdio bytes.Buffer
	defer func() {
		t.Log("node ouput:")
		t.Log(stdio.String())
	}()

	node, err := RunNodeProcess(ctx, gnolandBin, ProcessConfig{
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

	// Create a new HTTP client to interact with the integration node.
	cli, err := client.NewHTTPClient(node.Address)
	require.NoError(t, err)

	// Retreive node info.
	info, err := cli.ABCIInfo()
	require.NoError(t, err)
	assert.NotEmpty(t, info.Response.Data)

	// Attempt to stop the node
	err = node.Stop()
	require.NoError(t, err)

	// Attempt to stop the node a second time, should not fail
	err = node.Stop()
	require.NoError(t, err)
}
