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
		if isExternalURL(url) {
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

	// Scan external URLs
	scanResults, err := v.client.ScanURLs(ctx, externalURLs)
	if err != nil {
		v.logger.Error("failed to scan URLs", "error", err)
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

// isExternalURL checks if a URL is external (not a relative or gno.land URL).
func isExternalURL(url string) bool {
	// Empty or relative URLs are internal
	if url == "" || strings.HasPrefix(url, "/") || strings.HasPrefix(url, "#") {
		return false
	}

	// Data URIs are not external
	if strings.HasPrefix(url, "data:") {
		return false
	}

	// Check for scheme
	if strings.Contains(url, "://") {
		// gno.land URLs are internal
		if strings.Contains(url, "gno.land") {
			return false
		}
		return true
	}

	// No scheme - likely relative
	return false
}
