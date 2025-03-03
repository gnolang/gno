package integration

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/rogpeppe/go-internal/testscript"
)

// SetupGno prepares the given testscript environment for tests that utilize the gno command.
// If the `gno` binary doesn't exist, it's built using the `go build` command into the specified buildDir.
// The function also include the `gno` command into `p.Cmds` to and wrap environment into p.Setup
// to correctly set up the environment variables needed for the `gno` command.
func SetupGno(p *testscript.Params, homeDir, buildDir string) error {
	// Try to fetch `GNOROOT` from the environment variables
	gnoroot := gnoenv.RootDir()

	if !osm.DirExists(buildDir) {
		return fmt.Errorf("%q does not exist or is not a directory", buildDir)
	}

	// Determine the path to the gno binary within the build directory
	gnoBin := filepath.Join(buildDir, "gno")
	if _, err := os.Stat(gnoBin); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			// Handle other potential errors from os.Stat
			return err
		}

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

		env.Setenv("HOME", homeDir)
		// Avoids go command printing errors relating to lack of go.mod.
		env.Setenv("GO111MODULE", "off")

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
			ts.Logf("gno command error: %+v", err)
		}

		commandSucceeded := (err == nil)
		successExpected := !neg

		// Compare the command's success status with the expected outcome.
		if commandSucceeded != successExpected {
			ts.Fatalf("unexpected gno command outcome (err=%t expected=%t)", commandSucceeded, successExpected)
		}
	}

	return nil
}
