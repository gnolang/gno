package components

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
