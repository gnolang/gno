package gnoweb_test

import (
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

// --- ajouter EN DESSOUS de TestWebHandler_DirectoryViewExplorerMode ---

// // TestWebHandlerConfigValidate vérifie que validate() détecte les champs manquants.
// func TestWebHandlerConfigValidate(t *testing.T) {
// 	t.Parallel()

// 	// base config valide
// 	dummy := &gnoweb.MockPackage{Domain: "ex", Path: "/r/ex", Files: map[string]string{}}
// 	base := newTestHandlerConfig(t, dummy)

// 	// WebClient manquant
// 	cfg := *base
// 	cfg.WebClient = nil
// 	err := cfg.validate()
// 	require.Error(t, err)
// 	assert.Contains(t, err.Error(), "no `WebClient` configured")

// 	// MarkdownRenderer manquant
// 	cfg = *base
// 	cfg.MarkdownRenderer = nil
// 	err = cfg.validate()
// 	require.Error(t, err)
// 	assert.Contains(t, err.Error(), "no `MarkdownRenderer` configured")

// 	// Aliases manquants
// 	cfg = *base
// 	cfg.Aliases = nil
// 	err = cfg.validate()
// 	require.Error(t, err)
// 	assert.Contains(t, err.Error(), "no `Aliases` configured")
// }

// TestNewWebHandlerInvalidConfig s’assure que NewWebHandler échoue sur config invalide.
// TestNewWebHandlerInvalidConfig s’assure que NewWebHandler échoue sur config invalide.
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

			// Duplique la config valide et muter le champ
			cfg := *valid
			tc.mutate(&cfg)

			logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{}))
			_, err := gnoweb.NewWebHandler(logger, &cfg)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// TestIsHomePath couvre la fonction utilitaire.
func TestIsHomePath(t *testing.T) {
	assert.True(t, gnoweb.IsHomePath("/"))
	assert.False(t, gnoweb.IsHomePath("/foo"))
}

// TestServeHTTPMethodNotAllowed vérifie le 405 pour les méthodes POST/PUT/etc.
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

// TestWebHandler_AliasPath teste le rebouclage d’alias GnowebPath.
// func TestWebHandler_AliasPath(t *testing.T) {
// 	t.Parallel()

// 	// mock qui implémente Render
// 	mockPkg := &gnoweb.MockPackage{
// 		Domain: "example.com",
// 		Path:   "/r/mypath",
// 		Files:  map[string]string{"render.gno": `package main; func Render(path string) string { return "from-alias" }`},
// 		Functions: []*doc.JSONFunc{{
// 			Name:    "Render",
// 			Params:  []*doc.JSONField{{Name: "path", Type: "string"}},
// 			Results: []*doc.JSONField{{Name: "", Type: "string"}},
// 		}},
// 	}
// 	cfg := newTestHandlerConfig(t, mockPkg)
// 	// alias "GET /alias" → "/r/mypath"
// 	cfg.Aliases["/alias"] = gnoweb.AliasTarget{Value: "/r/mypath", Kind: gnoweb.GnowebPath}

// 	logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{}))
// 	handler, err := gnoweb.NewWebHandler(logger, cfg)
// 	require.NoError(t, err)

// 	req := httptest.NewRequest(http.MethodGet, "/alias", nil)
// 	rr := httptest.NewRecorder()
// 	handler.ServeHTTP(rr, req)

// 	assert.Equal(t, http.StatusOK, rr.Code)
// 	assert.Contains(t, rr.Body.String(), "from-alias")
// }

// TestWebHandler_StaticMarkdownAlias couvre l’alias qui renvoie du Markdown statique.
// func TestWebHandler_StaticMarkdownAlias(t *testing.T) {
// 	t.Parallel()

// 	dummy := &gnoweb.MockPackage{Domain: "ex", Path: "/r/ignore", Files: map[string]string{}}
// 	cfg := newTestHandlerConfig(t, dummy)

// 	const md = "# Hello\n\nWorld"
// 	// alias "/md" → contenu markdown statique
// 	cfg.Aliases["/md"] = gnoweb.AliasTarget{Value: md, Kind: gnoweb.StaticMarkdown}

// 	logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{}))
// 	handler, err := gnoweb.NewWebHandler(logger, cfg)
// 	require.NoError(t, err)

// 	req := httptest.NewRequest(http.MethodGet, "/md", nil)
// 	rr := httptest.NewRecorder()
// 	handler.ServeHTTP(rr, req)

// 	assert.Equal(t, http.StatusOK, rr.Code)
// 	body := rr.Body.String()
// 	// Goldmark génère <h1>Hello</h1> puis "World"
// 	assert.Contains(t, body, "<h1>Hello</h1>")
// 	assert.Contains(t, body, "World")
// }
