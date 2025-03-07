package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/require"
)

func TestSetupWeb(t *testing.T) {
	opts := defaultWebOptions
	opts.bind = "127.0.0.1:0" // random port
	stdio := commands.NewDefaultIO()

	// Open /dev/null as a write-only file
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
	handler := SecureHeadersMiddleware(http.HandlerFunc(dummyHandler), true)

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
	if !strings.Contains(body, "OK: response from dummy handler with detailed information") {
		t.Errorf("Unexpected response body: %s", body)
	}
}

func TestSecureHeadersMiddlewareNonStrict(t *testing.T) {
	handler := SecureHeadersMiddleware(http.HandlerFunc(dummyHandler), false)

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
	if !strings.Contains(body, "OK: response from dummy handler with detailed information") {
		t.Errorf("Unexpected response body: %s", body)
	}
}
