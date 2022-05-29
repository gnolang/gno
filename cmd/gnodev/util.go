package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func isGnoFile(f fs.DirEntry) bool {
	name := f.Name()
	return !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".gno") && !f.IsDir()
}

func gnoFilesFromArgs(args []string) ([]string, error) {
	paths := []string{}
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			return nil, fmt.Errorf("invalid file or package path: %w", err)
		}
		if !info.IsDir() {
			curpath := arg
			paths = append(paths, curpath)
		} else {
			err = filepath.WalkDir(arg, func(curpath string, f fs.DirEntry, err error) error {
				if err != nil {
					return fmt.Errorf("%s: walk dir: %w", arg, err)
				}

				if !isGnoFile(f) {
					return nil // skip
				}
				paths = append(paths, curpath)
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
	}
	return paths, nil
}

func gnoPackagesFromArgs(args []string) ([]string, error) {
	paths := []string{}
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			return nil, fmt.Errorf("invalid file or package path: %w", err)
		}
		if !info.IsDir() {
			paths = append(paths, arg)
		} else {
			// if the passed arg is a dir, then we'll recursively walk the dir
			// and look for directories containing at least one .gno file.

			visited := map[string]bool{} // used to run the builder only once per folder.
			err = filepath.WalkDir(arg, func(curpath string, f fs.DirEntry, err error) error {
				if err != nil {
					return fmt.Errorf("%s: walk dir: %w", arg, err)
				}
				if f.IsDir() {
					return nil // skip
				}
				if !isGnoFile(f) {
					return nil // skip
				}

				parentDir := filepath.Dir(curpath)
				if _, found := visited[parentDir]; found {
					return nil
				}
				visited[parentDir] = true

				// cannot use path.Join or filepath.Join, because we need
				// to ensure that ./ is the prefix to pass to go build.
				pkg := "./" + parentDir
				paths = append(paths, pkg)
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
	}
	return paths, nil
}

// CaptureStdoutAndStderr temporarily pipes os.Stdout and os.Stderr into a buffer.
// Imported from https://github.com/moul/u/blob/master/io.go.
func captureStdoutAndStderr() (func() string, error) {
	oldErr := os.Stderr
	oldOut := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	os.Stderr = w
	os.Stdout = w

	closer := func() string {
		w.Close()
		out, _ := ioutil.ReadAll(r)
		os.Stderr = oldErr
		os.Stdout = oldOut
		return string(out)
	}
	return closer, nil
}

func fmtDuration(d time.Duration) string {
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func guessRootDir() string {
	cmd := exec.Command("go", "list", "-m", "-mod=mod", "-f", "{{.Dir}}", "github.com/gnolang/gno")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("can't guess --root-dir, please fill it manually.")
	}
	rootDir := strings.TrimSpace(string(out))
	return rootDir
}
