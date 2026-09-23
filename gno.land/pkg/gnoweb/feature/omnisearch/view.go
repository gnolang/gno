package omnisearch

import (
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

// OmnisearchViewType tags the body view. Local because nothing in components
// renders this feature.
const OmnisearchViewType components.ViewType = "omnisearch-view"

// SearchData is the render payload for templates/page.html — field names must
// match the template.
type SearchData struct {
	Query string

	// PkgPath is the scoped package, gnoweb-relative; empty when the URL
	// names none.
	PkgPath string

	// FormAction is explicit rather than `action=""`, which would re-submit
	// the webargs already in the path and grow the URL on every search.
	FormAction string

	Groups []Group

	// Selectors is the honest inventory of what this deployment can answer.
	Selectors []*Selector

	// UnknownFilter names a qualifier that is neither a selector nor a
	// narrowing keyword, so the page can say so instead of showing nothing.
	UnknownFilter string

	// Indexer is set only when an indexer-backed group ran; nil drops the
	// provenance footer entirely.
	Indexer *IndexerStatus
}

// IndexerStatus is the provenance footer: indexer results are not consensus
// data, so the reader gets the endpoint and how far behind it is.
type IndexerStatus struct {
	URL       string
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
