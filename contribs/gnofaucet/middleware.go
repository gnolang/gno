package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/gnolang/faucet"
	"github.com/gnolang/faucet/spec"
)

// ipMiddleware returns the IP verification middleware, using the given subnet throttler
func ipMiddleware(behindProxy bool, st *ipThrottler) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")

				// Determine the remote address
				host, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					http.Error(
						w,
						fmt.Sprintf("invalid request IP and port, %s", err.Error()),
						http.StatusUnauthorized,
					)

					return
				}

				// Check if the request is behind a proxy
				if xff := r.Header.Get("X-Forwarded-For"); xff != "" && behindProxy {
					host = xff
				}

				// If the host is empty or IPv6 loopback, set it to IPv4 loopback
				switch host {
				case "", ipv6Loopback, ipv6ZeroAddr:
					host = ipv4Loopback
				}

				// Make sure the host IP is valid
				hostAddr, err := netip.ParseAddr(host)
				if err != nil {
					http.Error(
						w,
						fmt.Sprintf("invalid request IP, %s", err.Error()),
						http.StatusUnauthorized,
					)

					return
				}

				// Register the request with the throttler
				if err := st.registerNewRequest(hostAddr); err != nil {
					http.Error(
						w,
						fmt.Sprintf("unable to verify IP request, %s", err.Error()),
						http.StatusUnauthorized,
					)

					return
				}

				// Continue with serving the faucet request
				next.ServeHTTP(w, r)
			},
		)
	}
}

// captchaMiddleware returns the captcha middleware, if any
func captchaMiddleware(secret string) faucet.Middleware {
	return func(next faucet.HandlerFunc) faucet.HandlerFunc {
		return func(ctx context.Context, req *spec.BaseJSONRequest) *spec.BaseJSONResponse {
			// Parse the request meta to extract the captcha secret
			var meta struct {
				Captcha string `json:"captcha"`
			}

			// Decode the original request
			if err := json.NewDecoder(bytes.NewBuffer(req.Meta)).Decode(&meta); err != nil {
				return spec.NewJSONResponse(
					req.ID,
					nil,
					spec.NewJSONError("invalid captcha request", spec.InvalidRequestErrorCode),
				)
			}

			// Verify the captcha response
			if err := checkRecaptcha(secret, strings.TrimSpace(meta.Captcha)); err != nil {
				return spec.NewJSONResponse(
					req.ID,
					nil,
					spec.NewJSONError("invalid captcha", spec.InvalidParamsErrorCode),
				)
			}

			// Continue with serving the faucet request
			return next(ctx, req)
		}
	}
}

// checkRecaptcha checks the captcha challenge
func checkRecaptcha(secret, response string) error {
	// Create an HTTP client with a timeout
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	// Create the request
	req, err := http.NewRequest(
		http.MethodPost,
		siteVerifyURL,
		nil,
	)
	if err != nil {
		return fmt.Errorf("unable to create request, %w", err)
	}

	// Craft the request query string
	q := req.URL.Query()
	q.Add("secret", secret)
	q.Add("response", response)
	req.URL.RawQuery = q.Encode()

	// Execute the verify request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to execute request, %w", err)
	}
	defer resp.Body.Close()

	// Verify the captcha-verify response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code, %d", resp.StatusCode)
	}

	// Decode the response body
	var body SiteVerifyResponse
	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("failed to decode response, %w", err)
	}

	// Check if the recaptcha verification was successful
	if !body.Success {
		return errInvalidCaptcha
	}

	return nil
}
