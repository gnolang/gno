package packages

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/stretchr/testify/require"
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
	dir := StdlibDir(gnoenv.RootDir(), "foo/bar")
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

	gnoRoot := gnoenv.RootDir()

	tcs := []struct {
		name              string
		workroot          string
		patterns          []string
		conf              *LoadConfig
		res               []*pkgMatch
		errShouldContain  string
		warnShouldContain string
	}{
		{
			name:     "workspace-1-root",
			workroot: localFromSlash("./testdata/workspace-1"),
			patterns: []string{"."},
			res: []*pkgMatch{{
				Dir:   filepath.Join(cwd, "testdata", "workspace-1"),
				Match: []string{"."},
			}},
		},
		{
			name:     "workspace-1-recursive",
			workroot: localFromSlash("./testdata/workspace-1"),
			patterns: []string{"./..."},
			res: []*pkgMatch{{
				Dir:   filepath.Join(cwd, "testdata", "workspace-1"),
				Match: []string{"./..."},
			}, {
				Dir:   filepath.Join(cwd, "testdata", "workspace-1", "emptygnomod"),
				Match: []string{"./..."},
			}, {
				Dir:   filepath.Join(cwd, "testdata", "workspace-1", "invalidpkg"),
				Match: []string{"./..."},
			}},
		},
		{
			name:     "workspace-1-abs-root",
			workroot: localFromSlash("./testdata/workspace-1"),
			patterns: []string{filepath.Join(cwd, "testdata", "workspace-1")},
			res: []*pkgMatch{{
				Dir:   filepath.Join(cwd, "testdata", "workspace-1"),
				Match: []string{filepath.Join(cwd, "testdata", "workspace-1")},
			}},
		},
		{
			name:     "workspace-1-abs-recursive",
			workroot: localFromSlash("./testdata/workspace-1"),
			patterns: []string{filepath.Join(cwd, "testdata", "workspace-1", "...")},
			res: []*pkgMatch{{
				Dir:   filepath.Join(cwd, "testdata", "workspace-1"),
				Match: []string{filepath.Join(cwd, "testdata", "workspace-1", "...")},
			}, {
				Dir:   filepath.Join(cwd, "testdata", "workspace-1", "emptygnomod"),
				Match: []string{filepath.Join(cwd, "testdata", "workspace-1", "...")},
			}, {
				Dir:   filepath.Join(cwd, "testdata", "workspace-1", "invalidpkg"),
				Match: []string{filepath.Join(cwd, "testdata", "workspace-1", "...")},
			}},
		},
		{
			name:     "workspace-2-root",
			workroot: localFromSlash("./testdata/workspace-2"),
			patterns: []string{localFromSlash(".")},
			res: []*pkgMatch{{
				Dir:   filepath.Join(cwd, "testdata", "workspace-2"),
				Match: []string{localFromSlash(".")},
			}},
		},
		{
			name:     "workspace-2-recursive",
			workroot: localFromSlash("./testdata/workspace-2"),
			patterns: []string{localFromSlash("./...")},
			res: []*pkgMatch{{
				Dir:   filepath.Join(cwd, "testdata", "workspace-2"),
				Match: []string{localFromSlash("./...")},
			}, {
				Dir:   filepath.Join(cwd, "testdata", "workspace-2", "bar"),
				Match: []string{localFromSlash("./...")},
			}, {
				Dir:   filepath.Join(cwd, "testdata", "workspace-2", "foo"),
				Match: []string{localFromSlash("./...")},
			}},
		},
		{
			name:     "workspace-2-recursive-dup",
			workroot: localFromSlash("./testdata/workspace-2"),
			patterns: []string{localFromSlash("./..."), localFromSlash("./bar")},
			res: []*pkgMatch{{
				Dir:   filepath.Join(cwd, "testdata", "workspace-2"),
				Match: []string{localFromSlash("./...")},
			}, {
				Dir:   filepath.Join(cwd, "testdata", "workspace-2", "bar"),
				Match: []string{localFromSlash("./..."), localFromSlash("./bar")},
			}, {
				Dir:   filepath.Join(cwd, "testdata", "workspace-2", "foo"),
				Match: []string{localFromSlash("./...")},
			}},
		},
		{
			name:     "stdlibs",
			workroot: localFromSlash("./testdata/workspace-empty"),
			patterns: []string{"strings", "math"},
			res: []*pkgMatch{{
				Dir:   StdlibDir(gnoRoot, "math"),
				Match: []string{"math"},
			}, {
				Dir:   StdlibDir(gnoRoot, "strings"),
				Match: []string{"strings"},
			}},
		},
		{
			name:     "remote",
			workroot: localFromSlash("./testdata/workspace-empty"),
			patterns: []string{"gno.example.com/test/strings", "gno.example.com/test/math"},
			res: []*pkgMatch{{
				Dir:   PackageDir("gno.example.com/test/math"),
				Match: []string{"gno.example.com/test/math"},
			}, {
				Dir:   PackageDir("gno.example.com/test/strings"),
				Match: []string{"gno.example.com/test/strings"},
			}},
		},
		{
			name:             "err-stdlibs-recursive",
			workroot:         localFromSlash("./testdata/workspace-empty"), // XXX: allow to load stdlibs without a workspace
			patterns:         []string{"strings/..."},
			errShouldContain: "recursive remote patterns are not supported",
		},
		{
			name:             "err-outside-root",
			workroot:         localFromSlash("./testdata/workspace-empty"),
			patterns:         []string{".."},
			errShouldContain: `pattern ".." is not rooted in current workspace`,
		},
		{
			name:             "err-outside-root-abs",
			workroot:         localFromSlash("./testdata/workspace-empty"),
			patterns:         []string{filepath.Join(cwd, "..")},
			errShouldContain: `is not rooted in current workspace`,
		},
		{
			name:             "err-remote-recursive",
			workroot:         localFromSlash("./testdata/workspace-empty"),
			patterns:         []string{"gno.example.com/test/strings/..."},
			errShouldContain: "recursive remote patterns are not supported",
		},
		{
			name:             "err-recursive-noroot",
			workroot:         localFromSlash("./testdata/workspace-empty"),
			patterns:         []string{"./notexists/..."},
			errShouldContain: "no such file or directory",
		},
		{
			name:              "warn-recursive-nothing",
			workroot:          localFromSlash("./testdata/workspace-1"),
			patterns:          []string{localFromSlash("./emptydir/...")},
			warnShouldContain: fmt.Sprintf(`gno: warning: %q matched no packages`, localFromSlash("./emptydir/...")),
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require.NotEmpty(t, tc.workroot)
			workroot, err := filepath.Abs(tc.workroot)
			require.NoError(t, err)
			testChdir(t, workroot)

			warn := &strings.Builder{}
			// TODO: test single-package mode
			res, err := expandPatterns(gnoRoot, &loaderContext{IsWorkspace: true, Root: workroot}, warn, tc.patterns...)
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
