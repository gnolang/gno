package gnoweb_test

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/gnolang/gno/tm2/pkg/log"
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
	realmFunc     func(path, args string) ([]byte, error)
	fileFunc      func(path, filename string) ([]byte, gnoweb.FileMeta, error)
	docFunc       func(path string) (*doc.JSONDocumentation, error)
	listFilesFunc func(path string) ([]string, error)
	listPathsFunc func(prefix string, limit int) ([]string, error)
}

func (s *stubClient) Realm(path, args string) ([]byte, error) {
	if s.realmFunc != nil {
		return s.realmFunc(path, args)
	}
	return nil, errors.New("stubClient: Realm not implemented")
}

func (s *stubClient) File(path, filename string) ([]byte, gnoweb.FileMeta, error) {
	if s.fileFunc != nil {
		return s.fileFunc(path, filename)
	}
	return nil, gnoweb.FileMeta{}, errors.New("stubClient: File not implemented")
}

func (s *stubClient) Doc(path string) (*doc.JSONDocumentation, error) {
	if s.docFunc != nil {
		return s.docFunc(path)
	}
	return nil, errors.New("stubClient: Doc not implemented")
}

func (s *stubClient) ListFiles(path string) ([]string, error) {
	if s.listFilesFunc != nil {
		return s.listFilesFunc(path)
	}
	return nil, errors.New("stubClient: ListFiles not implemented")
}

func (s *stubClient) ListPaths(prefix string, limit int) ([]string, error) {
	if s.listPathsFunc != nil {
		return s.listPathsFunc(prefix, limit)
	}
	return nil, errors.New("stubClient: ListPaths not implemented")
}

// newTestHandlerConfig creates a HTTPHandlerConfig for tests using a stub client.
func newTestHandlerConfig(t *testing.T, client gnoweb.ClientAdapter) *gnoweb.HTTPHandlerConfig {
	t.Helper()

	renderer := gnoweb.NewHTMLRenderer(
		log.NewTestingLogger(t),
		gnoweb.NewDefaultRenderConfig(),
	)
	return &gnoweb.HTTPHandlerConfig{
		ClientAdapter: client,
		Renderer:      renderer,
		Aliases:       map[string]gnoweb.AliasTarget{},
	}
}

// TestWebHandler_Get tests the Get method of WebHandler using table-driven tests.
func TestWebHandler_Get(t *testing.T) {
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

// TestWebHandler_NoRender checks if gnoweb displays the `No Render` page properly.
// This happens when the render being queried does not have a Render function declared.
func TestWebHandler_NoRender(t *testing.T) {
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

func TestWebHandler_DirectoryViewExplorerMode(t *testing.T) {
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

// TestWebHandler_DirectoryViewPurePackage covers the pure "package" mode without error:
func TestWebHandler_DirectoryViewPurePackage(t *testing.T) {
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

// TestWebHandler_DirectoryViewErrorTotal covers the case where neither Sources nor QueryPaths return anything:
func TestWebHandler_DirectoryViewErrorTotal(t *testing.T) {
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

// TestNewWebHandlerInvalidConfig ensures that NewWebHandler fails on invalid config.
func TestNewWebHandlerInvalidConfig(t *testing.T) {
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
func TestServeHTTPMethodNotAllowed(t *testing.T) {
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

// TestWebHandler_DirectoryViewNoFiles covers the case where Sources returns
// no error but the list is empty (len(files)==0).
func TestWebHandler_DirectoryViewNoFiles(t *testing.T) {
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

// TestWebHandler_GetSourceView_Error covers the `if err != nil` branch of GetSourceView.
func TestWebHandler_GetSourceView_Error(t *testing.T) {
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

// TestWebHandler_GetSourceView_NoFiles covers the `if len(files)==0` of GetSourceView.
func TestWebHandler_GetSourceView_NoFiles(t *testing.T) {
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

func TestWebHandler_GetUserView(t *testing.T) {
	t.Parallel()

	client := &stubClient{
		listPathsFunc: func(prefix string, limit int) ([]string, error) {
			return []string{
				"/r/testuser/pkg1", "/r/testuser/pkg2",
			}, nil
		},
		realmFunc: func(path string, args string) ([]byte, error) {
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

func TestWebHandler_GetUserView_QueryPathsError(t *testing.T) {
	t.Parallel()

	client := &stubClient{
		listPathsFunc: func(prefix string, limit int) ([]string, error) {
			return nil, errors.New("fail to list paths")
		},
		realmFunc: func(path string, args string) ([]byte, error) {
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

func TestCreateUsernameFromBech32(t *testing.T) {
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

// TestWebHandler_GetSourceView_FilePreference tests the file preference logic
// when no specific file is requested in the source view.
func TestWebHandler_GetSourceView_FilePreference(t *testing.T) {
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

			// Create mock package with the test files
			mockPackage := &gnoweb.MockPackage{
				Domain: "example.com",
				Path:   "/r/test/path",
				Files:  make(map[string]string),
			}

			// Add all test files to the mock
			for _, file := range tc.files {
				mockPackage.Files[file] = "content of " + file
			}

			config := newTestHandlerConfig(t, mockPackage)
			logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{}))
			handler, err := gnoweb.NewWebHandler(logger, config)
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

// Ensure stubDirectoryClient implements gnoweb.WebClient
var _ gnoweb.WebClient = (*stubDirectoryClient)(nil)

// readmeFailClient is a lightweight mock for testing README.md failure in renderReadme.
type readmeFailClient struct{}

func (c *readmeFailClient) HasFile(pkgPath, fileName string) bool {
	return fileName == "README.md"
}

func (c *readmeFailClient) SourceFile(w io.Writer, pkgPath, fileName string, isRaw bool) (*gnoweb.FileMeta, error) {
	return nil, errors.New("mock readme fetch error")
}

// The remaining methods are no-ops or unused for this test:
func (c *readmeFailClient) Sources(path string) ([]string, error)                  { return []string{"README.md"}, nil }
func (c *readmeFailClient) SourceFileRaw(pkgPath, fileName string) ([]byte, error) { return nil, nil }
func (c *readmeFailClient) QueryPaths(prefix string, limit int) ([]string, error)  { return nil, nil }
func (c *readmeFailClient) Doc(path string) (*doc.JSONDocumentation, error)        { return nil, nil }
func (c *readmeFailClient) RenderRealm(w io.Writer, u *weburl.GnoURL, cr gnoweb.ContentRenderer) (*gnoweb.RealmMeta, error) {
	return nil, nil
}

func TestWebHandler_GetSourceView_ReadmeErrors(t *testing.T) {
	t.Parallel()

	mock := &gnoweb.MockPackage{
		Domain: "ex",
		Path:   "/r/test_readme",
		Files:  map[string]string{"README.md": "# Hello"},
	}

	cfg := newTestHandlerConfig(t, mock)
	cfg.WebClient = &readmeFailClient{}

	logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{Level: slog.LevelDebug}))
	handler, err := gnoweb.NewWebHandler(logger, cfg)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/r/test_readme$source&file=README.md", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "internal error")
}

// readmeSuccessClient simulates a client that successfully renders README.md
type readmeSuccessClient struct{}

func (c *readmeSuccessClient) HasFile(pkgPath, fileName string) bool {
	return fileName == "README.md"
}

func (c *readmeSuccessClient) SourceFile(w io.Writer, pkgPath, fileName string, isRaw bool) (*gnoweb.FileMeta, error) {
	if fileName == "README.md" {
		w.Write([]byte("# Hello World"))
		return &gnoweb.FileMeta{Lines: 1, SizeKb: 0.01}, nil
	}
	return nil, errors.New("file not found")
}

// The remaining methods are no-ops or unused for this test:
func (c *readmeSuccessClient) Sources(path string) ([]string, error) {
	return []string{"README.md"}, nil
}
func (c *readmeSuccessClient) SourceFileRaw(pkgPath, fileName string) ([]byte, error) {
	return nil, nil
}
func (c *readmeSuccessClient) QueryPaths(prefix string, limit int) ([]string, error) { return nil, nil }
func (c *readmeSuccessClient) Doc(path string) (*doc.JSONDocumentation, error)       { return nil, nil }
func (c *readmeSuccessClient) RenderRealm(w io.Writer, u *weburl.GnoURL, cr gnoweb.ContentRenderer) (*gnoweb.RealmMeta, error) {
	return nil, nil
}

func TestWebHandler_GetSourceView_ReadmeSuccess(t *testing.T) {
	t.Parallel()

	mock := &gnoweb.MockPackage{
		Domain: "ex",
		Path:   "/r/test_readme_success",
		Files:  map[string]string{"README.md": "# Hello"},
	}

	cfg := newTestHandlerConfig(t, mock)
	cfg.WebClient = &readmeSuccessClient{}

	logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{Level: slog.LevelDebug}))
	handler, err := gnoweb.NewWebHandler(logger, cfg)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/r/test_readme_success$source&file=README.md", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "README.md")
	// Should contain the rendered markdown content
	assert.Contains(t, rr.Body.String(), "Hello World")
}

func TestWebHandler_GetSourceView_DefaultCase(t *testing.T) {
	t.Parallel()

	mock := &gnoweb.MockPackage{
		Domain: "ex",
		Path:   "/r/test_default",
		Files:  map[string]string{"main.gno": "package main"},
	}

	cfg := newTestHandlerConfig(t, mock)

	logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{Level: slog.LevelDebug}))
	handler, err := gnoweb.NewWebHandler(logger, cfg)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/r/test_default$source&file=main.gno", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "main.gno")
	assert.Contains(t, rr.Body.String(), "package main")
}
