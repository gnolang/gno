package tests

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	gno "github.com/gnolang/gno/pkgs/gnolang"
)

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

// ignore are glob patterns to ignore
func runFileTests(t *testing.T, baseDir string, ignore []string, opts ...RunFileTestOption) {
	t.Helper()

	files, err := os.ReadDir(baseDir)
	if err != nil {
		t.Fatal(err)
	}
Upper:
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".gno" {
			continue
		}
		for _, is := range ignore {
			if match, err := path.Match(is, file.Name()); match {
				continue Upper
			} else if err != nil {
				t.Fatal(fmt.Errorf("error parsing glob pattern %q: %w", is, err))
			}
		}
		if testing.Short() && strings.Contains(file.Name(), "_long") {
			t.Logf("skipping test %s in short mode.", file.Name())
			continue
		}
		file := file
		t.Run(file.Name(), func(t *testing.T) {
			runFileTest(t, filepath.Join(baseDir, file.Name()), opts...)
		})
	}
}

func runFileTest(t *testing.T, path string, opts ...RunFileTestOption) {
	t.Helper()

	var logger loggerFunc
	if gno.IsDebug() && testing.Verbose() {
		logger = t.Log
	}
	err := RunFileTest("..", path, append(opts, WithLoggerFunc(logger))...)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
}
