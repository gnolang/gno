package packages

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"golang.org/x/mod/module"
)

type pkgMatch struct {
	Dir   string
	Match []string
}

func expandPatterns(warn io.Writer, patterns ...string) ([]*pkgMatch, error) {
	pkgMatches := []*pkgMatch(nil)

	addPkgDir := func(dir string, match *string) {
		idx := slices.IndexFunc(pkgMatches, func(sum *pkgMatch) bool { return sum.Dir == dir })
		if idx == -1 {
			matches := []string{}
			if match != nil {
				matches = append(matches, *match)
			}
			pkgMatches = append(pkgMatches, &pkgMatch{Dir: dir, Match: matches})
			return
		}
		if match == nil {
			return
		}
		if slices.Contains(pkgMatches[idx].Match, *match) {
			return
		}
		pkgMatches[idx].Match = append(pkgMatches[idx].Match, *match)
	}

	kinds := make([]patternKind, 0, len(patterns))
	for _, match := range patterns {
		patKind, err := getPatternKind(match)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", match, err)
		}
		kinds = append(kinds, patKind)
	}

	if slices.Contains(kinds, patternKindSingleFile) {
		remaining := []string{}
		remainingKinds := []patternKind{}

		files := make([]string, 0, len(patterns))
		for i, match := range patterns {
			kind := kinds[i]
			if kind != patternKindSingleFile {
				remaining = append(remaining, match)
				remainingKinds = append(remainingKinds, kind)
				continue
			}
			if !strings.HasSuffix(match, ".gno") {
				return nil, fmt.Errorf("named files must be .gno files: %s", match)
			}
			files = append(files, match)
		}

		pkgMatches = append(pkgMatches, &pkgMatch{Dir: "command-line-arguments", Match: files})

		patterns = remaining
		kinds = remainingKinds
	}

	for i, match := range patterns {
		patKind := kinds[i]

		switch patKind {
		case patternKindRecursiveRemote:
			return nil, fmt.Errorf("%s: recursive remote patterns are not supported", match)
		case patternKindSingleFile:
			return nil, fmt.Errorf("unexpected single pattern at this point")
		}

		pat, err := cleanPattern(match, patKind)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", match, err)
		}

		switch patKind {
		case patternKindDirectory:
			addPkgDir(pat, &match)

		case patternKindRemote:
			var dir string
			if gnolang.IsStdlib(pat) {
				dir = StdlibDir(pat)
			} else {
				dir = gnomod.PackageDir("", module.Version{Path: pat})
			}
			addPkgDir(dir, &match)

		case patternKindRecursiveLocal:
			dirs, err := expandRecursive(pat)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", pat, err)
			}
			if len(dirs) == 0 {
				_, _ = warn.Write([]byte(fmt.Sprintf(`gno: warning: %q matched no packages`, match)))
			}
			for _, dir := range dirs {
				addPkgDir(dir, &match)
			}
		}
	}

	sort.Slice(pkgMatches, func(i, j int) bool {
		return pkgMatches[i].Dir < pkgMatches[j].Dir
	})

	return pkgMatches, nil
}

func expandRecursive(pat string) ([]string, error) {
	root, _ := filepath.Split(pat)

	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, errors.New("glob root is not a directory")
	}

	// we swallow errors after this point as we want the most packages we can get
	dirs := []string{}
	_ = fs.WalkDir(os.DirFS(root), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		dir, base := filepath.Split(path)
		dir = filepath.Join(root, dir)
		if slices.Contains(dirs, dir) {
			return nil
		}
		if base == "gno.mod" || strings.HasSuffix(base, ".gno") || base == "LICENSE" || base == "README.md" {
			dirs = append(dirs, dir)
		}
		return nil
	})

	return dirs, nil
}

type patternKind int

const (
	patternKindUnknown = iota
	patternKindSingleFile
	patternKindDirectory
	patternKindRecursiveLocal
	patternKindRemote
	patternKindRecursiveRemote
)

func getPatternKind(pat string) (patternKind, error) {
	isLitteral := patternIsLiteral(pat)

	if !filepath.IsAbs(pat) && (patternIsRemote(pat) || gnolang.IsStdlib(pat)) {
		if isLitteral {
			return patternKindRemote, nil
		}
		dir, base := filepath.Split(pat)
		if base != "..." || strings.Contains(dir, "...") {
			return patternKindUnknown, fmt.Errorf("%s: partial globs are not supported", pat)
		}
		return patternKindRecursiveRemote, nil
	}

	if !isLitteral {
		dir, base := filepath.Split(pat)
		if base != "..." || strings.Contains(dir, "...") {
			return patternKindUnknown, fmt.Errorf("%s: partial globs are not supported", pat)
		}
		return patternKindRecursiveLocal, nil
	}

	if strings.HasSuffix(pat, ".gno") {
		return patternKindSingleFile, nil
	}

	return patternKindDirectory, nil
}

func cleanPattern(pat string, kind patternKind) (string, error) {
	switch kind {
	case patternKindSingleFile, patternKindDirectory, patternKindRecursiveLocal:
		return filepath.Abs(pat)
	case patternKindRemote, patternKindRecursiveRemote:
		return path.Clean(pat), nil
	default:
		return "", fmt.Errorf("unknown pattern kind %d", kind)
	}
}

// patternIsRemote reports wether a pattern is starting with a domain
func patternIsRemote(path string) bool {
	if len(path) == 0 {
		return false
	}
	if path[0] == '.' {
		return false
	}
	slashIdx := strings.IndexRune(path, '/')
	if slashIdx == -1 {
		return false
	}
	return strings.ContainsRune(path[:slashIdx], '.')
}

// patternIsLiteral reports whether the pattern is free of wildcards.
//
// A literal pattern must match at most one package.
func patternIsLiteral(pattern string) bool {
	return !strings.Contains(pattern, "...")
}
