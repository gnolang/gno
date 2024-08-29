package gnoenv

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGuessGnoRootDir_WithSetGnoRoot(t *testing.T) {
	originalGnoRoot := _GNOROOT
	defer func() { _GNOROOT = originalGnoRoot }() // Restore after test

	t.Setenv("GNOROOT", "")

	const testPath = "/path/to/gnoRoot"

	_GNOROOT = testPath
	root, err := GuessRootDir()
	require.NoError(t, err)
	require.Equal(t, testPath, root)
}

func TestGuessGnoRootDir_UsingCallerStack(t *testing.T) {
	originalGnoRoot := _GNOROOT
	defer func() { _GNOROOT = originalGnoRoot }()

	// Unset PATH should prevent InferGnoRootFromGoMod to works
	t.Setenv("GNOROOT", "")
	t.Setenv("PATH", "")

	_, err := exec.LookPath("go")
	require.Error(t, err)

	// gno/ .. /gnovm/ .. /pkg/ .. /gnoenv/gnoroot.go
	testPath, _ := filepath.Abs(filepath.Join(".", "..", "..", ".."))
	root, err := GuessRootDir()
	require.NoError(t, err)
	require.Equal(t, testPath, root)
}

func TestGuessGnoRootDir_Error(t *testing.T) {
	// XXX: Determine a method to test the GuessGnoRoot final error.
	// One approach might be to use `txtar` to build a test binary with -trimpath,
	// avoiding absolute paths in the call stack.
	t.Skip("not implemented; refer to the inline comment for more details.")
}

func TestGuessGnoRootDir_WithGoModList(t *testing.T) {
	// XXX: find a way to test `go mod list` phase.
	// One solution is to use txtar with embed go.mod file.
	// For now only `inferGnoRootFromGoMod` is tested.
	t.Skip("not implemented; refer to the inline comment for more details.")
}

func TestInferGnoRootFromGoMod(t *testing.T) {
	// gno/ .. /gnovm/ .. /pkg/ .. /gnoenv/gnoroot.go
	testPath, _ := filepath.Abs(filepath.Join(".", "..", "..", ".."))

	t.Run("go is present", func(t *testing.T) {
		root, err := inferRootFromGoMod()
		require.NoError(t, err)
		require.Equal(t, testPath, root)
	})

	t.Run("go is not present", func(t *testing.T) {
		// Unset PATH should prevent `inferGnoRootFromGoMod` to works
		t.Setenv("PATH", "")

		root, err := inferRootFromGoMod()
		require.Error(t, err)
		require.Empty(t, root)
	})
}
