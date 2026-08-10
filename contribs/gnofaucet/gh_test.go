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
	applyCalls  int
	applyUser   string
	applyAmount int
}

func (m *mockRewarderWithFn) GetReward(ctx context.Context, user string) (int, error) {
	if m.getRewardFn != nil {
		return m.getRewardFn(ctx, user)
	}

	return 0, nil
}

func (m *mockRewarderWithFn) Apply(ctx context.Context, user string, amount int) error {
	m.applyCalls++
	m.applyUser = user
	m.applyAmount = amount

	return nil
}

// TestGitHubClaimRewardsApplyTiming verifies that the contribution-reward debit
// runs only after the downstream claim handler succeeds. The previous ordering
// debited the reward before the cooldown check and before the actual on-chain
// drip, so any downstream failure permanently destroyed the user's earned
// balance with no refund path.
func TestGitHubClaimRewardsApplyTiming(t *testing.T) {
	t.Parallel()

	makeChain := func(reward int, cooldownAllows bool, downstreamErr bool) (*mockRewarderWithFn, *spec.BaseJSONResponse) {
		rewarder := &mockRewarderWithFn{
			getRewardFn: func(_ context.Context, _ string) (int, error) {
				return reward, nil
			},
		}

		mw := chainMiddlewares(getMiddlewares(rewarder, &mockCooldownLimiter{
			checkCooldownFn: func(_ context.Context, _ string, _ int64) (bool, error) {
				return cooldownAllows, nil
			},
		})...)

		next := func(_ context.Context, req *spec.BaseJSONRequest) *spec.BaseJSONResponse {
			if downstreamErr {
				return spec.NewJSONResponse(req.ID, nil,
					spec.NewJSONError("drip failed: simulated downstream failure", spec.ServerErrorCode))
			}

			return spec.NewJSONResponse(req.ID, "ok", nil)
		}

		ctx := context.WithValue(context.Background(), ghUsernameKey, "alice")
		req := spec.NewJSONRequest(1, claimRPCMethod, []any{"dst-addr"})

		return rewarder, mw(next)(ctx, req)
	}

	t.Run("apply does not run when cooldown rejects", func(t *testing.T) {
		t.Parallel()

		rewarder, resp := makeChain(100, false, false)

		assert.NotNil(t, resp.Error, "cooldown rejection must surface as an error")
		assert.Contains(t, resp.Error.Message, "cooldown")
		assert.Zero(t, rewarder.applyCalls,
			"reward must NOT be debited when cooldown rejects the claim")
	})

	t.Run("apply does not run when downstream drip fails", func(t *testing.T) {
		t.Parallel()

		rewarder, resp := makeChain(100, true, true)

		assert.NotNil(t, resp.Error, "downstream failure must surface as an error")
		assert.Contains(t, resp.Error.Message, "drip failed")
		assert.Zero(t, rewarder.applyCalls,
			"reward must NOT be debited when the downstream drip fails")
	})

	t.Run("apply runs exactly once on full success", func(t *testing.T) {
		t.Parallel()

		rewarder, resp := makeChain(100, true, false)

		assert.Nil(t, resp.Error, "successful claim must not return an error")
		assert.Equal(t, 1, rewarder.applyCalls,
			"reward must be debited exactly once on successful claim")
		assert.Equal(t, "alice", rewarder.applyUser)
		assert.Equal(t, 100, rewarder.applyAmount)
	})
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
