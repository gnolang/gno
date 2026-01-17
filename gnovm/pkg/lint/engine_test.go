package lint

import (
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// mockReporter captures reported issues for testing.
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

// alwaysIssueRule returns an issue for every function declaration.
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
	// Only report on specific node types to avoid too many issues
	if _, ok := node.(*gnolang.FuncDecl); ok {
		return []Issue{
			NewIssue("TEST001", SeverityWarning, "test issue", ctx.File.FileName, node.GetPos()),
		}
	}
	return nil
}

type neverIssueRule struct{}

func (neverIssueRule) Info() RuleInfo {
	return RuleInfo{
		ID:       "TEST002",
		Category: CategoryGeneral,
		Name:     "never-issue",
		Severity: SeverityWarning,
	}
}

func (neverIssueRule) Check(ctx *RuleContext, node gnolang.Node) []Issue {
	return nil
}

func TestNewEngine(t *testing.T) {
	cfg := DefaultConfig()
	reg := NewRegistry()
	rep := &mockReporter{}

	engine := NewEngine(cfg, reg, rep)

	if engine == nil {
		t.Fatal("NewEngine() returned nil")
	}
	if engine.config != cfg {
		t.Error("config not set correctly")
	}
	if engine.registry != reg {
		t.Error("registry not set correctly")
	}
	if engine.reporter != rep {
		t.Error("reporter not set correctly")
	}
}

func TestEngine_getEnabledRules(t *testing.T) {
	cfg := DefaultConfig()
	reg := NewRegistry()
	rep := &mockReporter{}

	// Register some rules
	reg.MustRegister(&alwaysIssueRule{})
	reg.MustRegister(&neverIssueRule{})

	engine := NewEngine(cfg, reg, rep)
	rules := engine.getEnabledRules()

	if len(rules) != 2 {
		t.Errorf("getEnabledRules() returned %d rules, want 2", len(rules))
	}
}

func TestEngine_Flush(t *testing.T) {
	cfg := DefaultConfig()
	reg := NewRegistry()
	rep := &mockReporter{}

	engine := NewEngine(cfg, reg, rep)

	err := engine.Flush()
	if err != nil {
		t.Fatalf("Flush() error = %v", err)
	}
	if !rep.flushed {
		t.Error("Flush() did not call reporter.Flush()")
	}
}

func TestEngine_Summary(t *testing.T) {
	cfg := DefaultConfig()
	reg := NewRegistry()
	rep := &mockReporter{
		info:     1,
		warnings: 2,
		errors:   3,
	}

	engine := NewEngine(cfg, reg, rep)

	info, warnings, errors := engine.Summary()
	if info != 1 {
		t.Errorf("Summary() info = %v, want 1", info)
	}
	if warnings != 2 {
		t.Errorf("Summary() warnings = %v, want 2", warnings)
	}
	if errors != 3 {
		t.Errorf("Summary() errors = %v, want 3", errors)
	}
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
