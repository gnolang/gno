package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIPMiddleware(t *testing.T) {
	t.Parallel()

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	sendRequest := func(mw func(http.Handler) http.Handler, remoteAddr string, headers map[string]string) int {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.RemoteAddr = remoteAddr
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		rr := httptest.NewRecorder()
		mw(okHandler).ServeHTTP(rr, req)
		return rr.Code
	}

	t.Run("no proxy uses RemoteAddr", func(t *testing.T) {
		t.Parallel()

		st := newIPThrottler(defaultRateLimitInterval, defaultCleanTimeout)
		mw := ipMiddleware(0, st)
		xff := map[string]string{"X-Forwarded-For": "9.9.9.9"}

		for i := 0; i < maxRequestsPerMinute; i++ {
			assert.Equal(t, http.StatusOK, sendRequest(mw, "192.168.1.1:12345", xff))
		}
		// Rate limited on RemoteAddr, even with a different XFF value
		assert.Equal(t, http.StatusUnauthorized, sendRequest(mw, "192.168.1.1:12345", map[string]string{
			"X-Forwarded-For": "1.2.3.4",
		}))
	})

	t.Run("single proxy uses rightmost XFF", func(t *testing.T) {
		t.Parallel()

		st := newIPThrottler(defaultRateLimitInterval, defaultCleanTimeout)
		mw := ipMiddleware(1, st)

		// Exhaust rate limit for 10.0.0.1 (rightmost) with different spoofed leftmost IPs
		for i := 0; i < maxRequestsPerMinute; i++ {
			assert.Equal(t, http.StatusOK, sendRequest(mw, "172.16.0.1:12345", map[string]string{
				"X-Forwarded-For": "8.8.8.8, 10.0.0.1",
			}))
		}
		// Rate limited on rightmost IP regardless of spoofed leftmost
		assert.Equal(t, http.StatusUnauthorized, sendRequest(mw, "172.16.0.1:12345", map[string]string{
			"X-Forwarded-For": "1.1.1.1, 10.0.0.1",
		}))
	})

	t.Run("single proxy with single XFF value", func(t *testing.T) {
		t.Parallel()

		st := newIPThrottler(defaultRateLimitInterval, defaultCleanTimeout)
		mw := ipMiddleware(1, st)

		assert.Equal(t, http.StatusOK, sendRequest(mw, "172.16.0.1:12345", map[string]string{
			"X-Forwarded-For": "10.0.0.2",
		}))
	})

	t.Run("no proxy ignores XFF", func(t *testing.T) {
		t.Parallel()

		st := newIPThrottler(defaultRateLimitInterval, defaultCleanTimeout)
		mw := ipMiddleware(0, st)

		for i := 0; i < maxRequestsPerMinute; i++ {
			assert.Equal(t, http.StatusOK, sendRequest(mw, "192.168.1.3:12345", nil))
		}
		// Still rate limited on RemoteAddr, XFF ignored since not behind proxy
		assert.Equal(t, http.StatusUnauthorized, sendRequest(mw, "192.168.1.3:12345", map[string]string{
			"X-Forwarded-For": "10.0.0.5",
		}))
	})

	t.Run("multi-proxy uses correct XFF entry", func(t *testing.T) {
		t.Parallel()

		st := newIPThrottler(defaultRateLimitInterval, defaultCleanTimeout)
		mw := ipMiddleware(2, st) // 2 trusted proxies

		// XFF: "spoofed, real-client, proxy1" — with 2 trusted proxies, idx = 3-2 = 1 → real-client
		for i := 0; i < maxRequestsPerMinute; i++ {
			assert.Equal(t, http.StatusOK, sendRequest(mw, "172.16.0.1:12345", map[string]string{
				"X-Forwarded-For": "8.8.8.8, 10.0.0.1, 192.168.1.1",
			}))
		}
		// Rate limited on real client IP (10.0.0.1) regardless of spoofed leftmost
		assert.Equal(t, http.StatusUnauthorized, sendRequest(mw, "172.16.0.1:12345", map[string]string{
			"X-Forwarded-For": "1.1.1.1, 10.0.0.1, 192.168.1.1",
		}))
	})

	t.Run("multi-proxy falls back to RemoteAddr when XFF too short", func(t *testing.T) {
		t.Parallel()

		st := newIPThrottler(defaultRateLimitInterval, defaultCleanTimeout)
		mw := ipMiddleware(3, st) // 3 trusted proxies

		// XFF has only 2 entries, fewer than trustedProxyCount=3 → fall back to RemoteAddr
		for i := 0; i < maxRequestsPerMinute; i++ {
			assert.Equal(t, http.StatusOK, sendRequest(mw, "192.168.1.4:12345", map[string]string{
				"X-Forwarded-For": "10.0.0.1, 10.0.0.2",
			}))
		}
		// Rate limited on RemoteAddr, not any XFF entry
		assert.Equal(t, http.StatusUnauthorized, sendRequest(mw, "192.168.1.4:12345", map[string]string{
			"X-Forwarded-For": "10.0.0.3, 10.0.0.4",
		}))
	})
}

// hCaptcha test credentials — always pass verification without a real browser.
// See: https://docs.hcaptcha.com/#integration-testing-test-keys
const (
	hcaptchaTestSecret   = "0x0000000000000000000000000000000000000000"
	hcaptchaTestResponse = "10000000-aaaa-bbbb-cccc-000000000001"
)

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

			// Verify no query params were used
			assert.Empty(t, r.URL.RawQuery)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(SiteVerifyResponse{Success: true})
		}))
		defer srv.Close()

		orig := siteVerifyURL
		siteVerifyURL = srv.URL
		defer func() { siteVerifyURL = orig }()

		require.NoError(t, checkHcaptcha("test-secret", "test-response"))
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

		err := checkHcaptcha("test-secret", "bad-token")
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

		err := checkHcaptcha("test-secret", "test-response")
		assert.ErrorContains(t, err, "unexpected status code")
	})

	t.Run("hcaptcha test credentials", func(t *testing.T) {
		// Calls the real hCaptcha siteverify endpoint using the official test
		// credentials that always return success.
		if testing.Short() {
			t.Skip("skipping network test in short mode")
		}

		require.NoError(t, checkHcaptcha(hcaptchaTestSecret, hcaptchaTestResponse))
	})
}
