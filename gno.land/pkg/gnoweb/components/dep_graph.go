package components

// DepsViewType identifies feature/chainmap's dependencies page, so
// layout_index.go gives it the dev chrome of the Source pages it extends
// without importing the feature, which already imports components.
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

// Shown reports whether the graph has anything to draw on the overview:
// imports, or importers to look up. The section, its jump link and its
// table-of-contents entry all follow it.
func (g DepGraph) Shown() bool { return len(g.Imports) > 0 || g.LookupURL != "" }

// Importers is the "imported by" side of a DepGraph.
type Importers struct {
	Links []ImportLink
	// Capped says the search stopped at a limit, so there may be more; a
	// reload cannot lift it.
	Capped bool
	// Retry says some candidates could not be read, which a reload may
	// complete.
	Retry bool
	// Unavailable says why there is no answer; Links is then empty.
	Unavailable string
}
