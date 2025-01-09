package gnoweb_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
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

// TestWebHandler_Get tests the Get method of WebHandler using table-driven tests.
func TestWebHandler_Get(t *testing.T) {
	// Set up a mock package with some files and functions
	mockPackage := &gnoweb.MockPackage{
		Domain: "example.com",
		Path:   "/r/mock/path",
		Files: map[string]string{
			"render.gno": `package main; func Render(path string) { return "one more time" }`,
			"gno.mod":    `module example.com/r/mock/path`,
			"LicEnse":    `my super license`,
		},
		Functions: []vm.FunctionSignature{
			{FuncName: "SuperRenderFunction", Params: []vm.NamedType{
				{Name: "my_super_arg", Type: "string"},
			}},
		},
	}

	// Create a mock web client with the mock package
	webclient := gnoweb.NewMockWebClient(mockPackage)

	// Create a WebHandlerConfig with the mock web client and static metadata
	config := gnoweb.WebHandlerConfig{
		WebClient: webclient,
	}

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
