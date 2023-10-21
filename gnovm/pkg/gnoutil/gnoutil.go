// Package gnoutil contains Gno development related utilities, common to many
// packages and binaries using Gno.
package gnoutil

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// RepoImport is the import path to the Gno repository.
	RepoImport = "github.com/gnolang/gno"

	// GnolangImport is the import path to the gnolang package.
	GnolangImport = RepoImport + "/gnovm/pkg/gnolang"
)

// IsGnoFile determines whether the given files matches all of the given patterns,
// with the same matching rules as [MatchPatterns].
//
// It is essentially a helper for MatchPatterns, implicitly adding the patterns
// "*.gno" and "!.*"
func IsGnoFile(name string, patterns ...string) (bool, error) {
	return MatchPatterns(name, append(patterns, "*.gno", "!.*")...)
}

// MatchPatterns returns whether the string s matches all of the given glob-like
// patterns.
//
//   - Without any modifiers, s matches the given pattern according to the rules
//     of [path.Match].
//   - If a pattern begins with !, it is negated.
//   - If a pattern is surrounded by forward slashes (/), it is interpreted as a
//     regular expression.
//   - A pattern may combine negation and regex; ie. "!/hello.*\.go/"
//   - Regular expressions receive the whole path; glob patterns only receive the
//     last element (path.Base).
//
// An error is returned only if the patterns have an invalid syntax.
func MatchPatterns(s string, patterns ...string) (bool, error) {
	// TODO: does a regex cache make sense here?
	bs := []byte(s)
	for _, pattern := range patterns {
		var negate bool
		if strings.HasPrefix(pattern, "!") {
			negate = true
			pattern = pattern[1:]
		}
		var res bool
		var err error
		if len(pattern) > 1 && pattern[0] == '/' && pattern[len(pattern)-1] == '/' {
			pattern = pattern[1 : len(pattern)-1]
			res, err = regexp.Match(pattern, bs)
		} else {
			res, err = path.Match(pattern, path.Base(s))
		}
		if err != nil {
			return false, fmt.Errorf("pattern %q: %w", pattern, err)
		}
		if res == negate {
			return false, nil
		}
	}
	return true, nil
}

type matchOptions struct {
	files           bool
	patterns        []string
	disableEllipsis bool
}

// MatchOption is an option to be passed to [Match] to modify its behavior.
type MatchOption func(c *matchOptions)

// MatchFiles instructs [Match] to find files instead of packages.
// The file names must match the given patterns, with the same rules/format as
// [MatchPatterns]. Implicitly, all files must not start with "." and must
// end with ".gno".
func MatchFiles(patterns ...string) MatchOption {
	return func(m *matchOptions) { m.files, m.patterns = true, patterns }
}

// MatchPackages instructs [Match] to find packages instead of files. This
// is the default behaviour. A package is defined as a directory containing at
// least one file ending with ".gno" and not starting with ".gno".
// Additional requirement patterns may be specified -- these apply to filenames,
// not directory names.
func MatchPackages(patterns ...string) MatchOption {
	return func(m *matchOptions) { m.files, m.patterns = false, patterns }
}

// MatchEllipsis sets whether to use the ellipsis syntax, as in Go, to match
// packages and files.
//
// When this is enabled, the string "/..." is treated as a wildcard and matches
// any string.
//
// The default behaviour is MatchEllipsis(true).
func MatchEllipsis(b bool) MatchOption {
	return func(m *matchOptions) { m.disableEllipsis = !b }
}

// Match is a central function to parse a set of arguments that expand to a set of
// Gno packages or files. [MatchOptions] may be provided to customise the
// matching behaviour of Match.
//
// By default, Match returns a list of packages matching the patterns in args,
// as well as any "explicit" file passed to it.
func Match(paths []string, opts ...MatchOption) ([]string, error) {
	var c matchOptions
	for _, opt := range opts {
		opt(&c)
	}

	var found []string

	for _, arg := range paths {
		// TODO: eventually we might want to support go-style arguments,
		// where we can pass in a package/realm path, ie:
		// go test gno.land/p/demo/avl
		// for now only work on local FS

		// normalize to /
		arg = strings.ReplaceAll(arg, string(os.PathSeparator), "/")
		if !path.IsAbs(arg) {
			arg = "./" + arg
		}
		if c.disableEllipsis || !strings.Contains(arg, "/...") {
			f, err := os.Stat(arg)
			if err != nil {
				// stat error will already contain path
				return nil, err
			}
			if f.IsDir() {
				files, _, err := collectMatchingFilesDirs(c, arg)
				if err != nil {
					return nil, err
				}
				if c.files {
					found = append(found, files...)
				} else {
					found = append(found, arg)
				}
			}
		}
		if !c.disableEllipsis && strings.Contains(arg, "/...") {
		}
	}

	return found, nil
}

func collectMatchingFilesDirs(c matchOptions, dir string) (files, dirs []string, err error) {
	des, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, err
	}

	for _, de := range des {
		if de.Name()[0] == '.' {
			continue
		}
		fullPath := path.Join(dir, de.Name())
		if de.IsDir() {
			dirs = append(dirs, fullPath)
			continue
		}
		if !IsGnoFile(fullPath, c.patterns...) {
			continue
		}
		files = append(files, fullPath)
		// break if we're only looking for packages on the first matching file we find.
		if !c.files {
			break
		}
	}
	return
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

func DefaultRootDir() string {
	// try to get the root directory from the GNOROOT environment variable.
	if rootdir := os.Getenv("GNOROOT"); rootdir != "" {
		return filepath.Clean(rootdir)
	}

	// if GNOROOT is not set, try to guess the root directory using the `go list` command.
	cmd := exec.Command("go", "list", "-m", "-mod=mod", "-f", "{{.Dir}}", "github.com/gnolang/gno")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("can't guess --root-dir, please fill it manually or define the GNOROOT environment variable globally.")
	}
	rootDir := strings.TrimSpace(string(out))
	return rootDir
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
