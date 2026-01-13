package packages

import (
	"bytes"
	"errors"
	"go/token"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogMiddleware(t *testing.T) {
	t.Parallel()

	mockResolver := NewMockResolver(&std.MemPackage{
		Path: "abc.xy/test/pkg",
		Name: "pkg",
		Files: []*std.MemFile{
			{Name: "file.gno", Body: "package pkg"},
		},
	})

	t.Run("logs package not found", func(t *testing.T) {
		t.Parallel()

		var buff bytes.Buffer

		logger := slog.New(slog.NewTextHandler(&buff, &slog.HandlerOptions{}))
		middleware := LogMiddleware(logger)

		resolver := MiddlewareResolver(mockResolver, middleware)
		pkg, err := resolver.Resolve(token.NewFileSet(), "abc.xy/invalid/pkg")
		require.Error(t, err)
		require.Nil(t, pkg)
		assert.Contains(t, buff.String(), "package not found")
	})

	t.Run("logs package resolution", func(t *testing.T) {
		t.Parallel()

		var buff bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buff, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		middleware := LogMiddleware(logger)

		resolver := MiddlewareResolver(mockResolver, middleware)
		pkg, err := resolver.Resolve(token.NewFileSet(), "abc.xy/test/pkg")
		require.NoError(t, err)
		require.NotNil(t, pkg)
		assert.Contains(t, buff.String(), "path resolved")
	})
}

func TestCacheMiddleware(t *testing.T) {
	t.Parallel()

	pkg := &std.MemPackage{Path: "abc.xy/cached/pkg", Name: "pkg"}
	t.Run("caches resolved packages", func(t *testing.T) {
		t.Parallel()

		mockResolver := NewMockResolver(pkg)
		cacheMiddleware := CacheMiddleware(CacheAll)
		cachedResolver := MiddlewareResolver(mockResolver, cacheMiddleware)

		// First call
		pkg1, err := cachedResolver.Resolve(token.NewFileSet(), pkg.Path)
		require.NoError(t, err)
		require.Equal(t, 1, mockResolver.resolveCalls[pkg.Path])

		// Second call
		pkg2, err := cachedResolver.Resolve(token.NewFileSet(), pkg.Path)
		require.NoError(t, err)
		require.Same(t, pkg1, pkg2)
		require.Equal(t, 1, mockResolver.resolveCalls[pkg.Path])
	})

	t.Run("no cache when shouldCache is false", func(t *testing.T) {
		t.Parallel()

		mockResolver := NewMockResolver(pkg)
		cacheMiddleware := CacheMiddleware(func(*Package) bool { return false })
		cachedResolver := MiddlewareResolver(mockResolver, cacheMiddleware)

		pkg1, err := cachedResolver.Resolve(token.NewFileSet(), pkg.Path)
		require.NoError(t, err)
		pkg2, err := cachedResolver.Resolve(token.NewFileSet(), pkg.Path)
		require.NoError(t, err)
		require.NotSame(t, pkg1, pkg2)
		require.Equal(t, 2, mockResolver.resolveCalls[pkg.Path])
	})
}

func TestFilterStdlibsMiddleware(t *testing.T) {
	t.Parallel()

	middleware := FilterStdlibs
	mockResolver := NewMockResolver(&std.MemPackage{
		Path: "abc.xy/t/pkg",
		Name: "pkg",
		Files: []*std.MemFile{
			{Name: "file.gno", Body: "package pkg"},
		},
	})
	filteredResolver := MiddlewareResolver(mockResolver, middleware)

	t.Run("filters stdlib paths", func(t *testing.T) {
		t.Parallel()

		_, err := filteredResolver.Resolve(token.NewFileSet(), "fmt")
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrResolverPackageSkip))
		require.Equal(t, 0, mockResolver.resolveCalls["fmt"])
	})

	t.Run("allows non-stdlib paths", func(t *testing.T) {
		t.Parallel()

		pkg, err := filteredResolver.Resolve(token.NewFileSet(), "abc.xy/t/pkg")
		require.NoError(t, err)
		require.NotNil(t, pkg)
		require.Equal(t, 1, mockResolver.resolveCalls["abc.xy/t/pkg"])
	})
}

func TestPackageCheckerMiddleware(t *testing.T) {
	t.Parallel()

	logger := log.NewTestingLogger(t)
	t.Run("valid package syntax", func(t *testing.T) {
		t.Parallel()

		validPkg := &std.MemPackage{
			Path: "abc.xy/r/valid/pkg",
			Name: "valid",
			Files: []*std.MemFile{
				{Name: "valid.gno", Body: "package valid; func Foo() {}"},
			},
		}
		mockResolver := NewMockResolver(validPkg)
		middleware := PackageCheckerMiddleware(logger)
		resolver := MiddlewareResolver(mockResolver, middleware)

		pkg, err := resolver.Resolve(token.NewFileSet(), validPkg.Path)
		require.NoError(t, err)
		require.NotNil(t, pkg)
	})

	t.Run("invalid package syntax", func(t *testing.T) {
		t.Parallel()

		invalidPkg := &std.MemPackage{
			Path: "abc.xy/r/invalid/pkg",
			Name: "invalid",
			Files: []*std.MemFile{
				{Name: "invalid.gno", Body: "package invalid\nfunc Foo() {"},
			},
		}
		mockResolver := NewMockResolver(invalidPkg)
		middleware := PackageCheckerMiddleware(logger)
		resolver := MiddlewareResolver(mockResolver, middleware)

		_, err := resolver.Resolve(token.NewFileSet(), invalidPkg.Path)
		require.Error(t, err)
		require.Contains(t, err.Error(), `file "invalid.gno" have error(s)`)
	})

	t.Run("ignores non-gno files", func(t *testing.T) {
		t.Parallel()

		nonGnoPkg := &std.MemPackage{
			Path: "abc.xy/r/non/gno/pkg",
			Name: "pkg",
			Files: []*std.MemFile{
				{Name: "README.md", Body: "# Documentation"},
			},
		}
		mockResolver := NewMockResolver(nonGnoPkg)
		middleware := PackageCheckerMiddleware(logger)
		resolver := MiddlewareResolver(mockResolver, middleware)

		_, err := resolver.Resolve(token.NewFileSet(), nonGnoPkg.Path)
		require.NoError(t, err)
	})
}

func TestResolverLocal_Resolve(t *testing.T) {
	t.Parallel()

	const anotherPath = "abc.xy/t/another/path"

	localResolver := NewLocalResolver(anotherPath, filepath.Join("./testdata", TestdataPkgA))

	t.Run("valid package", func(t *testing.T) {
		t.Parallel()

		pkg, err := localResolver.Resolve(token.NewFileSet(), anotherPath)
		require.NoError(t, err)
		require.NotNil(t, pkg)
		require.Equal(t, pkg.Name, "aa")
	})

	t.Run("invalid package", func(t *testing.T) {
		t.Parallel()

		pkg, err := localResolver.Resolve(token.NewFileSet(), "abc.xy/t/wrong/package")
		require.Nil(t, pkg)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrResolverPackageNotFound)
	})
}

func TestResolver_ResolveRemote(t *testing.T) {
	const targetPath = "gno.land/r/target/path"

	mempkg := std.MemPackage{
		Name: "path",
		Path: targetPath,
		Files: []*std.MemFile{
			{
				Name: "path.gno",
				Body: `package path; func Render(_ string) string { return "bar" }`,
			},
		},
	}
	mempkg.SetFile("gnomod.toml", gnolang.GenGnoModLatest(mempkg.Path))
	mempkg.Sort()

	rootdir := gnoenv.RootDir()
	cfg := integration.TestingMinimalNodeConfig(rootdir)
	logger := log.NewTestingLogger(t)

	// Setup genesis state
	privKey := secp256k1.GenPrivKey()
	cfg.Genesis.AppState = integration.GenerateTestingGenesisState(privKey, mempkg)

	_, address := integration.TestingInMemoryNode(t, logger, cfg)
	cl, err := client.NewHTTPClient(address)
	require.NoError(t, err)

	remoteResolver := NewRemoteResolver(address, cl)
	t.Run("valid package", func(t *testing.T) {
		pkg, err := remoteResolver.Resolve(token.NewFileSet(), mempkg.Path)
		require.NoError(t, err)
		require.NotNil(t, pkg)
		assert.Equal(t, mempkg, pkg.MemPackage)
	})

	t.Run("invalid package", func(t *testing.T) {
		pkg, err := remoteResolver.Resolve(token.NewFileSet(), "gno.land/r/not/a/valid/package")
		require.Nil(t, pkg)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrResolverPackageNotFound)
	})
}

func TestResolverRoot_Resolve(t *testing.T) {
	t.Parallel()

	fsResolver := NewRootResolver("./testdata")
	t.Run("valid packages", func(t *testing.T) {
		t.Parallel()

		for _, tpkg := range []string{TestdataPkgA, TestdataPkgB, TestdataPkgC} {
			t.Run(tpkg, func(t *testing.T) {
				t.Logf("resolving %q", tpkg)
				pkg, err := fsResolver.Resolve(token.NewFileSet(), tpkg)
				require.NoError(t, err)
				require.NotNil(t, pkg)
				require.Equal(t, tpkg, pkg.Path)
				require.Equal(t, filepath.Base(tpkg), pkg.Name)
			})
		}
	})

	t.Run("invalid packages", func(t *testing.T) {
		t.Parallel()

		pkg, err := fsResolver.Resolve(token.NewFileSet(), "abc.xy/wrong/package")
		require.Nil(t, pkg)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrResolverPackageNotFound)
	})
}
