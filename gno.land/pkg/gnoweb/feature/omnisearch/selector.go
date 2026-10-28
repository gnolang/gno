package omnisearch

import (
	"context"
	"html/template"
	"regexp"
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

// FilterKeyword describes a narrowing qualifier for the keyword lists: the
// results page's and the omnibar's, so both offer the same ones.
type FilterKeyword struct {
	Name  string `json:"name"`
	Hint  string `json:"hint"`
	Help  string `json:"help"`
	Group string `json:"group"`

	// Values is every value the filter takes, when there are few.
	Values []string `json:"values,omitempty"`

	// Complete names the list the omnibar suggests values from, one it
	// already holds: "namespace" or "path".
	Complete string `json:"complete,omitempty"`
}

// filterKeywords is the one list of narrowing qualifiers: isNarrowing reads
// it, so a keyword cannot be offered without being accepted.
var filterKeywords = []FilterKeyword{
	{Name: FilterAuthor, Hint: "author:<name>", Help: "One namespace or deployer", Group: GroupNarrow, Complete: "namespace"},
	{Name: FilterIn, Hint: "in:<path>", Help: "One package (/r/… or /p/…)", Group: GroupNarrow, Complete: "path"},
	{Name: FilterIs, Hint: "is:<realm|package>", Help: "Only realms, or only packages", Group: GroupNarrow, Values: []string{"realm", "package"}},
}

func isNarrowing(key string) bool {
	return slices.ContainsFunc(filterKeywords, func(f FilterKeyword) bool { return f.Name == key })
}

// The keyword lists group the keywords that work anywhere by what they do.
// The scope already groups those about a package.
const (
	// GroupText searches text: rendered pages, deployed source.
	GroupText = "text"
	// GroupLookup finds one thing by its key: a transaction, a block.
	GroupLookup = "lookup"
	// GroupNarrow is the narrowing qualifiers.
	GroupNarrow = "narrow"
)

// KeywordGroup heads one group of the keyword lists.
type KeywordGroup struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// keywordGroups is the one list of headings, in the order shown: the results
// page's popover ranges over it and the omnibar receives it.
var keywordGroups = []KeywordGroup{
	{GroupText, "Find text"},
	{GroupLookup, "Look up"},
	{GroupNarrow, "Narrow results"},
}

// Selector chooses what a query searches. Exactly one runs per query —
// running all of them would turn one keystroke into a dozen queries.
type Selector struct {
	// Name is the qualifier key, e.g. "tx" in `tx:9f2a`.
	Name string

	// Hint is the form shown in the omnibar's suggestion list.
	Hint string
	// Label heads the selector's results group; Help says what it does,
	// beside the hint in the keyword lists.
	Label string
	Help  string
	Scope Scope
	// Group heads the keyword in the keyword lists, when its scope is global.
	Group string

	// Source is stamped by register from the table the selector is in, not
	// by the literal. See Handler.register.
	Source Source

	// Bare marks a selector typed without a colon (`imports`, `activity`).
	Bare bool

	// MinTerm raises MinTermLen where the term's selectivity sets the cost.
	MinTerm int

	// PageOnly keeps a fan-out selector off the per-keystroke JSON path.
	PageOnly bool

	// Accepts matches a value this selector answers when it is typed alone,
	// without its key: a pasted address is an `account:` search. The JSON
	// answer carries it, so the omnibar asks the server for the same input.
	Accepts *regexp.Regexp

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

	// Truncated marks a partial answer: more may exist than Results holds.
	// Notice says why, in the reader's terms.
	Truncated bool
	Notice    string
}

// minTerm is the selector's own floor, falling back to the global one.
func (s *Selector) minTerm() int {
	if s.MinTerm > 0 {
		return s.MinTerm
	}
	return MinTermLen
}

// capResults clones rather than reslices: callers pre-size against the whole
// package, so a reslice would pin that array for the life of the page. It
// reports whether it cut, so the cut is never silent.
func capResults(rs []Result) ([]Result, bool) {
	if len(rs) > MaxResults {
		return slices.Clone(rs[:MaxResults]), true
	}
	return rs, false
}

// partialAnswerError is a resolver's verdict that its rows are not the whole
// answer. It travels as an error so every resolver keeps one signature, and
// Search turns it into a notice rather than a failure.
type partialAnswerError string

func (p partialAnswerError) Error() string { return string(p) }

// recentNotice is what a reader is told when the indexer walk stopped short
// of genesis.
const recentNotice = "Searched the most recent blocks only: older transactions may exist."
