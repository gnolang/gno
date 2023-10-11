package gnoroot

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jaekwon/testify/require"
)

func TestGuessGnoRootDir_WithSetGnoRoot(t *testing.T) {
	originalGnoRoot := _GNOROOT
	defer func() { _GNOROOT = originalGnoRoot }() // Restore after test

	restoreGnoroot := tBackupEnvironement("GNOROOT")
	defer restoreGnoroot()

	const testPath = "/path/to/gnoRoot"

	_GNOROOT = testPath
	root, err := GuessGnoRootDir()
	require.NoError(t, err)
	require.Equal(t, root, testPath)
}

func TestGuessGnoRootDir_UsingCallerStack(t *testing.T) {
	originalGnoRoot := _GNOROOT
	defer func() { _GNOROOT = originalGnoRoot }()

	restoreGnoroot := tBackupEnvironement("GNOROOT")
	defer restoreGnoroot()

	// Should prevent GuessGnoRootDir to find go binary
	restorePath := tBackupEnvironement("PATH")
	defer restorePath()

	_, err := exec.LookPath("go")
	require.Error(t, err)

	// gno/ .. /gnovm/ .. /pkg/ .. /gnoroot/gnoroot.go
	testPath, _ := filepath.Abs(filepath.Join(".", "..", "..", ".."))
	root, err := GuessGnoRootDir()
	require.NoError(t, err)
	require.Equal(t, root, testPath)
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
	// gno/ .. /gnovm/ .. /pkg/ .. /gnoroot/gnoroot.go
	testPath, _ := filepath.Abs(filepath.Join(".", "..", "..", ".."))

	root, err := inferGnoRootFromGoMod()
	require.NoError(t, err)
	require.Equal(t, root, testPath)

	restorePath := tBackupEnvironement("PATH")
	defer restorePath()

	root, err = inferGnoRootFromGoMod()
	require.Error(t, err)
	require.Empty(t, root)

}

func tBackupEnvironement(key string) (restore func()) {
	value := os.Getenv(key)
	os.Unsetenv(key)
	return func() { os.Setenv(key, value) }
}
