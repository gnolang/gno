package main

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartInitialize(t *testing.T) {
	t.Parallel()

	// NOTE: cannot be txtar tests as they use their own parsing for the
	// "gnoland" command line. See pkg/integration.

	var (
		nodeDir     = t.TempDir()
		genesisFile = filepath.Join(nodeDir, "test_genesis.json")

		args = []string{
			"start",
			"--skip-start",
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
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	cmd := newRootCmd(io)
	require.NoError(t, cmd.ParseAndRun(ctx, args))

	// Make sure the directory is created
	assert.DirExists(t, nodeDir)
	assert.FileExists(t, genesisFile)
}
