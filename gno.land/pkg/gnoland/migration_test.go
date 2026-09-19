package gnoland

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunInPlaceMigration_BackupExists(t *testing.T) {
	t.Parallel()

	// Create a fake data dir with an existing backup.
	dataDir := t.TempDir()
	dbDir := filepath.Join(dataDir, "db")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	// Simulate a backup already existing.
	bakPath := filepath.Join(dbDir, migBakName+".db")
	require.NoError(t, os.MkdirAll(bakPath, 0o755))

	cfg := MigrationConfig{
		DataRootDir: dataDir,
	}
	err := RunInPlaceMigration(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "migration backup already exists")
}

func TestRunInPlaceMigration_EmptyBlockStore(t *testing.T) {
	t.Parallel()

	// Create a fake data dir with no existing backup and an empty block store.
	dataDir := t.TempDir()
	dbDir := filepath.Join(dataDir, "db")
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	cfg := MigrationConfig{
		DataRootDir: dataDir,
		GenesisPath: "/does/not/exist/genesis.json",
	}
	err := RunInPlaceMigration(cfg)
	require.Error(t, err)
	// It should fail at opening the block store DB (no DB files) or because
	// the block store is empty – either is fine for this guard check.
}
