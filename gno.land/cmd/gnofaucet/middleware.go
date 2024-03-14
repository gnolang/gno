package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"
)

// getIPMiddleware returns the IP verification middleware, using the given subnet throttler
func getIPMiddleware(behindProxy bool, st *ipThrottler) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
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

// getCaptchaMiddleware returns the captcha middleware, if any
func getCaptchaMiddleware(secret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				// Check if the captcha is enabled
				if secret == "" {
					// Continue with serving the faucet request
					next.ServeHTTP(w, r)

					return
				}

				// Parse the request to extract the captcha secret
				var request struct {
					Captcha string `json:"captcha"`
				}

				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "unable to read request body", http.StatusInternalServerError)

					return
				}

				// Close the original body
				if err := r.Body.Close(); err != nil {
					http.Error(w, "unable to close request body", http.StatusInternalServerError)

					return
				}

				// Create a new ReadCloser from the read bytes
				// so that future middleware will be able to read
				r.Body = io.NopCloser(bytes.NewReader(body))

				// Decode the original request
				if err := json.NewDecoder(bytes.NewBuffer(body)).Decode(&request); err != nil {
					http.Error(w, "invalid captcha request", http.StatusBadRequest)

					return
				}

				// Verify the captcha response
				if err := checkRecaptcha(secret, strings.TrimSpace(request.Captcha)); err != nil {
					http.Error(w, "invalid captcha", http.StatusUnauthorized)

					return
				}

				// Continue with serving the faucet request
				next.ServeHTTP(w, r)
			},
		)
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
