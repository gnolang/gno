// Package chainmap draws what is on the chain under a path, in the listing's
// two renderings: every `$map` URL, a treemap of the packages, and the list,
// one card per package. It also gives the indexer-backed answers the
// dependency graph needs.
//
// The map shows exactly the listing the list shows. gnoweb lists the paths
// once and hands the same slice to either rendering, so the two can never
// disagree about what exists.
//
// With no indexer both are complete and say what they measure: the map draws
// one tile per package, grouped by the next path segment, and each card what
// the chain records about its package. An indexer adds call activity, deploy
// dates and the packages that reference a given one. On the map, whatever it
// adds is tagged as indexer data with its freshness; on a card, its tooltip
// names the source. A value it could not compute is shown as unknown, or not
// at all, never as zero.
package chainmap
