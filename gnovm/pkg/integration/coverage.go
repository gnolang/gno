package integration

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rogpeppe/go-internal/testscript"
)

var coverageEnv struct {
	coverdir string
}

func init() {
	flag.StringVar(&coverageEnv.coverdir,
		"test.gocoverdir-txtar", "", "write testscripts coverage intermediate files to this directory")
}

// SetupTestscriptsCoverageFromFlag checks the `test.gocoverdir-txtar` flag to determine
// whether to configure testscript parameters for coverage analysis. If the flag is not set,
// it will skip the setup process.
func SetupTestscriptsCoverageFromFlag(p *testscript.Params) error {
	if coverageEnv.coverdir == "" {
		// Skip coverage setup if `test.gocoverdir-txtar` flag wasn't specified
		return nil
	}

	return SetupTestscriptsCoverage(p, coverageEnv.coverdir)
}

// SetupTestscriptsCoverage sets up the given testscripts environment for coverage.
// It will mostly override `GOCOVERDIR` with the target cover directory
func SetupTestscriptsCoverage(p *testscript.Params, coverdir string) error {
	// Check if the given coverage directory exist
	info, err := os.Stat(coverdir)
	if err != nil {
		return fmt.Errorf("output directory %q inaccessible: %w", coverdir, err)
	} else if !info.IsDir() {
		return fmt.Errorf("output directory %q not a directory", coverdir)
	}

	// We need to have an absolute path here, because current directory
	// context will change while executing testscripts.
	if !filepath.IsAbs(coverdir) {
		var err error
		if coverdir, err = filepath.Abs(coverdir); err != nil {
			return fmt.Errorf("unable to determine absolute path of %q: %w", coverdir, err)
		}
	}

	// Backup the original setup function
	origSetup := p.Setup
	p.Setup = func(env *testscript.Env) error {
		if origSetup != nil {
			// Call previous setup first
			origSetup(env)
		}

		// Override `GOCOVEDIR` directory for sub-execution
		env.Setenv("GOCOVERDIR", coverdir)
		return nil
	}

	return nil
}
