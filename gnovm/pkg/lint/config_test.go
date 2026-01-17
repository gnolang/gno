package lint

import "testing"

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
