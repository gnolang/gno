package gnoenv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jaekwon/testify/require"
)

func TestHomeDir(t *testing.T) {
	t.Run("use GNOHOME if set", func(t *testing.T) {
		// Backup any related environment variables
		tBackupEnvironement(t, "GNOHOME", "GNO_HOME")

		expected := "/test/gno_home"
		os.Setenv("GNOHOME", expected)
		require.Equal(t, expected, HomeDir())
	})

	t.Run("fallback to GNO_HOME if set", func(t *testing.T) {
		// Backup any related environment variables
		tBackupEnvironement(t, "GNOHOME", "GNO_HOME")
		t.Log("`GNO_HOME` is deprecated")

		expected := "/test/gnohome"
		os.Setenv("GNO_HOME", expected)
		require.Equal(t, expected, HomeDir())
	})

	t.Run("use UserConfigDir with gno", func(t *testing.T) {
		// Backup any related environment variables
		tBackupEnvironement(t, "GNOHOME", "GNO_HOME")

		dir, err := os.UserConfigDir()
		require.NoError(t, err)
		expected := filepath.Join(dir, "gno")
		require.Equal(t, expected, HomeDir())
	})
}
