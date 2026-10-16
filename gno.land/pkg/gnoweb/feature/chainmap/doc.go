// Package chainmap draws what is on the chain as a map: every `$map` URL, a
// treemap of the packages under a path, and the indexer-backed answers the
// dependency graph needs.
//
// The map shows exactly the listing the directory view shows. gnoweb lists the
// paths once and hands the same slice to either rendering, so the two can
// never disagree about what exists.
//
// With no indexer the map is complete and says what it measures: one tile per
// package, grouped by the next path segment. An indexer adds call activity and
// the packages that reference a given one. Whatever it adds is tagged as
// indexer data with its freshness, and a value it could not compute is shown
// as unknown, never as zero.
package chainmap
