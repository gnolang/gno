package state

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"net/url"
	"strconv"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// recoverFetcher must be deferred in every fetcher goroutine so a panic
// from one fetch never crashes the whole render. The shared log key set
// also gives prod a single grep target. The recovered value is clipped
// so a hostile chain returning an enormous panic payload cannot turn
// the log line itself into an amplification vector.
func recoverFetcher(logger *slog.Logger, kind string, fields ...any) {
	if r := recover(); r != nil {
		logger.Error("fetcher panic recovered",
			append([]any{"kind", kind, "panic", fmt.Sprintf("%.512s", r)}, fields...)...)
	}
}

// isFuncKind reports whether n renders as a func or closure — the two
// share the lazy-expand path and the funcs-first preview priority.
func isFuncKind(n *StateNode) bool {
	return n.Kind == KindFunc || n.Kind == KindClosure
}

// collectPreviewCandidates gathers stored refs (ObjectID + Expandable,
// no inline children) for the bounded preview pass. Funcs/closures go
// first: their fetch is terminal (Kind-detection only, no cascade), so
// it earns budget priority over data-ref previews.
func collectPreviewCandidates(nodes []StateNode, out *[]*StateNode) {
	isCandidate := func(n *StateNode) bool {
		return n.ObjectID != "" && n.Expandable && len(n.Children) == 0
	}
	for i := range nodes {
		if n := &nodes[i]; isCandidate(n) && isFuncKind(n) {
			*out = append(*out, n)
		}
	}
	for i := range nodes {
		if n := &nodes[i]; isCandidate(n) && !isFuncKind(n) {
			*out = append(*out, n)
		}
	}
	for i := range nodes {
		if n := &nodes[i]; len(n.Children) > 0 {
			collectPreviewCandidates(n.Children, out)
		}
	}
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
