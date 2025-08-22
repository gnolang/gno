package common

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistersDontPanic(t *testing.T) {
	t.Parallel()

	if os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("Skipping until https://github.com/gnolang/gno/issues/4561 is resolved")
	}

	t.Run("auth flags", func(t *testing.T) {
		t.Parallel()

		assert.NotPanics(t, func() {
			authFlags := &AuthFlags{}
			authFlags.RegisterFlags(&flag.FlagSet{})
		})
	})

	t.Run("server flags", func(t *testing.T) {
		t.Parallel()

		if _, err := os.Getwd(); err == nil {
			assert.NotPanics(t, func() {
				serverFlags := &ServerFlags{}
				serverFlags.RegisterFlags(&flag.FlagSet{})
			})
		}
	})
}

func TestDefaultAuthKeysFile(t *testing.T) {
	t.Parallel()

	if os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("Skipping until https://github.com/gnolang/gno/issues/4561 is resolved")
	}

	t.Run("valid context", func(t *testing.T) {
		t.Parallel()

		assert.NotEmpty(t, defaultAuthKeysFile())
	})

	t.Run("no user directory", func(t *testing.T) {
		t.Parallel()

		// Reset the user directory environment variables.
		os.Setenv("XDG_CONFIG_HOME", "")
		os.Setenv("HOME", "")

		// Should fallback to the current directory.
		if wd, err := os.Getwd(); err == nil {
			assert.Contains(t, defaultAuthKeysFile(), wd)
		}
	})

	t.Run("no user directory and working dir", func(t *testing.T) {
		t.Parallel()

		// Reset the user directory environment variables.
		os.Setenv("XDG_CONFIG_HOME", "")
		os.Setenv("HOME", "")

		// Should fallback to the current directory.
		wd, err := os.Getwd()
		require.NoError(t, err)
		require.Contains(t, defaultAuthKeysFile(), wd)

		// Create a faulty working directory.
		workingDir := filepath.Join(t.TempDir(), "faulty")
		require.NoError(t, os.MkdirAll(workingDir, 0o700))
		require.NoError(t, os.Chdir(workingDir))
		require.NoError(t, os.Chmod(workingDir, 0o000))

		// Unset the PWD from environment.
		os.Setenv("PWD", "")

		// This should panic because the working directory is faulty.
		require.Panics(t, func() { defaultAuthKeysFile() })

		// Reset the working directory permission for cleanup.
		assert.NoError(t, os.Chmod(workingDir, 0o700))
	})
}
