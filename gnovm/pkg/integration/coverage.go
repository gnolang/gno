package integration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

var coverageEnv struct {
	coverdir string
	once     sync.Once
}

// SetupCoverage sets up the given test environment for coverage
func SetupCoverage(p *testscript.Params) error {
	coverdir := os.Getenv("GOCOVERDIR_TXTAR")
	if testing.CoverMode() == "" || coverdir == "" {
		return nil
	}

	var err error

	if !filepath.IsAbs(coverdir) {
		coverdir, err = filepath.Abs(coverdir)
		if err != nil {
			return fmt.Errorf("unable to determine absolute path of %q: %w", coverdir, err)
		}
	}

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

		// Override `GOCOVEDIR` directory
		env.Setenv("GOCOVERDIR", coverdir)
		return nil
	}

	return nil
}
