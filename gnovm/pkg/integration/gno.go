package integration

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

// SetupGno sets up the given testscripts environment for tests that use the gno
// command. It build `gno` using `go build` command into the given buildDir if
// not existing already.
// It will add `gno` command to p.Cmds. It also wraps p.Setup to set up the environment
// variables for running the go command appropriately.
func SetupGno(p *testscript.Params, buildDir string) error {
	gnoroot := os.Getenv("GNOROOT")
	if gnoroot == "" {
		// Get root location of github.com/gnolang/gno
		goModPath, err := exec.Command("go", "env", "GOMOD").CombinedOutput()
		if err != nil {
			return fmt.Errorf("unable to determine gno root directory")
		}

		gnoroot = filepath.Dir(string(goModPath))
	}

	info, err := os.Stat(buildDir)
	if err != nil {
		return fmt.Errorf("unable to stat: %q", buildDir)
	}

	if !info.IsDir() {
		return fmt.Errorf("given buildDir is not a directory: %q", buildDir)
	}

	gnoBin := filepath.Join(buildDir, "gno")
	if _, err = os.Stat(gnoBin); errors.Is(err, os.ErrNotExist) {
		// Build a fresh gno binary in a temp directory
		gnoArgsBuilder := []string{"build", "-o", gnoBin}

		// Add coverage if needed
		if coverMode := testing.CoverMode(); coverMode != "" {
			gnoArgsBuilder = append(gnoArgsBuilder, "-covermode", coverMode)
		}

		// Add target command
		gnoArgsBuilder = append(gnoArgsBuilder, filepath.Join(gnoroot, "gnovm", "cmd", "gno"))

		if err = exec.Command("go", gnoArgsBuilder...).Run(); err != nil {
			return fmt.Errorf("uanble to build gno binary: %w", err)
		}
	} else if err != nil {
		// Return any other errors
		return err
	}

	// Wrap setup scripts
	origSetup := p.Setup
	p.Setup = func(env *testscript.Env) error {
		if origSetup != nil {
			if err := origSetup(env); err != nil {
				return err
			}
		}

		env.Setenv("GNOROOT", gnoroot) // thx PR 1014 :)

		// by default, $HOME=/no-home, but we need an existing $HOME directory
		// because some commands needs to access $HOME/.cache/go-build
		home, err := os.MkdirTemp("", "gno")
		if err != nil {
			return fmt.Errorf("unable to create temporary home directory: %w", err)
		}
		env.Defer(func() {
			os.RemoveAll(home)
		})
		env.Setenv("HOME", home)

		return nil
	}

	if p.Cmds == nil {
		p.Cmds = make(map[string]func(ts *testscript.TestScript, neg bool, args []string))
	}

	// Set gno command
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
