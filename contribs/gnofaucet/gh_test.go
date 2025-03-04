package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-github/v64/github"
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

// Mock function for GitHub client
func mockGetUser(token string) (*github.User, error) {
	if token == "mock_token" {
		return &github.User{Login: github.String("testUser")}, nil
	}
	return nil, errors.New("invalid token")
}

func TestGitHubMiddleware(t *testing.T) {
	cooldown := 2 * time.Minute
	exchangeCodeForUser = mockExchangeCodeForToken
	t.Run("Midleware without credentials", func(t *testing.T) {
		middleware := getGithubMiddleware("", "", cooldown)
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
		middleware := getGithubMiddleware("mockClientID", "mockSecret", cooldown)
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
		middleware := getGithubMiddleware("mockClientID", "mockSecret", cooldown)
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
		middleware := getGithubMiddleware("mockClientID", "mockSecret", cooldown)
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
		middleware := getGithubMiddleware("mockClientID", "mockSecret", cooldown)
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
