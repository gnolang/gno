package integration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/rogpeppe/go-internal/testscript"
)

var coverageEnv struct {
	coverdir string
	once     sync.Once
}

// SetupCoverage sets up the given testscripts environment for coverage.
// It will mostly override `GOCOVERDIR` with the target cover directory
func SetupCoverage(p *testscript.Params, coverdir string) error {
	var err error

	// We need to have an absolute path here, because current directory
	// context will change while executing testscripts.
	if !filepath.IsAbs(coverdir) {
		abspath, err := filepath.Abs(coverdir)
		if err != nil {
			return fmt.Errorf("unable to determine absolute path of %q: %w", coverdir, err)
		}
		coverdir = abspath

	}

	// If the given coverage directory doesn't exist, create it
	info, err := os.Stat(coverdir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if mkErr := os.Mkdir(coverdir, 0o755); mkErr != nil {
				return fmt.Errorf("failed to testscripts coverage dir %q: %w", coverdir, mkErr)
			}
		} else {
			// Handle other potential errors from os.Stat
			return fmt.Errorf("failed to stat %q: %w", coverdir, err)
		}
	} else if !info.IsDir() {
		return fmt.Errorf("coverage: %q is not a directory", coverdir)
	}

	origSetup := p.Setup
	p.Setup = func(env *testscript.Env) error {
		if origSetup != nil {
			origSetup(env)
		}

		// Override `GOCOVEDIR` directory for sub-execution
		env.Setenv("GOCOVERDIR", coverdir)
		return nil
	}

	return nil
}
