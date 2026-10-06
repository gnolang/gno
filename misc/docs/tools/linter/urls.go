package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"mvdan.cc/xurls/v2"
)

var httpClient = &http.Client{
	Timeout: 15 * time.Second,
}

// extractUrls extracts urls from given file content
func extractUrls(fileContent []byte) []string {
	scanner := bufio.NewScanner(bytes.NewReader(fileContent))
	urls := make([]string, 0)

	// Scan file line by line
	for scanner.Scan() {
		line := scanner.Text()

		// Extract links
		rxStrict := xurls.Strict()
		url := rxStrict.FindString(line)

		// Check for empty links and skip them
		if url == " " || len(url) == 0 {
			continue
		}

		// Look for http & https only
		if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
			if shouldCheckURL(url) {
				urls = append(urls, url)
			}
		}
	}

	return urls
}

// skippedURLSubstrings are URLs we deliberately do not reach out to, because a
// failure would say nothing about whether the documentation link is correct.
var skippedURLSubstrings = []string{
	// Not routable from CI.
	"localhost",
	"127.0.0.1",
	// Placeholder for examples.
	"example.land",
	// archive.org subdomains rate-limit and intermittently 502 from CI
	// runners; the canonical Aaron Swartz Manifesto URL lives there and its
	// reachability is not a CI concern.
	"archive.org",
	// YouTube blocks/rate-limits requests from data-center (CI) IPs,
	// returning 404/429 even for live videos; its reachability is not a CI
	// concern.
	"youtube.com",
	"youtu.be",
	// Deployment-specific hosts whose uptime is not a CI concern.
	"staging.gno.land",
}

// shouldCheckURL reports whether url is worth reaching out to from CI.
func shouldCheckURL(rawURL string) bool {
	for _, s := range skippedURLSubstrings {
		if strings.Contains(rawURL, s) {
			return false
		}
	}

	return !isRPCEndpoint(rawURL)
}

// isRPCEndpoint reports whether url points at a chain RPC endpoint (an `rpc.*`
// host, e.g. https://rpc.gno.land:443).
//
// These are JSON-RPC APIs, not documentation pages: a plain GET tells us
// nothing about whether the documented endpoint is the right one, while it does
// make docs CI fail whenever a live chain is down, rotating, or — as happened
// with rpc.gno.land — serving an expired TLS certificate. Their availability is
// an infrastructure concern, not a documentation one.
func isRPCEndpoint(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		// Not parseable as a URL; leave the decision to the caller's other
		// checks rather than silently skipping it.
		return false
	}

	return strings.HasPrefix(u.Hostname(), "rpc.")
}

func lintURLs(ctx context.Context, filepathToURLs map[string][]string, treatAsError bool) (string, error) {
	// Setup parallel checking for links
	g, _ := errgroup.WithContext(ctx)

	var (
		lock            sync.Mutex
		output          bytes.Buffer
		hasInvalidLinks bool
	)

	for filePath, urls := range filepathToURLs {
		filePath := filePath
		for _, url := range urls {
			url := url
			g.Go(func() error {
				if err := checkUrl(url); err != nil {
					lock.Lock()
					if !hasInvalidLinks {
						output.WriteString("Remote links that need checking:\n")
						hasInvalidLinks = true
					}

					output.WriteString(fmt.Sprintf(">>> %s (found in file: %s)\n", url, filePath))
					lock.Unlock()
				}

				return nil
			})
		}
	}

	// Check for possible thread errors
	if err := g.Wait(); err != nil {
		return "", err
	}

	if !treatAsError {
		errFound404Links = nil
	}
	if hasInvalidLinks {
		return output.String(), errFound404Links
	}

	return "", nil
}

// checkUrl checks if a URL is a 404, retrying on transient errors.
func checkUrl(url string) error {
	const maxRetries = 3

	for attempt := range maxRetries {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}

		resp, err := httpClient.Get(url)
		if err != nil {
			// Network error: retry unless this is the last attempt.
			if attempt < maxRetries-1 {
				continue
			}

			return err404Link
		}
		resp.Body.Close()

		switch {
		case resp.StatusCode == http.StatusNotFound:
			// 404 is a definitive failure; no point retrying.
			return err404Link
		case resp.StatusCode == http.StatusTooManyRequests:
			// Rate-limited; retry.
			if attempt < maxRetries-1 {
				continue
			}

			return err404Link
		}

		// Treat everything else (including 5xx) as reachable.
		return nil
	}

	return nil
}
