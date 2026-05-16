package state

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"net/url"
	"strconv"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// recoverFetcher is deferred in fetch goroutines whose panic must be
// swallowed (errgroup doesn't recover; without this an amino panic on
// hostile chain data unwinds past errgroup → process crash). Panic
// payload is clipped to 512c so the log line itself can't amplify.
func recoverFetcher(logger *slog.Logger, kind string, fields ...any) {
	if r := recover(); r != nil {
		logger.Error("fetcher panic recovered",
			append([]any{"kind", kind, "panic", fmt.Sprintf("%.512s", r)}, fields...)...)
	}
}

// recoverToErr is recoverFetcher's fatal counterpart: writes a sentinel
// into errp (caller's named return) so errgroup.Wait surfaces a 500.
func recoverToErr(logger *slog.Logger, kind string, errp *error, fields ...any) {
	if r := recover(); r != nil {
		logger.Error("fetcher panic recovered",
			append([]any{"kind", kind, "panic", fmt.Sprintf("%.512s", r)}, fields...)...)
		*errp = fmt.Errorf("%s: panic recovered", kind)
	}
}

// isFuncKind reports whether n renders as a func or closure — the two
// share the lazy-expand path in the tree renderer (state/source-details).
func isFuncKind(n *StateNode) bool {
	return n.Kind == KindFunc || n.Kind == KindClosure
}

// oidTagSlotID derives a DOM-safe id for the per-card tag slot. The OID
// format is `<40-hex>:<index>`; ":" is not allowed in CSS selectors and
// brittle in querySelector, so swap it to "-". Used by state/decl + the
// frag=node OOB swap that injects the closure tag at hydration.
func oidTagSlotID(oid string) string {
	return "card-" + strings.ReplaceAll(oid, ":", "-") + "-tag"
}

// All four *Href builders MUST use gnoweb's `$webargs` grammar (the
// `$state&…` form). `?state&…` lands in u.Query, misses the WebQuery
// dispatch, and htmx swaps the full page into a fragment slot.

// stateObjectHref builds a `<pkgPath>$state&oid=...` permalink. tid keeps
// the destination page type-aware; viewMode keeps a tree-view hop in tree.
func stateObjectHref(pkgPath, oid, typeID, heightParam, viewMode string) template.URL {
	wq := url.Values{"state": {""}, "oid": {oid}}
	if typeID != "" {
		wq.Set("tid", typeID)
	}
	if heightParam != "" {
		wq.Set("height", heightParam)
	}
	if viewMode == "tree" {
		wq.Set("view", "tree")
	}
	u := weburl.GnoURL{Path: pkgPath, WebQuery: wq}
	return template.URL(u.EncodeWebURL()) //nolint:gosec
}

// stateFragNodeHref builds the hx-get URL for a lazy node fragment.
// depth is the row's tree depth — the server renders the fragment's
// children at depth+1 so they step in under the parent via the same
// --depth mechanism as the server-rendered tree (purely presentational).
// viewMode keeps a tree-view expansion in tree format; pretty (default)
// stays unstamped so the nginx cache key stays minimal.
func stateFragNodeHref(pkgPath, oid, typeID, heightParam string, depth int, viewMode string) template.URL {
	wq := url.Values{"state": {""}, "frag": {"node"}, "oid": {oid}}
	if typeID != "" {
		wq.Set("tid", typeID)
	}
	if heightParam != "" {
		wq.Set("height", heightParam)
	}
	if depth > 0 {
		wq.Set("depth", strconv.Itoa(depth))
	}
	if viewMode == "tree" {
		wq.Set("view", "tree")
	}
	u := weburl.GnoURL{Path: pkgPath, WebQuery: wq}
	return template.URL(u.EncodeWebURL()) //nolint:gosec
}

// stateFragSourceHref builds the hx-get URL for a lazy source fragment.
// endLine (the func's last line) lets the server slice the exact span;
// 0 falls back to a ±N context window.
func stateFragSourceHref(pkgPath, file string, line, endLine int, heightParam string) template.URL {
	wq := url.Values{"state": {""}, "frag": {"source"}, "file": {file}, "line": {strconv.Itoa(line)}}
	if endLine > 0 && endLine >= line {
		wq.Set("end", strconv.Itoa(endLine))
	}
	if heightParam != "" {
		wq.Set("height", heightParam)
	}
	u := weburl.GnoURL{Path: pkgPath, WebQuery: wq}
	return template.URL(u.EncodeWebURL()) //nolint:gosec
}

// stateRawJSONHref builds the `<pkgPath>$state&json[&height=N]` URL for
// the lazy "Copy package JSON" button — the raw qpkg_json payload the
// page no longer inlines (memory amp + cache poisoning risk).
func stateRawJSONHref(pkgPath, heightParam string) template.URL {
	wq := url.Values{"state": {""}, "json": {""}}
	if heightParam != "" {
		wq.Set("height", heightParam)
	}
	u := weburl.GnoURL{Path: pkgPath, WebQuery: wq}
	return template.URL(u.EncodeWebURL()) //nolint:gosec
}

// statePageHref builds the `<pkgPath>$state[&offset=N&limit=M&view=tree&
// height=H]` permalink for the pagination footer. offset=0 and the
// default limit are omitted to keep page-1 cache-key parity with the
// canonical unparameterized state URL.
func statePageHref(pkgPath, heightParam, viewMode string, offset, limit int) template.URL {
	wq := url.Values{"state": {""}}
	if offset > 0 {
		wq.Set("offset", strconv.Itoa(offset))
	}
	if limit > 0 && limit != maxTopLevelDecls {
		wq.Set("limit", strconv.Itoa(limit))
	}
	if heightParam != "" {
		wq.Set("height", heightParam)
	}
	if viewMode == ViewModeTree {
		wq.Set("view", ViewModeTree)
	}
	u := weburl.GnoURL{Path: pkgPath, WebQuery: wq}
	return template.URL(u.EncodeWebURL()) //nolint:gosec
}

// buildPagination computes the prev/next view-model from a paginated
// DecodePackage result. Returns nil when total ≤ limit at offset 0.
// Hrefs are gated on HasPrev/HasNext to skip work on edge pages.
func buildPagination(pkgPath, heightParam, viewMode string, total, offset, limit int) *Pagination {
	if total <= limit && offset <= 0 {
		return nil
	}
	start, end := clampSliceWindow(offset, limit, total)
	p := &Pagination{
		Total:       total,
		StartNumber: start + 1,
		EndNumber:   end,
		HasPrev:     start > 0,
		HasNext:     end < total,
	}
	if p.HasPrev {
		prev := max(start-limit, 0)
		p.FirstHref = statePageHref(pkgPath, heightParam, viewMode, 0, limit)
		p.PrevHref = statePageHref(pkgPath, heightParam, viewMode, prev, limit)
	}
	if p.HasNext {
		p.NextHref = statePageHref(pkgPath, heightParam, viewMode, end, limit)
		p.LastHref = statePageHref(pkgPath, heightParam, viewMode, lastPageOffset(total, limit), limit)
	}
	if end == start {
		p.StartNumber = 0 // empty page → honest "Showing 0–0"
	}
	return p
}

// lastPageOffset returns the offset of the last page. total=12, limit=5 → 10.
func lastPageOffset(total, limit int) int {
	if total <= 0 || limit <= 0 {
		return 0
	}
	return ((total - 1) / limit) * limit
}

// cacheControlForHeight returns the canonical Cache-Control shared by
// every state-feature response surface. Pinned heights are immutable
// (24h); "latest" gets a 1s window matching block time.
func cacheControlForHeight(height int64) string {
	if height > 0 {
		return "public, max-age=86400, immutable"
	}
	return "public, max-age=1"
}

// stateSourceHref builds the permanent `<pkgPath>$source&file=F` link to
// the canonical full-source view — the "See in code" target out of a
// frag=source fragment. Uses the `$webargs` grammar so it routes; the
// `#L` anchor is appended after EncodeWebURL like sourceHref does.
func stateSourceHref(pkgPath, file string, line int, heightParam string) template.URL {
	wq := url.Values{"source": {""}, "file": {file}}
	if heightParam != "" {
		wq.Set("height", heightParam)
	}
	u := weburl.GnoURL{Path: pkgPath, WebQuery: wq}
	href := u.EncodeWebURL()
	if line > 0 {
		href += "#L" + strconv.Itoa(line)
	}
	return template.URL(href) //nolint:gosec
}

// EnrichLinks walks the StateNode tree and populates Href + OwnerHref
// from ObjectID/OwnerID. Without this the template's `{{ if $n.Href }}`
// guards drop every Inspect / Owner / navlink button on the page.
func EnrichLinks(nodes []StateNode, pkgPath, heightParam, viewMode string) {
	for i := range nodes {
		n := &nodes[i]
		if n.ObjectID != "" {
			n.Href = stateObjectHref(pkgPath, n.ObjectID, n.TypeID, heightParam, viewMode)
		}
		if n.OwnerID != "" {
			n.OwnerHref = stateObjectHref(pkgPath, n.OwnerID, "", heightParam, viewMode)
		}
		if len(n.Children) > 0 {
			EnrichLinks(n.Children, pkgPath, heightParam, viewMode)
		}
	}
}

// sliceLines extracts a 1-based inclusive line range from content via a
// single-pass byte scan — avoids allocating the full file as a string +
// []string just to keep ~5 lines. Returns a sub-slice of content (no copy).
//   - startLine <= 0 returns the entire content.
//   - endLine clamped to the last line.
//   - startLine past end-of-file returns nil.
func sliceLines(content []byte, startLine, endLine int) []byte {
	if startLine <= 0 {
		return content
	}
	// Scan to the first byte of startLine.
	line, start := 1, 0
	for line < startLine && start < len(content) {
		nl := bytes.IndexByte(content[start:], '\n')
		if nl < 0 {
			return nil
		}
		start += nl + 1
		line++
	}
	if line < startLine {
		return nil
	}
	// Defensive: endLine < startLine → fall through to EOF.
	if endLine < startLine {
		return content[start:]
	}
	// Walk forward to the byte just before endLine's terminating newline
	// (or EOF if endLine is the last line).
	end := start
	for line <= endLine && end < len(content) {
		nl := bytes.IndexByte(content[end:], '\n')
		if nl < 0 {
			return content[start:]
		}
		if line == endLine {
			return content[start : end+nl]
		}
		end += nl + 1
		line++
	}
	return content[start:end]
}

// AttachDocs projects doc-index entries onto top-level StateNodes by Name.
// Only top-level nodes carry Names matchable to the doc index.
func AttachDocs(nodes []StateNode, vals []NamedDoc, funs []NamedDoc, typs []NamedDoc) {
	if len(nodes) == 0 {
		return
	}
	docs := make(map[string]string, len(vals)+len(funs)+len(typs))
	for _, d := range vals {
		if d.Doc != "" {
			docs[d.Name] = d.Doc
		}
	}
	for _, d := range funs {
		if d.Doc != "" {
			docs[d.Name] = d.Doc
		}
	}
	for _, d := range typs {
		if d.Doc != "" {
			docs[d.Name] = d.Doc
		}
	}
	for i := range nodes {
		if doc, ok := docs[nodes[i].Name]; ok {
			nodes[i].Doc = doc
		}
	}
}

// NamedDoc is the (Name, Doc) pair the handler extracts from the
// JSON doc index — kept lightweight so the handler doesn't need to
// import the gnovm/doc package transitively into other layers.
type NamedDoc struct {
	Name string
	Doc  string
}
