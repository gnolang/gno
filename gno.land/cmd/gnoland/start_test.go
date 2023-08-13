package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	os.Chdir(filepath.Join("..", "..")) // go to repo's root dir


	mockOut := bytes.NewBufferString("")
	mockErr := bytes.NewBufferString("")
	io := commands.NewTestIO()
	io.SetOut(commands.WriteNopCloser(mockOut))
	io.SetErr(commands.WriteNopCloser(mockErr))
	cmd := newRootCmd(io)

	args := []string{"init", "--skip-failing-genesis-txs"}
	t.Logf(`Running "gnoland %s"`, strings.Join(args, " "))
	err := cmd.ParseAndRun(context.Background(), args)
	require.NoError(t, err)

	stdout := mockOut.String()
	stderr := mockErr.String()

	require.Contains(t, stderr, "Node created.")
	require.NotContains(t, stdout, "panic:")
}

// TODO: test various configuration files?
