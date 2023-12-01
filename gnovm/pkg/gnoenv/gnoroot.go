package gnoenv

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var ErrUnableToGuessGnoRoot = errors.New("gno was unable to determine GNOROOT. Please set the GNOROOT environment variable")

// Can be set manually at build time using:
// -ldflags="-X github.com/gnolang/gno/gnovm/pkg/gnoenv._GNOROOT"
var _GNOROOT string

// RootDir guesses the Gno root directory and panics if it fails.
func RootDir() string {
	root, err := GuessRootDir()
	if err != nil {
		panic(err)
	}

	return root
}

var muGnoRoot sync.Mutex

// GuessRootDir attempts to determine the Gno root directory using various strategies:
// 1. First, It tries to obtain it from the `GNOROOT` environment variable.
// 2. If the env variable isn't set, It checks if `_GNOROOT` has been previously determined or set with -ldflags.
// 3. If not, it uses the `go list` command to infer from go.mod.
// 4. As a last resort, it determines `GNOROOT` based on the caller stack's file path.
func GuessRootDir() (string, error) {
	muGnoRoot.Lock()
	defer muGnoRoot.Unlock()

	// First try to get the root directory from the `GNOROOT` environment variable.
	if rootdir := os.Getenv("GNOROOT"); rootdir != "" {
		return strings.TrimSpace(rootdir), nil
	}

	var err error
	if _GNOROOT == "" {
		// Try to guess `GNOROOT` using various strategies
		_GNOROOT, err = guessRootDir()
	}

	return _GNOROOT, err
}

func guessRootDir() (string, error) {
	// Attempt to guess `GNOROOT` from go.mod by using the `go list` command.
	if rootdir, err := inferRootFromGoMod(); err == nil {
		return filepath.Clean(rootdir), nil
	}

	// If the above method fails, ultimately try to determine `GNOROOT` based
	// on the caller stack's file path.
	// Path need to be absolute here, that mostly mean that if `-trimpath`
	// as been passed this method will not works.
	if _, filename, _, ok := runtime.Caller(1); ok && filepath.IsAbs(filename) {
		if currentDir := filepath.Dir(filename); currentDir != "" {
			// Deduce Gno root directory relative from the current file's path.
			// gno/ .. /gnovm/ .. /pkg/ .. /gnoenv/gnoenv.go
			rootdir, err := filepath.Abs(filepath.Join(currentDir, "..", "..", ".."))
			if err == nil {
				return rootdir, nil
			}
		}
	}

	return "", ErrUnableToGuessGnoRoot
}

func inferRootFromGoMod() (string, error) {
	gobin, err := exec.LookPath("go")
	if err != nil {
		return "", fmt.Errorf("unable to find `go` binary: %w", err)
	}

	cmd := exec.Command(gobin, "list", "-m", "-mod=mod", "-f", "{{.Dir}}", "github.com/gnolang/gno")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("unable to infer GnoRoot from go.mod: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}
