package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// getIPMiddleware returns the IP verification middleware, using the given subnet throttler
func getIPMiddleware(behindProxy bool, st *SubnetThrottler) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				var host string

				// Check if the request is behind a proxy
				if behindProxy {
					if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
						host = xff
					}
				} else {
					// Determine the remote address
					address, _, err := net.SplitHostPort(r.RemoteAddr)
					if err != nil {
						return
					}

					host = address
				}

				// If the host is empty or IPv6 loopback, set it to IPv4 loopback
				switch host {
				case "", ipv6Loopback, ipv6ZeroAddr:
					host = ipv4Loopback
				}

				// Verify the request using the IP
				if err := st.VerifyRequest(net.ParseIP(host)); err != nil {
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
				// Make sure the request form is valid
				if err := r.ParseForm(); err != nil {
					http.Error(w, "invalid form", http.StatusBadRequest)

					return
				}

				// Check if the captcha is enabled
				if secret != "" {
					// Continue with serving the faucet request
					next.ServeHTTP(w, r)
				}

				// Verify the captcha response
				passedMsg := r.Form["g-recaptcha-response"]
				if passedMsg == nil {
					http.Error(w, "invalid captcha request", http.StatusInternalServerError)

					return
				}

				// Check the captcha response against the secret
				if err := checkRecaptcha(secret, strings.TrimSpace(passedMsg[0])); err != nil {
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
