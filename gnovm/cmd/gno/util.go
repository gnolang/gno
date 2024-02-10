package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/precompile"
)

func isFileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// TargetsFromPatterns returns a list of target paths that match the patterns.
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
			if f.IsDir() || !precompile.IsGnoFile(f) {
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

// makeTestGoMod creates the temporary go.mod for test
func makeTestGoMod(path string, packageName string, goversion string) error {
	content := fmt.Sprintf("module %s\n\ngo %s\n", packageName, goversion)
	return os.WriteFile(path, []byte(content), 0o644)
}

// resolvePath joins the output dir with relative pkg path
// e.g
// Output Dir: Temp/gno-precompile
// Pkg Path: ../example/gno.land/p/pkg
// Returns -> Temp/gno-precompile/example/gno.land/p/pkg
func resolvePath(output string, path precompile.ImportPath) (string, error) {
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
