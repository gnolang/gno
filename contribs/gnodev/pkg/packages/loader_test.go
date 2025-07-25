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

	const commonPath = "abc.yz/t/a"

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
		{"abc.xy/t/**", append(testdataPkgs, testdataNested...)},
		{"abc.xy/t/nested/*", []string{TestdataNestedA}},
		{"abc.xy/t/**/cc", []string{TestdataPkgA, TestdataPkgB, TestdataPkgC, TestdataNestedC}},
		{"abc.xy/t/*/aa", []string{TestdataNestedA}},
	}

	fsresolver := NewRootResolver("./testdata")
	globloader := NewGlobLoader("./testdata", fsresolver)

	for _, tc := range cases {
		t.Run(tc.GlobPath, func(t *testing.T) {
			pkgs, err := globloader.Load(tc.GlobPath)
			require.NoError(t, err)
			actuals := make([]string, 0, len(pkgs))
			for _, pkg := range pkgs {
				actuals = append(actuals, pkg.Path)
			}
			assert.Equal(t, tc.PkgResults, actuals)
		})
	}
}
