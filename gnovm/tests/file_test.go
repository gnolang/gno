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
	baseDir := filepath.Join(".", "challenges")
	runFileTests(t, baseDir, nil)
}

func filterFileTests(t *testing.T, files []fs.DirEntry, ignore []string) []fs.DirEntry {
	t.Helper()

	for i := 0; i < len(files); i++ {
		file := files[i]
		skip := func() { files = append(files[:i], files[i+1:]...); i-- }
		if filepath.Ext(file.Name()) != ".gno" {
			skip()
			continue
		}
		for _, is := range ignore {
			if match, err := path.Match(is, file.Name()); match {
				skip()
				continue
			} else if err != nil {
				t.Fatalf("error parsing glob pattern %q: %v", is, err)
			}
		}
		if testing.Short() && strings.Contains(file.Name(), "_long") {
			t.Logf("skipping test %s in short mode.", file.Name())
			skip()
			continue
		}
	}
	return files
}

// ignore are glob patterns to ignore
func runFileTests(t *testing.T, baseDir string, ignore []string, opts ...RunFileTestOption) {
	t.Helper()

	opts = append([]RunFileTestOption{WithSyncWanted(*withSync)}, opts...)

	files, err := os.ReadDir(baseDir)
	if err != nil {
		t.Fatal(err)
	}

	files = filterFileTests(t, files, ignore)

	for _, file := range files {
		file := file
		t.Run(file.Name(), func(t *testing.T) {
			if file.Name() == "composite0b.gno" {
				t.Skip("challenge failing")
			}
			if file.Name() == "composite1b.gno" {
				t.Skip("challenge failing")
			}
			if file.Name() == "not_a_type.gno" {
				t.Skip("challenge failing")
			}
			if file.Name() == "unused0.gno" {
				t.Skip("challenge failing")
			}

			runFileTest(t, filepath.Join(baseDir, file.Name()), opts...)
		})
	}
}

func runFileTest(t *testing.T, path string, opts ...RunFileTestOption) {
	t.Helper()

	opts = append([]RunFileTestOption{WithSyncWanted(*withSync)}, opts...)

	var logger loggerFunc
	if gno.IsDebug() && testing.Verbose() {
		logger = t.Log
	}
	rootDir := filepath.Join("..", "..")
	err := RunFileTest(rootDir, path, append(opts, WithLoggerFunc(logger))...)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
}
