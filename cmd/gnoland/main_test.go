package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/pkgs/testutils"
	"github.com/stretchr/testify/require"
)

func TestInitialize(t *testing.T) {
	cases := []struct {
		args []string
	}{
		// {[]string{"--skip-start", "--skip-failing-gensis-txs"}},
		{[]string{"--skip-start"}},
		// FIXME: test seems flappy as soon as we have multiple cases.
	}

	os.Chdir(filepath.Join("..", "..")) // go to repo's root dir

	for _, tc := range cases {
		name := strings.Join(tc.args, " ")
		t.Run(name, func(t *testing.T) {
			closer := testutils.CaptureStdoutAndStderr()

			err := runMain([]string{"--skip-failing-genesis-txs", "--skip-start"})
			stdouterr, bufErr := closer()
			require.NoError(t, bufErr)
			require.NoError(t, err)

			_ = stdouterr
			require.Contains(t, stdouterr, "Node created.", "failed to create node")
			require.Contains(t, stdouterr, "'--skip-start' is set. Exiting.", "not exited with skip-start")
			require.NotContains(t, stdouterr, "panic:")
		})
	}
}

// TODO: test various configuration files?
