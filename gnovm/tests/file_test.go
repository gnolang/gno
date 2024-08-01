package tests

import (
	"flag"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

var withSync = flag.Bool("update-golden-tests", false, "rewrite tests updating Realm: and Output: with new values where changed")

func TestFileStr(t *testing.T) {
	filePath := filepath.Join(".", "files", "str.gno")
	runFileTest(t, filePath, WithNativeLibs())
}

// Run tests in the `files` directory using shims from stdlib
// to native go standard library.
func TestFilesNative(t *testing.T) {
	baseDir := filepath.Join(".", "files")
	runFileTests(t, baseDir, []string{"*_stdlibs*"}, WithNativeLibs())
}

// Test files using standard library in stdlibs/.
func TestFiles(t *testing.T) {
	baseDir := filepath.Join(".", "files")
	runFileTests(t, baseDir, []string{"*_native*"})
}

func TestChallenges(t *testing.T) {
	t.Skip("Challenge tests, skipping.")
	baseDir := filepath.Join(".", "challenges")
	runFileTests(t, baseDir, nil)
}

type testFile struct {
	path string
	fs.DirEntry
}

// ignore are glob patterns to ignore
func runFileTests(t *testing.T, baseDir string, ignore []string, opts ...RunFileTestOption) {
	t.Helper()

	opts = append([]RunFileTestOption{WithSyncWanted(*withSync)}, opts...)

	files, err := readFiles(t, baseDir)
	if err != nil {
		t.Fatal(err)
	}

	files = filterFileTests(t, files, ignore)
	var path string
	var name string
	for _, file := range files {
		path = file.path
		name = strings.TrimPrefix(file.path, baseDir+string(os.PathSeparator))
		t.Run(name, func(t *testing.T) {
			runFileTest(t, path, opts...)
		})
	}
}

// it reads all files recursively in the directory
func readFiles(t *testing.T, dir string) ([]testFile, error) {
	t.Helper()
	var files []testFile

	err := filepath.WalkDir(dir, func(path string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if de.IsDir() && de.Name() == "extern" {
			return filepath.SkipDir
		}
		f := testFile{path: path, DirEntry: de}

		files = append(files, f)
		return nil
	})
	return files, err
}

func filterFileTests(t *testing.T, files []testFile, ignore []string) []testFile {
	t.Helper()
	filtered := make([]testFile, 0, 1000)
	var name string

	for _, f := range files {
		// skip none .gno files
		name = f.DirEntry.Name()
		if filepath.Ext(name) != ".gno" {
			continue
		}
		// skip ignored files
		if isIgnored(t, name, ignore) {
			continue
		}
		// skip _long file if we only want to test regular file.
		if testing.Short() && strings.Contains(name, "_long") {
			t.Logf("skipping test %s in short mode.", name)
			continue
		}
		filtered = append(filtered, f)
	}
	return filtered
}

func isIgnored(t *testing.T, name string, ignore []string) bool {
	t.Helper()
	isIgnore := false
	for _, is := range ignore {
		match, err := path.Match(is, name)
		if err != nil {
			t.Fatalf("error parsing glob pattern %q: %v", is, err)
		}
		if match {
			isIgnore = true
			break
		}
	}
	return isIgnore
}

func runFileTest(t *testing.T, path string, opts ...RunFileTestOption) {
	t.Helper()

	opts = append([]RunFileTestOption{WithSyncWanted(*withSync)}, opts...)

	var logger loggerFunc
	if gno.IsDebug() && testing.Verbose() {
		logger = t.Log
	}
	rootDir := filepath.Join("..", "..")
	_, err := RunFileTest(rootDir, path, append(opts, WithLoggerFunc(logger))...)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
}
