package gnomod_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/testutils"
)

func TestListAndNonDraftPkgs(t *testing.T) {
	for _, tc := range []struct {
		desc string
		in   []struct{ name, modfile string }

		outPkgList      []string
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
			outPkgList:      []string{"foo", "bar", "baz"},
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
			outPkgList:      []string{"foo", "baz", "qux"},
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
			pkgs, err := gno.ReadPkgListFromDir(dirPath)
			require.NoError(t, err)
			assert.Equal(t, len(tc.outPkgList), len(pkgs))
			for _, p := range pkgs {
				assert.Contains(t, tc.outPkgList, p.Name)
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
	require.NoError(t, err)
}

func TestSortPkgs(t *testing.T) {
	for _, tc := range []struct {
		desc      string
		in        gnomod.PkgList
		expected  []string
		shouldErr bool
	}{
		{
			desc:     "empty_input",
			in:       []gnomod.Pkg{},
			expected: make([]string, 0),
		}, {
			desc: "no_dependencies",
			in: []gnomod.Pkg{
				{Name: "pkg1", Dir: "/path/to/pkg1", Imports: []string{}},
				{Name: "pkg2", Dir: "/path/to/pkg2", Imports: []string{}},
				{Name: "pkg3", Dir: "/path/to/pkg3", Imports: []string{}},
			},
			expected: []string{"pkg1", "pkg2", "pkg3"},
		}, {
			desc: "circular_dependencies",
			in: []gnomod.Pkg{
				{Name: "pkg1", Dir: "/path/to/pkg1", Imports: []string{"pkg2"}},
				{Name: "pkg2", Dir: "/path/to/pkg2", Imports: []string{"pkg1"}},
			},
			shouldErr: true,
		}, {
			desc: "missing_dependencies",
			in: []gnomod.Pkg{
				{Name: "pkg1", Dir: "/path/to/pkg1", Imports: []string{"pkg2"}},
			},
			shouldErr: true,
		}, {
			desc: "valid_dependencies",
			in: []gnomod.Pkg{
				{Name: "pkg1", Dir: "/path/to/pkg1", Imports: []string{"pkg2"}},
				{Name: "pkg2", Dir: "/path/to/pkg2", Imports: []string{"pkg3"}},
				{Name: "pkg3", Dir: "/path/to/pkg3", Imports: []string{}},
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
