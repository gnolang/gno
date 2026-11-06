package gnolang

import "testing"

func TestSemanticsForVersion(t *testing.T) {
	t.Parallel()
	if s, ok := SemanticsForVersion(GnoVerLatest); !ok || s.Version != GnoVerLatest {
		t.Fatalf("latest version must be registered, got %+v ok=%v", s, ok)
	}
	if _, ok := SemanticsForVersion("nonesuch"); ok {
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
	pn.LangVersion = "nonesuch" // unregistered → dormant fallback, never fails
	if pn.Semantics().Version != GnoVerLatest {
		t.Fatalf("unknown version must fall back to latest")
	}
}

// TestPerPackageSemanticsDispatch is the load-bearing instrument: it
// proves the seam actually routes by each package's OWN version, not a
// global/constant. Since production registers a single version, it
// registers a SYNTHETIC second version and asserts two packages pinned
// to different versions resolve to different Semantics. Without this the
// dormant seam would be unfalsifiable (a stub always returning latest
// would pass every other test).
func TestPerPackageSemanticsDispatch(t *testing.T) {
	// not parallel: mutates the package-global registry.
	const synthetic = "test-v99"
	cleanup := registerSemanticsForTest(Semantics{Version: synthetic})
	defer cleanup()

	pkgLatest := NewPackageNode("a", "gno.land/p/a", nil) // defaults to latest
	pkgSynthetic := NewPackageNode("b", "gno.land/p/b", nil)
	pkgSynthetic.LangVersion = synthetic

	// Each package resolves to ITS OWN version — the whole point of
	// per-package pinning.
	if got := pkgLatest.Semantics().Version; got != GnoVerLatest {
		t.Fatalf("latest-pinned package resolved to %q, want %q", got, GnoVerLatest)
	}
	if got := pkgSynthetic.Semantics().Version; got != synthetic {
		t.Fatalf("synthetic-pinned package resolved to %q, want %q", got, synthetic)
	}
	if pkgLatest.Semantics().Version == pkgSynthetic.Semantics().Version {
		t.Fatal("two packages on different versions resolved to the SAME semantics — dispatch is not per-package")
	}

	// After cleanup the synthetic version is gone → falls back to latest.
	cleanup()
	if _, ok := SemanticsForVersion(synthetic); ok {
		t.Fatal("synthetic version leaked past cleanup")
	}
	if got := pkgSynthetic.Semantics().Version; got != GnoVerLatest {
		t.Fatalf("after deregistration, unknown version must fall back to latest, got %q", got)
	}
}
