package gnolang_test

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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
		o.BaseStore, o.TestStore = test.Store(
			rootDir, true,
			nopReader{}, o.WriterForStore(), io.Discard,
		)
		return o
	}
	// sharedOpts is used for all "short" tests.
	sharedOpts := newOpts()

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
		subTestName := path[len("files/"):]
		isLong := strings.HasSuffix(path, "_long.gno")
		if isLong && testing.Short() {
			t.Run(subTestName, func(t *testing.T) {
				t.Skip("skipping in -short")
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
				// Long tests are run in parallel, and with their own store.
				t.Parallel()
				opts = newOpts()
			}
			changed, err := opts.RunFiletest(path, content)
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
		opts = test.NewTestOptions(rootDir, nopReader{}, out, out)
		opts.Verbose = true
		return
	}
	sharedCapture, sharedOpts := newOpts()

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
		memPkg := gnolang.ReadMemPackage(fp, path)
		t.Run(strings.ReplaceAll(memPkg.Path, "/", "-"), func(t *testing.T) {
			capture, opts := sharedCapture, sharedOpts
			switch memPkg.Path {
			// Excluded in short
			case
				"bufio",
				"bytes",
				"strconv":
				if testing.Short() {
					t.Skip("Skipped because of -short, and this stdlib is very long currently.")
				}
				fallthrough
			// Run using separate store, as it's faster
			case
				"math/rand",
				"regexp",
				"regexp/syntax",
				"sort":
				t.Parallel()
				capture, opts = newOpts()
			}

			if capture != nil {
				capture.Reset()
			}

			err := test.Test(memPkg, "", opts)
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
}
