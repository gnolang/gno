package gnoweb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// countingClient counts the node queries a $download makes.
type countingClient struct {
	*MockClient
	queries atomic.Int64
}

func (c *countingClient) ListFiles(ctx context.Context, path string, height int64) ([]string, error) {
	c.queries.Add(1)
	return c.MockClient.ListFiles(ctx, path, height)
}

func (c *countingClient) File(ctx context.Context, pkgPath, fileName string, height int64) ([]byte, FileMeta, error) {
	c.queries.Add(1)
	return c.MockClient.File(ctx, pkgPath, fileName, height)
}

// TestServePackageText_Cache checks a repeated $download costs no node
// query until the advertised max-age runs out, and a missing package is
// asked again on the next request.
func TestServePackageText_Cache(t *testing.T) {
	t.Parallel()

	client := &countingClient{MockClient: NewMockClient(&MockPackage{
		Domain: "example.com",
		Path:   "/r/mock/path",
		Files:  map[string]string{"a.gno": "package a", "b.gno": "package a // b"},
	})}
	handler, err := NewHTTPHandler(slog.New(slog.DiscardHandler), &HTTPHandlerConfig{
		ClientAdapter: client,
		Renderer:      NewHTMLRenderer(slog.New(slog.DiscardHandler), NewDefaultRenderConfig(), client),
		Aliases:       map[string]AliasTarget{},
	})
	require.NoError(t, err)
	now := time.Unix(0, 0)
	handler.pkgFiles.now = func() time.Time { return now }

	get := func(target string) *httptest.ResponseRecorder {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, target, nil))
		return rr
	}

	rr := get("/r/mock/path$download")
	require.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "// file: b.gno")
	assert.Equal(t, "public, max-age=60", rr.Header().Get("Cache-Control"))
	assert.Equal(t, int64(3), client.queries.Load(), "one list, one query per file")

	rr = get("/r/mock/path$download")
	require.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "// file: b.gno")
	assert.Equal(t, int64(3), client.queries.Load(), "served from the cache")

	now = now.Add(packageTextTTL)
	get("/r/mock/path$download")
	assert.Equal(t, int64(6), client.queries.Load(), "rebuilt once the max-age ran out")

	for range 2 {
		assert.Equal(t, http.StatusNotFound, get("/r/does/not/exist$download").Code)
	}
	assert.Equal(t, int64(8), client.queries.Load(), "errors are not cached")

	// The build is shared, so it does not die with the request that
	// started it.
	now = now.Add(packageTextTTL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/r/mock/path$download", nil).WithContext(ctx))
	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestPackageFiles_SharedWithPlayground checks a playground copy reads the
// package through $download's bounds and cache: a copy after a $download
// costs no node query, and a package over the file cap is refused by both.
func TestPackageFiles_SharedWithPlayground(t *testing.T) {
	t.Parallel()

	files := map[string]string{"a.gno": "package a", "README.md": "# a"}
	big := map[string]string{}
	for i := range packageTextMaxFiles + 1 {
		big[fmt.Sprintf("f%d.gno", i)] = "package big"
	}
	client := &countingClient{MockClient: NewMockClient(
		&MockPackage{Domain: "example.com", Path: "/r/mock/path", Files: files},
		&MockPackage{Domain: "example.com", Path: "/r/mock/big", Files: big},
	)}
	handler, err := NewHTTPHandler(slog.New(slog.DiscardHandler), &HTTPHandlerConfig{
		ClientAdapter: client,
		Meta:          StaticMetadata{Domain: "example.com", RemoteHelp: "127.0.0.1:26657", ChainId: "dev"},
		Renderer:      NewHTMLRenderer(slog.New(slog.DiscardHandler), NewDefaultRenderConfig(), client),
		Aliases:       map[string]AliasTarget{},
	})
	require.NoError(t, err)

	get := func(target string) *httptest.ResponseRecorder {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, target, nil))
		return rr
	}

	require.Equal(t, http.StatusOK, get("/r/mock/path$download").Code)
	queries := client.queries.Load()
	rr := get("/_/play?from=r/mock/path")
	require.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `{"name":"a.gno","content":"package a"}`)
	assert.NotContains(t, rr.Body.String(), "README.md", "a copy holds sources only")
	assert.Equal(t, queries, client.queries.Load(), "the copy is served from $download's cache")

	assert.Equal(t, http.StatusRequestEntityTooLarge, get("/r/mock/big$download").Code)
	assert.Equal(t, http.StatusRequestEntityTooLarge, get("/_/play?from=r/mock/big").Code)
	assert.Equal(t, http.StatusNotFound, get("/_/play?from=r/does/not/exist").Code)
}

func TestPackageTextCache(t *testing.T) {
	t.Parallel()

	t.Run("coalesces concurrent builds", func(t *testing.T) {
		t.Parallel()

		var c packageFileCache
		var builds atomic.Int64
		release := make(chan struct{})
		var wg sync.WaitGroup
		for range 8 {
			wg.Go(func() {
				files, err := c.get("/r/a", func() ([]packageFile, error) {
					builds.Add(1)
					<-release
					return []packageFile{{name: "a.gno", source: []byte("a")}}, nil
				})
				assert.NoError(t, err)
				assert.Equal(t, []packageFile{{name: "a.gno", source: []byte("a")}}, files)
			})
		}
		// Give every caller time to reach the build before it returns; a
		// late one finds the stored text either way.
		time.Sleep(20 * time.Millisecond)
		close(release)
		wg.Wait()
		assert.Equal(t, int64(1), builds.Load())
	})

	t.Run("bounded memory", func(t *testing.T) {
		t.Parallel()

		var c packageFileCache
		big := []packageFile{{name: "a.gno", source: make([]byte, packageTextCacheBytes/2+1)}}
		build := func() ([]packageFile, error) { return big, nil }
		_, err := c.get("/r/a", build)
		require.NoError(t, err)
		_, err = c.get("/r/b", build)
		require.NoError(t, err)
		assert.Len(t, c.entries, 1, "a second big package would pass the cap")
		assert.LessOrEqual(t, c.size, packageTextCacheBytes)
	})

	t.Run("errors are returned, not cached", func(t *testing.T) {
		t.Parallel()

		var c packageFileCache
		boom := errors.New("boom")
		_, err := c.get("/r/a", func() ([]packageFile, error) { return nil, boom })
		assert.ErrorIs(t, err, boom)
		assert.Empty(t, c.entries)
	})

	// A deployed package never shrinks, so the costly rejection is kept.
	t.Run("too large is cached", func(t *testing.T) {
		t.Parallel()

		var c packageFileCache
		var builds int
		build := func() ([]packageFile, error) { builds++; return nil, errPackageTooLarge }
		for range 2 {
			_, err := c.get("/r/a", build)
			assert.ErrorIs(t, err, errPackageTooLarge)
		}
		assert.Equal(t, 1, builds)
	})
}
