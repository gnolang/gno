package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/go-github/v64/github"
	"github.com/redis/go-redis/v9"
)

// Mock function for exchangeCodeForToken
func mockExchangeCodeForToken(ctx context.Context, secret, clientID, code string) (*github.User, error) {
	login := "mock_login"
	if code == "valid" {
		return &github.User{Login: &login}, nil
	}
	return nil, errors.New("invalid code")
}

func TestGitHubMiddleware(t *testing.T) {
	cooldown := 2 * time.Minute
	exchangeCodeForUser = mockExchangeCodeForToken
	var tenGnots int64 = 10000000
	claimBody := fmt.Sprintf(`{"amount": "%dugnot"}`, tenGnots)
	t.Run("request without code", func(t *testing.T) {
		middleware := getGithubMiddleware("mockClientID", "mockSecret", getCooldownLimiter(t, cooldown, 0))
		req := httptest.NewRequest("GET", "http://localhost?code=", bytes.NewBufferString(claimBody))
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
		middleware := getGithubMiddleware("mockClientID", "mockSecret", getCooldownLimiter(t, cooldown, 0))
		req := httptest.NewRequest("GET", "http://localhost?code=invalid", bytes.NewBufferString(claimBody))
		rec := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status BadRequest, got %d", rec.Code)
		}
	})

	t.Run("Invalid amount", func(t *testing.T) {
		claimBody := `{"amount": 100000}`
		middleware := getGithubMiddleware("mockClientID", "mockSecret", getCooldownLimiter(t, cooldown, 0))
		req := httptest.NewRequest("GET", "http://localhost?code=valid", bytes.NewBufferString(claimBody))
		rec := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status OK, got %d", rec.Code)
		}
	})

	t.Run("OK", func(t *testing.T) {
		middleware := getGithubMiddleware("mockClientID", "mockSecret", getCooldownLimiter(t, cooldown, 0))
		req := httptest.NewRequest("GET", "http://localhost?code=valid", bytes.NewBufferString(claimBody))
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
		middleware := getGithubMiddleware("mockClientID", "mockSecret", getCooldownLimiter(t, cooldown, 0))
		req := httptest.NewRequest("GET", "http://localhost?code=valid", bytes.NewBufferString(claimBody))
		rec := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %d", rec.Code)
		}

		req = httptest.NewRequest("GET", "http://localhost?code=valid", bytes.NewBufferString(claimBody))
		rec = httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusTooManyRequests {
			t.Errorf("Expected status TooManyRequest, got %d", rec.Code)
		}
	})

	t.Run("User exceeded lifetime limit", func(t *testing.T) {
		cooldown = time.Millisecond
		// Max lifetime amount is 20 Gnots so we should be able to make 2 claims
		middleware := getGithubMiddleware("mockClientID", "mockSecret", getCooldownLimiter(t, cooldown, tenGnots*2))
		req := httptest.NewRequest("GET", "http://localhost?code=valid", bytes.NewBufferString(claimBody))
		rec := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rec, req)
		// First claim ok
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %d", rec.Code)
		}
		// Wait 2 times the cooldown
		time.Sleep(2 * cooldown)

		req = httptest.NewRequest("GET", "http://localhost?code=valid", bytes.NewBufferString(claimBody))
		rec = httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		// Second claim should also be ok
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %d", rec.Code)
		}

		// Third one should fail
		time.Sleep(2 * cooldown)

		req = httptest.NewRequest("GET", "http://localhost?code=valid", bytes.NewBufferString(claimBody))
		rec = httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		// third claim should fail
		if rec.Code != http.StatusTooManyRequests {
			t.Errorf("Expected status OK, got %d", rec.Code)
		}
	})
}

func getCooldownLimiter(t *testing.T, duration time.Duration, maxlifeTimeAmount int64) *CooldownLimiter {
	t.Helper()
	redisServer := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: redisServer.Addr(),
	})

	limiter := NewCooldownLimiter(duration, rdb, maxlifeTimeAmount)

	return limiter
}
