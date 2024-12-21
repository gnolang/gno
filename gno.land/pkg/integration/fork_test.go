package integration

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForkGnoland(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	tmpdir := t.TempDir()

	gnoRootDir := gnoenv.RootDir()

	gnolandBuildDir := filepath.Join(tmpdir, "build")
	gnolandBin := filepath.Join(gnolandBuildDir, "gnoland")

	err := buildGnoland(t, gnoRootDir, gnolandBin)
	require.NoError(t, err)

	cfg := TestingMinimalNodeConfig(gnoRootDir)
	remoteAddr, cmd, err := ExecuteForkBinary(ctx, gnolandBin, &ForkConfig{
		Verbose:  true,
		RootDir:  gnoRootDir,
		TMConfig: cfg.TMConfig,
		Genesis:  NewMarshalableGenesisDoc(cfg.Genesis),
	})
	require.NoError(t, err)

	defer cmd.Process.Kill()

	cli, err := client.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	info, err := cli.ABCIInfo()
	require.NoError(t, err)
	assert.NotEmpty(t, info.Response.Data)
}
