package safeurl

import (
	"context"
	"errors"
	"log/slog"
	"time"

	sdk "github.com/gnoverse/safeurl-sdk/go"
)

// Client wraps the SafeURL SDK Scanner with caching.
type Client struct {
	scanner *sdk.Scanner
	cache   *Cache
	logger  *slog.Logger
}

// NewClient creates a new SafeURL client with caching.
func NewClient(baseURL, apiKey string, cache *Cache, logger *slog.Logger, timeout time.Duration) (*Client, error) {
	if baseURL == "" {
		baseURL = sdk.DefaultBaseURL
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	scanner, err := sdk.NewScannerWithBaseURL(baseURL, apiKey,
		sdk.WithMaxWait(timeout),
		sdk.WithPollInterval(200*time.Millisecond),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		scanner: scanner,
		cache:   cache,
		logger:  logger,
	}, nil
}

// ScanURLs scans multiple URLs for safety, using the cache where possible.
// Returns a map of URL to ScanResult.
func (c *Client) ScanURLs(ctx context.Context, urls []string) (map[string]ScanResult, error) {
	if len(urls) == 0 {
		return make(map[string]ScanResult), nil
	}

	// Check cache first
	results, missing := c.cache.GetMulti(urls)

	if len(missing) == 0 {
		c.logger.Debug("all URLs found in cache", "count", len(urls))
		return results, nil
	}

	c.logger.Debug("scanning URLs", "cached", len(urls)-len(missing), "to_scan", len(missing))

	// Scan missing URLs - SDK handles batching and polling automatically
	scanResults, err := c.scanner.ScanURLs(ctx, missing)
	if err != nil {
		c.logger.Warn("SafeURL scan failed", "error", err, "url_count", len(missing))
		// Mark failed URLs as unavailable
		for _, url := range missing {
			results[url] = ScanResult{
				URL:       url,
				Status:    StatusUnavailable,
				ScannedAt: time.Now(),
				ExpiresAt: time.Now().Add(5 * time.Minute), // Short TTL for failures
			}
		}
		return results, nil
	}

	// Convert SDK results to our ScanResult type and cache them
	for url, scan := range scanResults {
		result := convertScanResult(url, scan)
		results[url] = result
		c.cache.Set(url, result)
	}

	// Handle any URLs that weren't in the response (shouldn't happen, but be safe)
	for _, url := range missing {
		if _, ok := results[url]; !ok {
			results[url] = ScanResult{
				URL:       url,
				Status:    StatusUnavailable,
				ScannedAt: time.Now(),
				ExpiresAt: time.Now().Add(5 * time.Minute),
			}
		}
	}

	return results, nil
}

// convertScanResult converts an SDK ScanResponse to our internal ScanResult type.
func convertScanResult(url string, scan *sdk.ScanResponse) ScanResult {
	status := StatusSafe
	verdict := string(scan.Verdict)

	switch {
	case scan.State == sdk.ScanStateFailed:
		status = StatusUnavailable
	case scan.Verdict.IsUnsafe():
		status = StatusUnsafe
	case !scan.Verdict.IsSafe() && scan.Verdict != sdk.VerdictUnknown:
		// Treat unexpected verdicts as unavailable
		status = StatusUnavailable
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	if scan.ExpiresAt != nil {
		expiresAt = *scan.ExpiresAt
	}

	return ScanResult{
		URL:       url,
		Status:    status,
		Verdict:   verdict,
		ScannedAt: time.Now(),
		ExpiresAt: expiresAt,
	}
}

// IsTimeout returns true if the error is a timeout error from the SDK.
func IsTimeout(err error) bool {
	return errors.Is(err, sdk.ErrTimeout)
}
