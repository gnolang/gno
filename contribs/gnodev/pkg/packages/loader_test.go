package packages

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_LoadWithDeps(t *testing.T) {
	t.Parallel()

	fsresolver := NewRootResolver("./testdata")
	loader := NewLoader(fsresolver)

	// package c depend on package b
	pkgs, err := loader.Load(TestdataPkgC)
	require.NoError(t, err)
	require.Len(t, pkgs, 3)
	for i, path := range []string{TestdataPkgA, TestdataPkgB, TestdataPkgC} {
		assert.Equal(t, path, pkgs[i].Path)
	}
}

func TestLoader_ResolverPriority(t *testing.T) {
	t.Parallel()

	const commonPath = "abc.yz/pkg/a"

	pkgA := std.MemPackage{Name: "pkga", Path: commonPath}
	resolverA := NewMockResolver(&pkgA)

	pkgB := std.MemPackage{Name: "pkgb", Path: commonPath}
	resolverB := NewMockResolver(&pkgB)

	t.Run("pkgA then pkgB", func(t *testing.T) {
		t.Parallel()

		loader := NewLoader(resolverA, resolverB)
		pkg, err := loader.Resolve(commonPath)
		require.NoError(t, err)
		require.Equal(t, pkgA.Name, pkg.Name)
		require.Equal(t, commonPath, pkg.Path)
	})

	t.Run("pkgB then pkgA", func(t *testing.T) {
		t.Parallel()

		loader := NewLoader(resolverB, resolverA)
		pkg, err := loader.Resolve(commonPath)
		require.NoError(t, err)
		require.Equal(t, pkgB.Name, pkg.Name)
		require.Equal(t, commonPath, pkg.Path)
	})
}

func TestLoader_Glob(t *testing.T) {
	const root = "./testdata"
	cases := []struct {
		GlobPath   string
		PkgResults []string
	}{
		{"abc.xy/pkg/*", []string{TestdataPkgA, TestdataPkgB, TestdataPkgC}},
		{"abc.xy/nested/*", []string{TestdataNestedA}},
		{"abc.xy/**/cc", []string{TestdataNestedC, TestdataPkgA, TestdataPkgB, TestdataPkgC}},
		{"abc.xy/*/aa", []string{TestdataNestedA, TestdataPkgA}},
	}

	fsresolver := NewRootResolver("./testdata")
	globloader := NewGlobLoader("./testdata", fsresolver)

	for _, tc := range cases {
		t.Run(tc.GlobPath, func(t *testing.T) {
			pkgs, err := globloader.Load(tc.GlobPath)
			require.NoError(t, err)
			require.Len(t, pkgs, len(tc.PkgResults))
			for i, expected := range tc.PkgResults {
				assert.Equal(t, expected, pkgs[i].Path)
			}
		})
	}
}
