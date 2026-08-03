package integration

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

// NewTestingParams setup and initialize base params for testing.
func NewTestingParams(t *testing.T, testdir string) testscript.Params {
	t.Helper()

	var params testscript.Params
	params.Dir = testdir

	params.UpdateScripts, _ = strconv.ParseBool(os.Getenv("UPDATE_SCRIPTS"))
	params.TestWork, _ = strconv.ParseBool(os.Getenv("TESTWORK"))
	if deadline, ok := t.Deadline(); ok && params.Deadline.IsZero() {
		params.Deadline = deadline
	}

	// Store the original setup scripts for potential wrapping
	params.Setup = func(env *testscript.Env) error {
		// Set the UPDATE_SCRIPTS environment variable
		env.Setenv("UPDATE_SCRIPTS", strconv.FormatBool(params.UpdateScripts))

		// Set the  environment variable
		env.Setenv("TESTWORK", strconv.FormatBool(params.TestWork))

		return nil
	}

	return params
}

// NewTestingParamsFromRoots is like [NewTestingParams], but discovers the
// scripts to run below each of roots (see [FindTestScripts]) instead of reading
// a single directory. It lets a suite keep its scripts next to the code they
// exercise.
func NewTestingParamsFromRoots(t *testing.T, roots ...string) (testscript.Params, error) {
	t.Helper()

	files, err := FindTestScripts(roots...)
	if err != nil {
		return testscript.Params{}, err
	}
	if len(files) == 0 {
		return testscript.Params{}, fmt.Errorf("no testscript file found below %v", roots)
	}

	// Params.Files is only honored when Params.Dir is empty.
	params := NewTestingParams(t, "")
	params.Files = files
	return params, nil
}

// FindTestScripts walks each of roots recursively and returns every testscript
// file below them, in walk order. Hidden directories are skipped.
//
// Only *.txtar is matched. testscript's own directory mode also accepts *.txt,
// but a bare .txt next to real code is far more likely to be a fixture than a
// script, so it is left alone deliberately.
//
// testscript names each subtest after the script's base name alone, silently
// disambiguating collisions as "name" and "name#1". FindTestScripts returns an
// error on collision instead — across all roots, since they share one subtest
// namespace — so that `-run` selectors stay unambiguous.
func FindTestScripts(roots ...string) ([]string, error) {
	var (
		files []string
		seen  = map[string]string{}
	)
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if path != root && strings.HasPrefix(d.Name(), ".") {
					return fs.SkipDir
				}
				return nil
			}
			name := d.Name()
			if !strings.HasSuffix(name, ".txtar") {
				return nil
			}
			if prev, dup := seen[name]; dup {
				return fmt.Errorf("duplicate testscript base name %q: %q and %q; base names must be unique because subtests are named after them", name, prev, path)
			}
			seen[name] = path
			files = append(files, path)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}
