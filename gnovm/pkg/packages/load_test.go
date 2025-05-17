package packages

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload/examplespkgfetcher"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/module"
)

func TestListAndNonDraftPkgs(t *testing.T) {
	for _, tc := range []struct {
		desc string
		in   []struct{ name, modfile string }

		outListPkgs     []string
		outNonDraftPkgs []string
	}{
		{
			desc: "no packages",
		},
		{
			desc: "no package depends on another package",
			in: []struct{ name, modfile string }{
				{
					"foo",
					`module foo`,
				},
				{
					"bar",
					`module bar`,
				},
				{
					"baz",
					`module baz`,
				},
			},
			outListPkgs:     []string{"foo", "bar", "baz"},
			outNonDraftPkgs: []string{"foo", "bar", "baz"},
		},
		{
			desc: "no package depends on draft package",
			in: []struct{ name, modfile string }{
				{
					"foo",
					`module foo`,
				},
				{
					"baz",
					`module baz`,
				},
				{
					"qux",
					`// Draft

					module qux`,
				},
			},
			outListPkgs:     []string{"foo", "baz", "qux"},
			outNonDraftPkgs: []string{"foo", "baz"},
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

			// List packages
			pkgs, err := Load(&LoadConfig{AllowEmpty: true, Fetcher: examplespkgfetcher.New("")}, filepath.Join(dirPath, "..."))
			require.NoError(t, err)
			assert.Equal(t, len(tc.outListPkgs), len(pkgs))
			for _, p := range pkgs {
				assert.Contains(t, tc.outListPkgs, p.ImportPath)
			}

			// Sort packages
			sorted, err := pkgs.Sort(false)
			require.NoError(t, err)

			// Non draft packages
			nonDraft := sorted.GetNonDraftPkgs()
			assert.Equal(t, len(tc.outNonDraftPkgs), len(nonDraft))
			for _, p := range nonDraft {
				assert.Contains(t, tc.outNonDraftPkgs, p.ImportPath)
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
	err = os.WriteFile(filepath.Join(pkgDirPath, "gno.mod"), []byte(modData), 0o644)
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
				{ImportPath: "pkg1", Dir: "/path/to/pkg1", Imports: map[FileKind][]string{}},
				{ImportPath: "pkg2", Dir: "/path/to/pkg2", Imports: map[FileKind][]string{}},
				{ImportPath: "pkg3", Dir: "/path/to/pkg3", Imports: map[FileKind][]string{}},
			},
			expected: []string{"pkg1", "pkg2", "pkg3"},
		}, {
			desc: "circular_dependencies",
			in: []*Package{
				{ImportPath: "pkg1", Dir: "/path/to/pkg1", Imports: map[FileKind][]string{FileKindPackageSource: {"pkg2"}}},
				{ImportPath: "pkg2", Dir: "/path/to/pkg2", Imports: map[FileKind][]string{FileKindPackageSource: {"pkg1"}}},
			},
			shouldErr: true,
		}, {
			desc: "missing_dependencies",
			in: []*Package{
				{ImportPath: "pkg1", Dir: "/path/to/pkg1", Imports: map[FileKind][]string{FileKindPackageSource: {"pkg2"}}},
			},
			shouldErr: true,
		}, {
			desc: "valid_dependencies",
			in: []*Package{
				{ImportPath: "pkg1", Dir: "/path/to/pkg1", Imports: map[FileKind][]string{FileKindPackageSource: {"pkg2"}}},
				{ImportPath: "pkg2", Dir: "/path/to/pkg2", Imports: map[FileKind][]string{FileKindPackageSource: {"pkg3"}}},
				{ImportPath: "pkg3", Dir: "/path/to/pkg3", Imports: map[FileKind][]string{}},
			},
			expected: []string{"pkg3", "pkg2", "pkg1"},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			sorted, err := tc.in.Sort(false)
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

func TestLoadNonDraftExamples(t *testing.T) {
	examples := filepath.Join("..", "..", "..", "examples", "...")
	conf := LoadConfig{
		Deps:    true,
		Fetcher: pkgdownload.NewNoopFetcher(),
	}

	pkgs, err := Load(&conf, examples)
	require.NoError(t, err)

	for _, pkg := range pkgs {
		if pkg.Draft {
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

	tcs := []struct {
		name string
		// workdir          string
		patterns           []string
		conf               *LoadConfig
		res                PkgList
		errShouldContain   string
		ioerrShouldContain string
	}{
		{
			name:     "workspace-1-root",
			patterns: []string{localFromSlash("./testdata/workspace-1")},
			res: PkgList{{
				ImportPath: "gno.example.com/r/wspace1/foo",
				Name:       "foo",
				Root:       filepath.Join(cwd, "testdata", "workspace-1"),
				ModPath:    "gno.example.com/r/wspace1/foo",
				Dir:        filepath.Join(cwd, "testdata", "workspace-1"),
				Match:      []string{localFromSlash("./testdata/workspace-1")},
				Files: FilesMap{
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
			patterns: []string{filepath.Join(cwd, "testdata", "workspace-1")},
			res: PkgList{{
				ImportPath: "gno.example.com/r/wspace1/foo",
				Name:       "foo",
				Root:       filepath.Join(cwd, "testdata", "workspace-1"),
				ModPath:    "gno.example.com/r/wspace1/foo",
				Dir:        filepath.Join(cwd, "testdata", "workspace-1"),
				Match:      []string{filepath.Join(cwd, "testdata", "workspace-1")},
				Files: FilesMap{
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
			patterns: []string{localFromSlash("./testdata/workspace-1/...")},
			res: PkgList{{
				ImportPath: "gno.example.com/r/wspace1/foo",
				Name:       "foo",
				Root:       filepath.Join(cwd, "testdata", "workspace-1"),
				ModPath:    "gno.example.com/r/wspace1/foo",
				Dir:        filepath.Join(cwd, "testdata", "workspace-1"),
				Match:      []string{localFromSlash("./testdata/workspace-1/...")},
				Files: FilesMap{
					FileKindPackageSource: {"foo.gno"},
					FileKindTest:          {"foo_test.gno"},
				},
				Imports: map[FileKind][]string{
					FileKindTest: {"testing"},
				},
			}},
		},
		{
			name:     "workspace-2-root",
			patterns: []string{localFromSlash("./testdata/workspace-2")},
			res: PkgList{{
				Name:  "main",
				Dir:   filepath.Join(cwd, "testdata", "workspace-2"),
				Match: []string{localFromSlash("./testdata/workspace-2")},
				Files: FilesMap{
					FileKindPackageSource: {"lib.gno", "main.gno"},
				},
			}},
		},
		{
			name:     "workspace-2-recursive",
			patterns: []string{localFromSlash("./testdata/workspace-2/...")},
			res: PkgList{{
				Name:  "main",
				Dir:   filepath.Join(cwd, "testdata", "workspace-2"),
				Match: []string{localFromSlash("./testdata/workspace-2/...")},
				Files: FilesMap{
					FileKindPackageSource: {"lib.gno", "main.gno"},
				},
			}, {
				ImportPath: "gno.example.com/r/wspace2/bar",
				Name:       "bar",
				Dir:        filepath.Join(cwd, "testdata", "workspace-2", "bar"),
				Root:       filepath.Join(cwd, "testdata", "workspace-2", "bar"),
				ModPath:    "gno.example.com/r/wspace2/bar",
				Match:      []string{localFromSlash("./testdata/workspace-2/...")},
				Files: FilesMap{
					FileKindPackageSource: {"bar.gno"},
					FileKindXTest:         {"bar_test.gno"},
				},
				Imports: FilesMap{
					FileKindPackageSource: {"gno.example.com/r/wspace2/foo"},
					FileKindXTest:         {"gno.example.com/r/wspace2/bar", "testing"},
				},
			}, {
				ImportPath: "gno.example.com/r/wspace2/bar/baz",
				Name:       "baz",
				Dir:        filepath.Join(cwd, "testdata", "workspace-2", "bar", "baz"),
				Root:       filepath.Join(cwd, "testdata", "workspace-2", "bar"),
				ModPath:    "gno.example.com/r/wspace2/bar",
				Match:      []string{localFromSlash("./testdata/workspace-2/...")},
				Files: FilesMap{
					FileKindPackageSource: {"baz.gno"},
				},
			}, {
				ImportPath: "gno.example.com/r/wspace2/foo",
				ModPath:    "gno.example.com/r/wspace2/foo",
				Name:       "foo",
				Dir:        filepath.Join(cwd, "testdata", "workspace-2", "foo"),
				Root:       filepath.Join(cwd, "testdata", "workspace-2", "foo"),
				Match:      []string{localFromSlash("./testdata/workspace-2/...")},
				Files: FilesMap{
					FileKindPackageSource: {"foo.gno"},
				},
			}},
		},
		{
			name:     "stdlibs",
			patterns: []string{"math/bits"},
			res: PkgList{{
				ImportPath: "math/bits",
				Name:       "bits",
				Dir:        filepath.Join(StdlibDir("math"), "bits"),
				Match:      []string{"math/bits"},
				Files: FilesMap{
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
			patterns: []string{"gno.example.com/p/demo/avl"},
			res: PkgList{{
				ImportPath: "gno.example.com/p/demo/avl",
				Name:       "avl",
				Dir:        gnomod.PackageDir("", module.Version{Path: "gno.example.com/p/demo/avl"}),
				Root:       gnomod.PackageDir("", module.Version{Path: "gno.example.com/p/demo/avl"}),
				ModPath:    "gno.example.com/p/demo/avl",
				Match:      []string{"gno.example.com/p/demo/avl"},
				Files: FilesMap{
					FileKindPackageSource: {"avl.gno"},
				},
			}},
			ioerrShouldContain: `gno: downloading gno.example.com/p/demo/avl`,
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
			name:               "warn-recursive-nothing",
			patterns:           []string{localFromSlash("./testdata/workspace-1/emptydir/...")},
			errShouldContain:   "no packages",
			ioerrShouldContain: fmt.Sprintf(`gno: warning: %q matched no packages`, localFromSlash("./testdata/workspace-1/emptydir/...")),
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			errBuf := &writeCloser{}
			conf := &LoadConfig{Out: errBuf, Fetcher: examplespkgfetcher.New(filepath.Join("testdata", "examples"))}

			res, err := Load(conf, tc.patterns...)

			if tc.errShouldContain == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.errShouldContain)
			}

			if tc.ioerrShouldContain != "" {
				require.Contains(t, errBuf.String(), tc.ioerrShouldContain)
			} else {
				require.Equal(t, errBuf.String(), "")
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
