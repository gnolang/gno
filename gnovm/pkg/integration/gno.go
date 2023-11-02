package integration

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/rogpeppe/go-internal/testscript"
)

// SetupGno prepares the given testscript environment for tests that utilize the gno command.
// If the `gno` binary doesn't exist, it's built using the `go build` command into the specified buildDir.
// The function also include the `gno` command into `p.Cmds` to and wrap environment into p.Setup
// to correctly set up the environment variables needed for the `gno` command.
func SetupGno(p *testscript.Params, buildDir string) error {
	// Try to fetch `GNOROOT` from the environment variables
	gnoroot := os.Getenv("GNOROOT")
	if gnoroot == "" {
		// If `GNOROOT` isn't set, determine the root directory of github.com/gnolang/gno
		goModPath, err := exec.Command("go", "env", "GOMOD").CombinedOutput()
		if err != nil {
			return fmt.Errorf("unable to determine gno root directory")
		}

		gnoroot = filepath.Dir(string(goModPath))
	}

	if !osm.DirExists(buildDir) {
		return fmt.Errorf("%q does not exist or is not a directory", buildDir)
	}

	// Determine the path to the gno binary within the build directory
	gnoBin := filepath.Join(buildDir, "gno")
	if _, err := os.Stat(gnoBin); errors.Is(err, os.ErrNotExist) {
		// Build a fresh gno binary in a temp directory
		gnoArgsBuilder := []string{"build", "-o", gnoBin}

		// Forward `-covermode` settings if set
		if coverMode := testing.CoverMode(); coverMode != "" {
			gnoArgsBuilder = append(gnoArgsBuilder, "-covermode", coverMode)
		}

		// Append the path to the gno command source
		gnoArgsBuilder = append(gnoArgsBuilder, filepath.Join(gnoroot, "gnovm", "cmd", "gno"))

		if err = exec.Command("go", gnoArgsBuilder...).Run(); err != nil {
			return fmt.Errorf("unable to build gno binary: %w", err)
		}
	} else if err != nil {
		// Handle other potential errors from os.Stat
		return err
	}

	// Store the original setup scripts for potential wrapping
	origSetup := p.Setup
	p.Setup = func(env *testscript.Env) error {
		// If there's an original setup, execute it
		if origSetup != nil {
			if err := origSetup(env); err != nil {
				return err
			}
		}

		// Set the GNOROOT environment variable
		env.Setenv("GNOROOT", gnoroot)

		// Create a temporary home directory because certain commands require access to $HOME/.cache/go-build
		home, err := os.MkdirTemp("", "gno")
		if err != nil {
			return fmt.Errorf("unable to create temporary home directory: %w", err)
		}
		env.Setenv("HOME", home)
		env.Defer(func() { os.RemoveAll(home) })

		return nil
	}

	// Initialize cmds map if needed
	if p.Cmds == nil {
		p.Cmds = make(map[string]func(ts *testscript.TestScript, neg bool, args []string))
	}

	// Register the gno command for testscripts
	p.Cmds["gno"] = func(ts *testscript.TestScript, neg bool, args []string) {
		err := ts.Exec(gnoBin, args...)
		if err != nil {
			ts.Logf("[%v]\n", err)
			if !neg {
				ts.Fatalf("unexpected gno command failure")
			}
		} else {
			if neg {
				ts.Fatalf("unexpected gno command success")
			}
		}
	}

	return nil
}
