package safeurl

import (
	"context"
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

// SubmitURLs submits URLs for scanning without waiting for completion.
// Returns immediately with pending status for URLs not in cache.
// Use GetScan to poll for results.
func (c *Client) SubmitURLs(ctx context.Context, urls []string) (map[string]ScanResult, error) {
	if len(urls) == 0 {
		return make(map[string]ScanResult), nil
	}

	// Check cache first
	results, missing := c.cache.GetMulti(urls)

	if len(missing) == 0 {
		c.logger.Debug("all URLs found in cache", "count", len(urls))
		return results, nil
	}

	c.logger.Debug("submitting URLs for scan", "cached", len(urls)-len(missing), "to_scan", len(missing))

	// Submit missing URLs - returns immediately without waiting
	scanResults, err := c.scanner.SubmitURLs(ctx, missing)
	if err != nil {
		c.logger.Warn("SafeURL submit failed", "error", err, "url_count", len(missing))
		// Mark failed URLs as unavailable
		for _, url := range missing {
			results[url] = ScanResult{
				URL:       url,
				Status:    StatusUnavailable,
				ScannedAt: time.Now(),
				ExpiresAt: time.Now().Add(5 * time.Minute),
			}
		}
		return results, nil
	}

	// Convert SDK results to our ScanResult type
	for url, scan := range scanResults {
		result := convertScanResult(url, scan)
		results[url] = result
		// Only cache completed results
		if result.Status != StatusPending {
			c.cache.Set(url, result)
		}
	}

	// Handle any URLs that weren't in the response
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

// GetScan retrieves the current status of a scan by ID.
func (c *Client) GetScan(ctx context.Context, scanID string) (*ScanResult, error) {
	scan, err := c.scanner.GetScan(ctx, scanID)
	if err != nil {
		return nil, err
	}

	result := convertScanResult(scan.URL, scan)

	// Cache completed results
	if result.Status != StatusPending {
		c.cache.Set(scan.URL, result)
	}

	return &result, nil
}

// convertScanResult converts an SDK ScanResponse to our internal ScanResult type.
func convertScanResult(url string, scan *sdk.ScanResponse) ScanResult {
	// Check if scan is still in progress
	if !scan.State.IsTerminal() {
		return ScanResult{
			ScanID:    scan.ID,
			URL:       url,
			Status:    StatusPending,
			ScannedAt: time.Now(),
			ExpiresAt: time.Now().Add(5 * time.Minute), // Short TTL for pending
		}
	}

	// Scan is complete, determine status
	status := StatusSafe
	verdict := scan.GetVerdict()

	switch {
	case scan.State == sdk.ScanStateFailed:
		status = StatusUnavailable
	case verdict.IsUnsafe():
		status = StatusUnsafe
	case !verdict.IsSafe() && verdict != sdk.VerdictUnknown:
		status = StatusUnavailable
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	if scan.ExpiresAt != nil {
		expiresAt = *scan.ExpiresAt
	}

	return ScanResult{
		ScanID:    scan.ID,
		URL:       url,
		Status:    status,
		Verdict:   string(verdict),
		ScannedAt: time.Now(),
		ExpiresAt: expiresAt,
	}
}
