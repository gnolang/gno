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

// GnoBuildOptions specifies options for building the gno binary.
type GnoBuildOptions struct {
	// BinaryName is the name of the output binary (default: "gno").
	BinaryName string
	// BuildTags are the build tags to use (e.g., "gnobench").
	BuildTags string
}

// buildGnoBinary builds the gno binary with the given options if it doesn't exist.
// Returns the path to the binary.
func buildGnoBinary(gnoroot, buildDir string, opts GnoBuildOptions) (string, error) {
	if !osm.DirExists(buildDir) {
		return "", fmt.Errorf("%q does not exist or is not a directory", buildDir)
	}

	if opts.BinaryName == "" {
		opts.BinaryName = "gno"
	}

	gnoBin := filepath.Join(buildDir, opts.BinaryName)
	if _, err := os.Stat(gnoBin); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}

		// Build the gno binary
		gnoArgsBuilder := []string{"build"}
		if opts.BuildTags != "" {
			gnoArgsBuilder = append(gnoArgsBuilder, "-tags", opts.BuildTags)
		}
		gnoArgsBuilder = append(gnoArgsBuilder, "-o", gnoBin)

		// Forward `-covermode` settings if set
		if coverMode := testing.CoverMode(); coverMode != "" {
			gnoArgsBuilder = append(gnoArgsBuilder, "-covermode", coverMode)
		}

		// Append the path to the gno command source
		gnoArgsBuilder = append(gnoArgsBuilder, filepath.Join(gnoroot, "gnovm", "cmd", "gno"))

		if err = exec.Command("go", gnoArgsBuilder...).Run(); err != nil {
			return "", fmt.Errorf("unable to build gno binary: %w", err)
		}
	}

	return gnoBin, nil
}

// setupGnoCommand configures testscript params with the gno command and environment.
func setupGnoCommand(p *testscript.Params, gnoBin, gnoroot, homeDir string) {
	origSetup := p.Setup
	p.Setup = func(env *testscript.Env) error {
		if origSetup != nil {
			if err := origSetup(env); err != nil {
				return err
			}
		}

		env.Setenv("GNOROOT", gnoroot)
		env.Setenv("HOME", homeDir)
		// Avoids go command printing errors relating to lack of go.mod.
		env.Setenv("GO111MODULE", "off")

		return nil
	}

	if p.Cmds == nil {
		p.Cmds = make(map[string]func(ts *testscript.TestScript, neg bool, args []string))
	}

	p.Cmds["gno"] = func(ts *testscript.TestScript, neg bool, args []string) {
		err := ts.Exec(gnoBin, args...)
		if err != nil {
			ts.Logf("gno command error: %+v", err)
		}

		commandSucceeded := (err == nil)
		successExpected := !neg

		if commandSucceeded != successExpected {
			ts.Fatalf("unexpected gno command outcome (err=%t expected=%t)", commandSucceeded, successExpected)
		}
	}
}

// SetupGno prepares the given testscript environment for tests that utilize the gno command.
// If the `gno` binary doesn't exist, it's built using the `go build` command into the specified buildDir.
// The function also include the `gno` command into `p.Cmds` to and wrap environment into p.Setup
// to correctly set up the environment variables needed for the `gno` command.
func SetupGno(p *testscript.Params, homeDir, buildDir string) error {
	gnoroot := gnoenv.RootDir()

	gnoBin, err := buildGnoBinary(gnoroot, buildDir, GnoBuildOptions{})
	if err != nil {
		return err
	}

	setupGnoCommand(p, gnoBin, gnoroot, homeDir)
	return nil
}
