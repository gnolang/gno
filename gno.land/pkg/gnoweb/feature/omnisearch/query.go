package omnisearch

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

// Input bounds. A search box is attacker-controlled and every resolver turns
// its terms into a chain or indexer query, so it is screened before any
// fetch. Mirrors ADR-003 §Resource bounds.
const (
	// MaxQueryLen fits a couple of qualifiers plus a bech32 address.
	MaxQueryLen = 256

	// Past a handful, qualifiers stop narrowing and start being free work.
	MaxFilters = 8

	// Below MinTermLen a query matches most of the chain.
	MinTermLen = 2

	// MaxResults caps one group: relevant, not exhaustive.
	MaxResults = 20
)

var (
	// ErrQueryTooLong reports a `q` over MaxQueryLen.
	ErrQueryTooLong = errors.New("query too long")

	// ErrTooManyFilters reports more than MaxFilters qualifiers.
	ErrTooManyFilters = errors.New("too many search qualifiers")

	// ErrQueryInvalid rejects rather than strips: the value round-trips into
	// HTML and log lines, where a CR or LF is someone else's bug.
	ErrQueryInvalid = errors.New("query contains control characters")
)

// Filter is one `key:value` qualifier.
type Filter struct {
	Key   string
	Value string
}

// Query is a parsed search input: any number of `key:value` qualifiers plus
// the free text left over.
type Query struct {
	// Raw is the validated input, as typed; Text is what is left once the
	// qualifiers are removed.
	Raw  string
	Text string

	Filters []Filter

	// PkgPath is gnoweb-relative ("/r/demo/boards"), the form Doc/ListFiles
	// take and every href is built from. ChainPath is the same package
	// domain-qualified, the form the chain and the indexer use. Not
	// interchangeable.
	PkgPath   string
	ChainPath string

	// jsonPath marks a query from the omnibar rather than a page navigation,
	// which is what gates the fan-out selectors.
	jsonPath bool

	// formAction is the bare path plus `$search`, never the current URL:
	// `action=""` would re-submit the webargs already in the path.
	formAction string
}

// ParseQuery validates and tokenises a raw `q` value. Each qualifier splits
// on its FIRST colon only: a value may legitimately contain one.
func ParseQuery(raw string) (*Query, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) > MaxQueryLen {
		return nil, ErrQueryTooLong
	}
	for _, r := range raw {
		if unicode.IsControl(r) {
			return nil, ErrQueryInvalid
		}
	}

	q := &Query{Raw: raw}
	var text []string
	for _, tok := range strings.Fields(raw) {
		key, value, found := strings.Cut(tok, ":")
		key = strings.ToLower(strings.TrimSpace(key))
		if !found || !isQualifierKey(key) || strings.HasPrefix(value, "//") {
			text = append(text, tok)
			continue
		}
		q.Filters = append(q.Filters, Filter{Key: key, Value: strings.TrimSpace(value)})
	}
	if len(q.Filters) > MaxFilters {
		return nil, ErrTooManyFilters
	}
	q.Text = strings.Join(text, " ")
	return q, nil
}

// qualifierKey: a qualifier's key is one lowercase bare word. This is what
// keeps a gno path with arguments (`/r/gnoland/pages:p/about`, which the
// omnibar prefills) and a pasted URL from parsing as a search.
var qualifierKey = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

func isQualifierKey(key string) bool {
	return qualifierKey.MatchString(key)
}

// Get returns the first value for a qualifier, and whether it was present.
func (q *Query) Get(key string) (string, bool) {
	for _, f := range q.Filters {
		if f.Key == key {
			return f.Value, true
		}
	}
	return "", false
}

// HasBareWord matches a no-argument selector typed as a bare word.
func (q *Query) HasBareWord(word string) bool {
	return strings.EqualFold(strings.TrimSpace(q.Text), word)
}
