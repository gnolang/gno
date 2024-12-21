package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGnolandIntegration tests the forking of a Gnoland node.
// XXX: For now keep this test sequential (no parallel execution is allowed).
func TestGnolandIntegration(t *testing.T) {
	// Set a timeout of 20 seconds for the test context.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	tmpdir := t.TempDir() // Create a temporary directory for the test.

	gnoRootDir := gnoenv.RootDir() // Get the root directory for Gnolang.

	// Define paths for the build directory and the gnoland binary.
	gnolandBuildDir := filepath.Join(tmpdir, "build")
	gnolandBin := filepath.Join(gnolandBuildDir, "gnoland")

	// Compile the gnoland binary.
	err := buildGnoland(t, gnoRootDir, gnolandBin)
	require.NoError(t, err)

	// Prepare a minimal node configuration for testing.
	cfg := TestingMinimalNodeConfig(gnoRootDir)
	remoteAddr, cmd, err := ExecuteNode(ctx, gnolandBin, &Config{
		Verbose:  true,
		RootDir:  gnoRootDir,
		TMConfig: cfg.TMConfig,
		Genesis:  NewMarshalableGenesisDoc(cfg.Genesis),
	}, os.Stderr)
	require.NoError(t, err)

	defer func() {
		// Ensure the process is killed after the test to clean up resources.
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	// Create a new HTTP client to interact with the integration node.
	cli, err := client.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Retreive node info.
	info, err := cli.ABCIInfo()
	require.NoError(t, err)
	assert.NotEmpty(t, info.Response.Data)
}
