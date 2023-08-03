package main

import (
	"fmt"
	"go/ast"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func isGnoFile(f fs.DirEntry) bool {
	name := f.Name()
	return !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".gno") && !f.IsDir()
}

func isFileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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
func getPathsFromImportSpec(importSpec []*ast.ImportSpec) (importPaths []importPath) {
	for _, i := range importSpec {
		path := i.Path.Value[1 : len(i.Path.Value)-1] // trim leading and trailing `"`
		if strings.HasPrefix(path, gno.ImportPrefix) {
			res := strings.TrimPrefix(path, gno.ImportPrefix)
			importPaths = append(importPaths, importPath("."+res))
		}
	}
	return
}

// ResolvePath joins the output dir with relative pkg path
// e.g
// Output Dir: Temp/gno-precompile
// Pkg Path: ../example/gno.land/p/pkg
// Returns -> Temp/gno-precompile/example/gno.land/p/pkg
func ResolvePath(output string, path importPath) (string, error) {
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

	return os.WriteFile(pathWithName, data, 0o644)
}

// copyDir copies the dir from src to dst, the paths have to be
// absolute to ensure consistent behavior.
func copyDir(src, dst string) error {
	if !filepath.IsAbs(src) || !filepath.IsAbs(dst) {
		return fmt.Errorf("src or dst path not absolute, src: %s dst: %s", src, dst)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("cannot read dir: %s", src)
	}

	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%w'", dst, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.Type().IsDir() {
			copyDir(srcPath, dstPath)
		} else if entry.Type().IsRegular() {
			copyFile(srcPath, dstPath)
		}
	}

	return nil
}

// copyFile copies the file from src to dst, the paths have
// to be absolute to ensure consistent behavior.
func copyFile(src, dst string) error {
	if !filepath.IsAbs(src) || !filepath.IsAbs(dst) {
		return fmt.Errorf("src or dst path not absolute, src: %s dst: %s", src, dst)
	}

	// verify if it's regular flile
	srcStat, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("cannot copy file: %w", err)
	}
	if !srcStat.Mode().IsRegular() {
		return fmt.Errorf("%s not a regular file", src)
	}

	// create dst file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// open src file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// copy srcFile -> dstFile
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}

// Adapted from https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func prettySize(nb int64) string {
	const unit = 1000
	if nb < unit {
		return fmt.Sprintf("%d", nb)
	}
	div, exp := int64(unit), 0
	for n := nb / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%c", float64(nb)/float64(div), "kMGTPE"[exp])
}
