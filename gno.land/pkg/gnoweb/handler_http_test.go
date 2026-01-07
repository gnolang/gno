package gnoweb_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	md "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testingLogger struct {
	*testing.T
}

func (t *testingLogger) Write(b []byte) (n int, err error) {
	t.T.Log(strings.TrimSpace(string(b)))
	return len(b), nil
}

// Top-level stubClient definition for use in error simulation/custom behavior tests
// stubClient simulates a client that can be customized per test by setting function fields.
type stubClient struct {
	realmFunc     func(ctx context.Context, path, args string) ([]byte, error)
	fileFunc      func(ctx context.Context, path, filename string) ([]byte, gnoweb.FileMeta, error)
	docFunc       func(ctx context.Context, path string) (*doc.JSONDocumentation, error)
	listFilesFunc func(ctx context.Context, path string) ([]string, error)
	listPathsFunc func(ctx context.Context, prefix string, limit int) ([]string, error)
}

func (s *stubClient) Realm(ctx context.Context, path, args string) ([]byte, error) {
	if s.realmFunc != nil {
		return s.realmFunc(ctx, path, args)
	}
	return nil, errors.New("stubClient: Realm not implemented")
}

func (s *stubClient) File(ctx context.Context, path, filename string) ([]byte, gnoweb.FileMeta, error) {
	if s.fileFunc != nil {
		return s.fileFunc(ctx, path, filename)
	}
	return nil, gnoweb.FileMeta{}, errors.New("stubClient: File not implemented")
}

func (s *stubClient) Doc(ctx context.Context, path string) (*doc.JSONDocumentation, error) {
	if s.docFunc != nil {
		return s.docFunc(ctx, path)
	}
	return nil, errors.New("stubClient: Doc not implemented")
}

func (s *stubClient) ListFiles(ctx context.Context, path string) ([]string, error) {
	if s.listFilesFunc != nil {
		return s.listFilesFunc(ctx, path)
	}
	return nil, errors.New("stubClient: ListFiles not implemented")
}

func (s *stubClient) ListPaths(ctx context.Context, prefix string, limit int) ([]string, error) {
	if s.listPathsFunc != nil {
		return s.listPathsFunc(ctx, prefix, limit)
	}
	return nil, errors.New("stubClient: ListPaths not implemented")
}

type rawRenderer struct{}

func (rawRenderer) RenderRealm(w io.Writer, u *weburl.GnoURL, src []byte, ctx gnoweb.RealmRenderContext) (md.Toc, error) {
	_, err := w.Write(src)
	return md.Toc{}, err
}

func (rawRenderer) RenderSource(w io.Writer, name string, src []byte) error {
	_, err := w.Write(src)
	return err
}

// newTestHandlerConfig creates a HTTPHandlerConfig for tests using a stub client.
func newTestHandlerConfig(t *testing.T, client gnoweb.ClientAdapter) *gnoweb.HTTPHandlerConfig {
	t.Helper()

	return &gnoweb.HTTPHandlerConfig{
		ClientAdapter: client,
		Renderer:      &rawRenderer{},
		Aliases:       map[string]gnoweb.AliasTarget{},
	}
}

// TestHTTPHandler_Get tests the Get method of WebHandler using table-driven tests.
func TestHTTPHandler_Get(t *testing.T) {
	t.Parallel()

	mockPackage := &gnoweb.MockPackage{
		Domain: "example.com",
		Path:   "/r/mock/path",
		Files: map[string]string{
			"render.gno": `package main; func Render(path string) string { return "one more time" }`,
			"gno.mod":    `module example.com/r/mock/path`,
			"LicEnse":    `my super license`,
		},
		Functions: []*doc.JSONFunc{
			{Name: "SuperRenderFunction", Params: []*doc.JSONField{{Name: "my_super_arg", Type: "string"}}},
			{Name: "Render", Params: []*doc.JSONField{{Name: "path", Type: "string"}}, Results: []*doc.JSONField{{Name: "", Type: "string"}}},
		},
	}

	config := newTestHandlerConfig(t, gnoweb.NewMockClient(mockPackage))

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
		{Path: "/~!1337", Status: http.StatusNotFound, Contain: "invalid path"},
	}

	for _, tc := range cases {
		t.Run(strings.TrimPrefix(tc.Path, "/"), func(t *testing.T) {
			t.Parallel()
			t.Logf("input: %+v", tc)

			// Initialize testing logger
			logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{}))

			// Create a new WebHandler
			handler, err := gnoweb.NewHTTPHandler(logger, config)
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

// TestHTTPHandler_NoRender checks if gnoweb displays the `No Render` page properly.
// This happens when the render being queried does not have a Render function declared.
func TestHTTPHandler_NoRender(t *testing.T) {
	t.Parallel()

	mockPackage := &gnoweb.MockPackage{
		Domain: "gno.land",
		Path:   "/r/mock/path",
		Files: map[string]string{
			"render.gno": `package main; func init() {}`,
			"gno.mod":    `module gno.land/r/mock/path`,
		},
		Functions: []*doc.JSONFunc{}, // No Render function
	}

	config := newTestHandlerConfig(t, gnoweb.NewMockClient(mockPackage))

	logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{}))
	handler, err := gnoweb.NewHTTPHandler(logger, config)
	require.NoError(t, err, "failed to create WebHandler")

	mockPath := "/r/mock/path"
	req, err := http.NewRequest(http.MethodGet, mockPath, nil)
	require.NoError(t, err, "failed to create HTTP request")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "unexpected status code")
	assert.Contains(t, rr.Body.String(), "gno.mod", "rendered body should contain the file list (gno.mod)")
	assert.Contains(t, rr.Body.String(), "render.gno", "rendered body should contain the file list (render.gno)")
}

// TestHTTPHandler_GetSourceDownload tests the source file download functionality
func TestHTTPHandler_GetSourceDownload(t *testing.T) {
	t.Parallel()

	mockPackage := &gnoweb.MockPackage{
		Domain: "example.com",
		Path:   "/r/mock/path",
		Files: map[string]string{
			"test.gno": `package main; func main() {}`,
		},
	}

	config := newTestHandlerConfig(t, gnoweb.NewMockClient(mockPackage))

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
			handler, err := gnoweb.NewHTTPHandler(logger, config)
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

func TestHTTPHandler_DirectoryViewExplorerMode(t *testing.T) {
	mockPackage := &gnoweb.MockPackage{
		Domain: "example.com",
		Path:   "/r/mock/explorer",
		Files: map[string]string{
			"file1.gno": `package main; func main() {}`,
			"file2.gno": `package main; func main() {}`,
		},
	}

	config := newTestHandlerConfig(t, gnoweb.NewMockClient(mockPackage))
	logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{}))
	handler, err := gnoweb.NewHTTPHandler(logger, config)
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

// TestHTTPHandler_DirectoryViewPurePackage covers the pure "package" mode without error:
func TestHTTPHandler_DirectoryViewPurePackage(t *testing.T) {
	t.Parallel()

	mockPackage := &gnoweb.MockPackage{
		Domain: "ex",
		Path:   "/p/pkg",
		Files: map[string]string{
			"only.gno": "package only;",
		},
	}

	cfg := newTestHandlerConfig(t, gnoweb.NewMockClient(mockPackage))

	handler, err := gnoweb.NewHTTPHandler(
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

// TestHTTPHandler_DirectoryViewErrorTotal covers the case where neither Sources nor QueryPaths return anything:
func TestHTTPHandler_DirectoryViewErrorTotal(t *testing.T) {
	t.Parallel()

	// For error simulation tests, instantiate the top-level stubClient and set the relevant function fields for each test. Do not redeclare methods or types inside the test functions.
	client := &stubClient{}
	cfg := newTestHandlerConfig(t, client)
	handler, err := gnoweb.NewHTTPHandler(slog.New(slog.NewTextHandler(&testingLogger{t}, nil)), cfg)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/r/y/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// GetClientErrorStatusPage by default should return 500
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "internal error")
}

// TestHTTPHandler_RealmExplorerWithRender tests realms with Render() show realm icon and Source button.
func TestHTTPHandler_RealmExplorerWithRender(t *testing.T) {
	t.Parallel()

	realmWithRender := &gnoweb.MockPackage{
		Domain: "gno.land",
		Path:   "/r/demo/withrender",
		Files:  map[string]string{"render.gno": `package withrender`},
		Functions: []*doc.JSONFunc{{
			Name:    "Render",
			Params:  []*doc.JSONField{{Name: "path", Type: "string"}},
			Results: []*doc.JSONField{{Type: "string"}},
		}},
	}

	handler, _ := gnoweb.NewHTTPHandler(slog.New(slog.NewTextHandler(&testingLogger{t}, nil)), newTestHandlerConfig(t, gnoweb.NewMockClient(realmWithRender)))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/r/demo/", nil))

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "#ico-realm")
	assert.Contains(t, rr.Body.String(), "Source")
}

// TestNewWebHandlerInvalidConfig ensures that NewWebHandler fails on invalid config.
func TestHTTPHandler_NewInvalidConfig(t *testing.T) {
	t.Parallel()

	minimalMock := gnoweb.NewMockClient(&gnoweb.MockPackage{Path: "/", Files: map[string]string{}})
	valid := newTestHandlerConfig(t, minimalMock)

	cases := []struct {
		name   string
		mutate func(cfg *gnoweb.HTTPHandlerConfig)
	}{
		{
			name: "missing Client",
			mutate: func(cfg *gnoweb.HTTPHandlerConfig) {
				cfg.ClientAdapter = nil
			},
		},
		{
			name: "missing Renderer",
			mutate: func(cfg *gnoweb.HTTPHandlerConfig) {
				cfg.Renderer = nil
			},
		},
		{
			name: "missing Aliases",
			mutate: func(cfg *gnoweb.HTTPHandlerConfig) {
				cfg.Aliases = nil
			},
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
			_, err := gnoweb.NewHTTPHandler(logger, &cfg)
			require.Error(t, err)
		})
	}
}

// TestServeHTTPMethodNotAllowed verifies 405 for HTTP methods.
func TestHTTPHandler_ServeHTTPMethodNotAllowed(t *testing.T) {
	t.Parallel()

	minimalMock := gnoweb.NewMockClient(&gnoweb.MockPackage{Path: "/", Files: map[string]string{}})
	cfg := newTestHandlerConfig(t, minimalMock)
	logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{Level: slog.LevelDebug}))
	handler, err := gnoweb.NewHTTPHandler(logger, cfg)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/r/ex", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Contains(t, rr.Body.String(), "method not allowed")
}

// TestHTTPHandler_DirectoryViewNoFiles covers the case where Sources returns
// no error but the list is empty (len(files)==0).
func TestHTTPHandler_DirectoryViewNoFiles(t *testing.T) {
	t.Parallel()

	mockPackage := &gnoweb.MockPackage{
		Domain: "ex",
		Path:   "/r/empty",
		Files:  map[string]string{},
	}

	cfg := newTestHandlerConfig(t, gnoweb.NewMockClient(mockPackage))

	handler, err := gnoweb.NewHTTPHandler(
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

// TestHTTPHandler_GetSourceView_Error covers the `if err != nil` branch of GetSourceView.
func TestHTTPHandler_GetSourceView_Error(t *testing.T) {
	t.Parallel()

	// For error simulation tests, instantiate the top-level stubClient and set the relevant function fields for each test. Do not redeclare methods or types inside the test functions.
	client := &stubClient{}

	cfg := newTestHandlerConfig(t, client)

	handler, err := gnoweb.NewHTTPHandler(
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

// TestHTTPHandler_GetSourceView_NoFiles covers the `if len(files)==0` of GetSourceView.
func TestHTTPHandler_GetSourceView_NoFiles(t *testing.T) {
	t.Parallel()

	mockPackage := &gnoweb.MockPackage{
		Domain: "ex",
		Path:   "/r/emptysrc",
		Files:  map[string]string{},
	}

	cfg := newTestHandlerConfig(t, gnoweb.NewMockClient(mockPackage))

	handler, err := gnoweb.NewHTTPHandler(
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

func TestHTTPHandler_GetClientErrorStatusPage(t *testing.T) {
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
			err:      gnoweb.ErrClientPackageNotFound,
			wantCode: http.StatusNotFound,
			wantView: true,
			wantMsg:  gnoweb.ErrClientPackageNotFound.Error(),
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

func TestHTTPHandler_GetUserView(t *testing.T) {
	t.Parallel()

	client := &stubClient{
		listPathsFunc: func(ctx context.Context, prefix string, limit int) ([]string, error) {
			return []string{
				"/r/testuser/pkg1", "/r/testuser/pkg2",
			}, nil
		},
		realmFunc: func(ctx context.Context, path string, args string) ([]byte, error) {
			if path != "/r/testuser/home" {
				return nil, fmt.Errorf("unknown path")
			}

			return []byte("# Welcome to testuser's profile"), nil
		},
	}

	cfg := newTestHandlerConfig(t, client)

	handler, err := gnoweb.NewHTTPHandler(
		slog.New(slog.NewTextHandler(&testingLogger{t}, nil)),
		cfg,
	)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/u/testuser", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	body := rr.Body.String()

	// The content from RenderRealm
	assert.Contains(t, body, "Welcome to testuser's profile")
	// The contributions
	assert.Contains(t, body, "pkg1")
	assert.Contains(t, body, "pkg2")
	// The username should be visible
	assert.Contains(t, body, "testuser")
}

func TestHTTPHandler_GetUserView_QueryPathsError(t *testing.T) {
	t.Parallel()

	client := &stubClient{
		listPathsFunc: func(ctx context.Context, prefix string, limit int) ([]string, error) {
			return nil, errors.New("fail to list paths")
		},
		realmFunc: func(ctx context.Context, path string, args string) ([]byte, error) {
			if path != "/r/testuser/home" {
				return nil, fmt.Errorf("unknown path")
			}

			return []byte("# Welcome to testuser's profile"), nil
		},
	}

	cfg := newTestHandlerConfig(t, client)

	handler, err := gnoweb.NewHTTPHandler(
		slog.New(slog.NewTextHandler(&testingLogger{t}, nil)),
		cfg,
	)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/u/testuser", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Should be 500 + internal error
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "internal error")
}

func TestHTTPHandler_CreateUsernameFromBech32(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid bech32 address",
			input:    "g1edq4dugw0sgat4zxcw9xardvuydqf6cgleuc8p",
			expected: "g1ed...uc8p",
		},
		{
			name:     "invalid bech32 address",
			input:    "invalid-address",
			expected: "invalid-address",
		},
		{
			name:     "empty address",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := gnoweb.CreateUsernameFromBech32(tt.input)
			assert.Equal(t, tt.expected, result, "CreateUsernameFromBech32(%q) = %q, want %q", tt.input, result, tt.expected)
		})
	}
}

// TestHTTPHandler_GetSourceView_FilePreference tests the file preference logic
// when no specific file is requested in the source view.
func TestHTTPHandler_GetSourceView_FilePreference(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		files          []string
		expectedFile   string
		expectedStatus int
	}{
		{
			name:           "prefer README.md over other files",
			files:          []string{"config.toml", "README.md", "main.gno"},
			expectedFile:   "README.md",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "prefer .gno file when no README.md",
			files:          []string{"config.toml", "main.gno", "test.toml"},
			expectedFile:   "main.gno",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "fallback to first file when no preferred files",
			files:          []string{"config.toml", "test.toml", "data.json"},
			expectedFile:   "config.toml",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "prefer first .gno file when multiple .gno files",
			files:          []string{"config.toml", "main.gno", "utils.gno", "test.gno"},
			expectedFile:   "main.gno",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := &stubClient{
				listFilesFunc: func(ctx context.Context, path string) ([]string, error) {
					return tc.files, nil
				},
				fileFunc: func(ctx context.Context, path string, filename string) ([]byte, gnoweb.FileMeta, error) {
					if slices.Contains(tc.files, filename) {
						content := fmt.Sprintf("content of %s", filename)
						return []byte(content), gnoweb.FileMeta{}, nil
					}

					return nil, gnoweb.FileMeta{}, gnoweb.ErrClientFileNotFound
				},
			}

			config := newTestHandlerConfig(t, client)
			handler, err := gnoweb.NewHTTPHandler(
				slog.New(slog.NewTextHandler(&testingLogger{t}, nil)),
				config,
			)
			handler.Renderer = &rawRenderer{}

			require.NoError(t, err)

			// Request source view without specifying a file
			req, err := http.NewRequest(http.MethodGet, "/r/test/path$source", nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Check status
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// Check that the expected file content is displayed
			expectedContent := "content of " + tc.expectedFile
			assert.Contains(t, rr.Body.String(), expectedContent,
				"should display content of preferred file: %s", tc.expectedFile)
		})
	}
}

func TestHTTPHandler_GetSourceView_ReadmeErrors(t *testing.T) {
	t.Parallel()

	client := &stubClient{
		fileFunc: func(ctx context.Context, path string, filename string) ([]byte, gnoweb.FileMeta, error) {
			return nil, gnoweb.FileMeta{}, errors.New("mock readme fetch error")
		},
	}

	cfg := newTestHandlerConfig(t, client)
	handler, err := gnoweb.NewHTTPHandler(
		slog.New(slog.NewTextHandler(&testingLogger{t}, nil)),
		cfg,
	)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/r/test_readme$source&file=README.md", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "internal error")
}

func TestHTTPHandler_GetSourceView_ReadmeSuccess(t *testing.T) {
	t.Parallel()

	client := &stubClient{
		fileFunc: func(ctx context.Context, path string, filename string) ([]byte, gnoweb.FileMeta, error) {
			if filename == "README.md" {
				return []byte("# Hello World"), gnoweb.FileMeta{}, nil
			}

			return nil, gnoweb.FileMeta{}, errors.New("uknown file")
		},
		listFilesFunc: func(ctx context.Context, path string) ([]string, error) {
			return []string{"README.md"}, nil
		},
	}

	cfg := newTestHandlerConfig(t, client)
	handler, err := gnoweb.NewHTTPHandler(
		slog.New(slog.NewTextHandler(&testingLogger{t}, nil)),
		cfg,
	)

	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/r/test_readme_success$source&file=README.md", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "README.md")
	// Should contain the rendered markdown content
	assert.Contains(t, rr.Body.String(), "Hello World")
}

func TestHTTPHandler_GetSourceView_DefaultCase(t *testing.T) {
	t.Parallel()

	pkg := &gnoweb.MockPackage{
		Domain: "ex",
		Path:   "/r/test_default",
		Files:  map[string]string{"main.gno": "package main"},
	}

	cfg := newTestHandlerConfig(t, gnoweb.NewMockClient(pkg))

	handler, err := gnoweb.NewHTTPHandler(
		slog.New(slog.NewTextHandler(&testingLogger{t}, nil)),
		cfg,
	)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/r/test_default$source&file=main.gno", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "main.gno")
	assert.Contains(t, rr.Body.String(), "package main")
}

func TestHTTPHandler_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := &stubClient{
		realmFunc: func(ctx context.Context, path, args string) ([]byte, error) {
			// Simulate a slow operation
			select {
			case <-time.After(100 * time.Millisecond):
				return []byte("slow response"), nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	}

	cfg := newTestHandlerConfig(t, client)
	handler, err := gnoweb.NewHTTPHandler(
		slog.New(slog.NewTextHandler(&testingLogger{t}, nil)),
		cfg,
	)
	require.NoError(t, err)

	// Create request with short timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/r/slow/realm", nil)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Should return an error status due to context timeout
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "internal error")
}

func TestHTTPHandler_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := &stubClient{
		listFilesFunc: func(ctx context.Context, path string) ([]string, error) {
			// Check if context is cancelled
			if err := ctx.Err(); err != nil {
				return nil, fmt.Errorf("context cancelled: %w", err)
			}
			return []string{"test.gno"}, nil
		},
		fileFunc: func(ctx context.Context, path, filename string) ([]byte, gnoweb.FileMeta, error) {
			// Check if context is cancelled
			if err := ctx.Err(); err != nil {
				return nil, gnoweb.FileMeta{}, fmt.Errorf("context cancelled: %w", err)
			}
			return []byte("package test"), gnoweb.FileMeta{}, nil
		},
	}

	cfg := newTestHandlerConfig(t, client)
	handler, err := gnoweb.NewHTTPHandler(
		slog.New(slog.NewTextHandler(&testingLogger{t}, nil)),
		cfg,
	)
	require.NoError(t, err)

	// Create request with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := httptest.NewRequest(http.MethodGet, "/r/test/path$source", nil)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Should return an error status due to cancelled context
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "internal error")
}

func TestHTTPHandler_ContextPropagation(t *testing.T) {
	t.Parallel()

	newClient := func(cr map[string]bool) gnoweb.ClientAdapter {
		return &stubClient{
			realmFunc: func(ctx context.Context, path, args string) ([]byte, error) {
				cr["realm"] = ctx != nil
				return []byte("realm content"), nil
			},
			listFilesFunc: func(ctx context.Context, path string) ([]string, error) {
				cr["listFiles"] = ctx != nil
				return []string{"test.gno"}, nil
			},
			fileFunc: func(ctx context.Context, path, filename string) ([]byte, gnoweb.FileMeta, error) {
				cr["file"] = ctx != nil
				return []byte("file content"), gnoweb.FileMeta{}, nil
			},
			docFunc: func(ctx context.Context, path string) (*doc.JSONDocumentation, error) {
				cr["doc"] = ctx != nil
				return &doc.JSONDocumentation{PackagePath: "test"}, nil
			},
			listPathsFunc: func(ctx context.Context, prefix string, limit int) ([]string, error) {
				cr["listPaths"] = ctx != nil
				return []string{"/r/test/path1", "/r/test/path2"}, nil
			},
		}
	}

	testCases := []struct {
		name             string
		path             string
		expectedContexts []string
	}{
		{
			name:             "realm view",
			path:             "/r/test/realm",
			expectedContexts: []string{"realm"},
		},
		{
			name:             "source view",
			path:             "/r/test/path$source",
			expectedContexts: []string{"listFiles"},
		},
		{
			name:             "help view",
			path:             "/r/test/path$help",
			expectedContexts: []string{"doc"},
		},
		{
			name:             "user view",
			path:             "/u/testuser",
			expectedContexts: []string{"realm", "listPaths"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			contextReceived := make(map[string]bool)

			cl := newClient(contextReceived)
			cfg := newTestHandlerConfig(t, cl)
			handler, err := gnoweb.NewHTTPHandler(
				slog.New(slog.NewTextHandler(&testingLogger{t}, nil)),
				cfg,
			)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Verify that context was received for expected operations
			for _, expectedCtx := range tc.expectedContexts {
				assert.True(t, contextReceived[expectedCtx],
					"Context should have been received for %s operation", expectedCtx)
			}
		})
	}
}

func TestHTTPHandler_DownloadWithContext(t *testing.T) {
	t.Parallel()

	const content = "file content for download"

	contextReceived := false
	client := &stubClient{
		fileFunc: func(ctx context.Context, path, filename string) ([]byte, gnoweb.FileMeta, error) {
			contextReceived = ctx != nil
			return []byte(content), gnoweb.FileMeta{}, nil
		},
	}

	cfg := newTestHandlerConfig(t, client)
	handler, err := gnoweb.NewHTTPHandler(
		slog.New(slog.NewTextHandler(&testingLogger{t}, nil)),
		cfg,
	)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/r/test/path$source&file=test.gno&download", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, contextReceived)
	assert.Contains(t, rr.Body.String(), content)
}

// TestHTTPHandler_Post_OpenRedirectBlocked tests that protocol-relative URLs
// are blocked as a defense-in-depth measure.
func TestHTTPHandler_Post_OpenRedirectBlocked(t *testing.T) {
	t.Parallel()

	mockPackage := &gnoweb.MockPackage{
		Domain: "example.com",
		Path:   "/r/test",
		Files: map[string]string{
			"render.gno": `package main`,
		},
	}

	config := newTestHandlerConfig(t, gnoweb.NewMockClient(mockPackage))

	cases := []struct {
		name       string
		path       string
		formData   string
		wantStatus int
		wantIn     string // substring that should be in response
	}{
		{
			name:       "valid path allowed",
			path:       "/r/test:validpath",
			formData:   "field=value",
			wantStatus: http.StatusSeeOther,
		},
		{
			// Defense-in-depth: block protocol-relative URLs that would redirect externally
			// This catches edge cases where the URL encodes to //evil.domain
			name:       "protocol relative URL blocked",
			path:       "/evil.domain",
			formData:   "field=value",
			wantStatus: http.StatusBadRequest,
			wantIn:     "invalid",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{}))
			handler, err := gnoweb.NewHTTPHandler(logger, config)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(tc.formData))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatus, rr.Code, "unexpected status code for path %s", tc.path)
			if tc.wantIn != "" {
				assert.Contains(t, rr.Body.String(), tc.wantIn)
			}
		})
	}
}

// TestHTTPHandler_Post_HiddenPathField tests that the __gno_path hidden form field
// is properly extracted and encoded in the redirect URL.
func TestHTTPHandler_Post_HiddenPathField(t *testing.T) {
	t.Parallel()

	mockPackage := &gnoweb.MockPackage{
		Domain: "example.com",
		Path:   "/r/test",
		Files: map[string]string{
			"render.gno": `package main`,
		},
	}

	config := newTestHandlerConfig(t, gnoweb.NewMockClient(mockPackage))

	cases := []struct {
		name            string
		urlPath         string
		formData        string
		wantStatus      int
		wantRedirectURL string
	}{
		{
			name:            "simple path from hidden field",
			urlPath:         "/r/test",
			formData:        "__gno_path=submit&name=test",
			wantStatus:      http.StatusSeeOther,
			wantRedirectURL: "/r/test:submit?name=test",
		},
		{
			name:            "path with slashes encoded",
			urlPath:         "/r/test",
			formData:        "__gno_path=foo/bar/baz&name=test",
			wantStatus:      http.StatusSeeOther,
			wantRedirectURL: "/r/test:foo%2Fbar%2Fbaz?name=test",
		},
		{
			name:            "path with dots encoded",
			urlPath:         "/r/test",
			formData:        "__gno_path=../../../foo&name=test",
			wantStatus:      http.StatusSeeOther,
			wantRedirectURL: "/r/test:..%2F..%2F..%2Ffoo?name=test",
		},
		{
			name:            "hidden field not included in query params",
			urlPath:         "/r/test",
			formData:        "__gno_path=mypath&field=value",
			wantStatus:      http.StatusSeeOther,
			wantRedirectURL: "/r/test:mypath?field=value",
		},
		{
			name:            "no hidden field - no args in redirect",
			urlPath:         "/r/test",
			formData:        "field=value",
			wantStatus:      http.StatusSeeOther,
			wantRedirectURL: "/r/test?field=value",
		},
		{
			name:            "query in path is encoded",
			urlPath:         "/r/test",
			formData:        "__gno_path=submit?evil=injection&field=value",
			wantStatus:      http.StatusSeeOther,
			wantRedirectURL: "/r/test:submit%3Fevil=injection?field=value",
		},
		{
			name:            "PoC path traversal attack neutralized",
			urlPath:         "/r/test",
			formData:        "__gno_path=user../../../../../evil.domain.com#&field=value",
			wantStatus:      http.StatusSeeOther,
			wantRedirectURL: "/r/test:user..%2F..%2F..%2F..%2F..%2Fevil.domain.com%23?field=value",
		},
		{
			name:            "protocol-relative URL encoded",
			urlPath:         "/r/test",
			formData:        "__gno_path=//evil.com/steal&data=test",
			wantStatus:      http.StatusSeeOther,
			wantRedirectURL: "/r/test:%2F%2Fevil.com%2Fsteal?data=test",
		},
		{
			name:            "full URL with protocol encoded",
			urlPath:         "/r/test",
			formData:        "__gno_path=https://evil.com/steal&data=test",
			wantStatus:      http.StatusSeeOther,
			wantRedirectURL: "/r/test:https:%2F%2Fevil.com%2Fsteal?data=test",
		},
		{
			name:            "javascript URI neutralized",
			urlPath:         "/r/test",
			formData:        "__gno_path=javascript:alert(1)&data=test",
			wantStatus:      http.StatusSeeOther,
			wantRedirectURL: "/r/test:javascript:alert%281%29?data=test",
		},
		{
			name:            "data URI neutralized",
			urlPath:         "/r/test",
			formData:        "__gno_path=data:text/html,<script>alert(1)</script>&data=test",
			wantStatus:      http.StatusSeeOther,
			wantRedirectURL: "/r/test:data:text%2Fhtml%2C%3Cscript%3Ealert%281%29%3C%2Fscript%3E?data=test",
		},
		{
			name:            "fragment in path encoded",
			urlPath:         "/r/test",
			formData:        "__gno_path=submit#fragment&field=value",
			wantStatus:      http.StatusSeeOther,
			wantRedirectURL: "/r/test:submit%23fragment?field=value",
		},
		{
			name:            "complex attack vector encoded",
			urlPath:         "/r/test",
			formData:        "__gno_path=../..//evil.com#@victim.com&field=value",
			wantStatus:      http.StatusSeeOther,
			wantRedirectURL: "/r/test:..%2F..%2F%2Fevil.com%23@victim.com?field=value",
		},
		{
			name:            "null byte injection stays encoded",
			urlPath:         "/r/test",
			formData:        "__gno_path=submit%00evil&field=value",
			wantStatus:      http.StatusSeeOther,
			wantRedirectURL: "/r/test:submit%00evil?field=value",
		},
		{
			name:            "unicode domain encoded",
			urlPath:         "/r/test",
			formData:        "__gno_path=submit/παράδειγμα.δοκιμή&field=value",
			wantStatus:      http.StatusSeeOther,
			wantRedirectURL: "/r/test:submit%2F%CF%80%CE%B1%CF%81%CE%AC%CE%B4%CE%B5%CE%B9%CE%B3%CE%BC%CE%B1.%CE%B4%CE%BF%CE%BA%CE%B9%CE%BC%CE%AE?field=value",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{}))
			handler, err := gnoweb.NewHTTPHandler(logger, config)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, tc.urlPath, strings.NewReader(tc.formData))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.wantStatus, rr.Code, "unexpected status code")
			if tc.wantStatus == http.StatusSeeOther {
				location := rr.Header().Get("Location")
				assert.Equal(t, tc.wantRedirectURL, location, "unexpected redirect URL")
			}
		})
	}
}
