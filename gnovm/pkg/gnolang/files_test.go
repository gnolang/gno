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
	rootDir, err := filepath.Abs("../../../")
	require.NoError(t, err)

	opts := &test.TestOptions{
		RootDir: rootDir,
		Output:  io.Discard,
		Error:   io.Discard,
		Sync:    *withSync,
	}
	opts.BaseStore, opts.TestStore = test.Store(
		rootDir, true,
		nopReader{}, opts.WriterForStore(), io.Discard,
	)

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
		if strings.HasSuffix(path, "_long.gno") && testing.Short() {
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
	rootDir, err := filepath.Abs("../../../")
	require.NoError(t, err)

	var capture bytes.Buffer
	out := io.Writer(&capture)
	if testing.Verbose() {
		out = os.Stdout
	}
	opts := test.NewTestOptions(rootDir, nopReader{}, out, out)
	opts.Verbose = true

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
			if testing.Short() {
				switch memPkg.Path {
				case "bytes", "strconv", "regexp/syntax":
					t.Skip("Skipped because of -short, and this stdlib is very long currently.")
				}
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
