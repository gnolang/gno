package gnoenv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jaekwon/testify/require"
)

func TestFixOldDefaultGnoHome(t *testing.T) {
	tempDir := t.TempDir()

	oldGnoHome := filepath.Join(tempDir, ".gno")
	newGnoHome := filepath.Join(tempDir, "gno")

	// Create a dummy old GNO_HOME
	os.Mkdir(oldGnoHome, 0o755)

	// Test migration
	fixOldDefaultGnoHome(newGnoHome)

	_, errOld := os.Stat(oldGnoHome)
	_, errNew := os.Stat(newGnoHome)
	require.True(t, os.IsNotExist(errOld))
	require.NoError(t, errNew)
}
