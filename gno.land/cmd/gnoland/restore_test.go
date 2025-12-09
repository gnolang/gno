package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/backup/v1"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRestore(t *testing.T) {
	backupDir := t.TempDir()
	generateBackup(t, backupDir, 2)

	restoreDir := t.TempDir()
	io := commands.NewTestIO()
	io.SetOut(os.Stdout)
	io.SetErr(os.Stderr)
	err := newRestoreCmd(io).ParseAndRun(context.Background(), []string{
		"--data-dir", filepath.Join(restoreDir, "chain-data"),
		"--backup-dir", backupDir,
		"--genesis", filepath.Join(backupDir, "genesis.json"),
		"--skip-genesis-sig-verification", "true",
		"--lazy",
	})
	require.NoError(t, err)
}

func generateBackup(t *testing.T, backupDir string, height int64) {
	t.Helper()

	// XXX: consider moving this into a gen command and commit a golden backup

	nodeDir := t.TempDir()

	io := commands.NewTestIO()
	io.SetOut(os.Stdout)
	io.SetErr(os.Stderr)

	cfg := &nodeCfg{}
	fs := flag.NewFlagSet("", flag.PanicOnError)
	cfg.RegisterFlags(fs)
	require.NoError(t, fs.Parse([]string{
		"--lazy",
		"--data-dir", nodeDir,
		"--genesis", filepath.Join(backupDir, "genesis.json"),
		"--skip-genesis-sig-verification", "true",
	}))

	node, err := createNode(context.Background(), cfg, io)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, node.Config().LocalApp.Close())
	})

	require.NoError(t, node.Start())
	for node.BlockStore().Height() < height {
		time.Sleep(1 * time.Second)
	}
	require.NoError(t, node.Stop())

	err = backup.WithWriter(backupDir, 0, height, zap.NewNop(), func(startHeight int64, write backup.Writer) error {
		for i := startHeight; i <= height; i++ {
			require.NoError(t, write(node.BlockStore().LoadBlock(i)))
		}
		return nil
	})
	require.NoError(t, err)
}
