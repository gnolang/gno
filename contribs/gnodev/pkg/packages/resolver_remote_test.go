package packages

import (
	"go/token"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_ResolveRemote(t *testing.T) {
	const targetPath = "gno.land/r/target/path"

	mempkg := gnovm.MemPackage{
		Name: "foo",
		Path: targetPath,
		Files: []*gnovm.MemFile{
			{
				Name: "foo.gno",
				Body: `package foo; func Render(_ string) string { return "bar" }`,
			},
			{Name: "gno.mod", Body: `module ` + targetPath},
		},
	}

	rootdir := gnoenv.RootDir()
	cfg := integration.TestingMinimalNodeConfig(rootdir)
	logger := log.NewTestingLogger(t)

	// Setup genesis state
	privKey := secp256k1.GenPrivKey()
	cfg.Genesis.AppState = integration.GenerateTestinGenesisState(privKey, mempkg)

	_, address := integration.TestingInMemoryNode(t, logger, cfg)
	cl, err := client.NewHTTPClient(address)
	require.NoError(t, err)

	remoteResolver := NewRemoteResolver(cl)
	t.Run("valid package", func(t *testing.T) {
		pkg, err := remoteResolver.Resolve(token.NewFileSet(), mempkg.Path)
		require.NoError(t, err)
		require.NotNil(t, pkg)
		assert.Equal(t, pkg.MemPackage, mempkg)
	})

	t.Run("invalid package", func(t *testing.T) {
		pkg, err := remoteResolver.Resolve(token.NewFileSet(), "gno.land/r/not/a/valid/package")
		require.Nil(t, pkg)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrResolverPackageNotFound)
	})
}
