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
		Path   string
		Status int
		Body   string
	}{
		// Found
		{Path: "/r/mock/path", Status: http.StatusOK, Body: "[example.com]/r/mock/path"},
		{Path: "/r/mock/path/", Status: http.StatusOK, Body: "render.gno"},
		{Path: "/r/mock/path/render.gno", Status: http.StatusOK, Body: "one more time"},
		{Path: "/r/mock/path$source&file=render.gno", Status: http.StatusOK, Body: "one more time"},
		{Path: "/r/mock/path/$source", Status: http.StatusOK, Body: "one more time"}, // `render.gno` by default
		{Path: "/r/mock/path$help", Status: http.StatusOK, Body: "SuperRenderFunction"},
		{Path: "/r/mock/path$help", Status: http.StatusOK, Body: "my_super_arg"},

		// Package not found
		{Path: "/r/invalid/path", Status: http.StatusNotFound, Body: "not found"},

		// Invalid path
		{Path: "/r", Status: http.StatusNotFound, Body: "invalid path"},
		{Path: "/r/~!1337", Status: http.StatusNotFound, Body: "invalid path"},
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
			assert.Containsf(t, rr.Body.String(), tc.Body, "rendered body should contain: %q", tc.Body)
		})
	}
}
