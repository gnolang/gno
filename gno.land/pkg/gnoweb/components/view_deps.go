package components

const DepsViewType ViewType = "deps-view"

// DepGraph is the dependency graph around one package (ui/pkg_graph): what it
// imports, read from the chain, and what imports it, which only an indexer
// can propose.
type DepGraph struct {
	// Name is the package's gnoweb path: at the centre of the graph a bare
	// "v0" would say nothing.
	Name    string
	Imports []ImportLink

	// LookupURL links to the page that computes importers. Set where this
	// deployment can compute them but the page does not.
	LookupURL string

	// Importers is set on the page that computed them.
	Importers *Importers
}

// Importers is the "imported by" side of a DepGraph.
type Importers struct {
	Links []ImportLink
	// AtLeast says some candidates went unchecked: there may be more.
	AtLeast bool
	// Unavailable says why there is no answer; Links is then empty.
	Unavailable string
}

// DepsData is the payload of the dependencies page.
type DepsData struct {
	PkgPath string
	// Title names the package in the heading, "ufmt/v0" for a versioned one.
	Title string
	Graph DepGraph
	// Indexer is the provenance footer, set when the indexer answered.
	Indexer *IndexerStatus
}

// DepsView renders a package's dependency graph with its importers.
func DepsView(data DepsData) *View {
	view := NewTemplateView(DepsViewType, "renderDeps", data)
	view.SkipTargetInBody = true // on the content header
	return view
}
