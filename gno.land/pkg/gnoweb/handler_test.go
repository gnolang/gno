package gnoweb_test

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testingLogger struct {
	*testing.T
}

func (t *testingLogger) Write(b []byte) (n int, err error) {
	t.Log(strings.TrimSpace(string(b)))
	return len(b), nil
}

// pureClient is a WebClient stub that always returns exactly one source file.
type pureClient struct {
	stubDirectoryClient
}

func (c *pureClient) Sources(path string) ([]string, error) {
	return []string{"only.gno"}, nil
}

func newTestHandlerConfig(t *testing.T, mockPackage *gnoweb.MockPackage) *gnoweb.WebHandlerConfig {
	t.Helper()

	webclient := gnoweb.NewMockWebClient(mockPackage)

	markdownRenderer := gnoweb.NewMarkdownRenderer(
		log.NewTestingLogger(t),
		gnoweb.NewDefaultMarkdownRendererConfig(nil),
	)

	return &gnoweb.WebHandlerConfig{
		WebClient:        webclient,
		MarkdownRenderer: markdownRenderer,
		Aliases:          map[string]gnoweb.AliasTarget{},
	}
}

// renderFailClient simulates a client that always fails on RenderRealm
// but provides valid paths via QueryPaths.
type renderFailClient struct {
	stubDirectoryClient
}

func (c *renderFailClient) RenderRealm(w io.Writer, u *weburl.GnoURL, cr gnoweb.ContentRenderer) (*gnoweb.RealmMeta, error) {
	return nil, errors.New("render failed")
}

// TestWebHandler_Get tests the Get method of WebHandler using table-driven tests.
func TestWebHandler_Get(t *testing.T) {
	t.Parallel()
	// Set up a mock package with some files and functions
	mockPackage := &gnoweb.MockPackage{
		Domain: "example.com",
		Path:   "/r/mock/path",
		Files: map[string]string{
			"render.gno": `package main; func Render(path string) string { return "one more time" }`,
			"gno.mod":    `module example.com/r/mock/path`,
			"LicEnse":    `my super license`,
		},
		Functions: []*doc.JSONFunc{
			{Name: "SuperRenderFunction", Params: []*doc.JSONField{
				{Name: "my_super_arg", Type: "string"},
			}},
			{
				Name: "Render", Params: []*doc.JSONField{{Name: "path", Type: "string"}},
				Results: []*doc.JSONField{{Name: "", Type: "string"}},
			},
		},
	}

	// Create a WebHandlerConfig with the mock web client and markdown renderer
	config := newTestHandlerConfig(t, mockPackage)

	// Define test cases
	cases := []struct {
		Path     string
		Status   int
		Contain  string   // optional
		Contains []string // optional
	}{
		// Found
		{Path: "/r/mock/path", Status: http.StatusOK, Contain: "[example.com]/r/mock/path"},

		// Source page
		{Path: "/r/mock/path/", Status: http.StatusOK, Contain: "Directory"},
		{Path: "/r/mock/path/render.gno", Status: http.StatusOK, Contain: "one more time"},
		{Path: "/r/mock/path/LicEnse", Status: http.StatusOK, Contain: "my super license"},
		{Path: "/r/mock/path$source&file=render.gno", Status: http.StatusOK, Contain: "one more time"},
		{Path: "/r/mock/path$source", Status: http.StatusOK, Contain: "module"}, // `gno.mod` by default
		{Path: "/r/mock/path/license", Status: http.StatusNotFound},

		// Help page
		{Path: "/r/mock/path$help", Status: http.StatusOK, Contains: []string{
			"my_super_arg",
			"SuperRenderFunction",
		}},

		// Package not found
		{Path: "/r/invalid/path", Status: http.StatusNotFound, Contain: "not found"},

		// Invalid path
		{Path: "/r", Status: http.StatusBadRequest, Contain: "invalid path"},
		{Path: "/r/~!1337", Status: http.StatusNotFound, Contain: "invalid path"},
	}

	for _, tc := range cases {
		t.Run(strings.TrimPrefix(tc.Path, "/"), func(t *testing.T) {
			t.Parallel()
			t.Logf("input: %+v", tc)

			// Initialize testing logger
			logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{}))

			// Create a new WebHandler
			handler, err := gnoweb.NewWebHandler(logger, config)
			require.NoError(t, err)

			// Create a new HTTP request for each test case
			req, err := http.NewRequest(http.MethodGet, tc.Path, nil)
			require.NoError(t, err)

			// Create a ResponseRecorder to capture the response
			rr := httptest.NewRecorder()

			// Invoke serve method
			handler.ServeHTTP(rr, req)

			// Assert result
			assert.Equal(t, tc.Status, rr.Code)
			assert.Containsf(t, rr.Body.String(), tc.Contain, "rendered body should contain: %q", tc.Contain)
			for _, contain := range tc.Contains {
				assert.Containsf(t, rr.Body.String(), contain, "rendered body should contain: %q", contain)
			}
		})
	}
}

// TestWebHandler_NoRender checks if gnoweb displays the `No Render` page properly.
// This happens when the render being queried does not have a Render function declared.
func TestWebHandler_NoRender(t *testing.T) {
	t.Parallel()

	mockPath := "/r/mock/path"
	mockPackage := &gnoweb.MockPackage{
		Domain: "gno.land",
		Path:   "/r/mock/path",
		Files: map[string]string{
			"render.gno": `package main; func init() {}`,
			"gno.mod":    `module gno.land/r/mock/path`,
		},
	}

	// Create a WebHandlerConfig with the mock web client and markdown renderer
	config := newTestHandlerConfig(t, mockPackage)

	logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{}))
	handler, err := gnoweb.NewWebHandler(logger, config)
	require.NoError(t, err, "failed to create WebHandler")

	req, err := http.NewRequest(http.MethodGet, mockPath, nil)
	require.NoError(t, err, "failed to create HTTP request")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "unexpected status code")
	expectedBody := "This realm does not implement a Render() function."
	assert.Contains(t, rr.Body.String(), expectedBody, "rendered body should contain: %q", expectedBody)
}

// TestWebHandler_GetSourceDownload tests the source file download functionality
func TestWebHandler_GetSourceDownload(t *testing.T) {
	t.Parallel()

	mockPackage := &gnoweb.MockPackage{
		Domain: "example.com",
		Path:   "/r/mock/path",
		Files: map[string]string{
			"test.gno": `package main; func main() {}`,
		},
	}

	// Create a WebHandlerConfig with the mock web client and markdown renderer
	config := newTestHandlerConfig(t, mockPackage)

	cases := []struct {
		Path    string
		Status  int
		Contain string
		Headers map[string]string
	}{
		{
			Path:    "/r/mock/path$source&file=test.gno&download",
			Status:  http.StatusOK,
			Contain: "package main",
			Headers: map[string]string{
				"Content-Type":        "text/plain; charset=utf-8",
				"Content-Disposition": `attachment; filename="test.gno"`,
			},
		},
		{
			Path:    "/r/mock/path$source&file=nonexistent.gno&download",
			Status:  http.StatusNotFound,
			Contain: "not found",
		},
		{
			Path:    "/r/mock/path$source&download",
			Status:  http.StatusNotFound,
			Contain: "not found",
		},
		{
			Path:    "/invalid/path$source&file=test.gno&download",
			Status:  http.StatusNotFound,
			Contain: "not found",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(strings.TrimPrefix(tc.Path, "/"), func(t *testing.T) {
			t.Parallel()
			t.Logf("input: %+v", tc)

			logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{}))
			handler, err := gnoweb.NewWebHandler(logger, config)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodGet, tc.Path, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.Status, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.Contain)

			if tc.Headers != nil {
				for k, v := range tc.Headers {
					assert.Equal(t, v, rr.Header().Get(k))
				}
			}
		})
	}
}

func TestWebHandler_DirectoryViewExplorerMode(t *testing.T) {
	mockPackage := &gnoweb.MockPackage{
		Domain: "example.com",
		Path:   "/r/mock/explorer",
		Files: map[string]string{
			"file1.gno": `package main; func main() {}`,
			"file2.gno": `package main; func main() {}`,
		},
	}

	config := newTestHandlerConfig(t, mockPackage)
	logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{}))
	handler, err := gnoweb.NewWebHandler(logger, config)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "/r/mock/explorer/", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Directory")
	assert.Contains(t, rr.Body.String(), "file1.gno")
	assert.Contains(t, rr.Body.String(), "file2.gno")
}

// stubDirectoryClient simulates a client that fails on Sources and, depending on the test,
// returns either paths or an error on QueryPaths.
type stubDirectoryClient struct {
	sourcesErr    error
	queryPaths    []string
	queryPathsErr error
}

func (c *stubDirectoryClient) RenderRealm(w io.Writer, u *weburl.GnoURL, cr gnoweb.ContentRenderer) (*gnoweb.RealmMeta, error) {
	return &gnoweb.RealmMeta{}, nil
}

func (c *stubDirectoryClient) SourceFile(w io.Writer, pkgPath, fileName string, isRaw bool) (*gnoweb.FileMeta, error) {
	return &gnoweb.FileMeta{}, nil
}

func (c *stubDirectoryClient) Doc(path string) (*doc.JSONDocumentation, error) {
	return &doc.JSONDocumentation{Funcs: []*doc.JSONFunc{}}, nil
}

func (c *stubDirectoryClient) Sources(path string) ([]string, error) {
	return nil, c.sourcesErr
}

func (c *stubDirectoryClient) QueryPaths(prefix string, limit int) ([]string, error) {
	return c.queryPaths, c.queryPathsErr
}

// TestWebHandler_DirectoryViewPurePackage covers the pure "package" mode without error:
func TestWebHandler_DirectoryViewPurePackage(t *testing.T) {
	t.Parallel()

	client := &pureClient{}

	cfg := newTestHandlerConfig(t, &gnoweb.MockPackage{
		Domain: "ex",
		Path:   "/p/pkg",
		Files:  map[string]string{},
	})
	cfg.WebClient = client

	handler, err := gnoweb.NewWebHandler(
		slog.New(slog.NewTextHandler(&testingLogger{t}, nil)),
		cfg,
	)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/p/pkg/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "only.gno")
	assert.Contains(t, rr.Body.String(), "/p/pkg/")
}

// TestWebHandler_DirectoryViewErrorTotal covers the case where neither Sources nor QueryPaths return anything:
func TestWebHandler_DirectoryViewErrorTotal(t *testing.T) {
	t.Parallel()
	client := &stubDirectoryClient{
		sourcesErr:    errors.New("fail"),
		queryPaths:    []string{}, // empty
		queryPathsErr: nil,
	}
	cfg := newTestHandlerConfig(t, &gnoweb.MockPackage{Domain: "ex", Path: "/r/y", Files: map[string]string{}})
	cfg.WebClient = client
	handler, err := gnoweb.NewWebHandler(slog.New(slog.NewTextHandler(&testingLogger{t}, nil)), cfg)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/r/y/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// GetClientErrorStatusPage by default should return 500
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "internal error")
}

// TestNewWebHandlerInvalidConfig ensures that NewWebHandler fails on invalid config.
func TestNewWebHandlerInvalidConfig(t *testing.T) {
	t.Parallel()

	dummy := &gnoweb.MockPackage{Domain: "ex", Path: "/r/ex", Files: map[string]string{}}
	valid := newTestHandlerConfig(t, dummy)

	cases := []struct {
		name    string
		mutate  func(cfg *gnoweb.WebHandlerConfig)
		wantErr string
	}{
		{
			name: "missing WebClient",
			mutate: func(cfg *gnoweb.WebHandlerConfig) {
				cfg.WebClient = nil
			},
			wantErr: "no `WebClient` configured",
		},
		{
			name: "missing MarkdownRenderer",
			mutate: func(cfg *gnoweb.WebHandlerConfig) {
				cfg.MarkdownRenderer = nil
			},
			wantErr: "no `MarkdownRenderer` configured",
		},
		{
			name: "missing Aliases",
			mutate: func(cfg *gnoweb.WebHandlerConfig) {
				cfg.Aliases = nil
			},
			wantErr: "no `Aliases` configured",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Duplicate the valid config and mutate the field
			cfg := *valid
			tc.mutate(&cfg)

			logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{}))
			_, err := gnoweb.NewWebHandler(logger, &cfg)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// TestIsHomePath covers the utility function.
func TestIsHomePath(t *testing.T) {
	assert.True(t, gnoweb.IsHomePath("/"))
	assert.False(t, gnoweb.IsHomePath("/foo"))
}

// TestServeHTTPMethodNotAllowed verifies 405 for POST/PUT/etc methods.
func TestServeHTTPMethodNotAllowed(t *testing.T) {
	t.Parallel()

	dummy := &gnoweb.MockPackage{Domain: "ex", Path: "/r/ex", Files: map[string]string{}}
	cfg := newTestHandlerConfig(t, dummy)
	logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{Level: slog.LevelDebug}))
	handler, err := gnoweb.NewWebHandler(logger, cfg)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/r/ex", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Contains(t, rr.Body.String(), "method not allowed")
}

// TestWebHandler_DirectoryViewNoFiles covers the case where Sources returns
// no error but the list is empty (len(files)==0).
func TestWebHandler_DirectoryViewNoFiles(t *testing.T) {
	t.Parallel()

	// stub that doesn't error on Sources, but returns an empty slice
	client := &stubDirectoryClient{
		sourcesErr:    nil,
		queryPaths:    []string{"shouldNotBeUsed"},
		queryPathsErr: nil,
	}

	cfg := newTestHandlerConfig(t, &gnoweb.MockPackage{
		Domain: "ex",
		Path:   "/r/empty",
		Files:  map[string]string{},
	})
	cfg.WebClient = client

	handler, err := gnoweb.NewWebHandler(
		slog.New(slog.NewTextHandler(&testingLogger{t}, nil)),
		cfg,
	)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/r/empty/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// We expect a 200 with the error component "no files available"
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "no files available")
}

// TestWebHandler_GetSourceView_Error covers the `if err != nil` branch of GetSourceView.
func TestWebHandler_GetSourceView_Error(t *testing.T) {
	t.Parallel()

	// stubDirectoryClient implements Sources with an error
	client := &stubDirectoryClient{
		sourcesErr:    errors.New("fail listing sources"),
		queryPaths:    nil,
		queryPathsErr: nil,
	}

	cfg := newTestHandlerConfig(t, &gnoweb.MockPackage{
		Domain: "ex",
		Path:   "/r/errsrc",
		Files:  map[string]string{},
	})
	cfg.WebClient = client

	handler, err := gnoweb.NewWebHandler(
		slog.New(slog.NewTextHandler(&testingLogger{t}, nil)),
		cfg,
	)
	require.NoError(t, err)

	// We use the URL that triggers GetSourceView
	req := httptest.NewRequest(http.MethodGet, "/r/errsrc$source", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Should be 500 + internal error
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "internal error")
}

// TestWebHandler_GetSourceView_NoFiles covers the `if len(files)==0` of GetSourceView.
func TestWebHandler_GetSourceView_NoFiles(t *testing.T) {
	t.Parallel()

	// stubDirectoryClient implements Sources without error but returns nil slice
	client := &stubDirectoryClient{
		sourcesErr:    nil,
		queryPaths:    nil,
		queryPathsErr: nil,
	}

	cfg := newTestHandlerConfig(t, &gnoweb.MockPackage{
		Domain: "ex",
		Path:   "/r/emptysrc",
		Files:  map[string]string{},
	})
	cfg.WebClient = client

	handler, err := gnoweb.NewWebHandler(
		slog.New(slog.NewTextHandler(&testingLogger{t}, nil)),
		cfg,
	)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/r/emptysrc$source", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Should be 200 + "no files available"
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "no files available")
}

func TestGetClientErrorStatusPage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		err      error
		wantCode int
		wantView bool
		wantMsg  string
	}{
		{
			name:     "nil error",
			err:      nil,
			wantCode: http.StatusOK,
			wantView: false,
		},
		{
			name:     "path not found",
			err:      gnoweb.ErrClientPathNotFound,
			wantCode: http.StatusNotFound,
			wantView: true,
			wantMsg:  gnoweb.ErrClientPathNotFound.Error(),
		},
		{
			name:     "bad request",
			err:      gnoweb.ErrClientBadRequest,
			wantCode: http.StatusInternalServerError,
			wantView: true,
			wantMsg:  "bad request",
		},
		{
			name:     "response error",
			err:      gnoweb.ErrClientResponse,
			wantCode: http.StatusInternalServerError,
			wantView: true,
			wantMsg:  "internal error",
		},
		{
			name:     "other error",
			err:      errors.New("foo"),
			wantCode: http.StatusInternalServerError,
			wantView: true,
			wantMsg:  "internal error",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			code, view := gnoweb.GetClientErrorStatusPage(nil, tc.err)
			assert.Equal(t, tc.wantCode, code)

			if !tc.wantView {
				assert.Nil(t, view)
				return
			}
			require.NotNil(t, view)

			// Render the component and check its output contains the expected message
			var buf bytes.Buffer
			err := view.Render(&buf)
			require.NoError(t, err)
			assert.Contains(t, buf.String(), tc.wantMsg)
		})
	}
}
