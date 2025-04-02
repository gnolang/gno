package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/require"
)

func TestRestore(t *testing.T) {
	tmpDir := t.TempDir()
	io := commands.NewTestIO()
	io.SetOut(os.Stdout)
	io.SetErr(os.Stderr)
	err := newRestoreCmd(io).ParseAndRun(context.Background(), []string{
		"--data-dir", filepath.Join(tmpDir, "chain-data"),
		"--backup-dir", filepath.FromSlash("testdata/backup-2blocks"),
		"--genesis", filepath.FromSlash("testdata/backup-2blocks/genesis.json"),
		"--lazy",
	})
	require.NoError(t, err)
}
