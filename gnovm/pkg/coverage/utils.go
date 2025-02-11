package coverage

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// findAbsFilePath finds the absolute path of a file given its relative path.
// It starts searching from root directory and recursively traverses directories.
func findAbsFilePath(c *Coverage, fpath string) (string, error) {
	cache, ok := c.pathCache[fpath]
	if ok {
		return cache, nil
	}

	var absPath string
	err := filepath.WalkDir(c.rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, fpath) {
			absPath = path
			return filepath.SkipAll
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	if absPath == "" {
		return "", fmt.Errorf("file %s not found", fpath)
	}

	c.pathCache[fpath] = absPath

	return absPath, nil
}

func findMatchingFiles(fileMap fileCoverageMap, pat string) []string {
	var files []string
	for file := range fileMap {
		if strings.Contains(file, pat) {
			files = append(files, file)
		}
	}
	return files
}

func IsTestFile(pkgPath string) bool {
	return strings.HasSuffix(pkgPath, "_test.gno") ||
		strings.HasSuffix(pkgPath, "_testing.gno") ||
		strings.HasSuffix(pkgPath, "_filetest.gno")
}

func isValidFile(currentPath, path string) bool {
	return strings.HasPrefix(path, currentPath) &&
		strings.HasSuffix(path, ".gno") &&
		!IsTestFile(path)
}
