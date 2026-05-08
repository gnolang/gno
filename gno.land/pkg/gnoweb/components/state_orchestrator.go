package components

import (
	"bytes"
	"html/template"
	"net/url"
	"strings"
	"sync"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// maxConcurrentFileFetches caps how many source files Enrich asks the
// FileFetcher (chain RPC in production) for in parallel. Realistic realms
// declare functions across 1-5 files, so 8 is plenty headroom; the bound
// matters only as a back-pressure safeguard against pathological inputs.
const maxConcurrentFileFetches = 8

// maxInlinePreviewFetches caps how many stored objects EnrichInlinePreviews
// will fetch in TOTAL across rounds on a single page render. The Pretty
// view's promise is "see actual content, not a navigation menu" — so the
// budget is sized to cover a realm's whole top-level surface (typically
// 5–15 decls × ~2 rounds for `*T` chains) plus a margin for nested refs,
// while staying bounded enough to avoid fan-out storms on huge realms.
const maxInlinePreviewFetches = 30

// maxInlinePreviewRounds caps the number of fetch rounds. Gno's heap→ref→
// struct indirection (typical for `*T` declarations) requires 2 rounds to
// reach the actual struct from a top-level pointer. More rounds get
// diminishing returns vs cost.
const maxInlinePreviewRounds = 2

// maxConcurrentObjectFetches bounds the in-flight StateObject calls during
// inline-preview enrichment. Same back-pressure rationale as
// maxConcurrentFileFetches.
const maxConcurrentObjectFetches = 8

// FileFetcher loads the full bytes of a source file living under a package.
// The state explorer uses it to fetch source files lazily for func/closure
// nodes — only files actually referenced by the rendered tree are read.
type FileFetcher interface {
	Fetch(pkgPath, fileName string) ([]byte, error)
}

// SnippetHighlighter renders a chunk of source code as syntax-highlighted
// HTML (typed as template.HTML so the template trusts it as already-safe
// markup). Implementations typically wrap chroma.
type SnippetHighlighter interface {
	Render(fileName string, source []byte) (template.HTML, error)
}

// StateObjectFetcher fetches a stored object's raw Amino JSON by ObjectID.
// Used by EnrichInlinePreviews to fetch top-level refs so users see one
// level of children inline without an extra click.
type StateObjectFetcher interface {
	FetchObject(oid string) ([]byte, error)
}

// StateTypeFetcher fetches a stored named type's raw Amino JSON by TypeID.
// Used by EnrichInlinePreviews together with StateObjectFetcher so the
// inline preview can label struct fields with their declared names — Amino
// strips named-type definitions during ExportValues, so we recover them
// via this companion fetch.
type StateTypeFetcher interface {
	FetchType(tid string) ([]byte, error)
}

// Enrich walks the StateNode tree and decorates each node with anything the
// walker can't compute on its own:
//
//   - Href       — pre-built navigation URL for stored object refs, encoded
//                  via weburl.GnoURL so the URL parser round-trips correctly.
//   - SourceHTML — chroma-highlighted source snippet for func/closure nodes.
//
// Source files are fetched in parallel (bounded at maxConcurrentFileFetches)
// so a realm with functions across N files takes ~one RTT to the chain RPC,
// not N. Each unique file is fetched at most once per Enrich call.
//
// Failure modes (file not found, render error) leave SourceHTML empty
// rather than aborting — the rest of the page must still render.
//
// Pass nil for fetcher/highlighter to skip source enrichment (links still
// build); the orchestrator never assumes both deps are present.
func Enrich(nodes []StateNode, pkgPath string, fetcher FileFetcher, highlighter SnippetHighlighter) {
	// Pass 1 — local CPU work: build hrefs and collect the unique set of
	// files referenced anywhere in the tree.
	files := make(map[string]struct{})
	walkLinksAndCollect(nodes, pkgPath, files)

	// Pass 2 — concurrent I/O: prefetch all referenced files. Skipped if
	// no fetcher provided (link-only enrichment, used by template tests).
	var cache map[string][]byte
	if fetcher != nil && len(files) > 0 {
		cache = fetchFilesConcurrent(pkgPath, files, fetcher)
	}

	// Pass 3 — local CPU work again: chroma-highlight every snippet from
	// the prefetched cache.
	if highlighter != nil && len(cache) > 0 {
		walkRenderSnippets(nodes, highlighter, cache)
	}
}

// walkLinksAndCollect populates Hrefs in place and collects every
// Source.File whose body the renderer will need for inline snippets —
// any node carrying a Source (funcs, closures). Pure CPU pass; the
// fetcher dedupes by filename, so N funcs in one file = 1 fetch.
func walkLinksAndCollect(nodes []StateNode, pkgPath string, files map[string]struct{}) {
	for i := range nodes {
		n := &nodes[i]
		if n.ObjectID != "" {
			n.Href = stateObjectHref(pkgPath, n.ObjectID, n.TypeID)
		}
		if n.Source != nil && n.Source.File != "" {
			files[n.Source.File] = struct{}{}
		}
		if len(n.Children) > 0 {
			walkLinksAndCollect(n.Children, pkgPath, files)
		}
	}
}

// fetchFilesConcurrent fetches every file in `files` in parallel through the
// fetcher, bounded at maxConcurrentFileFetches. Errors are silently dropped
// (the file is simply absent from the returned cache, leaving the
// corresponding SourceHTML empty in pass 3).
func fetchFilesConcurrent(pkgPath string, files map[string]struct{}, fetcher FileFetcher) map[string][]byte {
	cache := make(map[string][]byte, len(files))
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrentFileFetches)
	for file := range files {
		wg.Add(1)
		sem <- struct{}{}
		go func(name string) {
			defer wg.Done()
			defer func() { <-sem }()
			content, err := fetcher.Fetch(pkgPath, name)
			if err != nil {
				return
			}
			mu.Lock()
			cache[name] = content
			mu.Unlock()
		}(file)
	}
	wg.Wait()
	return cache
}

// walkRenderSnippets renders the cached file slices into SourceHTML for
// every node that carries a Source range (funcs and closures). The card
// then shows the body inline — same behavior as the initial Jae PR.
// CPU-only; safe to run sequentially after the concurrent fetch phase.
func walkRenderSnippets(nodes []StateNode, highlighter SnippetHighlighter, cache map[string][]byte) {
	for i := range nodes {
		n := &nodes[i]
		if n.Source != nil && n.Source.File != "" {
			if content, ok := cache[n.Source.File]; ok {
				snippet := sliceLines(content, n.Source.StartLine, n.Source.EndLine)
				if html, err := highlighter.Render(n.Source.File, snippet); err == nil {
					n.SourceHTML = html
				}
			}
		}
		if len(n.Children) > 0 {
			walkRenderSnippets(n.Children, highlighter, cache)
		}
	}
}

// EnrichInlinePreviews fetches up to maxInlinePreviewFetches stored objects
// referenced by ref-shaped nodes and embeds their decoded fields as Children
// so users get a one-level preview without a navigation click. Refs beyond
// the budget keep their plain ref shape (graceful degradation).
//
// When typeFetcher is non-nil, the named-type definitions are also fetched
// in parallel and used to resolve struct field names — so users see
// `name : string = "alice"` instead of `0 : string = "alice"`. Type fetches
// are deduplicated by TypeID so a realm with N refs to the same type pays
// for one type lookup.
//
// Top-level refs are prioritised over deeply-nested ones: a realm's main
// declarations are visible first; only if budget remains do nested refs get
// previewed. Same OID requested by multiple nodes triggers a single fetch.
//
// Failure modes (fetch error, decode error) leave the node's Children empty
// — the user sees the original navigation link, never an error.
func EnrichInlinePreviews(nodes []StateNode, objFetcher StateObjectFetcher, typeFetcher StateTypeFetcher) {
	if objFetcher == nil {
		return
	}

	// Multiple rounds: gno's typical `*T` storage chains heap-item → inner
	// ref → struct, so the first fetch reveals an inner ref that itself
	// needs fetching to expose fields. Each round picks up newly-revealed
	// refs without children and fetches them. maxInlinePreviewFetches caps
	// the cumulative cost across all rounds.
	fetchedTotal := 0
	for round := 0; round < maxInlinePreviewRounds; round++ {
		var candidates []*StateNode
		collectPreviewCandidates(nodes, &candidates)
		remaining := maxInlinePreviewFetches - fetchedTotal
		if remaining <= 0 || len(candidates) == 0 {
			return
		}
		if len(candidates) > remaining {
			candidates = candidates[:remaining]
		}
		fetchPreviewsConcurrent(candidates, objFetcher, typeFetcher)
		fetchedTotal += len(candidates)
	}
}

// collectPreviewCandidates walks the tree breadth-first, gathering nodes
// that are stored refs (ObjectID set, Expandable) without inline children.
func collectPreviewCandidates(nodes []StateNode, out *[]*StateNode) {
	// Process this level first (top-level priority).
	for i := range nodes {
		n := &nodes[i]
		if n.ObjectID != "" && n.Expandable && len(n.Children) == 0 {
			*out = append(*out, n)
		}
	}
	// Then recurse into children.
	for i := range nodes {
		n := &nodes[i]
		if len(n.Children) > 0 {
			collectPreviewCandidates(n.Children, out)
		}
	}
}

// fetchPreviewsConcurrent fetches the (object, type) pair for each unique
// candidate in parallel, then decodes children with the resolved type so
// struct fields surface with their declared names. Each fetch is silent on
// error — graceful degradation. All I/O happens in two parallel pools
// (objects + types) so total wall-clock stays at ~one RTT.
func fetchPreviewsConcurrent(candidates []*StateNode, objFetcher StateObjectFetcher, typeFetcher StateTypeFetcher) {
	// Dedupe by OID so the same object isn't fetched twice.
	byOID := make(map[string][]*StateNode)
	for _, n := range candidates {
		byOID[n.ObjectID] = append(byOID[n.ObjectID], n)
	}

	// Collect unique TypeIDs so each named type is fetched once even if
	// referenced by many objects (e.g. a map of N Users → 1 User type fetch).
	uniqueTypeIDs := make(map[string]struct{})
	if typeFetcher != nil {
		for _, n := range candidates {
			if n.TypeID != "" {
				uniqueTypeIDs[n.TypeID] = struct{}{}
			}
		}
	}

	objCache := make(map[string][]byte, len(byOID))
	typeCache := make(map[string][]byte, len(uniqueTypeIDs))
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrentObjectFetches)

	// Pool 1: stored objects.
	for oid := range byOID {
		wg.Add(1)
		sem <- struct{}{}
		go func(oid string) {
			defer wg.Done()
			defer func() { <-sem }()
			raw, err := objFetcher.FetchObject(oid)
			if err != nil {
				return
			}
			mu.Lock()
			objCache[oid] = raw
			mu.Unlock()
		}(oid)
	}

	// Pool 2: named types — runs alongside Pool 1, no extra wall-clock.
	for tid := range uniqueTypeIDs {
		wg.Add(1)
		sem <- struct{}{}
		go func(tid string) {
			defer wg.Done()
			defer func() { <-sem }()
			raw, err := typeFetcher.FetchType(tid)
			if err != nil {
				return
			}
			mu.Lock()
			typeCache[tid] = raw
			mu.Unlock()
		}(tid)
	}
	wg.Wait()

	// Apply: for each candidate, decode object children with its type
	// (when both are present), or fall back to positional indices.
	// After populating Children, re-compute Preview so the collapsed
	// row picks up the freshly-fetched fields — without it, lazily-
	// loaded refs would stay generic until the user expands.
	for oid, refs := range byOID {
		raw, ok := objCache[oid]
		if !ok {
			continue
		}
		for _, n := range refs {
			var typeRaw []byte
			if n.TypeID != "" {
				typeRaw = typeCache[n.TypeID]
			}
			decoded, err := DecodeObjectFull(raw, typeRaw)
			if err != nil {
				continue
			}
			n.Children = decoded.Nodes
			// Propagate the fetched object's ObjectInfo to the outer ref
			// node so the card header surfaces Hash/Owner/RefCount even
			// when the outer was just a pointer with no ObjectInfo of
			// its own. Only fill empties — don't clobber values already
			// set by a previous round.
			if n.Hash == "" {
				n.Hash = decoded.Info.Hash
			}
			if n.OwnerID == "" {
				n.OwnerID = decoded.Info.OwnerID
			}
			if n.ModTime == "" {
				n.ModTime = decoded.Info.ModTime
			}
			if n.RefCount == "" {
				n.RefCount = decoded.Info.RefCount
			}
			if n.LastObjectSize == "" {
				n.LastObjectSize = decoded.Info.LastObjectSize
			}
			if !n.IsEscaped {
				n.IsEscaped = decoded.Info.IsEscaped
			}
			n.Preview = buildStructPreview(decoded.Nodes)
		}
	}
}

// stateObjectHref builds the URL for `<pkgPath>$state&oid=<encoded-oid>` —
// optionally annotated with `&tid=<TypeID>` so the destination page can
// resolve struct field names without an extra round-trip away from the
// shared parser. Single source of truth for state-explorer hrefs.
func stateObjectHref(pkgPath, oid, typeID string) template.URL {
	wq := url.Values{"state": {""}, "oid": {oid}}
	if typeID != "" {
		wq.Set("tid", typeID)
	}
	u := weburl.GnoURL{Path: pkgPath, WebQuery: wq}
	return template.URL(u.EncodeWebURL())
}

// sliceLines extracts a 1-based inclusive line range from content.
//   - startLine <= 0 returns the entire content (treat as "no slicing").
//   - endLine > number of lines is clamped to the last line.
//   - endLine < startLine is treated as startLine..end (defensive).
//   - startLine past end-of-file returns nothing (no panic).
func sliceLines(content []byte, startLine, endLine int) []byte {
	if startLine <= 0 {
		return content
	}
	lines := strings.Split(string(content), "\n")
	if startLine > len(lines) {
		return nil
	}
	end := endLine
	if end > len(lines) || end < startLine {
		end = len(lines)
	}
	var buf bytes.Buffer
	for i := startLine - 1; i < end; i++ {
		if i > startLine-1 {
			buf.WriteByte('\n')
		}
		buf.WriteString(lines[i])
	}
	return buf.Bytes()
}

// AttachDocs projects doc-index entries onto top-level StateNodes by
// Name. Each match populates the node's Doc (markdown comment).
// Names that don't match are silently dropped — the doc index may
// include items handled elsewhere.
//
// Caller is the handler: it fetches Client.Doc(pkgPath) in parallel
// with qpkg_json so the page assembly stays at one RTT.
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
// import the gnovm/doc package transitively into the components tree.
type NamedDoc struct {
	Name string
	Doc  string
}
