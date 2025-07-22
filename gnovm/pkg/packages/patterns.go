package packages

import (
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
)

type pkgMatch struct {
	Dir   string
	Match []string
}

func expandPatterns(gnoRoot string, workspaceRoot string, out io.Writer, patterns ...string) ([]*pkgMatch, error) {
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

		if workspaceRoot == "" {
			continue
		}
		switch patKind {
		case patternKindDirectory, patternKindSingleFile, patternKindRecursiveLocal:
			absPat, err := filepath.Abs(match)
			if err != nil {
				return nil, fmt.Errorf("can't get absolute path to pattern %q: %w", match, err)
			}
			if !strings.HasPrefix(absPat, workspaceRoot) {
				return nil, fmt.Errorf("pattern %q is not rooted in current workspace (%q is not in %q)", match, absPat, workspaceRoot)
			}
		}
	}

	if slices.Contains(kinds, patternKindSingleFile) {
		return nil, fmt.Errorf("command-line-arguments package not supported yet")
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
			if _, err := os.Stat(pat); err != nil {
				return nil, fmt.Errorf("%s: %w", match, err)
			}
			addPkgDir(pat, &match)

		case patternKindRemote:
			var dir string
			if gnolang.IsStdlib(pat) {
				dir = StdlibDir(gnoRoot, pat)
			} else {
				dir = PackageDir(pat)
			}
			addPkgDir(dir, &match)

		case patternKindRecursiveLocal:
			dirs, err := expandRecursive(workspaceRoot, pat)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", match, err)
			}
			if len(dirs) == 0 {
				fmt.Fprintf(out, "gno: warning: %q matched no packages\n", match)
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

func expandRecursive(workspaceRoot string, pattern string) ([]string, error) {
	// this works because we only support ... at the end of patterns for now
	patternRoot, _ := filepath.Split(pattern)

	// check that the pattern root is a directory
	rootInfo, err := os.Stat(patternRoot)
	if err != nil {
		return nil, err
	}
	if !rootInfo.IsDir() {
		return nil, fmt.Errorf("recursive pattern root %q is not a directory", patternRoot)
	}

	pkgDirs := []string{}
	if err := fs.WalkDir(os.DirFS(patternRoot), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// get file's parent directory
		parentDir, base := filepath.Split(path)
		parentDir = filepath.Join(patternRoot, parentDir)

		// ignore file if it's containing directory is already marked as a package dir
		if slices.Contains(pkgDirs, parentDir) {
			return nil
		}

		switch base {
		case "gnomod.toml", "gno.mod":
			// add directories that contain gnomods as package dirs
			pkgDirs = append(pkgDirs, parentDir)
			return nil
		case "gnowork.toml":
			// ignore sub-tree if it's a sub-workspace
			if parentDir != workspaceRoot {
				pkgDirs = slices.DeleteFunc(pkgDirs, func(d string) bool { return d == parentDir })
				return fs.SkipDir
			}
		}

		// XXX: include directories with .gno files and automatically derive pkgpath from gnowork.toml paths

		return nil
	}); err != nil {
		return nil, err
	}

	return pkgDirs, nil
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

	if !filepath.IsAbs(pat) && patternIsRemote(pat) {
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

// patternIsRemote reports wether a pattern is starting with a domain or is a stdlib
func patternIsRemote(path string) bool {
	if gnolang.IsStdlib(strings.TrimSuffix(path, "/...")) {
		return true
	}
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
