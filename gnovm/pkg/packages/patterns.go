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
)

type pkgMatch struct {
	Dir   string
	Match []string
}

func expandPatterns(gnoRoot string, loaderCtx *loaderContext, out io.Writer, patterns ...string) ([]*pkgMatch, error) {
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

		// single package mode
		if !loaderCtx.IsWorkspace {
			switch patKind {
			case patternKindRecursiveLocal:
				return nil, errors.New("recursive pattern not supported in single-package mode, consider creating a gnowork.toml file")
			case patternKindDirectory:
				absPat, err := filepath.Abs(match)
				if err != nil {
					return nil, fmt.Errorf("can't get absolute path to pattern %q: %w", match, err)
				}
				if absPat != loaderCtx.Root {
					return nil, fmt.Errorf("pattern %q is not current package (%q is not %q)", match, absPat, loaderCtx.Root)
				}
			case patternKindSingleFile:
				absPat, err := filepath.Abs(match)
				if err != nil {
					return nil, fmt.Errorf("can't get absolute path to pattern %q: %w", match, err)
				}
				dir := filepath.Dir(absPat)
				if dir != loaderCtx.Root {
					return nil, fmt.Errorf("pattern %q is not current package (%q is not %q)", match, dir, loaderCtx.Root)
				}
			}
		}

		// workspace mode
		switch patKind {
		case patternKindDirectory, patternKindSingleFile, patternKindRecursiveLocal:
			absPat, err := filepath.Abs(match)
			if err != nil {
				return nil, fmt.Errorf("can't get absolute path to pattern %q: %w", match, err)
			}
			if !strings.HasPrefix(absPat, loaderCtx.Root) {
				return nil, fmt.Errorf("pattern %q is not rooted in current workspace (%q is not in %q)", match, absPat, loaderCtx.Root)
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
			// sanity assert
			if !loaderCtx.IsWorkspace {
				panic(fmt.Errorf("unexpected recursive pattern at this point"))
			}

			dirs, err := expandRecursive(loaderCtx.Root, pat)
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
			dir := filepath.Join(patternRoot, path)
			if dir == workspaceRoot {
				return nil
			}
			subwork := filepath.Join(dir, "gnowork.toml")
			_, err := os.Stat(subwork)
			switch {
			case os.IsNotExist(err):
				// not a sub-workspace, continue walking
				return nil
			case err != nil:
				return fmt.Errorf("check that dir is not a subworkspace: %w", err)
			default:
				return fs.SkipDir
			}
		}

		// get file's parent directory
		parentDir, base := filepath.Split(path)
		parentDir = filepath.Join(patternRoot, parentDir)

		switch base {
		case "gnomod.toml", "gno.mod":
			// add directories that contain gnomods as package dirs
			if !slices.Contains(pkgDirs, parentDir) {
				pkgDirs = append(pkgDirs, parentDir)
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
