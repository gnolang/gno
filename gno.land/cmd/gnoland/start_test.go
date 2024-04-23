package main

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartInitialize(t *testing.T) {
	t.Parallel()

	var (
		nodeDir = t.TempDir()

		args = []string{
			"start",
			"--skip-start",
			"--skip-failing-genesis-txs",
			"--data-dir",
			nodeDir,
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
}
