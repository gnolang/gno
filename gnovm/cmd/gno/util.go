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
	var paths []string

	for _, argPath := range args {
		info, err := os.Stat(argPath)
		if err != nil {
			return nil, fmt.Errorf("gno: invalid file or package path %q: %w", argPath, err)
		}

		if !info.IsDir() {
			if isGnoFile(fs.FileInfoToDirEntry(info)) {
				paths = append(paths, cleanPath(argPath))
			}

			continue
		}

		// Gather package paths from the directory
		err = walkDirForGnoFiles(argPath, func(path string) {
			paths = append(paths, cleanPath(path))
		})
		if err != nil {
			return nil, fmt.Errorf("unable to walk dir: %w", err)
		}
	}

	return paths, nil
}

func gnoFilesFromArgs(args []string) ([]string, error) {
	var paths []string

	for _, argPath := range args {
		info, err := os.Stat(argPath)
		if err != nil {
			return nil, fmt.Errorf("gno: invalid file or package path %q: %w", argPath, err)
		}

		if !info.IsDir() {
			if isGnoFile(fs.FileInfoToDirEntry(info)) {
				paths = append(paths, cleanPath(argPath))
			}
			continue
		}

		files, err := os.ReadDir(argPath)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			if isGnoFile(f) {
				path := filepath.Join(argPath, f.Name())
				paths = append(paths, cleanPath(path))
			}
		}
	}

	return paths, nil
}

// ensures that the path is absolute or starts with a dot.
// ensures that the path is a dir path.
func cleanPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if strings.HasPrefix(path, ".") {
		return path
	}
	// cannot use path.Join or filepath.Join, because we need
	// to ensure that ./ is the prefix to pass to go build.
	// if not absolute.
	return "." + string(filepath.Separator) + path
}

func walkDirForGnoFiles(root string, addPath func(path string)) error {
	visited := make(map[string]struct{})

	walkFn := func(currPath string, f fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%s: walk dir: %w", root, err)
		}

		if f.IsDir() || !isGnoFile(f) {
			return nil
		}

		parentDir := filepath.Dir(currPath)
		if _, found := visited[parentDir]; found {
			return nil
		}

		visited[parentDir] = struct{}{}

		addPath(parentDir)

		return nil
	}

	return filepath.WalkDir(root, walkFn)
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
			return nil, fmt.Errorf("gno: invalid file or package path %q: %w", dirToSearch, err)
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
	re = strings.ReplaceAll(re, `\.\.\.`, `.*`)
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

// ResolvePath determines the path where to place output files.
// output is the output directory provided by the user.
// dstPath is the desired output path by the gno program.
//
// If dstPath is relative non-local path (ie. contains ../), the dstPath will
// be made absolute and joined with output.
//
// Otherwise, the result is simply filepath.Join(output, dstPath).
//
// See related test for examples.
func ResolvePath(output, dstPath string) (string, error) {
	if filepath.IsAbs(dstPath) ||
		filepath.IsLocal(dstPath) {
		return filepath.Join(output, dstPath), nil
	}
	// Make dstPath absolute and join it with output.
	absDst, err := filepath.Abs(dstPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(output, absDst), nil
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
