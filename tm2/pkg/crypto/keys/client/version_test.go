package client

import (
	"testing"
	"context"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/version"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/require"
)

func TestVersionApp(t *testing.T) {
	t.Parallel()

	// make new test dir
	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
	defer kbCleanUp()

	// Set current home
	cfg := &BaseCfg{
		BaseOptions: BaseOptions{
			Home: kbHome,
		},
	}

	versionCmd := NewVersionCmd(cfg, commands.NewTestIO())
	versionValues := []string{"chain/test4.2", "develop", "master"}

	originalVersion := version.Version

	t.Cleanup(func() {
		version.Version = originalVersion
	})

	{
		// test: original version
		err := versionCmd.ParseAndRun(context.Background(), []string{})
		require.NoError(t, err)
	}

	{
		// test: version settled
		version.Version = versionValues[0]
		err := versionCmd.ParseAndRun(context.Background(), []string{})
		require.NoError(t, err)
	}
}
