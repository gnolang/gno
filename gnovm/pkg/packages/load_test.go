package packages

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload/examplespkgfetcher"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAndNonIgnoredPkgs(t *testing.T) {
	for _, tc := range []struct {
		desc              string
		in                []struct{ name, modfile string }
		errorContains     string
		outListPkgs       []string
		outNonIgnoredPkgs []string
	}{
		{
			desc: "no packages",
		},
		{
			desc: "no package depends on another package",
			in: []struct{ name, modfile string }{
				{
					"foo",
					`module = "foo"`,
				},
				{
					"bar",
					`module = "bar"`,
				},
				{
					"baz",
					`module = "baz"`,
				},
			},
			outListPkgs:       []string{"foo", "bar", "baz"},
			outNonIgnoredPkgs: []string{"foo", "bar", "baz"},
		},
		{
			desc: "no package depends on ignore package",
			in: []struct{ name, modfile string }{
				{
					"foo",
					`module = "foo"`,
				},
				{
					"baz",
					`module = "baz"`,
				},
				{
					"qux",
					`ignore = true
					module = "qux"`,
				},
			},
			outListPkgs:       []string{"foo", "baz", "qux"},
			outNonIgnoredPkgs: []string{"foo", "baz"},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			// Create test dir
			dirPath, cleanUpFn := testutils.NewTestCaseDir(t)
			require.NotNil(t, dirPath)
			defer cleanUpFn()

			// Create packages
			for _, p := range tc.in {
				createGnoModPkg(t, dirPath, p.name, p.modfile)
			}

			err := os.WriteFile(filepath.Join(dirPath, "gnowork.toml"), nil, 0o664)
			require.NoError(t, err)

			testChdir(t, dirPath)

			// List packages
			pkgs, err := Load(LoadConfig{
				AllowEmpty: true,
				Fetcher:    pkgdownload.NewNoopFetcher(),
			}, filepath.Join(dirPath, "..."))
			require.NoError(t, err)

			assert.Equal(t, len(tc.outListPkgs), len(pkgs))
			for _, p := range pkgs {
				assert.Contains(t, tc.outListPkgs, p.ImportPath)
			}

			// Sort packages
			sorted, err := pkgs.Sort()
			require.NoError(t, err)

			// Non draft packages
			nonDraft := sorted.GetNonIgnoredPkgs()
			assert.Equal(t, len(tc.outNonIgnoredPkgs), len(nonDraft))
			for _, p := range nonDraft {
				assert.Contains(t, tc.outNonIgnoredPkgs, p.ImportPath)
			}
		})
	}
}

func createGnoModPkg(t *testing.T, dirPath, pkgName, modData string) {
	t.Helper()

	// Create package dir
	pkgDirPath := filepath.Join(dirPath, pkgName)
	err := os.MkdirAll(pkgDirPath, 0o755)
	require.NoError(t, err)

	// Create gno.mod
	err = os.WriteFile(filepath.Join(pkgDirPath, "gnomod.toml"), []byte(modData), 0o644)
	require.NoError(t, err)
}

func TestSortPkgs(t *testing.T) {
	for _, tc := range []struct {
		desc      string
		in        PkgList
		expected  []string
		shouldErr bool
	}{
		{
			desc:     "empty_input",
			in:       []*Package{},
			expected: make([]string, 0),
		}, {
			desc: "no_dependencies",
			in: []*Package{
				{ImportPath: "test.land/r/pkg1", Dir: "/path/to/pkg1", Imports: map[FileKind][]string{}},
				{ImportPath: "test.land/r/pkg2", Dir: "/path/to/pkg2", Imports: map[FileKind][]string{}},
				{ImportPath: "test.land/r/pkg3", Dir: "/path/to/pkg3", Imports: map[FileKind][]string{}},
			},
			expected: []string{"test.land/r/pkg1", "test.land/r/pkg2", "test.land/r/pkg3"},
		}, {
			desc: "circular_dependencies",
			in: []*Package{
				{ImportPath: "test.land/r/pkg1", Dir: "/path/to/pkg1", Imports: map[FileKind][]string{FileKindPackageSource: {"test.land/r/pkg2"}}},
				{ImportPath: "test.land/r/pkg2", Dir: "/path/to/pkg2", Imports: map[FileKind][]string{FileKindPackageSource: {"test.land/r/pkg1"}}},
			},
			shouldErr: true,
		}, {
			desc: "missing_dependencies",
			in: []*Package{
				{ImportPath: "test.land/r/pkg1", Dir: "/path/to/pkg1", Imports: map[FileKind][]string{FileKindPackageSource: {"test.land/r/pkg2"}}},
			},
			shouldErr: true,
		}, {
			desc: "valid_dependencies",
			in: []*Package{
				{ImportPath: "test.land/r/pkg1", Dir: "/path/to/pkg1", Imports: map[FileKind][]string{FileKindPackageSource: {"test.land/r/pkg2"}}},
				{ImportPath: "test.land/r/pkg2", Dir: "/path/to/pkg2", Imports: map[FileKind][]string{FileKindPackageSource: {"test.land/r/pkg3"}}},
				{ImportPath: "test.land/r/pkg3", Dir: "/path/to/pkg3", Imports: map[FileKind][]string{}},
			},
			expected: []string{"test.land/r/pkg3", "test.land/r/pkg2", "test.land/r/pkg1"},
		}, {
			desc: "stdlib_imports_skipped",
			in: []*Package{
				{ImportPath: "test.land/r/pkg1", Dir: "/path/to/pkg1", Imports: map[FileKind][]string{FileKindPackageSource: {"std", "strings"}}},
			},
			expected: []string{"test.land/r/pkg1"},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			sorted, err := tc.in.Sort()
			if tc.shouldErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				for i := range tc.expected {
					assert.Equal(t, tc.expected[i], sorted[i].ImportPath)
				}
			}
		})
	}
}

func TestLoadNonIgnoredExamples(t *testing.T) {
	examples := filepath.Join("..", "..", "..", "examples")

	testChdir(t, examples)

	conf := LoadConfig{
		Fetcher: pkgdownload.NewNoopFetcher(),
		Deps:    true,
		Test:    true,
	}

	pkgs, err := Load(conf, "./...")
	require.NoError(t, err)

	for _, pkg := range pkgs {
		if pkg.Ignore {
			continue
		}
		require.Empty(t, pkg.Errors)
	}
}

func TestDataLoad(t *testing.T) {
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

	// XXX: this won't guarantee clean state, since we only have one remote test it's okay but we need to fix paralelization
	homeDir := t.TempDir()
	t.Setenv("GNOHOME", homeDir)

	workspace1Abs := filepath.Join(cwd, "testdata", "workspace-1")
	workspace2Abs := filepath.Join(cwd, "testdata", "workspace-2")
	workspace3Abs := filepath.Join(cwd, "testdata", "workspace-3")

	tcs := []struct {
		name             string
		workdir          string
		patterns         []string
		conf             *LoadConfig
		res              PkgList
		errShouldContain string
		outShouldContain string
		deps             bool
	}{
		{
			name:     "workspace-1-root",
			workdir:  localFromSlash("./testdata/workspace-1"),
			patterns: []string{"."},
			res: PkgList{{
				ImportPath: "gno.example.com/r/wspace1/foo",
				Name:       "foo",
				Dir:        workspace1Abs,
				Match:      []string{"."},
				Files: FilesMap{
					FileKindOther:         {"gnomod.toml", "gnowork.toml"},
					FileKindPackageSource: {"foo.gno"},
					FileKindTest:          {"foo_test.gno"},
				},
				Imports: map[FileKind][]string{
					FileKindTest: {"testing"},
				},
			}},
		},
		{
			name:     "workspace-1-root-abs",
			workdir:  localFromSlash("./testdata/workspace-1"),
			patterns: []string{workspace1Abs},
			res: PkgList{{
				ImportPath: "gno.example.com/r/wspace1/foo",
				Name:       "foo",
				Dir:        workspace1Abs,
				Match:      []string{workspace1Abs},
				Files: FilesMap{
					FileKindOther:         {"gnomod.toml", "gnowork.toml"},
					FileKindPackageSource: {"foo.gno"},
					FileKindTest:          {"foo_test.gno"},
				},
				Imports: map[FileKind][]string{
					FileKindTest: {"testing"},
				},
			}},
		},
		{
			name:     "workspace-1-recursive",
			workdir:  localFromSlash("./testdata/workspace-1"),
			patterns: []string{"./..."},
			res: PkgList{{
				ImportPath: "gno.example.com/r/wspace1/foo",
				Name:       "foo",
				Dir:        workspace1Abs,
				Match:      []string{"./..."},
				Files: FilesMap{
					FileKindOther:         {"gnomod.toml", "gnowork.toml"},
					FileKindPackageSource: {"foo.gno"},
					FileKindTest:          {"foo_test.gno"},
				},
				Imports: map[FileKind][]string{
					FileKindTest: {"testing"},
				},
			}, {
				Dir:   filepath.Join(workspace1Abs, "emptygnomod"),
				Match: []string{"./..."},
				Files: FilesMap{},
				Errors: []*Error{{
					Pos: filepath.Join(workspace1Abs, "emptygnomod"),
					Msg: "invalid gnomod.toml: 'module' is required (type: *errors.errorString)",
				}},
			}, {
				ImportPath: "gno.example.com/r/wspace1/invalidpkg",
				Dir:        filepath.Join(workspace1Abs, "invalidpkg"),
				Match:      []string{"./..."},
				Files:      FilesMap{},
				Errors: []*Error{{
					Pos: filepath.Join(workspace1Abs, "invalidpkg"),
					Msg: fmt.Sprintf("%s/b.gno:0: expected package name \"invalidpkga\" but got \"invalidpkgb\" (type: *errors.errorString)", filepath.Join(workspace1Abs, "invalidpkg")),
				}},
			}},
		},
		{
			name:     "workspace-1-root-multi-match",
			workdir:  localFromSlash("./testdata/workspace-1"),
			patterns: []string{"./...", workspace1Abs, filepath.Join(workspace1Abs, "...")},
			res: PkgList{{
				ImportPath: "gno.example.com/r/wspace1/foo",
				Name:       "foo",
				Dir:        workspace1Abs,
				Match:      []string{"./...", workspace1Abs, filepath.Join(workspace1Abs, "...")},
				Files: FilesMap{
					FileKindOther:         {"gnomod.toml", "gnowork.toml"},
					FileKindPackageSource: {"foo.gno"},
					FileKindTest:          {"foo_test.gno"},
				},
				Imports: map[FileKind][]string{
					FileKindTest: {"testing"},
				},
			}, {
				Dir:   filepath.Join(workspace1Abs, "emptygnomod"),
				Match: []string{"./...", filepath.Join(workspace1Abs, "...")},
				Files: FilesMap{},
				Errors: []*Error{{
					Pos: filepath.Join(workspace1Abs, "emptygnomod"),
					Msg: "invalid gnomod.toml: 'module' is required (type: *errors.errorString)",
				}},
			}, {
				ImportPath: "gno.example.com/r/wspace1/invalidpkg",
				Dir:        filepath.Join(workspace1Abs, "invalidpkg"),
				Match:      []string{"./...", filepath.Join(workspace1Abs, "...")},
				Files:      FilesMap{},
				Errors: []*Error{{
					Pos: filepath.Join(workspace1Abs, "invalidpkg"),
					Msg: fmt.Sprintf("%s/b.gno:0: expected package name \"invalidpkga\" but got \"invalidpkgb\" (type: *errors.errorString)", filepath.Join(workspace1Abs, "invalidpkg")),
				}},
			}},
		},
		{
			name:     "workspace-2-root",
			workdir:  localFromSlash("./testdata/workspace-2"),
			patterns: []string{"."},
			res: PkgList{{
				ImportPath: "gno.example.com/r/wspace2",
				Name:       "main",
				Dir:        workspace2Abs,
				Match:      []string{"."},
				Files: FilesMap{
					FileKindOther:         {"gnomod.toml", "gnowork.toml"},
					FileKindPackageSource: {"lib.gno", "main.gno"},
				},
			}},
		},
		{
			name:     "workspace-2-recursive",
			workdir:  localFromSlash("./testdata/workspace-2"),
			patterns: []string{"./..."},
			res: PkgList{{
				ImportPath: "gno.example.com/r/wspace2",
				Name:       "main",
				Dir:        workspace2Abs,
				Match:      []string{"./..."},
				Files: FilesMap{
					FileKindOther:         {"gnomod.toml", "gnowork.toml"},
					FileKindPackageSource: {"lib.gno", "main.gno"},
				},
			}, {
				ImportPath: "gno.example.com/r/wspace2/bar",
				Name:       "bar",
				Dir:        filepath.Join(cwd, "testdata", "workspace-2", "bar"),
				Match:      []string{"./..."},
				Files: FilesMap{
					FileKindOther:         {"gnomod.toml"},
					FileKindPackageSource: {"bar.gno"},
					FileKindXTest:         {"bar_test.gno"},
				},
				Imports: FilesMap{
					FileKindPackageSource: {"gno.example.com/r/wspace2/foo"},
					FileKindXTest:         {"gno.example.com/r/wspace2/bar", "testing"},
				},
			}, {
				ImportPath: "gno.example.com/r/wspace2/foo",
				Name:       "foo",
				Dir:        filepath.Join(cwd, "testdata", "workspace-2", "foo"),
				Match:      []string{"./..."},
				Files: FilesMap{
					FileKindOther:         {"gnomod.toml"},
					FileKindPackageSource: {"foo.gno"},
				},
			}},
		},
		{
			name:             "workspace-3-recursive", // this test that subworkspaces are properly ignored
			workdir:          localFromSlash("./testdata/workspace-3"),
			patterns:         []string{"./..."},
			deps:             true,
			outShouldContain: "gno: downloading gno.example.com/r/wspace3/subwork\ngno: downloading gno.example.com/r/wspace3/subwork/subworkpkg\n",
			res: PkgList{{
				Dir:        workspace3Abs,
				Name:       "wspace3",
				ImportPath: "gno.example.com/r/wspace3",
				Match:      []string{"./..."},
				Files: FilesMap{
					FileKindOther:         []string{"gnomod.toml", "gnowork.toml"},
					FileKindPackageSource: []string{"main.gno"},
				},
				Imports: map[FileKind][]string{
					FileKindPackageSource: {
						"gno.example.com/r/wspace3/subwork",
						"gno.example.com/r/wspace3/subwork/subworkpkg",
					},
				},
			}, {
				Dir:        PackageDir("gno.example.com/r/wspace3/subwork"),
				ImportPath: "gno.example.com/r/wspace3/subwork",
				Files:      FilesMap{},
				Errors: []*Error{{
					Pos: PackageDir("gno.example.com/r/wspace3/subwork"),
					Msg: "query files list for pkg \"gno.example.com/r/wspace3/subwork\": package \"gno.example.com/r/wspace3/subwork\" is not available",
				}},
			}, {
				Dir:        PackageDir("gno.example.com/r/wspace3/subwork/subworkpkg"),
				ImportPath: "gno.example.com/r/wspace3/subwork/subworkpkg",
				Files:      FilesMap{},
				Errors: []*Error{{
					Pos: PackageDir("gno.example.com/r/wspace3/subwork/subworkpkg"),
					Msg: "query files list for pkg \"gno.example.com/r/wspace3/subwork/subworkpkg\": package \"gno.example.com/r/wspace3/subwork/subworkpkg\" is not available",
				}},
			}},
		},
		{
			name:     "stdlibs",
			workdir:  localFromSlash("./testdata/workspace-empty"), // XXX: allow to load stdlibs without a workspace
			patterns: []string{"math/bits"},
			res: PkgList{{
				ImportPath: "math/bits",
				Name:       "bits",
				Dir:        StdlibDir(gnoenv.RootDir(), "math/bits"),
				Match:      []string{"math/bits"},
				Files: FilesMap{
					FileKindOther: {
						"gnomod.toml",
					},
					FileKindPackageSource: {
						"bits.gno",
						"bits_errors.gno",
						"bits_tables.gno",
					},
					FileKindTest: {
						"export_test.gno",
					},
					FileKindXTest: {
						"bits_test.gno",
					},
				},
				Imports: map[FileKind][]string{
					FileKindPackageSource: {"errors"},
					FileKindXTest:         {"math/bits", "testing"},
				},
			}},
		},
		{
			name:     "remote",
			workdir:  localFromSlash("./testdata/workspace-empty"),
			patterns: []string{"gno.example.com/p/demo/avl"},
			res: PkgList{{
				ImportPath: "gno.example.com/p/demo/avl",
				Name:       "avl",
				Dir:        PackageDir("gno.example.com/p/demo/avl"),
				Match:      []string{"gno.example.com/p/demo/avl"},
				Files: FilesMap{
					FileKindOther:         {"gnomod.toml"},
					FileKindPackageSource: {"avl.gno"},
				},
			}},
			outShouldContain: `gno: downloading gno.example.com/p/demo/avl`,
		},
		{
			name:             "err-stdlibs-recursive",
			workdir:          localFromSlash("./testdata/workspace-empty"), // XXX: allow to load stdlibs without a workspace
			patterns:         []string{"strings/..."},
			errShouldContain: "recursive remote patterns are not supported",
		},
		{
			name:             "err-remote-recursive",
			workdir:          localFromSlash("./testdata/workspace-empty"),
			patterns:         []string{"gno.example.com/test/strings/..."},
			errShouldContain: "recursive remote patterns are not supported",
		},
		{
			name:             "err-recursive-noroot",
			workdir:          localFromSlash("./testdata/emptydir"),
			patterns:         []string{"./..."},
			errShouldContain: ErrGnoworkNotFound.Error(),
		},
		{
			name:             "warn-recursive-nothing",
			workdir:          localFromSlash("./testdata/workspace-empty"),
			patterns:         []string{"./..."},
			errShouldContain: "no packages",
			outShouldContain: `gno: warning: "./..." matched no packages`,
		},
	}

	testExamplesAbs, err := filepath.Abs(filepath.Join("testdata", "examples"))
	require.NoError(t, err)

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if tc.workdir != "" {
				testChdir(t, tc.workdir)
			}

			outBuf := &writeCloser{}
			conf := LoadConfig{
				Deps:    tc.deps,
				Out:     outBuf,
				Fetcher: examplespkgfetcher.New(testExamplesAbs),
			}

			res, err := Load(conf, tc.patterns...)

			t.Log("loader output:")
			t.Log(outBuf.String())

			if tc.errShouldContain == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.errShouldContain)
			}

			if tc.outShouldContain != "" {
				require.Contains(t, outBuf.String(), tc.outShouldContain)
			} else {
				require.Equal(t, "", outBuf.String())
			}

			// normalize res
			for _, pkg := range res {
				pkg.ImportsSpecs = nil
				if len(pkg.Errors) == 0 {
					pkg.Errors = nil
				}
				if len(pkg.Imports) == 0 {
					pkg.Imports = nil
				}
			}

			require.EqualValues(t, tc.res, res)
		})
	}
}

type writeCloser struct {
	strings.Builder
}

func (wc *writeCloser) Close() error {
	return nil
}

// port of go1.24 T.Chdir
func testChdir(t *testing.T, dir string) {
	t.Helper()

	oldwd, err := os.Open(".")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	// On POSIX platforms, PWD represents “an absolute pathname of the
	// current working directory.” Since we are changing the working
	// directory, we should also set or update PWD to reflect that.
	switch runtime.GOOS {
	case "windows", "plan9":
		// Windows and Plan 9 do not use the PWD variable.
	default:
		if !filepath.IsAbs(dir) {
			dir, err = os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
		}
		t.Setenv("PWD", dir)
	}
	t.Cleanup(func() {
		err := oldwd.Chdir()
		oldwd.Close()
		if err != nil {
			// It's not safe to continue with tests if we can't
			// get back to the original working directory. Since
			// we are holding a dirfd, this is highly unlikely.
			panic("testing.Chdir: " + err.Error())
		}
	})
}
