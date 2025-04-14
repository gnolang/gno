package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v64/github"
)

// getGithubMiddleware sets up authentication middleware for GitHub OAuth.
// If clientID and secret are empty, the middleware does nothing.
//
// Parameters:
// - clientID: The OAuth client ID issued by GitHub when registering the application.
// - secret: The OAuth client secret used to securely authenticate API requests.
// - cooldown: A cooldown duration to prevent several claims from the same user.
//
// GitHub OAuth applications require a client ID and secret to authenticate users securely.
// These credentials are obtained when registering an application on GitHub at:
// https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authenticating-to-the-rest-api-with-an-oauth-app#registering-your-app
func getGithubMiddleware(clientID, secret string, coolDownLimiter *CooldownLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
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

				user, err := exchangeCodeForUser(r.Context(), secret, clientID, code)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				claimAmount, err := getClaimAmount(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				// Just check if given account have asked for faucet before the cooldown period
				allowedToClaim, err := coolDownLimiter.CheckCooldown(r.Context(), user.GetLogin(), claimAmount)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if !allowedToClaim {
					http.Error(w, "user is on cooldown", http.StatusTooManyRequests)
					return
				}

				// Possibility to have more conditions like accountAge, commits, pullRequest, etc.
				next.ServeHTTP(w, r)
			},
		)
	}
}

type request struct {
	Amount string `json:"amount"`
}

func getClaimAmount(r *http.Request) (int64, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return 0, err
	}

	var data request
	err = json.Unmarshal(body, &data)
	if err != nil {
		return 0, err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	// amount sent is a string, so we need to convert it to int64
	// Ex: "1000000ugnot" -> 1000000
	// Regex to extract leading digits
	re := regexp.MustCompile(`^\d+`)
	numericPart := re.FindString(data.Amount)

	value, err := strconv.ParseInt(numericPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid claim amount, %w", err)
	}
	return value, nil
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

var exchangeCodeForUser = func(ctx context.Context, secret, clientID, code string) (*github.User, error) {
	client := new(http.Client)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		githubTokenExchangeURL,
		strings.NewReader(fmt.Sprintf("client_id=%s&client_secret=%s&code=%s", clientID, secret, code)))
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
