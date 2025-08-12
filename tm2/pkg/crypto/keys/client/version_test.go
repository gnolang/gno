package client

import (
	"testing"
	"context"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/version"
	"github.com/stretchr/testify/require"
)

func TestClient_Version(t *testing.T) {
	t.Parallel()

	// Set current home
	cfg := &BaseCfg{}

	// Initialize version command
	versionCmd := NewVersionCmd(cfg, commands.NewTestIO())
	versionValues := []string{"chain/test4.2", "develop", "master"}

	// test: original version
	err := versionCmd.ParseAndRun(context.Background(), []string{})
	require.NoError(t, err)

	// test: version settled
	version.Version = versionValues[0]
	err = versionCmd.ParseAndRun(context.Background(), []string{})
	require.NoError(t, err)
}
