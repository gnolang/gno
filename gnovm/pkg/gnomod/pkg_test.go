package gnomod

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListPkgs(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		pkg      []struct{ name, modfile string }
		expected []string
	}{
		{
			desc: "no packages",
		},
		{
			desc: "no package depends on another package",
			pkg: []struct{ name, modfile string }{
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
			expected: []string{"foo", "bar", "baz"},
		},
		{
			desc: "no package depends on draft package",
			pkg: []struct{ name, modfile string }{
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
			expected: []string{"foo", "bar", "baz"},
		},
		{
			desc: "package directly depends on draft package",
			pkg: []struct{ name, modfile string }{
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
			expected: []string{"baz"},
		},
		{
			desc: "package indirectly depends on draft package",
			pkg: []struct{ name, modfile string }{
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
			expected: []string{"qux"},
		},
		{
			desc: "package indirectly depends on draft package (multiple levels)",
			pkg: []struct{ name, modfile string }{
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
			expected: []string{},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			// Create test dir
			dirPath, cleanUpFn := testutils.NewTestCaseDir(t)
			require.NotNil(t, dirPath)
			defer cleanUpFn()

			// Create packages
			for _, p := range tc.pkg {
				createGnoModPkg(t, dirPath, p.name, p.modfile)
			}

			// List packages
			pkgs, err := ListPkgs(dirPath)
			require.NoError(t, err)

			for _, p := range pkgs {
				fmt.Println(p.name)
			}

			// Check output
			assert.Equal(t, len(tc.expected), len(pkgs))
			for _, p := range pkgs {
				assert.Contains(t, tc.expected, p.name)
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
		in        []pkg
		expected  []string
		shouldErr bool
	}{
		{
			desc:     "empty_input",
			in:       []pkg{},
			expected: make([]string, 0),
		}, {
			desc: "no_dependencies",
			in: []pkg{
				{name: "pkg1", path: "/path/to/pkg1", requires: []string{}},
				{name: "pkg2", path: "/path/to/pkg2", requires: []string{}},
				{name: "pkg3", path: "/path/to/pkg3", requires: []string{}},
			},
			expected: []string{"pkg1", "pkg2", "pkg3"},
		}, {
			desc: "circular_dependencies",
			in: []pkg{
				{name: "pkg1", path: "/path/to/pkg1", requires: []string{"pkg2"}},
				{name: "pkg2", path: "/path/to/pkg2", requires: []string{"pkg1"}},
			},
			shouldErr: true,
		}, {
			desc: "missing_dependencies",
			in: []pkg{
				{name: "pkg1", path: "/path/to/pkg1", requires: []string{"pkg2"}},
			},
			shouldErr: true,
		}, {
			desc: "valid_dependencies",
			in: []pkg{
				{name: "pkg1", path: "/path/to/pkg1", requires: []string{"pkg2"}},
				{name: "pkg2", path: "/path/to/pkg2", requires: []string{"pkg3"}},
				{name: "pkg3", path: "/path/to/pkg3", requires: []string{}},
			},
			expected: []string{"pkg3", "pkg2", "pkg1"},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := SortPkgs(tc.in)
			if tc.shouldErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				for i := range tc.expected {
					assert.Equal(t, tc.expected[i], tc.in[i].name)
				}
			}
		})
	}
}
