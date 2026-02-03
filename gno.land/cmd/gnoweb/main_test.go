package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeRemoteURL(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"tcp protocol", "tcp://127.0.0.1:26657", "http://127.0.0.1:26657"},
		{"http protocol", "http://127.0.0.1:26657", "http://127.0.0.1:26657"},
		{"https protocol", "https://rpc.gno.land:443", "https://rpc.gno.land:443"},
		{"no protocol", "127.0.0.1:26657", "http://127.0.0.1:26657"},
		{"no protocol with domain", "rpc.gno.land:443", "http://rpc.gno.land:443"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := normalizeRemoteURL(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestNormalizeRemoteURL_UnsupportedProtocol(t *testing.T) {
	t.Parallel()

	unsupportedCases := []struct {
		name  string
		input string
	}{
		{"unix protocol", "unix://var/run/gno.sock"},
		{"file protocol", "file:///path/to/file"},
		{"ftp protocol", "ftp://example.com"},
		{"ws protocol", "ws://example.com"},
	}

	for _, tc := range unsupportedCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Panics(t, func() {
				normalizeRemoteURL(tc.input)
			}, "Expected panic for unsupported protocol: %s", tc.input)
		})
	}
}

func TestSetupWeb(t *testing.T) {
	opts := defaultWebOptions
	opts.bind = "127.0.0.1:0"
	stdio := commands.NewDefaultIO()

	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0o644)
	require.NoError(t, err)
	defer devNull.Close()

	stdio.SetOut(devNull)

	_, err = setupWeb(&opts, []string{}, stdio)
	require.NoError(t, err)
}

// Dummy handler to simulate the processing chain.
// It now returns a more detailed message in the response body.
func dummyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func TestSecureHeadersMiddlewareStrict(t *testing.T) {
	handler := SecureHeadersMiddleware(http.HandlerFunc(dummyHandler), true, "http://example.com")

	req := httptest.NewRequest("GET", "http://example.com", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	res := rec.Result()

	// Check common headers.
	if res.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("Expected X-Content-Type-Options 'nosniff', got '%s'", res.Header.Get("X-Content-Type-Options"))
	}
	if res.Header.Get("X-Frame-Options") != "DENY" {
		t.Errorf("Expected X-Frame-Options 'DENY', got '%s'", res.Header.Get("X-Frame-Options"))
	}
	if res.Header.Get("Referrer-Policy") != "no-referrer" {
		t.Errorf("Expected Referrer-Policy 'no-referrer', got '%s'", res.Header.Get("Referrer-Policy"))
	}

	// Check headers specific to strict mode.
	csp := res.Header.Get("Content-Security-Policy")
	if !strings.Contains(csp, "https://assets.gnoteam.com") {
		t.Errorf("Expected Content-Security-Policy to contain 'https://assets.gnoteam.com', got '%s'", csp)
	}
	if res.Header.Get("Strict-Transport-Security") != "max-age=31536000" {
		t.Errorf("Expected Strict-Transport-Security 'max-age=31536000', got '%s'", res.Header.Get("Strict-Transport-Security"))
	}

	// Optionally, verify the response body.
	body := rec.Body.String()
	if !strings.Contains(body, "OK") {
		t.Errorf("Unexpected response body: %s", body)
	}
}

func TestSecureHeadersMiddlewareNonStrict(t *testing.T) {
	handler := SecureHeadersMiddleware(http.HandlerFunc(dummyHandler), false, "http://example.com")

	req := httptest.NewRequest("GET", "http://example.com", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	res := rec.Result()

	// Check that the common headers are set.
	if res.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("Expected X-Content-Type-Options 'nosniff', got '%s'", res.Header.Get("X-Content-Type-Options"))
	}
	if res.Header.Get("X-Frame-Options") != "DENY" {
		t.Errorf("Expected X-Frame-Options 'DENY', got '%s'", res.Header.Get("X-Frame-Options"))
	}
	if res.Header.Get("Referrer-Policy") != "no-referrer" {
		t.Errorf("Expected Referrer-Policy 'no-referrer', got '%s'", res.Header.Get("Referrer-Policy"))
	}

	// In non-strict mode, CSP and HSTS should not be defined.
	if csp := res.Header.Get("Content-Security-Policy"); csp != "" {
		t.Errorf("Did not expect Content-Security-Policy in non-strict mode, got '%s'", csp)
	}
	if hsts := res.Header.Get("Strict-Transport-Security"); hsts != "" {
		t.Errorf("Did not expect Strict-Transport-Security in non-strict mode, got '%s'", hsts)
	}

	// Optionally, verify the response body.
	body := rec.Body.String()
	if !strings.Contains(body, "OK") {
		t.Errorf("Unexpected response body: %s", body)
	}
}

func TestParseAliases(t *testing.T) {
	t.Parallel()

	var (
		existingStatic = path.Join(t.TempDir(), "existing")
		missingStatic  = path.Join(t.TempDir(), "missing")
	)

	_, err := os.Create(existingStatic)
	require.NoError(t, err)

	cases := []struct {
		name       string
		aliasesStr string
		mapSize    int
		error      bool
	}{
		{
			name:       "empty",
			aliasesStr: "",
			mapSize:    0,
			error:      true,
		},
		{
			name:       "only whitespaces",
			aliasesStr: "    ",
			mapSize:    0,
			error:      true,
		},
		{
			name:       "no separator",
			aliasesStr: "alias1",
			mapSize:    0,
			error:      true,
		},
		{
			name:       "too many separators",
			aliasesStr: "alias1 = = target1",
			mapSize:    0,
			error:      true,
		},
		{
			name:       "empty entry",
			aliasesStr: "alias1 = target1, , alias3 = target3",
			mapSize:    0,
			error:      true,
		},
		{
			name:       "valid entry",
			aliasesStr: "alias1 = target1, alias2 = target2, alias3 = target3",
			mapSize:    3,
			error:      false,
		},
		{
			name:       "alias existing static file",
			aliasesStr: "alias1 = static:" + existingStatic,
			mapSize:    1,
			error:      false,
		},
		{
			name:       "alias multiple static files",
			aliasesStr: "alias1 = static:" + existingStatic + ", alias2 = static:" + existingStatic + ", alias3 = static:" + existingStatic,
			mapSize:    3,
			error:      false,
		},
		{
			name:       "alias missing static file",
			aliasesStr: "alias1 = static:" + missingStatic,
			mapSize:    0,
			error:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			aliases, err := parseAliases(tc.aliasesStr)
			if tc.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.mapSize, len(aliases))
		})
	}
}
