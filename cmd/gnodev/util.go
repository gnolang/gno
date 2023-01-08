package main

import (
	"fmt"
	"go/ast"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	gno "github.com/gnolang/gno/pkgs/gnolang"
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

// makeTestGoMod creates the temporary go.mod for test
func makeTestGoMod(path string, packageName string, goversion string) error {
	content := fmt.Sprintf("module %s\n\ngo %s\n", packageName, goversion)
	return os.WriteFile(path, []byte(content), 0o644)
}

// getPathsFromImportSpec derive and returns ImportPaths
// without ImportPrefix from *ast.ImportSpec
func getPathsFromImportSpec(importSpec []*ast.ImportSpec) (importPaths []ImportPath) {
	for _, i := range importSpec {
		importPath := i.Path.Value[1 : len(i.Path.Value)-1] // trim leading and trailing `"`
		if strings.HasPrefix(importPath, gno.ImportPrefix) {
			res := strings.TrimPrefix(importPath, gno.ImportPrefix)
			importPaths = append(importPaths, ImportPath("."+res))
		}
	}
	return
}

// ResolvePath joins the output dir with relative pkg path
// e.g
// Output Dir: Temp/gno-precompile
// Pkg Path: ../example/gno.land/p/pkg
// Returns -> Temp/gno-precompile/example/gno.land/p/pkg
func ResolvePath(output string, path ImportPath) (string, error) {
	absOutput, err := filepath.Abs(output)
	if err != nil {
		return "", err
	}
	absPkgPath, err := filepath.Abs(string(path))
	if err != nil {
		return "", err
	}
	pkgPath := strings.TrimPrefix(absPkgPath, guessRootDir())

	return filepath.Join(absOutput, pkgPath), nil
}

// WriteDirFile write file to the path and also create
// directory if needed. with:
// Dir perm -> 0755; File perm -> 0o644
func WriteDirFile(pathWithName string, data []byte) error {
	path := filepath.Dir(pathWithName)

	// Create Dir if not exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0o755)
	}

	return os.WriteFile(pathWithName, []byte(data), 0o644)
}
