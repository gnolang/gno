package gnolang

import "testing"

func TestSemanticsForVersion(t *testing.T) {
	t.Parallel()
	if s, ok := SemanticsForVersion(GnoVerLatest); !ok || s.Version != GnoVerLatest {
		t.Fatalf("latest version must be registered, got %+v ok=%v", s, ok)
	}
	if _, ok := SemanticsForVersion("0.0"); ok {
		t.Fatalf("unregistered version must report ok=false")
	}
}

func TestPackageNodeSemanticsDefaultsDormant(t *testing.T) {
	t.Parallel()
	pn := NewPackageNode("main", "main", nil)
	if pn.LangVersion != GnoVerLatest {
		t.Fatalf("new package must default to latest, got %q", pn.LangVersion)
	}
	if pn.Semantics().Version != GnoVerLatest {
		t.Fatalf("semantics must resolve to latest")
	}
	// Unrecognized version falls back to latest (dormant, never fails).
	pn.LangVersion = "0.0"
	if pn.Semantics().Version != GnoVerLatest {
		t.Fatalf("unknown version must fall back to latest")
	}
}
