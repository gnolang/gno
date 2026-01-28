//go:build gnobench

package gnolang_test

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/stretchr/testify/require"
)

// TestBenchOpsFiles tests gasprofile files in "gnovm/tests/files".
// Requires -tags gnobench build tag.
//
// Cheatsheet:
//
//	run all benchops file tests:
//		go test -tags gnobench -run TestBenchOpsFiles -short
//	run a specific test:
//		go test -tags gnobench -run TestBenchOpsFiles/gasprofile_basic
//	update golden tests:
//		go test -tags gnobench -run TestBenchOpsFiles -short -update-golden-tests
func TestBenchOpsFiles(t *testing.T) {
	t.Parallel()

	rootDir, err := filepath.Abs("../../../")
	require.NoError(t, err)

	newOpts := func() *test.TestOptions {
		o := &test.TestOptions{
			RootDir: rootDir,
			Output:  io.Discard,
			Error:   io.Discard,
			Sync:    *withSync,
		}
		o.BaseStore, o.TestStore = test.StoreWithOptions(
			rootDir, o.WriterForStore(),
			test.StoreOptions{WithExtern: true, WithExamples: true, Testing: true},
		)
		return o
	}
	sharedOpts := newOpts()

	dir := "../../tests/files"
	fsys := os.DirFS(dir)
	err = fs.WalkDir(fsys, ".", func(path string, de fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case path == "extern":
			return fs.SkipDir
		case de.IsDir():
			return nil
		}

		// Only process gasprofile_*.gno files
		if !strings.HasPrefix(filepath.Base(path), "gasprofile_") {
			return nil
		}
		if !strings.HasSuffix(path, ".gno") {
			return nil
		}

		subTestName := path
		isLong := strings.HasSuffix(path, "_long.gno")
		if isLong && testing.Short() {
			t.Run(subTestName, func(t *testing.T) {
				t.Skip("skipping long (-short)")
			})
			return nil
		}

		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		var criticalError error
		t.Run(subTestName, func(t *testing.T) {
			opts := sharedOpts
			if isLong {
				t.Parallel()
				opts = newOpts()
			}
			changed, err := opts.RunFiletest(path, content, opts.TestStore)
			if err != nil {
				t.Fatal(err.Error())
			}
			if changed != "" {
				err = os.WriteFile(filepath.Join(dir, path), []byte(changed), de.Type())
				if err != nil {
					criticalError = fmt.Errorf("could not fix golden file: %w", err)
				}
			}
		})

		return criticalError
	})
	if err != nil {
		t.Fatal(err)
	}
}
