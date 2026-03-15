package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	"github.com/gnolang/faucet"
	"github.com/gnolang/faucet/spec"
)

// contextKey is an unexported type for context keys in this package.
type contextKey int

const remoteIPContextKey contextKey = iota

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

				// Store the resolved IP in the context for use by RPC middlewares
				ctx := context.WithValue(r.Context(), remoteIPContextKey, hostAddr.String())

				// Continue with serving the faucet request
				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}

// captchaMiddleware returns the captcha middleware, if any
func captchaMiddleware(secret, sitekey string, logger *slog.Logger) faucet.Middleware {
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

			// Extract the resolved client IP stored by ipMiddleware
			remoteIP, _ := ctx.Value(remoteIPContextKey).(string)

			// Verify the captcha response
			if err := checkHcaptcha(secret, strings.TrimSpace(meta.Captcha), remoteIP, sitekey, logger); err != nil {
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

// checkHcaptcha checks the captcha challenge
func checkHcaptcha(secret, response, remoteIP, sitekey string, logger *slog.Logger) error {
	// Create an HTTP client with a timeout
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	// Craft the form-encoded request body
	form := url.Values{}
	form.Set("secret", secret)
	form.Set("response", response)
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}
	if sitekey != "" {
		form.Set("sitekey", sitekey)
	}

	logger.Debug("sending hcaptcha verification request",
		slog.String("remoteip", remoteIP),
		slog.Bool("secret_set", secret != ""),
		slog.Bool("sitekey_set", sitekey != ""),
	)

	// Create the request
	req, err := http.NewRequest(
		http.MethodPost,
		siteVerifyURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return fmt.Errorf("unable to create request, %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

	logger.Debug("received hcaptcha verification response",
		slog.String("remoteip", remoteIP),
		slog.Bool("success", body.Success),
		slog.String("hostname", body.Hostname),
		slog.Any("error_codes", body.ErrorCodes),
	)

	// Check if the hcaptcha verification was successful
	if !body.Success {
		return errInvalidCaptcha
	}

	return nil
}
