package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v64/github"
	"github.com/stretchr/testify/assert"
)

func TestGitHubUsernameMiddleware(t *testing.T) {
	t.Parallel()

	const (
		clientID = "clientID"
		secret   = "secret"
	)

	tests := []struct {
		name             string
		query            string
		exchangeFn       ghExchangeFn
		nextShouldRun    bool
		expectedStatus   int
		expectedUsername string
	}{
		{
			name:           "missing code",
			query:          "",
			exchangeFn:     nil,
			nextShouldRun:  false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "exchange error",
			query: "?code=foo",
			exchangeFn: func(_ context.Context, _ string) (*github.User, error) {
				return nil, errors.New("boom")
			},
			nextShouldRun:  false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "successful flow",
			query: "?code=ok",
			exchangeFn: func(_ context.Context, _ string) (*github.User, error) {
				login := "alice"

				return &github.User{Login: &login}, nil
			},
			nextShouldRun:    true,
			expectedStatus:   http.StatusOK,
			expectedUsername: "alice",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			exchangeFn := defaultGHExchange
			if testCase.exchangeFn != nil {
				exchangeFn = testCase.exchangeFn
			}

			var (
				mw     = gitHubUsernameMiddleware(clientID, secret, exchangeFn)
				called = false

				next = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					called = true

					username := r.Context().Value(ghUsernameKey)

					assert.Equal(t, testCase.expectedUsername, username.(string))
				})
			)

			handler := mw(next)
			req := httptest.NewRequest("GET", "/cb"+testCase.query, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, testCase.nextShouldRun, called)
			assert.Equal(t, testCase.expectedStatus, rr.Code)
		})
	}
}
