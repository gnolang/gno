package components

import (
	"strings"
	"testing"
)

func renderGraph(t *testing.T, g DepGraph) string {
	t.Helper()
	var b strings.Builder
	if err := NewTemplateComponent("ui/pkg_graph", g).Render(&b); err != nil {
		t.Fatalf("render: %v", err)
	}
	return b.String()
}

// An answer with unchecked candidates and no confirmed importer is not an
// answer of zero: the side must not say nothing imports the package.
func TestPkgGraphNeverClaimsNoneWhenUnchecked(t *testing.T) {
	t.Parallel()

	for _, imp := range []*Importers{{Capped: true}, {Retry: true}, {Capped: true, Retry: true}} {
		out := renderGraph(t, DepGraph{Name: "/p/a", Importers: imp})
		if strings.Contains(out, "No live package imports it.") {
			t.Errorf("%+v: a partial answer with no confirmed importer claims there are none", imp)
		}
		if strings.Contains(out, "at least 0") {
			t.Errorf("%+v: a partial answer must not show a count of at least 0", imp)
		}
		if !strings.Contains(out, "None confirmed among the candidates checked.") {
			t.Errorf("%+v: a partial empty answer must say none was confirmed so far", imp)
		}
	}

	// Each cause is said as such: a limit a reload cannot lift, reads a
	// reload may complete, or both.
	capped := renderGraph(t, DepGraph{Name: "/p/a", Importers: &Importers{Capped: true}})
	if !strings.Contains(capped, "stopped at its limit") || strings.Contains(capped, "Reload") {
		t.Error("a capped answer must say the search stopped, without offering a reload")
	}
	retry := renderGraph(t, DepGraph{Name: "/p/a", Importers: &Importers{Retry: true}})
	if !strings.Contains(retry, "could not be checked") || !strings.Contains(retry, "Reload") || strings.Contains(retry, "stopped at its limit") {
		t.Error("failed reads must say so and offer a reload, and nothing else")
	}
	both := renderGraph(t, DepGraph{Name: "/p/a", Importers: &Importers{Capped: true, Retry: true}})
	if !strings.Contains(both, "stopped at its limit") || !strings.Contains(both, "could not be checked") {
		t.Error("both causes must be said when both apply")
	}

	complete := renderGraph(t, DepGraph{Name: "/p/a", Importers: &Importers{}})
	if !strings.Contains(complete, "No live package imports it.") {
		t.Error("a complete empty answer should say none imports it")
	}
}

// Without importers or a lookup link the graph has no "imported by" side at
// all: nothing mentions what this deployment cannot answer.
func TestPkgGraphWithoutIndexerShowsImportsOnly(t *testing.T) {
	t.Parallel()

	out := renderGraph(t, DepGraph{Name: "/p/a", Imports: []ImportLink{{Path: "strings", Kind: "stdlib"}}})
	if strings.Contains(out, "Imported by") || strings.Contains(out, "indexer") {
		t.Error("the graph mentions importers with no indexer")
	}
}
