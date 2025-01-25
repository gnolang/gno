package packages

import (
	"bytes"
	"errors"
	"go/token"
	"log/slog"
	"testing"

	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogMiddleware(t *testing.T) {
	t.Parallel()

	mockResolver := NewMockResolver(&gnovm.MemPackage{
		Path: "abc.xy/test/pkg",
		Name: "pkg",
		Files: []*gnovm.MemFile{
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

	pkg := &gnovm.MemPackage{Path: "abc.xy/cached/pkg", Name: "pkg"}
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
	mockResolver := NewMockResolver(&gnovm.MemPackage{
		Path: "abc.xy/pkg",
		Name: "pkg",
		Files: []*gnovm.MemFile{
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

		pkg, err := filteredResolver.Resolve(token.NewFileSet(), "abc.xy/pkg")
		require.NoError(t, err)
		require.NotNil(t, pkg)
		require.Equal(t, 1, mockResolver.resolveCalls["abc.xy/pkg"])
	})
}

func TestPackageCheckerMiddleware(t *testing.T) {
	t.Parallel()

	logger := log.NewTestingLogger(t)
	t.Run("valid package syntax", func(t *testing.T) {
		t.Parallel()

		validPkg := &gnovm.MemPackage{
			Path: "abc.xy/r/valid/pkg",
			Name: "valid",
			Files: []*gnovm.MemFile{
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

		invalidPkg := &gnovm.MemPackage{
			Path: "abc.xy/r/invalid/pkg",
			Name: "invalid",
			Files: []*gnovm.MemFile{
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

		nonGnoPkg := &gnovm.MemPackage{
			Path: "abc.xy/r/non/gno/pkg",
			Name: "pkg",
			Files: []*gnovm.MemFile{
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
