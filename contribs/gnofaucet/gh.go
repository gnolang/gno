package main

import (
	"bytes"
	"context"
	"encoding/json"
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

type gitHubTokenResponse struct {
	AccessToken string `json:"access_token"`
}

//nolint:gosec
const githubTokenExchangeURL = "https://github.com/login/oauth/access_token"

var exchangeCodeForUser = func(ctx context.Context, secret, clientID, code string) (*github.User, error) {
	body := fmt.Sprintf("client_id=%s&client_secret=%s&code=%s", clientID, secret, code)
	req, err := http.NewRequest("POST", githubTokenExchangeURL, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenResponse gitHubTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, err
	}

	if tokenResponse.AccessToken == "" {
		return nil, fmt.Errorf("unable to exchange code for token")
	}

	ghClient := github.NewClient(http.DefaultClient).WithAuthToken(tokenResponse.AccessToken)
	user, _, err := ghClient.Users.Get(ctx, "")
	return user, err
}
