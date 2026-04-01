// Package safeurl provides URL safety validation using the SafeURL API.
package safeurl

import "time"

// SafetyStatus represents the safety classification of a URL.
type SafetyStatus int

const (
	// StatusUnknown indicates the URL has not been scanned yet.
	StatusUnknown SafetyStatus = iota
	// StatusPending indicates the URL scan is in progress.
	StatusPending
	// StatusSafe indicates the URL has been verified as safe.
	StatusSafe
	// StatusUnsafe indicates the URL has been identified as potentially malicious.
	StatusUnsafe
	// StatusUnavailable indicates the safety check failed (API timeout, error, etc.).
	StatusUnavailable
)

// String returns a human-readable representation of the safety status.
func (s SafetyStatus) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusSafe:
		return "safe"
	case StatusUnsafe:
		return "unsafe"
	case StatusUnavailable:
		return "unavailable"
	default:
		return "unknown"
	}
}

// ScanResult holds the result of a URL safety scan.
type ScanResult struct {
	// URL is the scanned URL.
	URL string
	// ScanID is the API scan ID (for polling pending scans).
	ScanID string
	// Status is the safety classification.
	Status SafetyStatus
	// Verdict is the raw verdict string from the API (e.g., "safe", "malicious").
	Verdict string
	// ScannedAt is when the scan was performed.
	ScannedAt time.Time
	// ExpiresAt is when this result should be considered stale.
	ExpiresAt time.Time
}

// IsExpired returns true if the scan result has expired.
func (r ScanResult) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

// ValidatorConfig holds configuration for the URL safety validator.
type ValidatorConfig struct {
	// Enabled determines if URL validation is active.
	Enabled bool
	// BaseURL is the SafeURL API base URL (default: https://api.safeurl.ai).
	BaseURL string
	// APIKey is the SafeURL API key.
	APIKey string
	// CacheSize is the maximum number of entries in the cache (default: 10000).
	CacheSize int
	// CacheTTL is how long scan results are cached (default: 24h).
	CacheTTL time.Duration
	// ScanTimeout is the timeout for API requests (default: 5s).
	ScanTimeout time.Duration
}

// DefaultValidatorConfig returns a ValidatorConfig with sensible defaults.
func DefaultValidatorConfig() ValidatorConfig {
	return ValidatorConfig{
		Enabled:     false,
		BaseURL:     "https://api.safeurl.ai",
		CacheSize:   10000,
		CacheTTL:    24 * time.Hour,
		ScanTimeout: 5 * time.Second,
	}
}
