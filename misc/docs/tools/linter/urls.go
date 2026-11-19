package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
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
func extractUrls(fileContent []byte) []urlRef {
	scanner := bufio.NewScanner(bytes.NewReader(fileContent))
	urls := make([]urlRef, 0)

	// Scan file line by line
	for lineNo := 1; scanner.Scan(); lineNo++ {
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
				urls = append(urls, urlRef{url: url, line: lineNo})
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

// urlRef is a URL found in a file, with the 1-based line it sits on.
type urlRef struct {
	url  string
	line int
}

const (
	// maxConcurrentChecks bounds the requests in flight. Nearly every remote
	// link in docs/ points at github.com, so firing them all at once from a
	// shared CI runner is what gets the runner throttled.
	maxConcurrentChecks = 8

	// maxAttempts and initialBackoff give a throttled or flaky URL four tries
	// over about 7s (1s, 2s, 4s), stretched by Retry-After up to maxBackoff.
	maxAttempts    = 4
	initialBackoff = time.Second
	maxBackoff     = 10 * time.Second
)

// urlResult is the outcome of checking one URL.
type urlResult int

const (
	urlOK         urlResult = iota
	urlDead                 // definitive: 404, 410, or a host that does not resolve
	urlUnverified           // still throttled, 5xx or unreachable after every attempt
)

// urlChecker checks remote URLs. It is a value so tests can shrink the
// backoff and point it at a stub server.
type urlChecker struct {
	client   *http.Client
	attempts int
	backoff  time.Duration
}

func lintURLs(ctx context.Context, filepathToURLs map[string][]urlRef, treatAsError bool) (string, error) {
	c := urlChecker{client: httpClient, attempts: maxAttempts, backoff: initialBackoff}
	return c.lint(ctx, filepathToURLs, treatAsError)
}

func (c urlChecker) lint(ctx context.Context, filepathToURLs map[string][]urlRef, treatAsError bool) (string, error) {
	// Check each distinct URL once, however many places cite it.
	where := make(map[string][]string) // url > ["file:line"]
	for filePath, refs := range filepathToURLs {
		for _, ref := range refs {
			where[ref.url] = append(where[ref.url], fmt.Sprintf("%s:%d", filePath, ref.line))
		}
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(maxConcurrentChecks)

	var (
		lock             sync.Mutex
		lines            []string
		dead, unverified int
	)

	for url, locs := range where {
		g.Go(func() error {
			res, reason := c.check(ctx, url)
			if res == urlOK {
				return nil
			}

			label := "dead"
			if res == urlUnverified {
				label = "unverified"
			}

			lock.Lock()
			defer lock.Unlock()
			if res == urlDead {
				dead++
			} else {
				unverified++
			}
			for _, loc := range locs {
				lines = append(lines, fmt.Sprintf("%s: %s %s: %s\n", loc, label, url, reason))
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return "", err
	}

	if len(lines) == 0 {
		return "", nil
	}

	sort.Strings(lines)

	var output bytes.Buffer
	output.WriteString("Remote links that need checking:\n")
	for _, l := range lines {
		output.WriteString(l)
	}
	output.WriteString(fmt.Sprintf(
		"%d of %d remote URLs failed: %d dead, %d unverified\n",
		dead+unverified, len(where), dead, unverified,
	))
	if unverified > 0 {
		output.WriteString("Unverified URLs (throttled, 5xx or unreachable after retries) are reported but never fail the lint.\n")
	}

	if dead > 0 && treatAsError {
		return output.String(), errFound404Links
	}

	return output.String(), nil
}

// check fetches url, retrying with exponential backoff on 429, 5xx and
// transport errors. It returns why a URL failed, so the report can say it.
func (c urlChecker) check(ctx context.Context, url string) (urlResult, string) {
	var (
		reason string
		tried  int
	)

	for attempt := range c.attempts {
		tried++

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return urlDead, err.Error()
		}

		var hint time.Duration

		resp, err := c.client.Do(req)
		switch {
		case err != nil:
			var dnsErr *net.DNSError
			if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
				return urlDead, err.Error()
			}

			reason = err.Error()
		case resp.StatusCode == http.StatusNotFound, resp.StatusCode == http.StatusGone:
			resp.Body.Close()
			return urlDead, "HTTP " + resp.Status
		case resp.StatusCode == http.StatusTooManyRequests, resp.StatusCode >= 500:
			resp.Body.Close()
			reason = "HTTP " + resp.Status
			hint = retryAfter(resp)
		default:
			// Everything else (2xx, 401, 403, ...) says the page exists.
			resp.Body.Close()
			return urlOK, ""
		}

		if !c.wait(ctx, attempt, hint) {
			break
		}
	}

	return urlUnverified, fmt.Sprintf("%s (after %d attempts)", reason, tried)
}

// wait sleeps before the next attempt, at least hint when the server asked for
// it. It returns false when there is no next attempt or ctx is done.
func (c urlChecker) wait(ctx context.Context, attempt int, hint time.Duration) bool {
	if attempt+1 >= c.attempts {
		return false
	}

	d := max(c.backoff<<attempt, hint)
	d = min(d, maxBackoff)

	select {
	case <-time.After(d):
		return true
	case <-ctx.Done():
		return false
	}
}

// retryAfter parses a Retry-After header given in seconds.
func retryAfter(resp *http.Response) time.Duration {
	secs, err := strconv.Atoi(resp.Header.Get("Retry-After"))
	if err != nil || secs < 0 {
		return 0
	}

	return time.Duration(secs) * time.Second
}
