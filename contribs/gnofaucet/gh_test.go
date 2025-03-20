package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v64/github"
	"github.com/stretchr/testify/require"
)

// Mock function for exchangeCodeForToken
func mockExchangeCodeForToken(ctx context.Context, secret, clientID, code string) (*github.User, error) {
	login := "mock_login"
	if code == "valid" {
		fmt.Println("mockExchangeCodeForToken: valid")
		return &github.User{Login: &login}, nil
	}
	return nil, errors.New("invalid code")
}

func TestGitHubMiddleware(t *testing.T) {
	cooldown := 2 * time.Minute
	exchangeCodeForUser = mockExchangeCodeForToken
	t.Run("Midleware without credentials", func(t *testing.T) {
		middleware := getGithubMiddleware("", "", getCooldownLimiter(t, cooldown))
		// Test missing clientID and secret, middleware does nothing
		req := httptest.NewRequest("GET", "http://localhost", nil)
		rec := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %d", rec.Code)
		}
	})
	t.Run("request without code", func(t *testing.T) {
		middleware := getGithubMiddleware("mockClientID", "mockSecret", getCooldownLimiter(t, cooldown))
		req := httptest.NewRequest("GET", "http://localhost?code=", nil)
		rec := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status BadRequest, got %d", rec.Code)
		}
	})

	t.Run("request invalid code", func(t *testing.T) {
		middleware := getGithubMiddleware("mockClientID", "mockSecret", getCooldownLimiter(t, cooldown))
		req := httptest.NewRequest("GET", "http://localhost?code=invalid", nil)
		rec := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status BadRequest, got %d", rec.Code)
		}
	})

	t.Run("OK", func(t *testing.T) {
		middleware := getGithubMiddleware("mockClientID", "mockSecret", getCooldownLimiter(t, cooldown))
		req := httptest.NewRequest("GET", "http://localhost?code=valid", nil)
		rec := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %d", rec.Code)
		}
	})

	t.Run("Cooldown active", func(t *testing.T) {
		middleware := getGithubMiddleware("mockClientID", "mockSecret", getCooldownLimiter(t, cooldown))
		req := httptest.NewRequest("GET", "http://localhost?code=valid", nil)
		rec := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %d", rec.Code)
		}

		req = httptest.NewRequest("GET", "http://localhost?code=valid", nil)
		rec = httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusTooManyRequests {
			t.Errorf("Expected status TooManyRequest, got %d", rec.Code)
		}
	})
}

func getCooldownLimiter(t *testing.T, duration time.Duration) *CooldownLimiter {
	t.Helper()
	testDir := os.TempDir()

	file, err := os.CreateTemp(testDir, strings.ReplaceAll(t.Name(), "/", ""))
	require.NoError(t, err)

	limiter, err := NewCooldownLimiter(duration, path.Join(testDir, file.Name()))
	require.NoError(t, err)
	t.Cleanup(func() {
		os.Remove(file.Name())
	})
	return limiter
}
