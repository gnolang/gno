package gnoenv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFixOldDefaultGnoHome(t *testing.T) {
	tempHomeDir := t.TempDir()
	t.Setenv("HOME", tempHomeDir)

	oldGnoHome := filepath.Join(tempHomeDir, ".gno")
	newGnoHome := filepath.Join(tempHomeDir, "gno")

	// Create a dummy old GNO_HOME
	os.Mkdir(oldGnoHome, 0o755)

	// Test migration
	fixOldDefaultGnoHome(newGnoHome)

	_, errOld := os.Stat(oldGnoHome)
	require.NotNil(t, errOld)
	_, errNew := os.Stat(newGnoHome)
	require.True(t, os.IsNotExist(errOld), "invalid errors", errOld)
	require.NoError(t, errNew)
}
