package lint

import (
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// testRule is a simple rule for testing the registry.
type testRule struct {
	id string
}

func (r *testRule) Info() RuleInfo {
	return RuleInfo{
		ID:       r.id,
		Category: CategoryGeneral,
		Name:     "test-rule",
		Severity: SeverityWarning,
	}
}

func (r *testRule) Check(ctx *RuleContext, node gnolang.Node) []Issue {
	return nil
}

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry() returned nil")
	}
	if r.rules == nil {
		t.Error("rules map should be initialized")
	}
	if len(r.rules) != 0 {
		t.Error("new registry should be empty")
	}
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()

	rule1 := &testRule{id: "TEST001"}
	err := r.Register(rule1)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Verify rule is registered
	got, ok := r.Get("TEST001")
	if !ok {
		t.Error("registered rule should be retrievable")
	}
	if got != rule1 {
		t.Error("retrieved rule should match registered rule")
	}
}

func TestRegistry_Register_Duplicate(t *testing.T) {
	r := NewRegistry()

	rule1 := &testRule{id: "TEST001"}
	rule2 := &testRule{id: "TEST001"}

	err := r.Register(rule1)
	if err != nil {
		t.Fatalf("first Register() error = %v", err)
	}

	err = r.Register(rule2)
	if err == nil {
		t.Error("duplicate registration should return error")
	}
}

func TestRegistry_MustRegister(t *testing.T) {
	r := NewRegistry()

	rule := &testRule{id: "TEST001"}

	// Should not panic
	r.MustRegister(rule)

	// Verify registration
	_, ok := r.Get("TEST001")
	if !ok {
		t.Error("MustRegister should register the rule")
	}
}

func TestRegistry_MustRegister_Panic(t *testing.T) {
	r := NewRegistry()

	rule1 := &testRule{id: "TEST001"}
	rule2 := &testRule{id: "TEST001"}

	r.MustRegister(rule1)

	defer func() {
		if recover() == nil {
			t.Error("MustRegister should panic on duplicate")
		}
	}()

	r.MustRegister(rule2)
}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry()

	rule := &testRule{id: "TEST001"}
	r.MustRegister(rule)

	tests := []struct {
		name   string
		id     string
		wantOk bool
	}{
		{"existing", "TEST001", true},
		{"non-existing", "TEST999", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := r.Get(tt.id)
			if ok != tt.wantOk {
				t.Errorf("Get(%q) ok = %v, want %v", tt.id, ok, tt.wantOk)
			}
		})
	}
}

func TestRegistry_All(t *testing.T) {
	r := NewRegistry()

	// Empty registry
	all := r.All()
	if len(all) != 0 {
		t.Error("All() on empty registry should return empty slice")
	}

	// Add rules
	rule1 := &testRule{id: "TEST001"}
	rule2 := &testRule{id: "TEST002"}
	rule3 := &testRule{id: "TEST003"}

	r.MustRegister(rule1)
	r.MustRegister(rule2)
	r.MustRegister(rule3)

	all = r.All()
	if len(all) != 3 {
		t.Errorf("All() returned %d rules, want 3", len(all))
	}

	// Verify all rules are present (order may vary)
	ids := make(map[string]bool)
	for _, rule := range all {
		ids[rule.Info().ID] = true
	}
	for _, id := range []string{"TEST001", "TEST002", "TEST003"} {
		if !ids[id] {
			t.Errorf("All() missing rule %s", id)
		}
	}
}

func TestDefaultRegistry(t *testing.T) {
	if DefaultRegistry == nil {
		t.Error("DefaultRegistry should be initialized")
	}
}

func TestPackageLevelRegister(t *testing.T) {
	// Create a new registry for testing to avoid polluting DefaultRegistry
	oldRegistry := DefaultRegistry
	DefaultRegistry = NewRegistry()
	defer func() { DefaultRegistry = oldRegistry }()

	rule := &testRule{id: "PKG001"}

	err := Register(rule)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	_, ok := DefaultRegistry.Get("PKG001")
	if !ok {
		t.Error("package-level Register should add to DefaultRegistry")
	}
}

func TestPackageLevelMustRegister(t *testing.T) {
	oldRegistry := DefaultRegistry
	DefaultRegistry = NewRegistry()
	defer func() { DefaultRegistry = oldRegistry }()

	rule := &testRule{id: "PKG002"}

	MustRegister(rule)

	_, ok := DefaultRegistry.Get("PKG002")
	if !ok {
		t.Error("package-level MustRegister should add to DefaultRegistry")
	}
}
