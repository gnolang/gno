package lint

import "github.com/gnolang/gno/gnovm/pkg/gnolang"

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
