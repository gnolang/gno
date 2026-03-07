package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
		mw := ipMiddleware(false, st)
		xff := map[string]string{"X-Forwarded-For": "9.9.9.9"}

		for i := 0; i < maxRequestsPerMinute; i++ {
			assert.Equal(t, http.StatusOK, sendRequest(mw, "192.168.1.1:12345", xff))
		}
		// Rate limited on RemoteAddr, even with a different XFF value
		assert.Equal(t, http.StatusUnauthorized, sendRequest(mw, "192.168.1.1:12345", map[string]string{
			"X-Forwarded-For": "1.2.3.4",
		}))
	})

	t.Run("behind proxy uses rightmost XFF", func(t *testing.T) {
		t.Parallel()

		st := newIPThrottler(defaultRateLimitInterval, defaultCleanTimeout)
		mw := ipMiddleware(true, st)

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

	t.Run("behind proxy with single XFF value", func(t *testing.T) {
		t.Parallel()

		st := newIPThrottler(defaultRateLimitInterval, defaultCleanTimeout)
		mw := ipMiddleware(true, st)

		assert.Equal(t, http.StatusOK, sendRequest(mw, "172.16.0.1:12345", map[string]string{
			"X-Forwarded-For": "10.0.0.2",
		}))
	})

	t.Run("not behind proxy ignores XFF", func(t *testing.T) {
		t.Parallel()

		st := newIPThrottler(defaultRateLimitInterval, defaultCleanTimeout)
		mw := ipMiddleware(false, st)

		for i := 0; i < maxRequestsPerMinute; i++ {
			assert.Equal(t, http.StatusOK, sendRequest(mw, "192.168.1.3:12345", nil))
		}
		// Still rate limited on RemoteAddr, XFF ignored since not behind proxy
		assert.Equal(t, http.StatusUnauthorized, sendRequest(mw, "192.168.1.3:12345", map[string]string{
			"X-Forwarded-For": "10.0.0.5",
		}))
	})
}
