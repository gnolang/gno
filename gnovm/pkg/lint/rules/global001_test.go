package rules

import (
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/lint"
)

func TestGLOBAL001_Info(t *testing.T) {
	rule := &GLOBAL001{}
	info := rule.Info()

	if info.ID != "GLOBAL001" {
		t.Errorf("ID = %v, want GLOBAL001", info.ID)
	}
	if info.Category != lint.CategoryGeneral {
		t.Errorf("Category = %v, want CategoryGeneral", info.Category)
	}
	if info.Name != "exported-global-variable" {
		t.Errorf("Name = %v, want exported-global-variable", info.Name)
	}
	if info.Severity != lint.SeverityWarning {
		t.Errorf("Severity = %v, want SeverityWarning", info.Severity)
	}
}

func TestGLOBAL001_Check_NotValueDecl(t *testing.T) {
	rule := &GLOBAL001{}
	ctx := &lint.RuleContext{}

	// GLOBAL001 only checks ValueDecl nodes
	// Passing nil should return nil issues
	issues := rule.Check(ctx, nil)
	if issues != nil {
		t.Errorf("Check(nil) = %v, want nil", issues)
	}
}

func TestIsExported(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"uppercase", "Foo", true},
		{"lowercase", "foo", false},
		{"uppercase with numbers", "Foo123", true},
		{"lowercase with numbers", "foo123", false},
		{"underscore start", "_Foo", false},
		{"single uppercase", "F", true},
		{"single lowercase", "f", false},
		{"empty string", "", false},
		{"unicode uppercase", "Über", true},
		{"unicode lowercase", "über", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isExported(tt.input)
			if got != tt.expected {
				t.Errorf("isExported(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsFileLevelDecl(t *testing.T) {
	// Test with nil parents
	if isFileLevelDecl(nil) {
		t.Error("isFileLevelDecl(nil) = true, want false")
	}

	// Note: Testing with actual FileNode requires constructing gnolang nodes,
	// which is complex. Full testing is done through integration tests (txtar).
}
