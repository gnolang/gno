package gnolang_test

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/stretchr/testify/require"
)

var withSync = flag.Bool("update-golden-tests", false, "rewrite tests updating Realm: and Output: with new values where changed")

// TestFiles tests all the files in "gnovm/tests/files".
//
// Cheatsheet:
//
//	fail on the first test:
//		go test -run TestFiles -failfast
//	run a specific test:
//		go test -run TestFiles/addr0b
//	fix a specific test:
//		go test -run TestFiles/'^bin1.gno' -short -v -update-golden-tests .
func TestFiles(t *testing.T) {
	rootDir, err := filepath.Abs("../../../")
	require.NoError(t, err)
	var opts test.FileTestOptions
	var stdin, stderr bytes.Buffer
	opts.BaseStore, opts.Store = test.TestStore(rootDir, true, &stdin, &opts.Stdout, &stderr)
	opts.Store.SetStrictGo2GnoMapping(true) // in gno.land, natives must be registered.
	// XXX: Using opts like this is a bit funky, replacing the state each time; maybe we can re-create each time
	// if we don't require usage of opts.Stdout.
	// XXX: Maybe testStore imports should use baseStore directly, so we directly load into store all package loads.

	dir := "../../tests/"
	fsys := os.DirFS(dir)
	err = fs.WalkDir(fsys, "files", func(path string, de fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case path == "files/extern":
			return fs.SkipDir
		case de.IsDir():
			return nil
		}
		if strings.HasSuffix(path, "_long.gno") && testing.Short() {
			t.Run(path, func(t *testing.T) {
				t.Log("skipping in -short")
			})
			return nil
		}

		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		var criticalError error
		t.Run(path[len("files/"):], func(t *testing.T) {
			if *withSync {
				changed, err := opts.RunSync(path, content)
				if err != nil {
					t.Fatal(err.Error())
				}
				if changed != "" {
					err = os.WriteFile(filepath.Join(dir, path), []byte(changed), de.Type())
					if err != nil {
						criticalError = fmt.Errorf("could not fix golden file: %w", err)
					}
				}
				return
			}

			err := opts.Run(path, content)
			if err != nil {
				t.Error(err.Error())
			}
		})

		return criticalError
	})
	if err != nil {
		t.Fatal(err)
	}
}
