package main

import (
	"fmt"
	"go/ast"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/transpiler"
)

func isGnoFile(f fs.DirEntry) bool {
	name := f.Name()
	return !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".gno") && !f.IsDir()
}

func isFileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func gnoFilesFromArgsRecursively(args []string) ([]string, error) {
	paths := []string{}
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			return nil, fmt.Errorf("invalid file or package path: %w", err)
		}
		if !info.IsDir() {
			curpath := arg
			paths = append(paths, curpath)
			continue
		}

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
	return paths, nil
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
			files, err := os.ReadDir(arg)
			if err != nil {
				return nil, err
			}
			for _, f := range files {
				if isGnoFile(f) {
					curpath := filepath.Join(arg, f.Name())
					paths = append(paths, curpath)
				}
			}
		}
	}
	return paths, nil
}

func gnoPackagesFromArgsRecursively(args []string) ([]string, error) {
	paths := []string{}
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			return nil, fmt.Errorf("invalid file or package path: %w", err)
		}

		if !info.IsDir() {
			paths = append(paths, arg)
			continue
		}

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

			pkg := parentDir
			if !filepath.IsAbs(parentDir) {
				// cannot use path.Join or filepath.Join, because we need
				// to ensure that ./ is the prefix to pass to go build.
				// if not absolute.
				pkg = "./" + parentDir
			}

			paths = append(paths, pkg)
			return nil
		})

		if err != nil {
			return nil, err
		}

	}
	return paths, nil
}

// targetsFromPatterns returns a list of target paths that match the patterns.
// Each pattern can represent a file or a directory, and if the pattern
// includes "/...", the "..." is treated as a wildcard, matching any string.
// Intended to be used by gno commands such as `gno test`.
func targetsFromPatterns(patterns []string) ([]string, error) {
	paths := []string{}
	for _, p := range patterns {
		var match func(string) bool
		patternLookup := false
		dirToSearch := p

		// Check if the pattern includes `/...`
		if strings.Contains(p, "/...") {
			index := strings.Index(p, "/...")
			if index != -1 {
				dirToSearch = p[:index] // Extract the directory path to search
			}
			match = matchPattern(strings.TrimPrefix(p, "./"))
			patternLookup = true
		}

		info, err := os.Stat(dirToSearch)
		if err != nil {
			return nil, fmt.Errorf("invalid file or package path: %w", err)
		}

		// If the pattern is a file or a directory
		// without `/...`, add it to the list.
		if !info.IsDir() || !patternLookup {
			paths = append(paths, p)
			continue
		}

		// the pattern is a dir containing `/...`, walk the dir recursively and
		// look for directories containing at least one .gno file and match pattern.
		visited := map[string]bool{} // used to run the builder only once per folder.
		err = filepath.WalkDir(dirToSearch, func(curpath string, f fs.DirEntry, err error) error {
			if err != nil {
				return fmt.Errorf("%s: walk dir: %w", dirToSearch, err)
			}
			// Skip directories and non ".gno" files.
			if f.IsDir() || !isGnoFile(f) {
				return nil
			}

			parentDir := filepath.Dir(curpath)
			if _, found := visited[parentDir]; found {
				return nil
			}

			visited[parentDir] = true
			if match(parentDir) {
				paths = append(paths, parentDir)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return paths, nil
}

// matchPattern(pattern)(name) reports whether
// name matches pattern.  Pattern is a limited glob
// pattern in which '...' means 'any string' and there
// is no other special syntax.
// Simplified version of go source's matchPatternInternal
// (see $GOROOT/src/cmd/internal/pkgpattern)
func matchPattern(pattern string) func(name string) bool {
	re := regexp.QuoteMeta(pattern)
	re = strings.Replace(re, `\.\.\.`, `.*`, -1)
	// Special case: foo/... matches foo too.
	if strings.HasSuffix(re, `/.*`) {
		re = re[:len(re)-len(`/.*`)] + `(/.*)?`
	}
	reg := regexp.MustCompile(`^` + re + `$`)
	return func(name string) bool {
		return reg.MatchString(name)
	}
}

func fmtDuration(d time.Duration) string {
	return fmt.Sprintf("%.2fs", d.Seconds())
}

// getPathsFromImportSpec derive and returns ImportPaths
// without ImportPrefix from *ast.ImportSpec
func getPathsFromImportSpec(importSpec []*ast.ImportSpec) (importPaths []importPath) {
	for _, i := range importSpec {
		path := i.Path.Value[1 : len(i.Path.Value)-1] // trim leading and trailing `"`
		if strings.HasPrefix(path, transpiler.ImportPrefix) {
			res := strings.TrimPrefix(path, transpiler.ImportPrefix)
			importPaths = append(importPaths, importPath("."+res))
		}
	}
	return
}

// ResolvePath joins the output dir with relative pkg path
// e.g
// Output Dir: Temp/gno-transpile
// Pkg Path: ../example/gno.land/p/pkg
// Returns -> Temp/gno-transpile/example/gno.land/p/pkg
func ResolvePath(output string, path importPath) (string, error) {
	absOutput, err := filepath.Abs(output)
	if err != nil {
		return "", err
	}
	absPkgPath, err := filepath.Abs(string(path))
	if err != nil {
		return "", err
	}
	pkgPath := strings.TrimPrefix(absPkgPath, gnoenv.RootDir())

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
