package components

import (
	"bytes"
	"context"
	"html/template"
	"net/url"
	"strconv"
	"sync"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// Bounds for per-render fan-out. The concurrency caps act as back-pressure
// on the chain RPC; the total caps protect against realms that try to
// amplify a single GET into a flood of fetches.
const (
	maxConcurrentFileFetches   = 8
	maxFilesPerRender          = 50
	maxConcurrentObjectFetches = 8
	maxInlinePreviewFetches    = 30
	maxInlinePreviewRounds     = 2 // covers Gno's heap→ref→struct indirection
)

type FileFetcher interface {
	Fetch(ctx context.Context, pkgPath, fileName string) ([]byte, error)
}

// SnippetHighlighter returns template.HTML so the result is treated as
// already-safe markup by html/template.
type SnippetHighlighter interface {
	Render(fileName string, source []byte) (template.HTML, error)
}

type StateObjectFetcher interface {
	FetchObject(ctx context.Context, oid string) ([]byte, error)
}

// StateTypeFetcher recovers named-type definitions that ExportValues
// strips, so inline previews can label struct fields by declared name.
type StateTypeFetcher interface {
	FetchType(ctx context.Context, tid string) ([]byte, error)
}

// Enrich decorates a StateNode tree with Href and SourceHTML, walking
// the tree first to collect referenced files, then fetching them in
// parallel, then highlighting. Failures degrade gracefully — SourceHTML
// is left empty. Passing nil fetcher or highlighter skips source.
// A canceled ctx aborts in-flight fetches and skips highlighting.
func Enrich(ctx context.Context, nodes []StateNode, pkgPath string, height int64, fetcher FileFetcher, highlighter SnippetHighlighter) {
	files := make(map[string]struct{})
	walkLinksAndCollect(nodes, pkgPath, height, files)

	var cache map[string][]byte
	if fetcher != nil && len(files) > 0 {
		cache = fetchFilesConcurrent(ctx, pkgPath, files, fetcher)
	}

	if highlighter != nil && len(cache) > 0 && ctx.Err() == nil {
		walkRenderSnippets(nodes, highlighter, cache)
	}
}

// walkLinksAndCollect populates Hrefs in place and collects the unique
// set of source files referenced by the tree.
func walkLinksAndCollect(nodes []StateNode, pkgPath string, height int64, files map[string]struct{}) {
	for i := range nodes {
		n := &nodes[i]
		if n.ObjectID != "" {
			n.Href = stateObjectHref(pkgPath, n.ObjectID, n.TypeID, height)
		}
		if n.OwnerID != "" {
			n.OwnerHref = stateObjectHref(pkgPath, n.OwnerID, "", height)
		}
		if n.Source != nil && n.Source.File != "" && len(files) < maxFilesPerRender {
			files[n.Source.File] = struct{}{}
		}
		if len(n.Children) > 0 {
			walkLinksAndCollect(n.Children, pkgPath, height, files)
		}
	}
}

// fetchFilesConcurrent fetches each file via the bounded semaphore;
// errors and ctx cancellation leave the file absent from the cache.
func fetchFilesConcurrent(ctx context.Context, pkgPath string, files map[string]struct{}, fetcher FileFetcher) map[string][]byte {
	cache := make(map[string][]byte, len(files))
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrentFileFetches)
	for file := range files {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			defer func() { _ = recover() }() // fetcher panics must not crash the process
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()
			content, err := fetcher.Fetch(ctx, pkgPath, name)
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

// walkRenderSnippets fills SourceHTML on every node with a Source range
// using the prefetched file cache.
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
// referenced by ref-shaped nodes and embeds their decoded fields as Children.
// `typeFetcher` (if non-nil) resolves struct field names via the named-type
// definitions Amino strips during ExportValues. Failures degrade silently.
// A canceled ctx aborts before the next round and during semaphore acquire.
func EnrichInlinePreviews(ctx context.Context, nodes []StateNode, objFetcher StateObjectFetcher, typeFetcher StateTypeFetcher) {
	if objFetcher == nil {
		return
	}

	// Multiple rounds: a `*T` storage chains heap-item → inner ref → struct,
	// so the first fetch reveals refs that themselves need fetching.
	fetchedTotal := 0
	for round := 0; round < maxInlinePreviewRounds; round++ {
		if ctx.Err() != nil {
			return
		}
		var candidates []*StateNode
		collectPreviewCandidates(nodes, &candidates)
		remaining := maxInlinePreviewFetches - fetchedTotal
		if remaining <= 0 || len(candidates) == 0 {
			return
		}
		if len(candidates) > remaining {
			candidates = candidates[:remaining]
		}
		fetchPreviewsConcurrent(ctx, candidates, objFetcher, typeFetcher)
		fetchedTotal += len(candidates)
	}
}

// collectPreviewCandidates gathers stored refs (ObjectID + Expandable,
// no inline children) breadth-first so top-level refs get priority.
func collectPreviewCandidates(nodes []StateNode, out *[]*StateNode) {
	for i := range nodes {
		n := &nodes[i]
		if n.ObjectID != "" && n.Expandable && len(n.Children) == 0 {
			*out = append(*out, n)
		}
	}
	for i := range nodes {
		n := &nodes[i]
		if len(n.Children) > 0 {
			collectPreviewCandidates(n.Children, out)
		}
	}
}

// fetchPreviewsConcurrent fetches objects and types under one shared
// concurrency cap so the total in-flight RPC per render matches the
// documented maxConcurrentObjectFetches. ctx cancellation aborts before
// semaphore acquire.
func fetchPreviewsConcurrent(ctx context.Context, candidates []*StateNode, objFetcher StateObjectFetcher, typeFetcher StateTypeFetcher) {
	byOID := make(map[string][]*StateNode)
	for _, n := range candidates {
		byOID[n.ObjectID] = append(byOID[n.ObjectID], n)
	}

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

	for oid := range byOID {
		wg.Add(1)
		go func(oid string) {
			defer wg.Done()
			defer func() { _ = recover() }() // fetcher panics must not crash the process
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()
			raw, err := objFetcher.FetchObject(ctx, oid)
			if err != nil {
				return
			}
			mu.Lock()
			objCache[oid] = raw
			mu.Unlock()
		}(oid)
	}

	for tid := range uniqueTypeIDs {
		wg.Add(1)
		go func(tid string) {
			defer wg.Done()
			defer func() { _ = recover() }() // fetcher panics must not crash the process
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()
			raw, err := typeFetcher.FetchType(ctx, tid)
			if err != nil {
				return
			}
			mu.Lock()
			typeCache[tid] = raw
			mu.Unlock()
		}(tid)
	}
	wg.Wait()

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
			// Fill the outer ref's ObjectInfo from the fetched payload
			// only when empty — preserves values from earlier rounds.
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
			n.Preview = buildChildrenPreview(decoded.Nodes)
		}
	}
}

// stateObjectHref builds a `<pkgPath>$state&oid=...` URL, propagating
// optional tid + height so time-travel and type resolution survive hops.
func stateObjectHref(pkgPath, oid, typeID string, height int64) template.URL {
	wq := url.Values{"state": {""}, "oid": {oid}}
	if typeID != "" {
		wq.Set("tid", typeID)
	}
	if height > 0 {
		wq.Set("height", strconv.FormatInt(height, 10))
	}
	u := weburl.GnoURL{Path: pkgPath, WebQuery: wq}
	return template.URL(u.EncodeWebURL()) //nolint:gosec
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
// Must run before Enrich — only top-level nodes carry Names matchable to
// the doc index.
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
