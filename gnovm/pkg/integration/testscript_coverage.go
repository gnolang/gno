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
		"txtarcoverdir", "", "write testscripts coverage intermediate files to this directory")
}

// ResolveCoverageDir attempts to resolve the coverage directory from the 'TXTARCOVERDIR'
// environment variable first, and if not set, from the 'test.txtarcoverdir' flag.
// It returns the resolved directory and a boolean indicating if the resolution was successful.
func ResolveCoverageDir() (string, bool) {
	// Attempt to resolve the cover directory from the environment variable or flag
	coverdir := os.Getenv("TXTARCOVERDIR")
	if coverdir == "" {
		coverdir = coverageEnv.coverdir
	}

	return coverdir, coverdir != ""
}

// SetupTestscriptsCoverage sets up the given testscripts environment for coverage.
// It will mostly override `GOCOVERDIR` with the target cover directory
func SetupTestscriptsCoverage(p *testscript.Params, coverdir string) error {
	// Check if the given coverage directory exist
	info, err := os.Stat(coverdir)
	if err != nil {
		return fmt.Errorf("output directory %q inaccessible: %w", coverdir, err)
	} else if !info.IsDir() {
		return fmt.Errorf("output %q not a directory", coverdir)
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
