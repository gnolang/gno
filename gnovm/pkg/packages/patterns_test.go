package packages

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/module"
)

func TestPatternKind(t *testing.T) {
	tcs := []struct {
		name             string
		pat              string
		kind             patternKind
		errShouldContain string
	}{{
		name: "single-file",
		pat:  "strings.gno",
		kind: patternKindSingleFile,
	}, {
		name: "abs-single-file",
		pat:  "/strings.gno",
		kind: patternKindSingleFile,
	}, {
		name: "dir",
		pat:  "./strings",
		kind: patternKindDirectory,
	}, {
		name: "dir-recursive",
		pat:  "./strings/...",
		kind: patternKindRecursiveLocal,
	}, {
		name: "parent-dir",
		pat:  "../strings",
		kind: patternKindDirectory,
	}, {
		name: "parent-dir-recursive",
		pat:  "../strings/...",
		kind: patternKindRecursiveLocal,
	}, {
		name: "abs-dir",
		pat:  "/strings",
		kind: patternKindDirectory,
	}, {
		name: "abs-dir-recursive",
		pat:  "/strings/...",
		kind: patternKindRecursiveLocal,
	}, {
		name: "stdlib",
		pat:  "strings",
		kind: patternKindRemote,
	}, {
		name: "stdlib-recursive",
		pat:  "strings/...",
		kind: patternKindRecursiveRemote,
	}, {
		name: "remote",
		pat:  "gno.example.com/r/test/foo",
		kind: patternKindRemote,
	}, {
		name: "remote-recursive",
		pat:  "gno.example.com/r/test/foo/...",
		kind: patternKindRecursiveRemote,
	}, {
		name:             "err-partial-recursive",
		pat:              "./foo/.../bar",
		errShouldContain: "./foo/.../bar: partial globs are not supported",
	}, {
		name:             "err-partial-recursive-2",
		pat:              "./foo/...bar",
		errShouldContain: "./foo/...bar: partial globs are not supported",
	}, {
		name:             "err-partial-remote-recursive",
		pat:              "gno.example.com/r/test/.../foo",
		errShouldContain: "gno.example.com/r/test/.../foo: partial globs are not supported",
	}, {
		name:             "err-partial-stdlib-recursive",
		pat:              "test/.../foo",
		errShouldContain: "test/.../foo: partial globs are not supported",
	}}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			kind, err := getPatternKind(tc.pat)
			if tc.errShouldContain == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.errShouldContain)
			}
			require.Equal(t, tc.kind, kind)
		})
	}
}

func TestStdlibDir(t *testing.T) {
	dir := StdlibDir("foo/bar")
	expectedDir := filepath.Join("..", "..", "stdlibs", "foo", "bar")

	absExpectedDir, err := filepath.Abs(expectedDir)
	require.NoError(t, err)

	require.Equal(t, absExpectedDir, dir)
}

func TestDataExpandPatterns(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	localFromSlash := func(p string) string {
		cp := path.Clean(p)
		fp := filepath.FromSlash(cp)
		if filepath.IsAbs(fp) {
			return fp
		}
		return "." + string(filepath.Separator) + fp
	}

	tcs := []struct {
		name string
		// workdir          string
		patterns          []string
		conf              *LoadConfig
		res               []*pkgMatch
		errShouldContain  string
		warnShouldContain string
	}{
		{
			name:     "workspace-1-root",
			patterns: []string{localFromSlash("./testdata/workspace-1")},
			res: []*pkgMatch{{
				Dir:   filepath.Join(cwd, "testdata", "workspace-1"),
				Match: []string{localFromSlash("./testdata/workspace-1")},
			}},
		},
		{
			name:     "workspace-1-recursive",
			patterns: []string{localFromSlash("./testdata/workspace-1/...")},
			res: []*pkgMatch{{
				Dir:   filepath.Join(cwd, "testdata", "workspace-1"),
				Match: []string{localFromSlash("./testdata/workspace-1/...")},
			}},
		},
		{
			name:     "workspace-1-abs-root",
			patterns: []string{filepath.Join(cwd, "testdata", "workspace-1")},
			res: []*pkgMatch{{
				Dir:   filepath.Join(cwd, "testdata", "workspace-1"),
				Match: []string{filepath.Join(cwd, "testdata", "workspace-1")},
			}},
		},
		{
			name:     "workspace-1-abs-recursive",
			patterns: []string{filepath.Join(cwd, "testdata", "workspace-1", "...")},
			res: []*pkgMatch{{
				Dir:   filepath.Join(cwd, "testdata", "workspace-1"),
				Match: []string{filepath.Join(cwd, "testdata", "workspace-1", "...")},
			}},
		},
		{
			name:     "workspace-2-root",
			patterns: []string{localFromSlash("./testdata/workspace-2")},
			res: []*pkgMatch{{
				Dir:   filepath.Join(cwd, "testdata", "workspace-2"),
				Match: []string{localFromSlash("./testdata/workspace-2")},
			}},
		},
		{
			name:     "workspace-2-recursive",
			patterns: []string{localFromSlash("./testdata/workspace-2/...")},
			res: []*pkgMatch{{
				Dir:   filepath.Join(cwd, "testdata", "workspace-2"),
				Match: []string{localFromSlash("./testdata/workspace-2/...")},
			}, {
				Dir:   filepath.Join(cwd, "testdata", "workspace-2", "bar"),
				Match: []string{localFromSlash("./testdata/workspace-2/...")},
			}, {
				Dir:   filepath.Join(cwd, "testdata", "workspace-2", "bar", "baz"),
				Match: []string{localFromSlash("./testdata/workspace-2/...")},
			}, {
				Dir:   filepath.Join(cwd, "testdata", "workspace-2", "foo"),
				Match: []string{localFromSlash("./testdata/workspace-2/...")},
			}},
		},
		{
			name:     "workspace-2-recursive-dup",
			patterns: []string{localFromSlash("./testdata/workspace-2/..."), localFromSlash("./testdata/workspace-2/bar")},
			res: []*pkgMatch{{
				Dir:   filepath.Join(cwd, "testdata", "workspace-2"),
				Match: []string{localFromSlash("./testdata/workspace-2/...")},
			}, {
				Dir:   filepath.Join(cwd, "testdata", "workspace-2", "bar"),
				Match: []string{localFromSlash("./testdata/workspace-2/..."), localFromSlash("./testdata/workspace-2/bar")},
			}, {
				Dir:   filepath.Join(cwd, "testdata", "workspace-2", "bar", "baz"),
				Match: []string{localFromSlash("./testdata/workspace-2/...")},
			}, {
				Dir:   filepath.Join(cwd, "testdata", "workspace-2", "foo"),
				Match: []string{localFromSlash("./testdata/workspace-2/...")},
			}},
		},
		{
			name:     "stdlibs",
			patterns: []string{"strings", "math"},
			res: []*pkgMatch{{
				Dir:   StdlibDir("math"),
				Match: []string{"math"},
			}, {
				Dir:   StdlibDir("strings"),
				Match: []string{"strings"},
			}},
		},
		{
			name:     "remote",
			patterns: []string{"gno.example.com/test/strings", "gno.example.com/test/math"},
			res: []*pkgMatch{{
				Dir:   gnomod.PackageDir("", module.Version{Path: "gno.example.com/test/math"}),
				Match: []string{"gno.example.com/test/math"},
			}, {
				Dir:   gnomod.PackageDir("", module.Version{Path: "gno.example.com/test/strings"}),
				Match: []string{"gno.example.com/test/strings"},
			}},
		},
		{
			name:             "err-stdlibs-recursive",
			patterns:         []string{"strings/..."},
			errShouldContain: "recursive remote patterns are not supported",
		},
		{
			name:             "err-remote-recursive",
			patterns:         []string{"gno.example.com/test/strings/..."},
			errShouldContain: "recursive remote patterns are not supported",
		},
		{
			name:             "err-recursive-noroot",
			patterns:         []string{"./testdata/notexists/..."},
			errShouldContain: "no such file or directory",
		},
		{
			name:              "warn-recursive-nothing",
			patterns:          []string{localFromSlash("./testdata/workspace-1/emptydir/...")},
			warnShouldContain: fmt.Sprintf(`gno: warning: %q matched no packages`, localFromSlash("./testdata/workspace-1/emptydir/...")),
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			warn := &strings.Builder{}
			res, err := expandPatterns(warn, tc.patterns...)
			if tc.errShouldContain == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.errShouldContain)
			}
			if tc.warnShouldContain != "" {
				require.Contains(t, warn.String(), tc.warnShouldContain)
			} else {
				require.Equal(t, warn.String(), "")
			}
			require.EqualValues(t, tc.res, res)
		})
	}
}
