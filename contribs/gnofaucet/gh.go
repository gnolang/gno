package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gnolang/faucet"
	"github.com/gnolang/faucet/spec"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/google/go-github/v64/github"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	igh "github.com/gnolang/gno/contribs/gnofaucet/github"
)

// type to scope the context.Context values to this package.
type faucetContextKey string

// ghUsernameKey is the context key for storing the GH username between
// http and RPC GitHub middleware handlers
const ghUsernameKey faucetContextKey = "gh-username"

// claimRPCMethod is a method that is exactly like the default drip method but without any amount set
const claimRPCMethod = "claim"

// getClaimRPCMethod is a method that returns the amount that the user will be able to claim
const getClaimRPCMethod = "checkClaim"

var allowedMethods = []string{faucet.DefaultDripMethod, claimRPCMethod, getClaimRPCMethod}

func getMiddlewares(rr igh.Rewarder, cooldownLimiter cooldownLimiter) []faucet.Middleware {
	return []faucet.Middleware{
		invalidMethodMiddleware(),
		gitHubClaimRewardsMiddleware(rr),
		gitHubClaimMiddleware(cooldownLimiter),
		gitHubCheckRewardsMiddleware(rr),
	}
}

// gitHubUsernameMiddleware sets up authentication middleware for GitHub OAuth.
// If clientID and secret are empty, the middleware does nothing.
//
// Parameters:
// - clientID: The OAuth client ID issued by GitHub when registering the application.
// - secret: The OAuth client secret used to securely authenticate API requests.
//
// GitHub OAuth applications require a client ID and secret to authenticate users securely.
// These credentials are obtained when registering an application on GitHub at:
// https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authenticating-to-the-rest-api-with-an-oauth-app#registering-your-app
func gitHubUsernameMiddleware(clientID, secret string, exchangeFn ghExchangeFn, logger *slog.Logger, rdb *redis.Client) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				coo, err := r.Cookie(string(ghUsernameKey))
				if coo != nil {
					logger.Debug("cookie obtained", "name", coo.Name, "value", coo.Value)

					username, rerr := rdb.Get(r.Context(), redisSessionKey(coo.Value)).Result()
					if rerr == nil && username != "" {
						updatedCtx := context.WithValue(r.Context(), ghUsernameKey, username)
						next.ServeHTTP(w, r.WithContext(updatedCtx))
						return
					}
					if rerr != nil && !errors.Is(rerr, redis.Nil) {
						logger.Debug("user from cookie not found", "err", rerr)
					}
				}

				if errors.Is(err, http.ErrNoCookie) {
					logger.Debug("cookie not present", "err", err)
				}
				if err != nil {
					logger.Debug("cookie error", "err", err)
				}

				w.Header().Set("Content-Type", "text/plain")

				// Extracts the authorization code returned by the GitHub OAuth flow.
				//
				// When a user successfully authenticates via GitHub OAuth, GitHub redirects them
				// to the registered callback URL with a `code` query parameter. This code is then
				// exchanged for an access token.
				//
				// Reference: https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps#2-users-are-redirected-back-to-your-site-by-github
				code := r.URL.Query().Get("code")
				if code == "" {
					http.Error(w, "missing code", http.StatusBadRequest)

					return
				}

				user, err := exchangeFn(
					r.Context(),
					fmt.Sprintf("client_id=%s&client_secret=%s&code=%s", clientID, secret, code),
				)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				sessionID := uuid.NewString()
				ttlSeconds := 3600
				if err := rdb.Set(
					r.Context(),
					redisSessionKey(sessionID),
					user.GetLogin(),
					time.Duration(ttlSeconds)*time.Second,
				).Err(); err != nil {
					http.Error(w, "unable to persist session", http.StatusInternalServerError)
				}

				c := &http.Cookie{
					Name:     string(ghUsernameKey),
					Value:    sessionID,
					MaxAge:   ttlSeconds,
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteLaxMode,
				}

				http.SetCookie(w, c)

				logger.Debug("cookie set!", "cookie", c)

				// Save the username in the context
				updatedCtx := context.WithValue(r.Context(), ghUsernameKey, user.GetLogin())

				// Possibility to have more conditions like accountAge, commits, pullRequest, etc.
				next.ServeHTTP(w, r.WithContext(updatedCtx))
			},
		)
	}
}

func redisSessionKey(id string) string { return "gh-session:" + id }

type cooldownLimiter interface {
	checkCooldown(context.Context, string, int64) (bool, error)
}

// invalidMethodMiddleware is the middleware that checks if the method is valid.
// If the method is not valid, it will return an error.
// If the method is valid, it will call the next middleware.
func invalidMethodMiddleware() faucet.Middleware {
	return func(next faucet.HandlerFunc) faucet.HandlerFunc {
		return func(ctx context.Context, req *spec.BaseJSONRequest) *spec.BaseJSONResponse {
			if !slices.Contains(allowedMethods, req.Method) {
				return spec.NewJSONResponse(
					req.ID,
					nil,
					spec.NewJSONError("invalid method requested", spec.InvalidRequestErrorCode),
				)
			}

			return next(ctx, req)
		}
	}
}

// gitHubCheckRewardsMiddleware is the middleware that is being called
// from the frontend to check if the pending user's rewards.
func gitHubCheckRewardsMiddleware(rewarder igh.Rewarder) faucet.Middleware {
	return func(next faucet.HandlerFunc) faucet.HandlerFunc {
		return func(ctx context.Context, req *spec.BaseJSONRequest) *spec.BaseJSONResponse {
			if req.Method != getClaimRPCMethod {
				return next(ctx, req)
			}

			// Grab the username from the context
			username, ok := ctx.Value(ghUsernameKey).(string)
			if !ok {
				return spec.NewJSONResponse(
					req.ID,
					nil,
					spec.NewJSONError("invalid username value", spec.InvalidRequestErrorCode),
				)
			}

			reward, err := rewarder.GetReward(ctx, username)
			if err != nil {
				return spec.NewJSONResponse(
					req.ID,
					nil,
					spec.NewJSONError(fmt.Sprintf("unable to get reward: %v", err), spec.ServerErrorCode),
				)
			}

			return spec.NewJSONResponse(req.ID, reward, nil)
		}
	}
}

// gitHubClaimRewardsMiddleware is the middleware that claims rewards for the user.
// Using the github username, it will check Redis to see if the user has already claimed rewards.
func gitHubClaimRewardsMiddleware(rewarder igh.Rewarder) faucet.Middleware {
	return func(next faucet.HandlerFunc) faucet.HandlerFunc {
		return func(ctx context.Context, req *spec.BaseJSONRequest) *spec.BaseJSONResponse {
			if req.Method != claimRPCMethod {
				return next(ctx, req)
			}

			// Grab the username from the context
			username, ok := ctx.Value(ghUsernameKey).(string)
			if !ok {
				return spec.NewJSONResponse(
					req.ID,
					nil,
					spec.NewJSONError("invalid username value", spec.InvalidRequestErrorCode),
				)
			}

			if len(req.Params) != 1 {
				return spec.NewJSONResponse(
					req.ID,
					nil,
					spec.NewJSONError("params must contain only the address", spec.InvalidParamsErrorCode),
				)
			}
			reward, err := rewarder.GetReward(ctx, username)
			if err != nil {
				return spec.NewJSONResponse(
					req.ID,
					nil,
					spec.NewJSONError(fmt.Sprintf("unable to get reward: %v", err), spec.ServerErrorCode),
				)
			}

			if reward == 0 {
				return spec.NewJSONResponse(
					req.ID,
					nil,
					spec.NewJSONError("no GNOTs to reward.", spec.ServerErrorCode),
				)
			}

			req.Method = faucet.DefaultDripMethod

			c := std.NewCoin("ugnot", int64(reward)).String()
			req.Params = append(req.Params, c)

			// TODO: this should be called AFTER the faucet successfully gives tokens to the requester
			if err := rewarder.Apply(ctx, username, reward); err != nil {
				return spec.NewJSONResponse(
					req.ID,
					nil,
					spec.NewJSONError(fmt.Sprintf("unable to apply reward: %v", err), spec.ServerErrorCode),
				)
			}

			return next(ctx, req)
		}
	}
}

// gitHubClaimMiddleware is the GitHub claim validation middleware, based on the provided username
func gitHubClaimMiddleware(coolDownLimiter cooldownLimiter) faucet.Middleware {
	return func(next faucet.HandlerFunc) faucet.HandlerFunc {
		return func(ctx context.Context, req *spec.BaseJSONRequest) *spec.BaseJSONResponse {
			if req.Method != faucet.DefaultDripMethod {
				return next(ctx, req)
			}

			// Grab the username from the context
			username, ok := ctx.Value(ghUsernameKey).(string)
			if !ok {
				return spec.NewJSONResponse(
					req.ID,
					nil,
					spec.NewJSONError("invalid username value", spec.InvalidRequestErrorCode),
				)
			}

			// Grab the claim amount
			if len(req.Params) < 2 {
				return spec.NewJSONResponse(
					req.ID,
					nil,
					spec.NewJSONError("amount not provided", spec.InvalidParamsErrorCode),
				)
			}

			claimAmount, err := std.ParseCoin(req.Params[1].(string))
			if err != nil {
				return spec.NewJSONResponse(
					req.ID,
					nil,
					spec.NewJSONError("invalid amount", spec.InvalidParamsErrorCode),
				)
			}

			// Just check if given account have asked for faucet before the cooldown period
			allowedToClaim, err := coolDownLimiter.checkCooldown(ctx, username, claimAmount.Amount)
			if err != nil {
				return spec.NewJSONResponse(
					req.ID,
					nil,
					spec.NewJSONError("unable to check cooldown", spec.ServerErrorCode),
				)
			}

			if !allowedToClaim {
				return spec.NewJSONResponse(
					req.ID,
					nil,
					spec.NewJSONError("user is on cooldown", spec.ServerErrorCode),
				)
			}

			return next(ctx, req)
		}
	}
}

// ghTokenResponse is the GitHub OAuth response
// for successful code exchanges
type ghTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// ghExchangeErrorResponse is the GitHub OAuth error response
type ghExchangeErrorResponse struct {
	Error       string `json:"error"`
	Description string `json:"error_description"`
	URI         string `json:"error_uri"`
}

//nolint:gosec
const githubTokenExchangeURL = "https://github.com/login/oauth/access_token"

type ghExchangeFn func(context.Context, string) (*github.User, error)

func defaultGHExchange(ctx context.Context, body string) (*github.User, error) {
	client := new(http.Client)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		githubTokenExchangeURL,
		strings.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create HTTP request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to post HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body into a byte slice so we can use it multiple times
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Attempt to decode as an error response.
	// The GitHub API returns 200 both for errors and valid response types
	var errorResponse ghExchangeErrorResponse
	if err := json.Unmarshal(respBody, &errorResponse); err == nil && errorResponse.Error != "" {
		return nil, fmt.Errorf("GitHub OAuth error: %s - %s", errorResponse.Error, errorResponse.Description)
	}

	// Attempt to decode as a token response
	var tokenResponse ghTokenResponse
	if err := json.Unmarshal(respBody, &tokenResponse); err != nil {
		return nil, err
	}

	// Make sure the response is set
	if tokenResponse.AccessToken == "" {
		return nil, errors.New("unable to exchange GitHub code for OAuth token")
	}

	// Fetch the user
	ghClient := github.NewClient(http.DefaultClient).WithAuthToken(tokenResponse.AccessToken)
	user, _, err := ghClient.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("unable to fetch GitHub user: %w", err)
	}

	return user, nil
}
