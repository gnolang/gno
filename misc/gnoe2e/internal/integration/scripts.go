package integration

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
)

// ResolveScriptFiles turns command arguments into the scripts to run. Accepts
// no arguments (the suite that ships with the checkout), directories, .txtar
// files, or a mix.
//
// A directory contributes its scripts in filename order and a named file
// contributes itself, so the caller's order survives as the run order.
func ResolveScriptFiles(args []string) ([]string, error) {
	if len(args) == 0 {
		// Anchored on the checkout rather than the working directory: the
		// scripts belong to the harness, so where the binary was invoked from
		// must not decide which suite runs. Resolved here because
		// gnoenv.RootDir panics when it cannot find a checkout, which a
		// caller naming its own scripts should never pay for.
		args = []string{filepath.Join(gnoenv.RootDir(), "misc", "gnoe2e", "testdata", "integration")}
	}

	var scripts []string
	for _, arg := range args {
		abs, err := filepath.Abs(arg)
		if err != nil {
			return nil, fmt.Errorf("invalid path %q: %w", arg, err)
		}
		info, err := os.Stat(abs)
		if err != nil {
			return nil, fmt.Errorf("not found: %s", arg)
		}

		if !info.IsDir() {
			// A named file that is not a script is a mistake worth reporting:
			// silently skipping it would run a suite the caller did not ask
			// for and report success.
			if filepath.Ext(abs) != ".txtar" {
				return nil, fmt.Errorf("expected .txtar file, got: %s", arg)
			}
			scripts = append(scripts, abs)
			continue
		}

		entries, err := os.ReadDir(abs)
		if err != nil {
			return nil, fmt.Errorf("read directory %s: %w", arg, err)
		}
		found := 0
		for _, e := range entries {
			if e.IsDir() || filepath.Ext(e.Name()) != ".txtar" {
				continue
			}
			scripts = append(scripts, filepath.Join(abs, e.Name()))
			found++
		}
		if found == 0 {
			return nil, fmt.Errorf("no .txtar scripts in %s", arg)
		}
	}
	return scripts, nil
}
