package omnisearch

import (
	"context"
	"html/template"
	"slices"
)

// Source travels to the template because the distinction is not cosmetic:
// chain results are consensus data, indexer results are one operator's best
// effort.
type Source string

const (
	// SourceChain marks a result answered by the RPC node.
	SourceChain Source = "chain"

	// SourceIndexer is never consensus data.
	SourceIndexer Source = "indexer"
)

// Scope says what a selector needs in order to answer.
type Scope string

const (
	// ScopePackage needs a package path, from the URL or from `in:`.
	ScopePackage Scope = "package"

	// ScopeGlobal answers from the term alone.
	ScopeGlobal Scope = "global"
)

// Narrowing qualifiers refine a search without choosing what is searched.
// Listed so an unrecognised one is reported as a typo, not ignored.
const (
	FilterAuthor = "author"
	FilterIn     = "in"
	FilterIs     = "is"
)

func isNarrowing(key string) bool {
	switch key {
	case FilterAuthor, FilterIn, FilterIs:
		return true
	}
	return false
}

// Selector chooses what a query searches. Exactly one runs per query —
// running all of them would turn one keystroke into a dozen queries.
type Selector struct {
	// Name is the qualifier key, e.g. "tx" in `tx:9f2a`.
	Name string

	// Hint is the form shown in the omnibar's suggestion list.
	Hint  string
	Label string
	Scope Scope

	// Source is stamped by register from the table the selector is in, not
	// by the literal. See Handler.register.
	Source Source

	// Bare marks a selector typed without a colon (`imports`, `activity`).
	Bare bool

	// MinTerm raises MinTermLen where the term's selectivity sets the cost.
	MinTerm int

	// PageOnly keeps a fan-out selector off the per-keystroke JSON path.
	PageOnly bool

	// resolve answers the query; term is the selector's own value.
	resolve func(ctx context.Context, h *Handler, q *Query, term string) ([]Result, error)
}

// Result is one row in the results list.
type Result struct {
	Title  string
	Detail string

	// Href is typed so escaping is enforced by construction. Empty when the
	// row has no destination.
	Href template.URL

	// Tags are short badges. Provenance is not among them — it belongs to
	// the group, and Group.Source carries it once.
	Tags []string
}

// Group is one answered subject: several for a discovery search, one when a
// selector chose the subject.
type Group struct {
	Label   string
	Source  Source
	Results []Result

	// Err renders as a failure rather than being dropped: "could not ask" is
	// not "nothing matched".
	Err error

	// Truncated marks an answer built from a capped listing.
	Truncated bool
}

// minTerm is the selector's own floor, falling back to the global one.
func (s *Selector) minTerm() int {
	if s.MinTerm > 0 {
		return s.MinTerm
	}
	return MinTermLen
}

// capResults clones rather than reslices: callers pre-size against the whole
// package, so a reslice would pin that array for the life of the page.
func capResults(rs []Result) []Result {
	if len(rs) > MaxResults {
		return slices.Clone(rs[:MaxResults])
	}
	return rs
}
