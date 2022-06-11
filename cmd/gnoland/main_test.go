package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/pkgs/testutils"
	"github.com/stretchr/testify/require"
)

func TestInitialize(t *testing.T) {
	os.Chdir(filepath.Join("..", "..")) // go to repo's root dir

	closer := testutils.CaptureStdoutAndStderr()

	err := runMain([]string{"--skip-failing-genesis-txs", "--skip-start"})
	stdouterr, bufErr := closer()
	require.NoError(t, bufErr)
	require.NoError(t, err)

	_ = stdouterr
	require.Contains(t, stdouterr, "Node created.", "failed to create node")
	require.Contains(t, stdouterr, "'--skip-start' is set. Exiting.", "not exited with skip-start")
}

// TODO: test various configuration files?
