package omnisearch

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// Request budgets: a page has a reader who committed to a navigation, the
// omnibar has one typing.
const (
	pageTimeout = 6 * time.Second
	jsonTimeout = 3 * time.Second
)

// Handle is the entry point for every `$search` URL. A nil view return means
// the body has already been written.
func (h *Handler) Handle(ctx context.Context, w http.ResponseWriter, r *http.Request, u *weburl.GnoURL) (int, *components.View) {
	isJSON := u.WebQuery.Has("json")

	if h.deps.Limiter != nil && !h.deps.Limiter.AllowRequest(r) {
		return writeRateLimited(w, isJSON), nil
	}

	q, err := ParseQuery(queryTerm(u))
	if err != nil {
		if isJSON {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return http.StatusBadRequest, nil
		}
		return http.StatusBadRequest, components.StatusErrorComponent(err.Error())
	}
	h.scope(q, u)
	q.jsonPath = isJSON

	if isJSON {
		ctx, cancel := context.WithTimeout(ctx, jsonTimeout)
		defer cancel()
		return h.serveJSON(ctx, w, r, q), nil
	}

	ctx, cancel := context.WithTimeout(ctx, pageTimeout)
	defer cancel()
	return http.StatusOK, NewPageView(h.build(ctx, q))
}

// queryTerm reads `q` from either grammar: gnoweb's links build `$search&q=`,
// a plain GET form can only produce `?q=`.
func queryTerm(u *weburl.GnoURL) string {
	if v := u.WebQuery.Get("q"); v != "" {
		return v
	}
	return u.Query.Get("q")
}

// scope resolves which package the query runs against: `in:` when given,
// otherwise the URL's path. Only realm and pure paths carry one.
func (h *Handler) scope(q *Query, u *weburl.GnoURL) {
	base := u.Path
	if base == "" {
		base = "/"
	}
	q.formAction = base + "$search"

	rel := ""
	if in, ok := q.Get(FilterIn); ok {
		rel = normalizePkgPath(in, h.deps.Domain)
	} else if u.IsRealm() || u.IsPure() {
		rel = u.Path
	}
	if rel == "" {
		return
	}
	q.PkgPath = rel
	q.ChainPath = path.Join(h.deps.Domain, rel)
}

// normalizePkgPath accepts "/r/demo/x", "r/demo/x" and "gno.land/r/demo/x",
// and rejects anything else.
func normalizePkgPath(in, domain string) string {
	rel := strings.TrimPrefix(in, domain)
	if !strings.HasPrefix(rel, "/") {
		rel = "/" + rel
	}
	rel = path.Clean(rel)
	if !isSafeRelPath(rel) {
		return ""
	}
	return rel
}

// build runs the query and assembles the render payload.
func (h *Handler) build(ctx context.Context, q *Query) SearchData {
	data := SearchData{
		Query:     q.Raw,
		PkgPath:   q.PkgPath,
		Selectors: h.selectors,
	}
	data.FormAction = q.formAction
	data.Groups, data.UnknownFilter = h.Search(ctx, q)
	data.Indexer = h.indexerStatus(ctx, data.Groups)
	return data
}

// Search answers a query. At most one selector runs: they name different
// subjects, so running all of them would turn one keystroke into a dozen
// queries. Without a selector the query falls to the discovery search.
func (h *Handler) Search(ctx context.Context, q *Query) (groups []Group, unknown string) {
	if q.Raw == "" {
		return nil, ""
	}
	if bad := h.unknownFilter(q); bad != "" {
		// A typo is worth naming. Falling through to a discovery search would
		// silently answer a different question than the one that was asked.
		return nil, bad
	}

	sel, term := h.selectorFor(q)
	if sel == nil {
		return h.discover(ctx, q), ""
	}

	if sel.PageOnly && q.jsonPath {
		// A fan-out selector costs one navigation, never one per keystroke.
		return []Group{{
			Label:  sel.Label,
			Source: sel.Source,
			Err:    fmt.Errorf("%s runs on the results page — press Enter", sel.Name),
		}}, ""
	}

	if sel.Scope == ScopePackage && q.PkgPath == "" {
		return []Group{{
			Label:  sel.Label,
			Source: sel.Source,
			Err:    errors.New("this search needs a realm or package path — open one, or add in:/r/…"),
		}}, ""
	}
	if floor := sel.minTerm(); !sel.Bare && len(term) < floor {
		return []Group{{
			Label:  sel.Label,
			Source: sel.Source,
			Err:    fmt.Errorf("%s needs at least %d characters", sel.Name, floor),
		}}, ""
	}

	g := Group{Label: sel.Label, Source: sel.Source}
	results, err := sel.resolve(ctx, h, q, term)
	if err != nil {
		// Reported, never fatal: a dead indexer degrades to a visible
		// "could not answer", not a 500.
		h.deps.Logger.Warn("omnisearch: resolver failed",
			"selector", sel.Name, "term_length", len(term), "error", err)
		g.Err = err
	}
	g.Results = results
	return []Group{g}, ""
}

// selectorFor picks the one selector a query names.
func (h *Handler) selectorFor(q *Query) (*Selector, string) {
	for _, f := range q.Filters {
		sel, ok := h.byName[f.Key]
		if !ok {
			continue
		}
		// A bare selector takes no argument, but `activity:foo` plainly
		// wants activity — better than a path search that explains nothing.
		if sel.Bare {
			return sel, ""
		}
		return sel, f.Value
	}
	for _, sel := range h.selectors {
		if sel.Bare && q.HasBareWord(sel.Name) {
			return sel, ""
		}
	}
	return nil, ""
}

// unknownFilter returns the first qualifier that is neither a selector nor a
// narrowing keyword. An unregistered indexer selector lands here on purpose:
// `tx:abc` reads as "not a qualifier here", not "no such transaction".
func (h *Handler) unknownFilter(q *Query) string {
	for _, f := range q.Filters {
		if _, ok := h.byName[f.Key]; ok {
			continue
		}
		if isNarrowing(f.Key) {
			continue
		}
		return f.Key
	}
	return ""
}

// indexerStatus reads the tip only when an indexer group ran: a chain-only
// answer must not query the indexer for a footer nobody will see.
func (h *Handler) indexerStatus(ctx context.Context, groups []Group) *IndexerStatus {
	if h.deps.Indexer == nil {
		return nil
	}
	used := false
	for _, g := range groups {
		if g.Source == SourceIndexer {
			used = true
			break
		}
	}
	if !used {
		return nil
	}

	st := &IndexerStatus{URL: h.deps.Indexer.URL()}
	height, err := h.deps.Indexer.LatestBlockHeight(ctx)
	if err != nil {
		st.Err = err
		return st
	}
	st.LastBlock = height
	return st
}

// writeRateLimited answers in the shape the caller can read.
func writeRateLimited(w http.ResponseWriter, isJSON bool) int {
	w.Header().Set("Retry-After", "60")
	if isJSON {
		writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return http.StatusTooManyRequests
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusTooManyRequests)
	_, _ = w.Write([]byte("rate limit exceeded\n"))
	return http.StatusTooManyRequests
}

// isSafeRelPath validates against the router's own grammar rather than a
// hand-written charset, which was looser and minted links weburl then
// refused. Deployment is permissionless, so a chain path is attacker data;
// requiring /r/ or /p/ also rules out a scheme.
func isSafeRelPath(rel string) bool {
	u := weburl.GnoURL{Path: rel}
	return u.IsRealm() || u.IsPure()
}

// safePathHref links a chain-listed package path, or nothing.
func safePathHref(rel string) template.URL {
	if !isSafeRelPath(rel) {
		return ""
	}
	return template.URL(rel) //nolint:gosec // G203: isSafeRelPath admits only /r/ and /p/ paths of validated segments.
}

// safeUserHref links a namespace, or nothing.
func safeUserHref(ns string) template.URL {
	// weburl's grammar allows "/" inside a path, so IsUser() alone accepts
	// "/u/a/b" and even "/u/". A namespace is exactly one segment.
	if ns == "" || strings.Contains(ns, "/") {
		return ""
	}
	u := weburl.GnoURL{Path: "/u/" + ns}
	if !u.IsUser() {
		return ""
	}
	return template.URL(u.Path) //nolint:gosec // G203: single segment, validated against weburl's own grammar.
}
