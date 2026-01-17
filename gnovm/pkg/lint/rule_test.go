package lint

import (
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// mockRule is a test implementation of the Rule interface.
type mockRule struct {
	id       string
	category RuleCategory
	name     string
	severity Severity
	issues   []Issue
}

func (r *mockRule) Info() RuleInfo {
	return RuleInfo{
		ID:       r.id,
		Category: r.category,
		Name:     r.name,
		Severity: r.severity,
	}
}

func (r *mockRule) Check(ctx *RuleContext, node gnolang.Node) []Issue {
	return r.issues
}

func TestRuleInfo(t *testing.T) {
	rule := &mockRule{
		id:       "TEST001",
		category: CategoryAVL,
		name:     "test-rule",
		severity: SeverityWarning,
	}

	info := rule.Info()

	if info.ID != "TEST001" {
		t.Errorf("ID = %v, want TEST001", info.ID)
	}
	if info.Category != CategoryAVL {
		t.Errorf("Category = %v, want CategoryAVL", info.Category)
	}
	if info.Name != "test-rule" {
		t.Errorf("Name = %v, want test-rule", info.Name)
	}
	if info.Severity != SeverityWarning {
		t.Errorf("Severity = %v, want SeverityWarning", info.Severity)
	}
}

func TestRuleContext(t *testing.T) {
	ctx := &RuleContext{
		File:    nil,
		Source:  "package main\n\nfunc main() {}\n",
		Parents: []gnolang.Node{},
	}

	if ctx.Source == "" {
		t.Error("Source should not be empty")
	}
	if ctx.Parents == nil {
		t.Error("Parents should not be nil")
	}
}

func TestRuleCategories(t *testing.T) {
	tests := []struct {
		category RuleCategory
		expected string
	}{
		{CategoryAVL, "AVL"},
		{CategoryGeneral, "General"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.category) != tt.expected {
				t.Errorf("category = %v, want %v", tt.category, tt.expected)
			}
		})
	}
}
