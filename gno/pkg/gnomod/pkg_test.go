package gnomod

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
					"bar",
					`module bar

					require foo v0.0.0`,
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
			outListPkgs:     []string{"foo", "bar", "baz", "qux"},
			outNonDraftPkgs: []string{"foo", "bar", "baz"},
		},
		{
			desc: "package directly depends on draft package",
			in: []struct{ name, modfile string }{
				{
					"foo",
					`// Draft

					module foo`,
				},
				{
					"bar",
					`module bar
					require foo v0.0.0`,
				},
				{
					"baz",
					`module baz`,
				},
			},
			outListPkgs:     []string{"foo", "bar", "baz"},
			outNonDraftPkgs: []string{"baz"},
		},
		{
			desc: "package indirectly depends on draft package",
			in: []struct{ name, modfile string }{
				{
					"foo",
					`// Draft

					module foo`,
				},
				{
					"bar",
					`module bar

					require foo v0.0.0`,
				},
				{
					"baz",
					`module baz

					require bar v0.0.0`,
				},
				{
					"qux",
					`module qux`,
				},
			},
			outListPkgs:     []string{"foo", "bar", "baz", "qux"},
			outNonDraftPkgs: []string{"qux"},
		},
		{
			desc: "package indirectly depends on draft package (multiple levels - 1)",
			in: []struct{ name, modfile string }{
				{
					"foo",
					`// Draft

					module foo`,
				},
				{
					"bar",
					`module bar

					require foo v0.0.0`,
				},
				{
					"baz",
					`module baz

					require bar v0.0.0`,
				},
				{
					"qux",
					`module qux

					require baz v0.0.0`,
				},
			},
			outListPkgs:     []string{"foo", "bar", "baz", "qux"},
			outNonDraftPkgs: []string{},
		},
		{
			desc: "package indirectly depends on draft package (multiple levels - 2)",
			in: []struct{ name, modfile string }{
				{
					"foo",
					`// Draft

					module foo`,
				},
				{
					"bar",
					`module bar

					require qux v0.0.0`,
				},
				{
					"baz",
					`module baz

					require foo v0.0.0`,
				},
				{
					"qux",
					`module qux

					require baz v0.0.0`,
				},
			},
			outListPkgs:     []string{"foo", "bar", "baz", "qux"},
			outNonDraftPkgs: []string{},
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
			pkgs, err := ListPkgs(dirPath)
			require.NoError(t, err)
			assert.Equal(t, len(tc.outListPkgs), len(pkgs))
			for _, p := range pkgs {
				assert.Contains(t, tc.outListPkgs, p.Name)
			}

			// Sort packages
			sorted, err := pkgs.Sort()
			require.NoError(t, err)

			// Non draft packages
			nonDraft := sorted.GetNonDraftPkgs()
			assert.Equal(t, len(tc.outNonDraftPkgs), len(nonDraft))
			for _, p := range nonDraft {
				assert.Contains(t, tc.outNonDraftPkgs, p.Name)
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
			in:       []Pkg{},
			expected: make([]string, 0),
		}, {
			desc: "no_dependencies",
			in: []Pkg{
				{Name: "pkg1", Dir: "/path/to/pkg1", Requires: []string{}},
				{Name: "pkg2", Dir: "/path/to/pkg2", Requires: []string{}},
				{Name: "pkg3", Dir: "/path/to/pkg3", Requires: []string{}},
			},
			expected: []string{"pkg1", "pkg2", "pkg3"},
		}, {
			desc: "circular_dependencies",
			in: []Pkg{
				{Name: "pkg1", Dir: "/path/to/pkg1", Requires: []string{"pkg2"}},
				{Name: "pkg2", Dir: "/path/to/pkg2", Requires: []string{"pkg1"}},
			},
			shouldErr: true,
		}, {
			desc: "missing_dependencies",
			in: []Pkg{
				{Name: "pkg1", Dir: "/path/to/pkg1", Requires: []string{"pkg2"}},
			},
			shouldErr: true,
		}, {
			desc: "valid_dependencies",
			in: []Pkg{
				{Name: "pkg1", Dir: "/path/to/pkg1", Requires: []string{"pkg2"}},
				{Name: "pkg2", Dir: "/path/to/pkg2", Requires: []string{"pkg3"}},
				{Name: "pkg3", Dir: "/path/to/pkg3", Requires: []string{}},
			},
			expected: []string{"pkg3", "pkg2", "pkg1"},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			sorted, err := tc.in.Sort()
			if tc.shouldErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				for i := range tc.expected {
					assert.Equal(t, tc.expected[i], sorted[i].Name)
				}
			}
		})
	}
}
