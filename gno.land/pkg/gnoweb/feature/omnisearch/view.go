package omnisearch

import (
	"html/template"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

// OmnisearchViewType tags the body view.
const OmnisearchViewType = components.SearchViewType

// SearchData is the render payload for templates/page.html — field names must
// match the template.
type SearchData struct {
	Query string
	// QueryIsLabel: the query is a bare keyword its group's heading already
	// names, so the title does not repeat it.
	QueryIsLabel bool

	// PkgPath is the scoped package, gnoweb-relative; empty when the URL
	// names none.
	PkgPath string

	// WholeChainHref, set on a scoped discovery search, repeats it without
	// the scope.
	WholeChainHref string

	Groups []Group

	// Selectors is the honest inventory of what this deployment can answer.
	Selectors []*Selector

	// UnknownFilter names a qualifier that is neither a selector nor a
	// narrowing keyword, so the page can say so instead of showing nothing.
	UnknownFilter string

	// Indexer is set only when an indexer-backed group ran; nil drops the
	// provenance footer entirely.
	Indexer *IndexerStatus

	// RailTitle and RailLinks are what the rail offers under the counts: the
	// package's keywords on a package page, a few queries to try when there
	// is no answer. Empty otherwise: the counts are enough.
	RailTitle string
	RailLinks []RailLink
}

// RailLink is one search the rail runs in a click.
type RailLink struct {
	Label string
	Query string
	Href  template.URL
}

// IndexerStatus is the provenance footer: indexer results are not consensus
// data, so the reader gets how far behind it is. Not the endpoint, which is
// often a host the reader cannot reach.
type IndexerStatus struct {
	LastBlock int
	// Err makes the footer say the freshness is unknown rather than omit it.
	Err error
}

// HasResults tells "nothing matched" from "nothing was asked".
func (d SearchData) HasResults() bool {
	for _, g := range d.Groups {
		if len(g.Results) > 0 {
			return true
		}
	}
	return false
}

// HasErrors reports a group that could not answer: "Nothing matched." under
// "Could not answer" would contradict it.
func (d SearchData) HasErrors() bool {
	for _, g := range d.Groups {
		if g.Err != nil {
			return true
		}
	}
	return false
}
