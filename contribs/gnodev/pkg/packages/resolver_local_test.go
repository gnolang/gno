package packages

import (
	"go/token"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolverLocal_Resolve(t *testing.T) {
	const anotherPath = "abc.xy/another/path"
	localResolver := NewLocalResolver(anotherPath, filepath.Join("./testdata", TestdataPkgA))

	t.Run("valid package", func(t *testing.T) {
		pkg, err := localResolver.Resolve(token.NewFileSet(), anotherPath)
		require.NoError(t, err)
		require.NotNil(t, pkg)
		require.Equal(t, pkg.Name, "a")
	})

	t.Run("invalid package", func(t *testing.T) {
		pkg, err := localResolver.Resolve(token.NewFileSet(), "abc.xy/wrong/package")
		require.Nil(t, pkg)
		require.Error(t, err)
		require.ErrorAs(t, err, &ErrResolverPackageNotFound)
	})
}
