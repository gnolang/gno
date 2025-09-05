package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gnolang/faucet"
	"github.com/gnolang/faucet/spec"
	"github.com/google/go-github/v64/github"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	igh "github.com/gnolang/gno/contribs/gnofaucet/github"
)

var noopLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestGitHubUsernameMiddleware(t *testing.T) {
	t.Parallel()

	const (
		clientID = "clientID"
		secret   = "secret"
	)

	testTable := []struct {
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

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			exchangeFn := defaultGHExchange
			if testCase.exchangeFn != nil {
				exchangeFn = testCase.exchangeFn
			}

			redisServer := miniredis.RunT(t)
			rdb := redis.NewClient(&redis.Options{
				Addr: redisServer.Addr(),
			})

			var (
				mw     = gitHubUsernameMiddleware(clientID, secret, exchangeFn, noopLogger, rdb)
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

type checkCooldownDelegate func(context.Context, string, int64) (bool, error)

type mockCooldownLimiter struct {
	checkCooldownFn checkCooldownDelegate
}

func (m *mockCooldownLimiter) checkCooldown(ctx context.Context, key string, amountClaimed int64) (bool, error) {
	if m.checkCooldownFn != nil {
		return m.checkCooldownFn(ctx, key, amountClaimed)
	}

	return false, nil
}

type mockRewarder struct{}

func (m *mockRewarder) GetReward(ctx context.Context, user string) (int, error) {
	return 0, nil
}

func (m *mockRewarder) Apply(ctx context.Context, user string, amount int) error {
	return nil
}

func TestGitHubClaimMiddleware(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		ctx             context.Context
		name            string
		limiter         cooldownLimiter
		req             *spec.BaseJSONRequest
		nextShouldRun   bool
		expectedError   string
		expectedErrCode int
	}{
		{
			name: "no username in ctx",
			limiter: &mockCooldownLimiter{
				checkCooldownFn: func(_ context.Context, _ string, _ int64) (bool, error) {
					return false, errors.New("error")
				},
			},
			ctx:             context.Background(),
			req:             spec.NewJSONRequest(1, faucet.DefaultDripMethod, []any{"foo", "1000ugnot"}),
			nextShouldRun:   false,
			expectedError:   "invalid username value",
			expectedErrCode: spec.InvalidRequestErrorCode,
		},
		{
			name: "invalid method",
			limiter: &mockCooldownLimiter{
				checkCooldownFn: func(_ context.Context, _ string, _ int64) (bool, error) {
					return true, nil
				},
			},
			ctx:             context.WithValue(context.Background(), ghUsernameKey, "bob"),
			req:             spec.NewJSONRequest(2, "random_method", []any{"foo", "1000ugnot"}),
			nextShouldRun:   false,
			expectedError:   "invalid method requested",
			expectedErrCode: spec.InvalidRequestErrorCode,
		},
		{
			name: "missing amount param",
			limiter: &mockCooldownLimiter{
				checkCooldownFn: func(_ context.Context, _ string, _ int64) (bool, error) {
					return true, nil
				},
			},
			ctx:             context.WithValue(context.Background(), ghUsernameKey, "bob"),
			req:             spec.NewJSONRequest(3, faucet.DefaultDripMethod, []any{"only_one"}),
			nextShouldRun:   false,
			expectedError:   "amount not provided",
			expectedErrCode: spec.InvalidParamsErrorCode,
		},
		{
			name: "invalid amount parse",
			limiter: &mockCooldownLimiter{
				checkCooldownFn: func(_ context.Context, _ string, _ int64) (bool, error) {
					return true, nil
				},
			},
			ctx:             context.WithValue(context.Background(), ghUsernameKey, "bob"),
			req:             spec.NewJSONRequest(4, faucet.DefaultDripMethod, []any{"foo", "notacoins"}),
			nextShouldRun:   false,
			expectedError:   "invalid amount",
			expectedErrCode: spec.InvalidParamsErrorCode,
		},
		{
			name: "cooldown check error",
			limiter: &mockCooldownLimiter{
				checkCooldownFn: func(_ context.Context, _ string, _ int64) (bool, error) {
					return false, errors.New("cooldown error")
				},
			},
			ctx:             context.WithValue(context.Background(), ghUsernameKey, "bob"),
			req:             spec.NewJSONRequest(5, faucet.DefaultDripMethod, []any{"foo", "100atom"}),
			nextShouldRun:   false,
			expectedError:   "unable to check cooldown",
			expectedErrCode: spec.ServerErrorCode,
		},
		{
			name: "cooldown active",
			limiter: &mockCooldownLimiter{
				checkCooldownFn: func(_ context.Context, _ string, _ int64) (bool, error) {
					return false, nil
				},
			},
			ctx:             context.WithValue(context.Background(), ghUsernameKey, "bob"),
			req:             spec.NewJSONRequest(6, faucet.DefaultDripMethod, []any{"foo", "100atom"}),
			nextShouldRun:   false,
			expectedError:   "user is on cooldown",
			expectedErrCode: spec.ServerErrorCode,
		},
		{
			name: "no cooldown",
			limiter: &mockCooldownLimiter{
				checkCooldownFn: func(_ context.Context, _ string, _ int64) (bool, error) {
					return true, nil
				},
			},
			ctx:           context.WithValue(context.Background(), ghUsernameKey, "bob"),
			req:           spec.NewJSONRequest(7, faucet.DefaultDripMethod, []any{"foo", "100atom"}),
			nextShouldRun: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				mw     = chainMiddlewares(getMiddlewares(&mockRewarder{}, testCase.limiter)...)
				called = false
				next   = func(ctx context.Context, req *spec.BaseJSONRequest) *spec.BaseJSONResponse {
					called = true

					return spec.NewJSONResponse(req.ID, "ok", nil)
				}
			)

			handler := mw(next)
			resp := handler(testCase.ctx, testCase.req)

			assert.Equal(t, testCase.nextShouldRun, called)

			if testCase.nextShouldRun {
				assert.Nil(t, resp.Error)

				assert.Equal(t, "ok", resp.Result.(string))

				return
			}

			assert.NotNil(t, resp.Error)
			assert.Contains(t, resp.Error.Message, testCase.expectedError)
			assert.Equal(t, resp.Error.Code, testCase.expectedErrCode)
		})
	}
}

type mockRewarderWithFn struct {
	getRewardFn func(context.Context, string) (int, error)
}

func (m *mockRewarderWithFn) GetReward(ctx context.Context, user string) (int, error) {
	if m.getRewardFn != nil {
		return m.getRewardFn(ctx, user)
	}

	return 0, nil
}

func (m *mockRewarderWithFn) Apply(ctx context.Context, user string, amount int) error {
	return nil
}

func TestGitHubCheckRewardsMiddleware(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name            string
		ctx             context.Context
		req             *spec.BaseJSONRequest
		rewarder        igh.Rewarder
		nextShouldRun   bool
		expectedError   string
		expectedErrCode int
		expectedResult  any
	}{
		{
			name:            "no username in ctx",
			ctx:             context.Background(),
			req:             spec.NewJSONRequest(1, getClaimRPCMethod, nil),
			rewarder:        &mockRewarderWithFn{},
			nextShouldRun:   false,
			expectedError:   "invalid username value",
			expectedErrCode: spec.InvalidRequestErrorCode,
		},
		{
			name: "invalid method",
			ctx:  context.WithValue(context.Background(), ghUsernameKey, "ajnavarro"),
			req:  spec.NewJSONRequest(2, "boo", nil),
			rewarder: &mockRewarderWithFn{
				getRewardFn: func(_ context.Context, _ string) (int, error) {
					return 0, nil
				},
			},
			nextShouldRun:   false,
			expectedError:   "invalid method requested",
			expectedErrCode: spec.InvalidRequestErrorCode,
		},
		{
			name: "rewarder error",
			ctx:  context.WithValue(context.Background(), ghUsernameKey, "ajnavarro"),
			req:  spec.NewJSONRequest(3, getClaimRPCMethod, nil),
			rewarder: &mockRewarderWithFn{
				getRewardFn: func(_ context.Context, _ string) (int, error) {
					return 0, errors.New("boom")
				},
			},
			nextShouldRun:   false,
			expectedError:   "unable to get reward",
			expectedErrCode: spec.ServerErrorCode,
		},
		{
			name: "successful getClaim",
			ctx:  context.WithValue(context.Background(), ghUsernameKey, "ajnavarro"),
			req:  spec.NewJSONRequest(4, getClaimRPCMethod, nil),
			rewarder: &mockRewarderWithFn{
				getRewardFn: func(_ context.Context, _ string) (int, error) {
					return 10, nil
				},
			},
			nextShouldRun:  false,
			expectedResult: 10,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				mw     = chainMiddlewares(getMiddlewares(tc.rewarder, nil)...)
				called = false
				next   = func(ctx context.Context, req *spec.BaseJSONRequest) *spec.BaseJSONResponse {
					called = true

					return spec.NewJSONResponse(req.ID, "next", nil)
				}
			)

			handler := mw(next)
			resp := handler(tc.ctx, tc.req)

			assert.Equal(t, tc.nextShouldRun, called)

			if tc.expectedError != "" {
				assert.NotNil(t, resp.Error)
				assert.Contains(t, resp.Error.Message, tc.expectedError)
				assert.Equal(t, tc.expectedErrCode, resp.Error.Code)

				return
			}

			assert.Nil(t, resp.Error)
			assert.Equal(t, tc.expectedResult, resp.Result)
		})
	}
}

// chainMiddlewares combines the given JSON-RPC middlewares
func chainMiddlewares(mw ...faucet.Middleware) faucet.Middleware {
	return func(final faucet.HandlerFunc) faucet.HandlerFunc {
		h := final

		for i := len(mw) - 1; i >= 0; i-- {
			h = mw[i](h)
		}

		return h
	}
}
