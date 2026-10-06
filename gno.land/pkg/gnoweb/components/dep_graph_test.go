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

	out := renderGraph(t, DepGraph{Name: "/p/a", Importers: &Importers{AtLeast: true}})
	if strings.Contains(out, "No live package imports it.") {
		t.Error("a partial answer with no confirmed importer claims there are none")
	}
	if strings.Contains(out, "at least 0") {
		t.Error("a partial answer must not show a count of at least 0")
	}
	if !strings.Contains(out, "could not be checked") {
		t.Error("a partial answer must say candidates went unchecked")
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
