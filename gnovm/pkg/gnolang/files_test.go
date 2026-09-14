package gnolang_test

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/stretchr/testify/require"
)

var withSync = flag.Bool("update-golden-tests", false, "rewrite tests updating Realm: and Output: with new values where changed")

type nopReader struct{}

func (nopReader) Read(p []byte) (int, error) { return 0, io.EOF }

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

	// Tests run in parallel, drawing a TestOptions (with its store) from this
	// pool. Reusing stores lets tests share previously loaded packages.
	// When syncing golden files, run sequentially to keep walk-order writes
	// and error propagation deterministic.
	poolSize := runtime.GOMAXPROCS(0)
	if *withSync {
		poolSize = 1
	}
	optsPool := make(chan *test.TestOptions, poolSize)
	for range poolSize {
		optsPool <- newOpts()
	}

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
		subTestName := path
		isHidden := strings.HasPrefix(path, ".")
		if isHidden {
			t.Run(subTestName, func(t *testing.T) {
				t.Skip("skipping hidden")
			})
			return nil
		}
		isLong := strings.HasSuffix(path, "_long.gno")
		if isLong && testing.Short() {
			t.Run(subTestName, func(t *testing.T) {
				t.Skip("skipping long (-short)")
			})
			return nil
		}
		isKnown := strings.HasSuffix(path, "_known.gno")
		if isKnown {
			t.Run(subTestName, func(t *testing.T) {
				t.Skip("skipping known issue")
			})
			return nil
		}
		if strings.HasSuffix(path, ".swp") ||
			strings.HasSuffix(path, ".swo") ||
			strings.HasSuffix(path, ".swn") {
			return nil
		}

		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		var criticalError error
		t.Run(subTestName, func(t *testing.T) {
			if !*withSync {
				t.Parallel()
			}
			var opts *test.TestOptions
			if isLong || strings.HasPrefix(subTestName, "alloc") {
				// Long tests get their own store, so they don't hold
				// up a pool slot for their whole (long) duration.
				// Allocator-golden tests (alloc_*) also get their own store:
				// they assert absolute allocator byte counts, which depend on
				// which packages a pooled, previously-used store has already
				// loaded — scheduling-dependent, so their goldens flake under
				// the pool (and always diverge in filtered runs). A fresh
				// store makes the counted allocations deterministic.
				opts = newOpts()
			} else {
				opts = <-optsPool
				defer func() { optsPool <- opts }()
			}
			changed, _, _, err := opts.RunFiletest(path, content, opts.TestStore)
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

// TestStdlibs tests all the standard library packages.
func TestStdlibs(t *testing.T) {
	t.Parallel()

	rootDir, err := filepath.Abs("../../../")
	require.NoError(t, err)

	newOpts := func() (capture *bytes.Buffer, opts *test.TestOptions) {
		var out io.Writer
		if testing.Verbose() {
			out = os.Stdout
		} else {
			capture = new(bytes.Buffer)
			out = capture
		}
		opts = test.NewTestOptions(rootDir, out, out, nil)
		opts.Verbose = true
		return
	}
	dir := "../../stdlibs/"
	fsys := os.DirFS(dir)
	err = fs.WalkDir(fsys, ".", func(path string, de fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case !de.IsDir() || path == ".":
			return nil
		}

		fp := filepath.Join(dir, path)

		// Exclude empty directories.
		files, err := os.ReadDir(fp)
		hasFiles := false
		if err != nil {
			return err
		}
		for _, file := range files {
			if !file.IsDir() &&
				strings.HasSuffix(file.Name(), ".gno") {
				hasFiles = true
			}
		}
		if !hasFiles {
			return nil
		}

		// Read and run tests.
		mpkg := gnolang.MustReadMemPackage(fp, path, gnolang.MPStdlibAll)
		t.Run(strings.ReplaceAll(mpkg.Path, "/", "-"), func(t *testing.T) {
			switch mpkg.Path {
			// Excluded in short
			case
				"bufio",
				"bytes",
				"strconv":
				if testing.Short() {
					t.Skip("Skipped because of -short, and this stdlib is very long currently.")
				}
			}

			// Run in parallel, each package using its own store.
			t.Parallel()
			capture, opts := newOpts()

			err := test.Test(mpkg, "", opts)
			if !testing.Verbose() {
				t.Log(capture.String())
			}
			if err != nil {
				t.Error(err)
			}
		})

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Shared by the tests/stdlibs walk below, which runs sequentially.
	sharedCapture, sharedOpts := newOpts()

	testDir := "../../tests/stdlibs/"
	testFs := os.DirFS(testDir)
	err = fs.WalkDir(testFs, ".", func(path string, de fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case !de.IsDir() || path == ".":
			return nil
		}
		if _, err := os.Stat(filepath.Join(dir, path)); err == nil {
			// skip; this dir exists already in the normal stdlibs and we
			// currently don't support testing these "mixed stdlibs".
			return nil
		}

		fp := filepath.Join(testDir, path)
		mpkg := gnolang.MustReadMemPackage(fp, path, gnolang.MPStdlibAll)
		t.Run("test-"+strings.ReplaceAll(mpkg.Path, "/", "-"), func(t *testing.T) {
			if sharedCapture != nil {
				sharedCapture.Reset()
			}

			err := test.Test(mpkg, "", sharedOpts)
			if !testing.Verbose() {
				t.Log(sharedCapture.String())
			}
			if err != nil {
				t.Error(err)
			}
		})

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
