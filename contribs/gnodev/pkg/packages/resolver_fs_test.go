package packages

import (
	"go/token"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolverFS_Resolve(t *testing.T) {
	fsResolver := NewFSResolver("./testdata")

	t.Run("valid packages", func(t *testing.T) {
		for _, tpkg := range []string{TestdataPkgA, TestdataPkgB, TestdataPkgC} {
			t.Run(tpkg, func(t *testing.T) {
				pkg, err := fsResolver.Resolve(token.NewFileSet(), tpkg)
				require.NoError(t, err)
				require.NotNil(t, pkg)
				require.Equal(t, tpkg, pkg.Path)
				require.Equal(t, filepath.Base(tpkg), pkg.Name)
			})
		}
	})

	t.Run("invalid packages", func(t *testing.T) {
		pkg, err := fsResolver.Resolve(token.NewFileSet(), "abc.xy/wrong/package")
		require.Nil(t, pkg)
		require.Error(t, err)
		require.ErrorAs(t, err, &ErrResolverPackageNotFound)
	})
}
