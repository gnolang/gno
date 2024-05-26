package importer

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// Filter filters the given files s with patterns, following the rules
// of [MatchPatterns]. It does not add the implicit gno patterns of [IsGnoFile].
//
// Filter panics if any of the patterns is invalid.
func Filter(s []string, patterns ...string) []string {
	ret := make([]string, 0, len(s))
	for _, v := range s {
		ok, err := MatchPatterns(v, patterns...)
		if err != nil {
			panic(err)
		}
		if ok {
			ret = append(ret, v)
		}
	}
	return ret
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
	noImplicit      bool
	disableEllipsis bool
	fs              fs.FS
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

// MatchNoImplicit removes the implicit gno filters of [Match].
// This allows Match to be used, ie., for go files.
func MatchNoImplicit() MatchOption {
	return func(m *matchOptions) { m.noImplicit = true }
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

// matches ellipsis as separators.
var (
	reEllipsis    = regexp.MustCompile(`(?:^|/)\.\.\.(?:/|$)`)
	reEllipsisEsc = regexp.MustCompile(`(?:^|/)\\\.\\\.\\\.(/|$)`)
)

// Match is a central function to parse a set of arguments that expand to a set of
// Gno packages or files. [MatchOptions] may be provided to customise the
// matching behaviour of Match.
//
// By default, Match returns a list of packages matching the patterns in args,
// as well as any "explicit" file passed to it.
func Match(paths []string, opts ...MatchOption) ([]string, error) {
	// Determine options.
	var c matchOptions
	for _, opt := range opts {
		opt(&c)
	}

	// Set defaults for c.fs, append standard patterns.
	var root string
	if c.fs == nil {
		c.fs = os.DirFS("/")
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		root = strings.ReplaceAll(wd, string(os.PathSeparator), "/")
		// remove leading /
		root = root[1:]
	}
	if !c.noImplicit {
		c.patterns = append(c.patterns, "*.gno", "!.*")
	}

	// found will contain the set of matched packages or files.
	var found []string

	for _, arg := range paths {
		// TODO: eventually we might want to support go-style arguments,
		// where we can pass in a package/realm path, ie:
		// go test gno.land/p/demo/avl
		// for now only work on local FS

		// Normalize separator to /, clean the path, and join it with the root
		// if it's relative.
		clean := path.Clean(strings.ReplaceAll(arg, string(os.PathSeparator), "/"))
		originalClean := clean
		if path.IsAbs(clean) {
			clean = clean[1:]
			if clean == "" {
				clean = "."
			}
		} else {
			clean = path.Join(root, clean)
		}
		if !fs.ValidPath(clean) {
			return nil, fmt.Errorf("invalid path: %q", arg)
		}

		// Find any valid ellipsis syntax.
		ellipsis := reEllipsis.FindStringIndex(clean)
		if c.disableEllipsis || ellipsis == nil {
			// No ellipsis, or they are disabled -- stat the path directly.
			f, err := fs.Stat(c.fs, clean)
			if err != nil {
				// stat error will already contain path
				return nil, err
			}
			// Explicit file.
			if !f.IsDir() {
				found = append(found, revertPath(root, originalClean, clean))
				continue
			}
			// Directory, collect all files matching our patterns.
			files, err := collectMatchingFiles(c, clean)
			if err != nil {
				return nil, err
			}
			switch {
			case len(files) == 0:
				if c.noImplicit {
					return nil, fmt.Errorf("dir %s: no matching files", arg)
				}
				return nil, fmt.Errorf("dir %s: no valid gno files", arg)
			case c.files:
				for _, file := range files {
					found = append(found, revertPath(root, originalClean, file))
				}
			default:
				if clean == "." {
					clean = ""
				}
				found = append(found, revertPath(root, originalClean, clean))
			}
			continue
		}

		// Find directory to walk.
		baseDir := clean[:ellipsis[0]]
		if baseDir == "" {
			baseDir = "."
		}

		// Use regexp for linear-time matching
		// Change wildcards to be regex. They will match all directories except for hidden dirs.
		// Note that the regex matches only directory names, not filenames.
		reString := reEllipsisEsc.ReplaceAllString(
			regexp.QuoteMeta(clean), "(?:(?:/|^)[^./][^/]*)*$1",
		)
		// for single triple dot, also allow a single "." as a valid package path.
		if clean == "..." {
			reString = `(?:\.|` + reString + `)`
		}
		reString = "^" + reString + "$"
		re := regexp.MustCompile(reString)
		pathTrim := strings.TrimSuffix(baseDir, "/")
		fi, err := fs.Stat(c.fs, path.Clean(pathTrim))
		if err != nil {
			return nil, err
		}
		err = walkDir(c.fs, pathTrim, &statDirEntry{fi}, func(fsPath string, entry fs.DirEntry, err error) error {
			// BFS guarantees that we get a dir, its files, then its subdirs.
			if entry.IsDir() {
				if !re.MatchString(fsPath) {
					return fs.SkipDir
				}
				return nil
			}
			ok, err := MatchPatterns(fsPath, c.patterns...)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
			if c.files {
				found = append(found, revertPath(root, originalClean, fsPath))
				return nil
			}
			dirPath := path.Dir(fsPath)
			if dirPath == "." {
				dirPath = ""
			}
			found = append(found, revertPath(root, originalClean, dirPath))
			// we found a gno file in this dir; so let's go look at subdirs.
			return fs.SkipDir
		})
		if err != nil {
			return nil, err
		}
	}

	return found, nil
}

// Reverts a "clean" path to be closer to its original form, helper function for Match.
// Original absolute -> add a slash to clean path.
// Original relative -> determine the number of prefixed `../` in the clean path,
// store root + n*`../` as prefix, save back as n*`../` + trimprefix(fullpath, prefix)
// cwd: user CWD for relative paths: home/user
// original: original path specified by user, cleaned: ../user2/file.gno
// clean: clean matched path, absolute but without leading slash: home/user2/file.gno
func revertPath(cwd, original, clean string) string {
	if path.IsAbs(original) {
		return filepath.Join(filepath.VolumeName(original), filepath.FromSlash("/"+clean))
	}
	var dotdot int
	for ; strings.HasPrefix(original[dotdot*3:], "../"); dotdot++ { //nolint:revive
	}

	dots := original[:dotdot*3]  // ../
	pref := path.Join(cwd, dots) // /home
	if pref == clean {
		return "."
	}
	res := dots + strings.TrimPrefix(clean, pref+"/")
	return res
}

type statDirEntry struct {
	info fs.FileInfo
}

func (d *statDirEntry) Name() string               { return d.info.Name() }
func (d *statDirEntry) IsDir() bool                { return d.info.IsDir() }
func (d *statDirEntry) Type() fs.FileMode          { return d.info.Mode().Type() }
func (d *statDirEntry) Info() (fs.FileInfo, error) { return d.info, nil }

// walkDir is mostly copied from fs.WalkDir, with a few modifications:
//
//   - search is performed breadth-first instead of depth-first.
//   - fs.SkipDir is not recursive, and only skips processing the current directory.
//   - the argument to this (recursive) function must be a directory.
func walkDir(fsys fs.FS, name string, d fs.DirEntry, walkDirFn fs.WalkDirFunc) error {
	var skipDir bool
	if err := walkDirFn(name, d, nil); err != nil {
		if !errors.Is(err, fs.SkipDir) {
			return err
		}
		skipDir = true
	}

	files, err := fs.ReadDir(fsys, name)
	if err != nil {
		// Second call, to report ReadDir error.
		err = walkDirFn(name, d, err)
		if err != nil {
			if errors.Is(err, fs.SkipDir) {
				err = nil
			}
			return err
		}
	}

	if !skipDir {
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			name1 := path.Join(name, file.Name())
			if err := walkDirFn(name1, file, nil); err != nil {
				if errors.Is(err, fs.SkipDir) {
					break
				}
				return err
			}
		}
	}

	for _, dir := range files {
		if !dir.IsDir() {
			continue
		}
		name1 := path.Join(name, dir.Name())
		if err := walkDir(fsys, name1, dir, walkDirFn); err != nil {
			return err
		}
	}
	return nil
}

func collectMatchingFiles(c matchOptions, dir string) (files []string, err error) {
	des, err := fs.ReadDir(c.fs, dir)
	if err != nil {
		return nil, err
	}

	for _, de := range des {
		if de.IsDir() {
			continue
		}
		fullPath := path.Join(dir, de.Name())
		ok, err := MatchPatterns(fullPath, c.patterns...)
		if err != nil {
			return nil, err
		}
		if ok {
			files = append(files, fullPath)
			// break if we're only looking for packages on the first matching file we find.
			if !c.files {
				break
			}
		}
	}
	return
}
