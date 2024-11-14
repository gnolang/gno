package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnofiles"
)

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
			if f.IsDir() || !gnofiles.DirEntryIsGnoFile(f) {
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
