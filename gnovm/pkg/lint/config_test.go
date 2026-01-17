package lint

import "testing"

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}
	if cfg.Mode != ModeDefault {
		t.Errorf("Mode = %v, want %v", cfg.Mode, ModeDefault)
	}
	if cfg.Format != "text" {
		t.Errorf("Format = %v, want 'text'", cfg.Format)
	}
}

func TestMode_Constants(t *testing.T) {
	tests := []struct {
		mode     Mode
		expected string
	}{
		{ModeDefault, "default"},
		{ModeStrict, "strict"},
		{ModeWarnOnly, "warn-only"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.mode) != tt.expected {
				t.Errorf("mode = %v, want %v", tt.mode, tt.expected)
			}
		})
	}
}

func TestConfig_IsRuleEnabled(t *testing.T) {
	cfg := DefaultConfig()

	// Currently all rules are enabled by default
	tests := []struct {
		ruleID string
		want   bool
	}{
		{"AVL001", true},
		{"GLOBAL001", true},
		{"NONEXISTENT", true}, // Currently returns true for all
	}
	for _, tt := range tests {
		t.Run(tt.ruleID, func(t *testing.T) {
			if got := cfg.IsRuleEnabled(tt.ruleID); got != tt.want {
				t.Errorf("IsRuleEnabled(%q) = %v, want %v", tt.ruleID, got, tt.want)
			}
		})
	}
}

func TestConfig_EffectiveSeverity(t *testing.T) {
	tests := []struct {
		name     string
		mode     Mode
		input    Severity
		expected Severity
	}{
		// Default mode - no changes
		{"default/info", ModeDefault, SeverityInfo, SeverityInfo},
		{"default/warning", ModeDefault, SeverityWarning, SeverityWarning},
		{"default/error", ModeDefault, SeverityError, SeverityError},

		// Strict mode - warnings become errors
		{"strict/info", ModeStrict, SeverityInfo, SeverityInfo},
		{"strict/warning", ModeStrict, SeverityWarning, SeverityError},
		{"strict/error", ModeStrict, SeverityError, SeverityError},

		// Warn-only mode - errors become warnings
		{"warn-only/info", ModeWarnOnly, SeverityInfo, SeverityInfo},
		{"warn-only/warning", ModeWarnOnly, SeverityWarning, SeverityWarning},
		{"warn-only/error", ModeWarnOnly, SeverityError, SeverityWarning},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Mode: tt.mode}
			got := cfg.EffectiveSeverity(tt.input)
			if got != tt.expected {
				t.Errorf("EffectiveSeverity(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
