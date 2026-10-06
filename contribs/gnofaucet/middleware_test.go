package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// discardLogger is a no-op logger for use in tests.
var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

// hCaptcha test credentials — always pass verification without a real browser.
// See: https://docs.hcaptcha.com/#integration-testing-test-keys
const (
	hcaptchaTestSecret   = "0x0000000000000000000000000000000000000000"
	hcaptchaTestResponse = "10000000-aaaa-bbbb-cccc-000000000001"
)

func TestIPMiddleware_XForwardedFor(t *testing.T) {
	t.Parallel()

	// Helper to run ipMiddleware and return the captured IP and status code.
	runMiddleware := func(t *testing.T, trustedProxyCount int, remoteAddr, xff string) (capturedIP string, statusCode int) {
		t.Helper()

		st := newIPThrottler(defaultRateLimitInterval, defaultCleanTimeout)
		handler := ipMiddleware(discardLogger, trustedProxyCount, st)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedIP, _ = r.Context().Value(remoteIPContextKey).(string)
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.RemoteAddr = remoteAddr
		if xff != "" {
			req.Header.Set("X-Forwarded-For", xff)
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		return capturedIP, rr.Code
	}

	t.Run("count=0 uses RemoteAddr and ignores XFF", func(t *testing.T) {
		t.Parallel()

		ip, code := runMiddleware(t, 0, "10.0.0.1:1234", "203.0.113.50")
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, "10.0.0.1", ip)
	})

	t.Run("count=1 selects rightmost XFF entry (client IP)", func(t *testing.T) {
		t.Parallel()

		// 1 proxy: proxy appends client IP → rightmost is client
		ip, code := runMiddleware(t, 1, "10.0.0.1:1234", "203.0.113.50")
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, "203.0.113.50", ip)
	})

	t.Run("count=1 ignores spoofed leftmost entries", func(t *testing.T) {
		t.Parallel()

		// Client spoofed "1.1.1.1", 1 proxy appended real client IP
		ip, code := runMiddleware(t, 1, "10.0.0.1:1234", "1.1.1.1, 203.0.113.50")
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, "203.0.113.50", ip)
	})

	t.Run("count=2 skips one trusted proxy entry", func(t *testing.T) {
		t.Parallel()

		// 2 proxies: proxy1 appended client, proxy2 appended proxy1
		// XFF: "spoofed, client, proxy1" → pick index len-2 = client
		ip, code := runMiddleware(t, 2, "10.0.0.1:1234", "1.1.1.1, 203.0.113.50, 70.41.3.18")
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, "203.0.113.50", ip)
	})

	t.Run("count exceeds entries uses leftmost", func(t *testing.T) {
		t.Parallel()

		// count=3 but only 2 entries: all from trusted proxies, leftmost is client
		ip, code := runMiddleware(t, 3, "10.0.0.1:1234", "203.0.113.50, 70.41.3.18")
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, "203.0.113.50", ip)
	})

	t.Run("count>0 with empty XFF falls back to RemoteAddr", func(t *testing.T) {
		t.Parallel()

		ip, code := runMiddleware(t, 1, "10.0.0.1:1234", "")
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, "10.0.0.1", ip)
	})

	t.Run("handles whitespace in XFF entries", func(t *testing.T) {
		t.Parallel()

		ip, code := runMiddleware(t, 2, "10.0.0.1:1234", "  1.1.1.1 , 203.0.113.50 , 70.41.3.18 ")
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, "203.0.113.50", ip)
	})
}

func TestCheckHcaptcha(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify method
			assert.Equal(t, http.MethodPost, r.Method)

			// Verify Content-Type is form-encoded (not query params)
			assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

			// Parse and verify form body
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			vals, err := url.ParseQuery(string(body))
			require.NoError(t, err)

			assert.Equal(t, "test-secret", vals.Get("secret"))
			assert.Equal(t, "test-response", vals.Get("response"))
			assert.Empty(t, vals.Get("remoteip"))
			assert.Empty(t, vals.Get("sitekey"))

			// Verify no query params were used
			assert.Empty(t, r.URL.RawQuery)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(SiteVerifyResponse{Success: true})
		}))
		defer srv.Close()

		orig := siteVerifyURL
		siteVerifyURL = srv.URL
		defer func() { siteVerifyURL = orig }()

		require.NoError(t, checkHcaptcha("test-secret", "test-response", "", "", discardLogger))
	})

	t.Run("success with remoteip and sitekey", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			vals, err := url.ParseQuery(string(body))
			require.NoError(t, err)

			assert.Equal(t, "1.2.3.4", vals.Get("remoteip"))
			assert.Equal(t, "test-sitekey", vals.Get("sitekey"))

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(SiteVerifyResponse{Success: true})
		}))
		defer srv.Close()

		orig := siteVerifyURL
		siteVerifyURL = srv.URL
		defer func() { siteVerifyURL = orig }()

		require.NoError(t, checkHcaptcha("test-secret", "test-response", "1.2.3.4", "test-sitekey", discardLogger))
	})

	t.Run("verification failure", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(SiteVerifyResponse{Success: false, ErrorCodes: []string{"invalid-input-response"}})
		}))
		defer srv.Close()

		orig := siteVerifyURL
		siteVerifyURL = srv.URL
		defer func() { siteVerifyURL = orig }()

		err := checkHcaptcha("test-secret", "bad-token", "", "", discardLogger)
		assert.Equal(t, errInvalidCaptcha, err)
	})

	t.Run("non-200 status", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		orig := siteVerifyURL
		siteVerifyURL = srv.URL
		defer func() { siteVerifyURL = orig }()

		err := checkHcaptcha("test-secret", "test-response", "", "", discardLogger)
		assert.ErrorContains(t, err, "unexpected status code")
	})

	t.Run("hcaptcha test credentials", func(t *testing.T) {
		// Calls the real hCaptcha siteverify endpoint using the official test
		// credentials that always return success.
		if testing.Short() {
			t.Skip("skipping network test in short mode")
		}

		require.NoError(t, checkHcaptcha(hcaptchaTestSecret, hcaptchaTestResponse, "", "", discardLogger))
	})
}
