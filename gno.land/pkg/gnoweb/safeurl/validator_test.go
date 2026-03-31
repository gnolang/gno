package safeurl

import (
	"testing"
	"time"
)

func TestValidator_Disabled(t *testing.T) {
	// Create disabled validator
	v, err := NewValidator(ValidatorConfig{Enabled: false}, nil)
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}

	if v.IsEnabled() {
		t.Error("expected validator to be disabled")
	}

	// Should return unknown status for all URLs
	results := v.ValidateURLs(nil, []string{"https://example.com", "https://test.com"})

	for url, result := range results {
		if result.Status != StatusUnknown {
			t.Errorf("URL %q: expected StatusUnknown, got %v", url, result.Status)
		}
	}
}

func TestValidator_DisabledWithEmptyAPIKey(t *testing.T) {
	// Even if Enabled=true, empty API key should disable
	v, err := NewValidator(ValidatorConfig{Enabled: true, APIKey: ""}, nil)
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}

	if v.IsEnabled() {
		t.Error("expected validator to be disabled when API key is empty")
	}
}

func TestIsExternalURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		// Internal URLs
		{"", false},
		{"#anchor", false},
		{"/relative/path", false},
		{"./relative", false},
		{"data:image/png;base64,abc", false},
		{"https://gno.land/r/demo", false},
		{"http://gno.land/p/demo", false},
		{"https://test.gno.land/r/demo", false},

		// External URLs
		{"https://example.com", true},
		{"http://malicious.site", true},
		{"https://google.com", true},
		{"ftp://files.example.com", true},
		// Note: Protocol-relative URLs (//...) are not handled as external
		// because they could resolve to gno.land on the same domain
	}

	for _, tt := range tests {
		got := isExternalURL(tt.url)
		if got != tt.expected {
			t.Errorf("isExternalURL(%q) = %v, want %v", tt.url, got, tt.expected)
		}
	}
}

func TestSafetyStatus_String(t *testing.T) {
	tests := []struct {
		status   SafetyStatus
		expected string
	}{
		{StatusUnknown, "unknown"},
		{StatusSafe, "safe"},
		{StatusUnsafe, "unsafe"},
		{StatusUnavailable, "unavailable"},
	}

	for _, tt := range tests {
		got := tt.status.String()
		if got != tt.expected {
			t.Errorf("SafetyStatus(%d).String() = %q, want %q", tt.status, got, tt.expected)
		}
	}
}

func TestScanResult_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		result   ScanResult
		expected bool
	}{
		{
			name:     "not expired",
			result:   ScanResult{ExpiresAt: timeNow().Add(time.Hour)},
			expected: false,
		},
		{
			name:     "expired",
			result:   ScanResult{ExpiresAt: timeNow().Add(-time.Hour)},
			expected: true,
		},
		{
			name:     "zero time (expired)",
			result:   ScanResult{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.IsExpired()
			if got != tt.expected {
				t.Errorf("IsExpired() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDefaultValidatorConfig(t *testing.T) {
	cfg := DefaultValidatorConfig()

	if cfg.Enabled {
		t.Error("default should be disabled")
	}
	if cfg.BaseURL != "https://api.safeurl.ai" {
		t.Errorf("unexpected BaseURL: %q", cfg.BaseURL)
	}
	if cfg.CacheSize != 10000 {
		t.Errorf("unexpected CacheSize: %d", cfg.CacheSize)
	}
	if cfg.CacheTTL != 24*time.Hour {
		t.Errorf("unexpected CacheTTL: %v", cfg.CacheTTL)
	}
	if cfg.ScanTimeout != 5*time.Second {
		t.Errorf("unexpected ScanTimeout: %v", cfg.ScanTimeout)
	}
}

// Helper to get current time (for testing)
func timeNow() time.Time {
	return time.Now()
}
