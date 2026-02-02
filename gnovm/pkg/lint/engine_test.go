package lint

import (
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

type mockReporter struct {
	issues   []Issue
	flushed  bool
	info     int
	warnings int
	errors   int
}

func (r *mockReporter) Report(issue Issue) {
	r.issues = append(r.issues, issue)
	switch issue.Severity {
	case SeverityInfo:
		r.info++
	case SeverityWarning:
		r.warnings++
	case SeverityError:
		r.errors++
	}
}

func (r *mockReporter) Flush() error {
	r.flushed = true
	return nil
}

func (r *mockReporter) Summary() (info, warnings, errors int) {
	return r.info, r.warnings, r.errors
}

type alwaysIssueRule struct{}

func (alwaysIssueRule) Info() RuleInfo {
	return RuleInfo{
		ID:       "TEST001",
		Category: CategoryGeneral,
		Name:     "always-issue",
		Severity: SeverityWarning,
	}
}

func (alwaysIssueRule) Check(ctx *RuleContext, node gnolang.Node) []Issue {
	if _, ok := node.(*gnolang.FuncDecl); ok {
		return []Issue{
			NewIssue("TEST001", SeverityWarning, "test issue", ctx.File.FileName, node.GetPos()),
		}
	}
	return nil
}

func TestEngine_Run_NoRules(t *testing.T) {
	cfg := DefaultConfig()
	reg := NewRegistry() // empty registry
	rep := &mockReporter{}

	engine := NewEngine(cfg, reg, rep)

	fset := &gnolang.FileSet{}
	sources := map[string]string{}

	count := engine.Run(fset, sources)
	if count != 0 {
		t.Errorf("Run() returned %d, want 0 for no rules", count)
	}
}

func TestEngine_Run_NoFiles(t *testing.T) {
	cfg := DefaultConfig()
	reg := NewRegistry()
	reg.MustRegister(&alwaysIssueRule{})
	rep := &mockReporter{}

	engine := NewEngine(cfg, reg, rep)

	fset := &gnolang.FileSet{} // empty fileset
	sources := map[string]string{}

	count := engine.Run(fset, sources)
	if count != 0 {
		t.Errorf("Run() returned %d, want 0 for no files", count)
	}
}
