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

		require.NoError(t, checkHcaptcha("test-secret", "test-response", "", ""))
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

		require.NoError(t, checkHcaptcha("test-secret", "test-response", "1.2.3.4", "test-sitekey"))
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

		err := checkHcaptcha("test-secret", "bad-token", "", "")
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

		err := checkHcaptcha("test-secret", "test-response", "", "")
		assert.ErrorContains(t, err, "unexpected status code")
	})

	t.Run("hcaptcha test credentials", func(t *testing.T) {
		// Calls the real hCaptcha siteverify endpoint using the official test
		// credentials that always return success.
		if testing.Short() {
			t.Skip("skipping network test in short mode")
		}

		require.NoError(t, checkHcaptcha(hcaptchaTestSecret, hcaptchaTestResponse, "", ""))
	})
}
