package safeurl

import (
	"context"
	"log/slog"
	"strings"
)

// Validator provides URL safety validation with graceful degradation.
type Validator struct {
	client  *Client
	enabled bool
	logger  *slog.Logger
}

// NewValidator creates a new URL safety validator.
// If cfg.Enabled is false or cfg.APIKey is empty, the validator will be disabled
// and all URLs will be returned as StatusUnknown (allowing them through).
func NewValidator(cfg ValidatorConfig, logger *slog.Logger) (*Validator, error) {
	if logger == nil {
		logger = slog.Default()
	}

	// Disabled mode - no API calls
	if !cfg.Enabled || cfg.APIKey == "" {
		logger.Info("SafeURL validation disabled")
		return &Validator{
			enabled: false,
			logger:  logger,
		}, nil
	}

	// Apply defaults
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultValidatorConfig().BaseURL
	}
	if cfg.CacheSize <= 0 {
		cfg.CacheSize = DefaultValidatorConfig().CacheSize
	}
	if cfg.CacheTTL <= 0 {
		cfg.CacheTTL = DefaultValidatorConfig().CacheTTL
	}
	if cfg.ScanTimeout <= 0 {
		cfg.ScanTimeout = DefaultValidatorConfig().ScanTimeout
	}

	cache := NewCache(cfg.CacheSize, cfg.CacheTTL)
	client, err := NewClient(cfg.BaseURL, cfg.APIKey, cache, logger, cfg.ScanTimeout)
	if err != nil {
		return nil, err
	}

	logger.Info("SafeURL validation enabled",
		"base_url", cfg.BaseURL,
		"cache_size", cfg.CacheSize,
		"cache_ttl", cfg.CacheTTL,
		"scan_timeout", cfg.ScanTimeout,
	)

	return &Validator{
		client:  client,
		enabled: true,
		logger:  logger,
	}, nil
}

// IsEnabled returns whether SafeURL validation is active.
func (v *Validator) IsEnabled() bool {
	return v.enabled
}

// ValidateURLs validates multiple URLs and returns their safety status.
// If the validator is disabled, all URLs are returned as StatusUnknown.
// External URLs are validated via the SafeURL API; internal URLs are returned as StatusSafe.
// This method uses async scanning - it submits scans and returns immediately.
// Pending scans will have StatusPending and include a ScanID for polling.
func (v *Validator) ValidateURLs(ctx context.Context, urls []string) map[string]ScanResult {
	results := make(map[string]ScanResult, len(urls))

	if !v.enabled {
		// Return unknown status for all URLs (pass-through mode)
		for _, url := range urls {
			results[url] = ScanResult{
				URL:    url,
				Status: StatusUnknown,
			}
		}
		return results
	}

	// Separate external URLs from internal ones
	var externalURLs []string
	for _, url := range urls {
		if IsExternalURL(url) {
			externalURLs = append(externalURLs, url)
		} else {
			// Internal URLs are always safe
			results[url] = ScanResult{
				URL:    url,
				Status: StatusSafe,
			}
		}
	}

	if len(externalURLs) == 0 {
		return results
	}

	// Submit external URLs for scanning (async - returns immediately)
	scanResults, err := v.client.SubmitURLs(ctx, externalURLs)
	if err != nil {
		v.logger.Error("failed to submit URLs for scan", "error", err)
		// Mark all external URLs as unavailable on error
		for _, url := range externalURLs {
			results[url] = ScanResult{
				URL:    url,
				Status: StatusUnavailable,
			}
		}
		return results
	}

	// Merge scan results
	for url, result := range scanResults {
		results[url] = result
	}

	return results
}

// GetScanStatus retrieves the current status of a scan by ID.
func (v *Validator) GetScanStatus(ctx context.Context, scanID string) (*ScanResult, error) {
	if !v.enabled {
		return nil, nil
	}
	return v.client.GetScan(ctx, scanID)
}

// IsExternalURL checks if a URL is external (requires safety validation).
// A URL is considered external if it has a scheme (http://, https://, etc.)
// and is not a gno.land domain. Relative URLs, anchors, data URIs, and
// protocol-relative URLs resolving to gno.land are considered internal.
func IsExternalURL(url string) bool {
	// Empty or anchor-only URLs are internal
	if url == "" || strings.HasPrefix(url, "#") {
		return false
	}

	// Relative URLs are internal (but not protocol-relative)
	if strings.HasPrefix(url, "/") && !strings.HasPrefix(url, "//") {
		return false
	}

	// Data URIs don't need external validation
	if strings.HasPrefix(url, "data:") {
		return false
	}

	// Check for scheme
	if strings.Contains(url, "://") {
		// gno.land URLs are internal, all others are external
		lowerURL := strings.ToLower(url)
		return !strings.Contains(lowerURL, "gno.land")
	}

	// Protocol-relative URLs (//example.com) are external
	if strings.HasPrefix(url, "//") {
		return true
	}

	// No scheme - could be relative
	return false
}
