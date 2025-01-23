package packages

import (
	"testing"

	"github.com/gnolang/gno/gnovm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_LoadWithDeps(t *testing.T) {
	t.Parallel()

	fsresolver := NewFSResolver("./testdata")
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

	pkgA := gnovm.MemPackage{Name: "pkga", Path: commonPath}
	resolverA := NewMockResolver(&pkgA)

	pkgB := gnovm.MemPackage{Name: "pkgb", Path: commonPath}
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
