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

func TestConfig_IsRuleEnabled(t *testing.T) {
	tests := []struct {
		name     string
		disable  map[string]bool
		ruleID   string
		expected bool
	}{
		{"no disabled rules", nil, "AVL001", true},
		{"empty disabled rules", map[string]bool{}, "AVL001", true},
		{"rule not in disabled list", map[string]bool{"GLOBAL001": true}, "AVL001", true},
		{"rule in disabled list", map[string]bool{"AVL001": true}, "AVL001", false},
		{"rule in disabled list with others", map[string]bool{"GLOBAL001": true, "AVL001": true, "OTHER": true}, "AVL001", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Disable: tt.disable}
			got := cfg.IsRuleEnabled(tt.ruleID)
			if got != tt.expected {
				t.Errorf("IsRuleEnabled(%q) = %v, want %v", tt.ruleID, got, tt.expected)
			}
		})
	}
}
