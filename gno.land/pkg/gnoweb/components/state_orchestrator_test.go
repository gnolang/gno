package components

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// fakeFetcher records calls so tests can assert on caching behaviour.
type fakeFetcher struct {
	files map[string][]byte
	err   error
	calls int32
}

func (f *fakeFetcher) Fetch(_ context.Context, pkgPath, fileName string) ([]byte, error) {
	atomic.AddInt32(&f.calls, 1)
	if f.err != nil {
		return nil, f.err
	}
	if b, ok := f.files[pkgPath+"/"+fileName]; ok {
		return b, nil
	}
	return nil, fmt.Errorf("not found: %s/%s", pkgPath, fileName)
}

// fakeHighlighter renders a deterministic HTML envelope so tests can
// assert on what got passed in. Mirrors what Chroma would produce shape-wise.
type fakeHighlighter struct {
	err error
}

func (h *fakeHighlighter) Render(fileName string, source []byte) (template.HTML, error) {
	if h.err != nil {
		return "", h.err
	}
	return template.HTML(fmt.Sprintf(`<pre data-file="%s">%s</pre>`, fileName, string(source))), nil
}

// TestEnrich_Basic populates SourceHTML for a closure node with a
// Source location — every node carrying a Source range gets its body
// rendered inline (closure or regular func), matching the initial Jae
// PR's behavior.
func TestEnrich_Basic(t *testing.T) {
	t.Parallel()

	const fileBody = "package x\n\nfunc Foo() int {\n\treturn 42\n}\n"
	fetcher := &fakeFetcher{files: map[string][]byte{"/r/demo/foo/foo.gno": []byte(fileBody)}}
	hl := &fakeHighlighter{}

	nodes := []StateNode{{
		Name: "Foo", Type: "func() int", Kind: "closure",
		Source: &SourceLocation{File: "foo.gno", StartLine: 3, EndLine: 5},
	}}

	Enrich(context.Background(), nil, nodes, "/r/demo/foo", 0, fetcher, hl)

	assert.NotEmpty(t, nodes[0].SourceHTML, "SourceHTML must be populated")
	assert.Contains(t, string(nodes[0].SourceHTML), "func Foo()",
		"snippet must include the function declaration line")
	assert.Contains(t, string(nodes[0].SourceHTML), "return 42",
		"snippet must include the body lines through endLine")
	assert.NotContains(t, string(nodes[0].SourceHTML), "package x",
		"lines before startLine must be excluded")
}

// TestEnrich_FuncKindGetsSourceHTML pins the design choice: regular
// `func` nodes also get their body rendered inline, not just a source
// link — same as the initial Jae PR. Skipping the fetch would force
// users into the Source tab to read trivial functions.
func TestEnrich_FuncKindGetsSourceHTML(t *testing.T) {
	t.Parallel()

	const fileBody = "package x\n\nfunc Foo() int { return 42 }\n"
	fetcher := &fakeFetcher{files: map[string][]byte{"/r/demo/foo/foo.gno": []byte(fileBody)}}
	hl := &fakeHighlighter{}

	nodes := []StateNode{{
		Name: "Foo", Type: "func() int", Kind: "func",
		Source: &SourceLocation{File: "foo.gno", StartLine: 3, EndLine: 3},
	}}

	Enrich(context.Background(), nil, nodes, "/r/demo/foo", 0, fetcher, hl)

	assert.NotEmpty(t, nodes[0].SourceHTML,
		"regular funcs must get inline source rendered, like closures")
	assert.Contains(t, string(nodes[0].SourceHTML), "func Foo()",
		"snippet must include the func declaration line")
	assert.Equal(t, int32(1), atomic.LoadInt32(&fetcher.calls),
		"file must be fetched exactly once")
}

// TestEnrich_NoSource leaves nodes without a Source untouched —
// the orchestrator must not fabricate empty highlights or error.
func TestEnrich_NoSource(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{files: map[string][]byte{}}
	hl := &fakeHighlighter{}

	nodes := []StateNode{
		{Name: "Counter", Type: "int", Kind: "primitive", Value: "1"},
		{Name: "Users", Type: "map[string]User", Kind: "ref", ObjectID: "ff:8", Expandable: true},
	}

	Enrich(context.Background(), nil, nodes, "/r/demo/foo", 0, fetcher, hl)

	for i, n := range nodes {
		assert.Empty(t, n.SourceHTML, "node %d had no Source — SourceHTML must stay empty", i)
	}
	assert.Equal(t, int32(0), atomic.LoadInt32(&fetcher.calls),
		"fetcher must not be called when there are no Source nodes")
}

// TestEnrich_Recurses ensures that closures with captures (and
// any nested children) get their inner Source nodes enriched too.
func TestEnrich_Recurses(t *testing.T) {
	t.Parallel()

	const file = "func() int {\n\treturn n\n}\n"
	fetcher := &fakeFetcher{files: map[string][]byte{"/r/demo/foo/foo.gno": []byte(file)}}
	hl := &fakeHighlighter{}

	nodes := []StateNode{{
		Name: "outer", Type: "struct{...}", Kind: "struct", Expandable: true,
		Children: []StateNode{{
			Name: "stepper", Type: "func() int", Kind: "closure", Expandable: true,
			Source: &SourceLocation{File: "foo.gno", StartLine: 1, EndLine: 3},
		}},
	}}

	Enrich(context.Background(), nil, nodes, "/r/demo/foo", 0, fetcher, hl)

	assert.Empty(t, nodes[0].SourceHTML, "parent struct has no Source — left alone")
	assert.NotEmpty(t, nodes[0].Children[0].SourceHTML, "nested closure gets its source")
}

// TestEnrich_FileCache confirms repeated references to the same
// file resolve to a single fetch — important under load (many closures
// declared in the same file shouldn't multiply I/O).
func TestEnrich_FileCache(t *testing.T) {
	t.Parallel()

	body := []byte("line1\nline2\nline3\nline4\n")
	fetcher := &fakeFetcher{files: map[string][]byte{"/r/demo/foo/foo.gno": body}}
	hl := &fakeHighlighter{}

	nodes := []StateNode{
		{Name: "a", Kind: "closure", Source: &SourceLocation{File: "foo.gno", StartLine: 1, EndLine: 1}},
		{Name: "b", Kind: "closure", Source: &SourceLocation{File: "foo.gno", StartLine: 2, EndLine: 2}},
		{Name: "c", Kind: "closure", Source: &SourceLocation{File: "foo.gno", StartLine: 3, EndLine: 3}},
	}

	Enrich(context.Background(), nil, nodes, "/r/demo/foo", 0, fetcher, hl)

	assert.Equal(t, int32(1), atomic.LoadInt32(&fetcher.calls),
		"three nodes pointing to the same file should produce one fetch")
	for _, n := range nodes {
		assert.NotEmpty(t, n.SourceHTML)
	}
}

// TestEnrich_FetchError leaves SourceHTML empty rather than
// propagating the error or panicking — a missing source file is recoverable
// (the rest of the page still renders).
func TestEnrich_FetchError(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{err: errors.New("disk on fire")}
	hl := &fakeHighlighter{}

	nodes := []StateNode{{
		Name: "Foo", Kind: "closure",
		Source: &SourceLocation{File: "foo.gno", StartLine: 1, EndLine: 5},
	}}

	Enrich(context.Background(), nil, nodes, "/r/demo/foo", 0, fetcher, hl)

	assert.Empty(t, nodes[0].SourceHTML, "fetch error → SourceHTML stays empty (graceful)")
}

// TestEnrich_RenderError same fallback when the highlighter fails
// (e.g., chroma chokes on input).
func TestEnrich_RenderError(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{files: map[string][]byte{"/r/demo/foo/foo.gno": []byte("x")}}
	hl := &fakeHighlighter{err: errors.New("chroma kaput")}

	nodes := []StateNode{{
		Name: "Foo", Kind: "closure",
		Source: &SourceLocation{File: "foo.gno", StartLine: 1, EndLine: 1},
	}}

	Enrich(context.Background(), nil, nodes, "/r/demo/foo", 0, fetcher, hl)

	assert.Empty(t, nodes[0].SourceHTML, "render error → SourceHTML stays empty (graceful)")
}

// TestEnrich_FetchesFilesInParallel locks in the parallelism guarantee.
// A timed fetcher records the peak number of concurrent in-flight calls;
// for N>1 distinct files, the peak must exceed 1 — otherwise we've
// regressed back to sequential I/O. Each fetch sleeps a small amount so
// goroutines actually overlap on CI.
func TestEnrich_FetchesFilesInParallel(t *testing.T) {
	t.Parallel()

	const distinctFiles = 4
	fetcher := &concurrentFetcher{
		delay: 30 * time.Millisecond,
		body:  []byte("x\n"),
	}
	hl := &fakeHighlighter{}

	nodes := make([]StateNode, distinctFiles)
	for i := 0; i < distinctFiles; i++ {
		nodes[i] = StateNode{
			Name:   fmt.Sprintf("f%d", i),
			Kind:   "closure",
			Source: &SourceLocation{File: fmt.Sprintf("file%d.gno", i), StartLine: 1, EndLine: 1},
		}
	}

	start := time.Now()
	Enrich(context.Background(), nil, nodes, "/r/demo/foo", 0, fetcher, hl)
	elapsed := time.Since(start)

	peak := atomic.LoadInt32(&fetcher.peak)
	assert.Greater(t, peak, int32(1),
		"at least 2 fetches must overlap — sequential I/O would peak at 1")
	// Sequential would be ~ distinctFiles * delay; parallel should be roughly delay.
	assert.Less(t, elapsed, time.Duration(distinctFiles)*fetcher.delay,
		"elapsed time must be sub-sequential (proves overlap)")
}

// concurrentFetcher records the maximum number of overlapping Fetch calls.
type concurrentFetcher struct {
	delay   time.Duration
	body    []byte
	current int32
	peak    int32
}

func (c *concurrentFetcher) Fetch(_ context.Context, _, _ string) ([]byte, error) {
	cur := atomic.AddInt32(&c.current, 1)
	defer atomic.AddInt32(&c.current, -1)
	for {
		old := atomic.LoadInt32(&c.peak)
		if cur <= old || atomic.CompareAndSwapInt32(&c.peak, old, cur) {
			break
		}
	}
	time.Sleep(c.delay)
	return c.body, nil
}

// fakeObjectFetcher returns canned JSON per OID and tracks call count.
type fakeObjectFetcher struct {
	bodies map[string][]byte
	calls  int32
	err    error
}

func (f *fakeObjectFetcher) FetchObject(_ context.Context, oid string) ([]byte, error) {
	atomic.AddInt32(&f.calls, 1)
	if f.err != nil {
		return nil, f.err
	}
	if b, ok := f.bodies[oid]; ok {
		return b, nil
	}
	return nil, fmt.Errorf("not found: %s", oid)
}

// fakeStructResponse builds a qobject_json shape for an empty struct with
// two int fields, sufficient to exercise the decode + attach path without
// burdening the test with full Amino realism.
func fakeStructResponse(oid string, val0, val1 int) []byte {
	return []byte(fmt.Sprintf(`{
		"objectid": %q,
		"value": {
			"@type": "/gno.StructValue",
			"Fields": [
				{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "%s"},
				{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "%s"}
			]
		}
	}`, oid, encodeInt64LE(int64(val0)), encodeInt64LE(int64(val1))))
}

// encodeInt64LE returns the base64 of an int64 little-endian, the way
// Amino encodes the N field of a TypedValue.
func encodeInt64LE(v int64) string {
	buf := make([]byte, 8)
	for i := 0; i < 8; i++ {
		buf[i] = byte(uint64(v) >> (8 * i))
	}
	return base64Encode(buf)
}

// TestEnrichInlinePreviews_AttachesChildren validates the happy path: a top-
// level ref node gets its decoded fields embedded as Children, with one
// fetcher call per unique OID.
func TestEnrichInlinePreviews_AttachesChildren(t *testing.T) {
	t.Parallel()

	const oid = "ffffffffffffffffffffffffffffffffffffffff:1"
	fetcher := &fakeObjectFetcher{
		bodies: map[string][]byte{oid: fakeStructResponse(oid, 7, 11)},
	}

	nodes := []StateNode{{
		Name: "Counter", Type: "Counter", Kind: "ref",
		ObjectID: oid, Expandable: true,
	}}

	EnrichInlinePreviews(context.Background(), nil, nodes, fetcher, nil)

	require.Len(t, nodes[0].Children, 2, "ref must be enriched with the object's fields")
	assert.Equal(t, "7", nodes[0].Children[0].Value)
	assert.Equal(t, "11", nodes[0].Children[1].Value)
	assert.Equal(t, int32(1), atomic.LoadInt32(&fetcher.calls))
}

// TestEnrichInlinePreviews_RespectsBudget caps how many refs get prefetched.
// The first maxInlinePreviewFetches (top-level priority) get children; the
// rest stay as bare refs the user can click into.
func TestEnrichInlinePreviews_RespectsBudget(t *testing.T) {
	t.Parallel()

	bodies := make(map[string][]byte, maxInlinePreviewFetches+5)
	nodes := make([]StateNode, maxInlinePreviewFetches+5)
	for i := range nodes {
		oid := fmt.Sprintf("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:%d", i+1)
		bodies[oid] = fakeStructResponse(oid, i, 0)
		nodes[i] = StateNode{
			Name: fmt.Sprintf("R%d", i), Kind: "ref",
			ObjectID: oid, Expandable: true,
		}
	}
	fetcher := &fakeObjectFetcher{bodies: bodies}

	EnrichInlinePreviews(context.Background(), nil, nodes, fetcher, nil)

	enriched, leftAsLink := 0, 0
	for _, n := range nodes {
		if len(n.Children) > 0 {
			enriched++
		} else {
			leftAsLink++
		}
	}
	assert.Equal(t, maxInlinePreviewFetches, enriched, "top N must get inline children")
	assert.Equal(t, 5, leftAsLink, "overflow refs must keep their plain link rendering")
	assert.Equal(t, int32(maxInlinePreviewFetches), atomic.LoadInt32(&fetcher.calls),
		"fetcher called exactly budget times — no over-fetch")
}

// TestEnrichInlinePreviews_DedupesByOID: when two nodes reference the same
// stored object, fetch happens once. Saves chain RPC load when an object is
// referenced multiple times in the same view.
func TestEnrichInlinePreviews_DedupesByOID(t *testing.T) {
	t.Parallel()

	const oid = "ffffffffffffffffffffffffffffffffffffffff:1"
	fetcher := &fakeObjectFetcher{
		bodies: map[string][]byte{oid: fakeStructResponse(oid, 1, 2)},
	}

	nodes := []StateNode{
		{Name: "A", Kind: "ref", ObjectID: oid, Expandable: true},
		{Name: "B", Kind: "ref", ObjectID: oid, Expandable: true},
	}

	EnrichInlinePreviews(context.Background(), nil, nodes, fetcher, nil)

	assert.Equal(t, int32(1), atomic.LoadInt32(&fetcher.calls),
		"two refs to the same OID → one fetch")
	require.Len(t, nodes[0].Children, 2)
	require.Len(t, nodes[1].Children, 2)
}

// TestEnrichInlinePreviews_FetchError leaves the node as a plain ref (no
// Children) so the user can still click into it. The page must keep
// rendering even when one ref fails to load.
func TestEnrichInlinePreviews_FetchError(t *testing.T) {
	t.Parallel()

	fetcher := &fakeObjectFetcher{err: errors.New("rpc down")}
	nodes := []StateNode{{
		Name: "Counter", Kind: "ref",
		ObjectID: "ffffffffffffffffffffffffffffffffffffffff:1", Expandable: true,
	}}

	EnrichInlinePreviews(context.Background(), nil, nodes, fetcher, nil)

	assert.Empty(t, nodes[0].Children, "fetch error → no inline children, ref stays clickable")
}

// TestEnrichInlinePreviews_TopLevelPriority validates breadth-first ordering:
// when there are more candidates than the budget allows, top-level refs win
// and deeper nested refs starve. Critical so the most-visible refs always
// get previewed first.
func TestEnrichInlinePreviews_TopLevelPriority(t *testing.T) {
	t.Parallel()

	bodies := make(map[string][]byte)

	// Saturate the budget with top-level refs (one per slot).
	nodes := make([]StateNode, maxInlinePreviewFetches)
	for i := 0; i < maxInlinePreviewFetches; i++ {
		oid := fmt.Sprintf("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:%d", i+1)
		bodies[oid] = fakeStructResponse(oid, i, 0)
		nodes[i] = StateNode{Name: fmt.Sprintf("top%d", i), Kind: "ref", ObjectID: oid, Expandable: true}
	}

	// Append a wrapper containing a deep ref — its budget slot should be
	// stolen by the top-level refs ahead of it.
	const deepOID = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb:1"
	bodies[deepOID] = fakeStructResponse(deepOID, 99, 0)
	nodes = append(nodes, StateNode{
		Name: "wrapper", Kind: "struct", Expandable: true,
		Children: []StateNode{
			{Name: "deepRef", Kind: "ref", ObjectID: deepOID, Expandable: true},
		},
	})

	EnrichInlinePreviews(context.Background(), nil, nodes, &fakeObjectFetcher{bodies: bodies}, nil)

	for i := 0; i < maxInlinePreviewFetches; i++ {
		assert.NotEmpty(t, nodes[i].Children,
			"top-level ref %d must be preview-enriched", i)
	}
	assert.Empty(t, nodes[len(nodes)-1].Children[0].Children,
		"deep ref must starve once the budget is consumed by top-level refs")
}

// fakeTypeFetcher returns a canned StructType for a TypeID — deterministic
// shape so tests can assert on the resolved field names.
type fakeTypeFetcher struct {
	bodies map[string][]byte
	calls  int32
}

func (f *fakeTypeFetcher) FetchType(_ context.Context, tid string) ([]byte, error) {
	atomic.AddInt32(&f.calls, 1)
	if b, ok := f.bodies[tid]; ok {
		return b, nil
	}
	return nil, fmt.Errorf("type not found: %s", tid)
}

// TestEnrichInlinePreviews_ResolvesFieldNamesViaType is the load-bearing
// UX assertion: when a previewed object's TypeID is set and a type fetcher
// is provided, the embedded children carry their declared field names —
// not "0", "1", "2". This is what turns "0:string='gnome#0'" into
// "name:string='gnome#0'" in the realm view.
func TestEnrichInlinePreviews_ResolvesFieldNamesViaType(t *testing.T) {
	t.Parallel()

	const oid = "ffffffffffffffffffffffffffffffffffffffff:11"
	const tid = "gno.land/r/demo/x.User"

	objJSON := []byte(fmt.Sprintf(`{
		"objectid": %q,
		"value": {
			"@type": "/gno.HeapItemValue",
			"Value": {
				"T": {"@type": "/gno.RefType", "ID": %q},
				"V": {"@type": "/gno.StructValue", "Fields": [
					{"T": {"@type": "/gno.PrimitiveType", "value": "16"}, "V": {"@type": "/gno.StringValue", "value": "alice"}},
					{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "HgAAAAAAAAA="}
				]}
			}
		}
	}`, oid, tid))
	typeJSON := []byte(fmt.Sprintf(`{
		"typeid": %q,
		"type": {"@type": "/gno.StructType", "PkgPath": "gno.land/r/demo/x", "Fields": [
			{"Name": "Name", "Type": {"@type": "/gno.PrimitiveType", "value": "16"}, "Embedded": false, "Tag": ""},
			{"Name": "Age", "Type": {"@type": "/gno.PrimitiveType", "value": "32"}, "Embedded": false, "Tag": ""}
		]}
	}`, tid))

	objFetcher := &fakeObjectFetcher{bodies: map[string][]byte{oid: objJSON}}
	typeFetcher := &fakeTypeFetcher{bodies: map[string][]byte{tid: typeJSON}}

	nodes := []StateNode{{
		Name: "user", Type: "*x.User", Kind: "pointer",
		ObjectID: oid, TypeID: tid, Expandable: true,
	}}

	EnrichInlinePreviews(context.Background(), nil, nodes, objFetcher, typeFetcher)

	require.Len(t, nodes[0].Children, 2, "preview unwraps the heap item to expose the struct fields")
	assert.Equal(t, "Name", nodes[0].Children[0].Name,
		"resolved field name from qtype_json — not positional '0'")
	assert.Equal(t, `"alice"`, nodes[0].Children[0].Value)
	assert.Equal(t, "Age", nodes[0].Children[1].Name)
	assert.Equal(t, "30", nodes[0].Children[1].Value)
	assert.Equal(t, int32(1), atomic.LoadInt32(&typeFetcher.calls),
		"one type fetch — even if N refs share the same TypeID we'd dedupe")
}

// TestEnrichInlinePreviews_FollowsHeapToRef reproduces gno's standard
// `*T` storage chain: a top-level pointer points to a heap item whose
// inner Value is itself a RefValue to the actual struct. Naive 1-round
// preview only fetches the heap item and stops, leaving the user with a
// "value : Type → :N" wrapper instead of the struct fields. With multi-
// round preview, the second round fetches the inner ref and exposes
// fields directly.
func TestEnrichInlinePreviews_FollowsHeapToRef(t *testing.T) {
	t.Parallel()

	const outerOID = "ffffffffffffffffffffffffffffffffffffffff:11"
	const innerOID = "ffffffffffffffffffffffffffffffffffffffff:12"
	const tid = "gno.land/r/demo/x.User"

	// :11 = HeapItemValue whose inner TypedValue.V is a RefValue → :12.
	heapJSON := []byte(fmt.Sprintf(`{
		"objectid": %q,
		"value": {
			"@type": "/gno.HeapItemValue",
			"Value": {
				"T": {"@type": "/gno.RefType", "ID": %q},
				"V": {"@type": "/gno.RefValue", "ObjectID": %q}
			}
		}
	}`, outerOID, tid, innerOID))

	// :12 = the actual StructValue.
	structJSON := []byte(fmt.Sprintf(`{
		"objectid": %q,
		"value": {
			"@type": "/gno.StructValue",
			"Fields": [
				{"T": {"@type": "/gno.PrimitiveType", "value": "16"}, "V": {"@type": "/gno.StringValue", "value": "alice"}},
				{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "HgAAAAAAAAA="}
			]
		}
	}`, innerOID))

	typeJSON := []byte(fmt.Sprintf(`{
		"typeid": %q,
		"type": {"@type": "/gno.StructType", "PkgPath": "gno.land/r/demo/x", "Fields": [
			{"Name": "Name", "Type": {"@type": "/gno.PrimitiveType", "value": "16"}, "Embedded": false, "Tag": ""},
			{"Name": "Age", "Type": {"@type": "/gno.PrimitiveType", "value": "32"}, "Embedded": false, "Tag": ""}
		]}
	}`, tid))

	objFetcher := &fakeObjectFetcher{bodies: map[string][]byte{outerOID: heapJSON, innerOID: structJSON}}
	typeFetcher := &fakeTypeFetcher{bodies: map[string][]byte{tid: typeJSON}}

	nodes := []StateNode{{
		Name: "user", Type: "*x.User", Kind: "pointer",
		ObjectID: outerOID, TypeID: tid, Expandable: true,
	}}

	EnrichInlinePreviews(context.Background(), nil, nodes, objFetcher, typeFetcher)

	// Round 1 fetches :11 → reveals one ref-only child pointing at :12.
	// Round 2 picks up that ref and fetches :12 → struct fields surface.
	require.NotEmpty(t, nodes[0].Children, "preview must reach the inner struct")
	require.Len(t, nodes[0].Children, 1, "outer ref unwraps to one inner ref before round 2")
	innerNode := &nodes[0].Children[0]
	require.Len(t, innerNode.Children, 2,
		"round 2 must expose the inner struct's fields, not stop at the ref node")
	assert.Equal(t, "Name", innerNode.Children[0].Name,
		"inner-struct fields carry their declared names from qtype_json")
	assert.Equal(t, `"alice"`, innerNode.Children[0].Value)
	assert.Equal(t, "Age", innerNode.Children[1].Name)
	assert.Equal(t, "30", innerNode.Children[1].Value)
}

// TestEnrich_BuildsHrefViaGnoURL verifies that Enrich computes Href for
// every node carrying an ObjectID, going through weburl.GnoURL — not a
// hand-rolled string template — so URL encoding stays consistent across
// gnoweb. Catches: a regression where ":" in the OID is left unencoded
// would break the URL parser (it splits paths on ":" before "$").
func TestEnrich_BuildsHrefViaGnoURL(t *testing.T) {
	t.Parallel()

	nodes := []StateNode{
		{Name: "Users", Kind: "ref", ObjectID: "ffffffffffffffffffffffffffffffffffffffff:42", Expandable: true},
		{Name: "leaf", Kind: "primitive", Value: "1"},
		{Name: "Branch", Kind: "struct", Expandable: true, Children: []StateNode{
			{Name: "nested", Kind: "ref", ObjectID: "abcdef0123456789abcdef0123456789abcdef01:7", Expandable: true},
		}},
	}

	Enrich(context.Background(), nil, nodes, "/r/demo/foo", 0, nil, nil)

	assert.NotEmpty(t, nodes[0].Href, "ref nodes must get an Href")
	assert.Empty(t, nodes[1].Href, "leaf without ObjectID has no Href")
	assert.NotEmpty(t, nodes[2].Children[0].Href, "nested ref nodes also get Href (recursion)")

	// Critical: the encoded ":" in the OID must round-trip through gnoweb's URL parser.
	gnourl, err := weburl.Parse("https://gno.land" + string(nodes[0].Href))
	require.NoError(t, err, "Href must be parsable by weburl — the very thing it routes to")
	assert.Equal(t, "/r/demo/foo", gnourl.Path)
	assert.Equal(t, "ffffffffffffffffffffffffffffffffffffffff:42", gnourl.WebQuery.Get("oid"),
		"ObjectID must round-trip via the URL parser without truncation at ':'")
	assert.True(t, gnourl.WebQuery.Has("state"), "state flag preserved")
}

// TestSliceLines covers the line-range slicer's edge cases — wrong line
// numbers from upstream must not crash or produce nonsense.
func TestSliceLines(t *testing.T) {
	t.Parallel()

	src := []byte("a\nb\nc\nd\ne\n")
	cases := []struct {
		name       string
		start, end int
		want       string
	}{
		{"normal range", 2, 4, "b\nc\nd"},
		{"single line", 3, 3, "c"},
		{"start past end-of-file", 99, 99, ""},
		{"start zero -> all", 0, 0, "a\nb\nc\nd\ne\n"},
		{"end past EOF clamps to end", 4, 999, "d\ne\n"},
		{"end < start — treat as start..eof", 4, 1, "d\ne\n"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := sliceLines(src, c.start, c.end)
			assert.Equal(t, c.want, string(got))
		})
	}
}

// panicFetcher panics on Fetch; it exists so we can prove the orchestrator
// recovers per-goroutine and does not crash the process.
type panicFetcher struct{}

func (panicFetcher) Fetch(_ context.Context, pkgPath, fileName string) ([]byte, error) {
	panic("boom from file fetcher")
}

type panicObjectFetcher struct{}

func (panicObjectFetcher) FetchObject(_ context.Context, oid string) ([]byte, error) {
	panic("boom from object fetcher")
}

type panicTypeFetcher struct{}

func (panicTypeFetcher) FetchType(_ context.Context, tid string) ([]byte, error) {
	panic("boom from type fetcher")
}

// TestEnrich_RecoversFromFetcherPanic — fetcher panics must not propagate
// past the orchestrator's per-goroutine recover, and must be logged.
func TestEnrich_RecoversFromFetcherPanic(t *testing.T) {
	t.Parallel()

	nodes := []StateNode{{
		Name: "Foo", Kind: "closure",
		Source: &SourceLocation{File: "foo.gno", StartLine: 1, EndLine: 1},
	}}

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	require.NotPanics(t, func() {
		Enrich(context.Background(), logger, nodes, "/r/demo/foo", 0, panicFetcher{}, &fakeHighlighter{})
	}, "panic in FileFetcher.Fetch must be contained in the orchestrator goroutine")

	assert.Empty(t, nodes[0].SourceHTML, "panicked fetch → SourceHTML empty (graceful)")
	log := buf.String()
	assert.Contains(t, log, "boom from file fetcher", "recovered panic value must appear in the log")
	assert.Contains(t, log, "foo.gno", "log must identify which file fetch panicked")
}

// TestEnrichInlinePreviews_RecoversFromObjectFetcherPanic — same invariant
// for the object-fetch pool inside fetchPreviewsConcurrent.
func TestEnrichInlinePreviews_RecoversFromObjectFetcherPanic(t *testing.T) {
	t.Parallel()

	const oid = "ffffffffffffffffffffffffffffffffffffffff:1"
	nodes := []StateNode{{
		Name: "R", Kind: "ref", Expandable: true,
		ObjectID: oid,
	}}

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	require.NotPanics(t, func() {
		EnrichInlinePreviews(context.Background(), logger, nodes, panicObjectFetcher{}, nil)
	}, "panic in StateObjectFetcher.FetchObject must be contained")

	assert.Empty(t, nodes[0].Children, "panicked object fetch → no children attached")
	log := buf.String()
	assert.Contains(t, log, "boom from object fetcher", "recovered panic value must appear in the log")
	assert.Contains(t, log, oid, "log must identify which OID fetch panicked")
}

// blockingFetcher blocks Fetch* until ctx is canceled, then returns ctx.Err.
// Lets a test pass a pre-canceled ctx and assert the orchestrator never
// commits a successful fetch.
type blockingFetcher struct {
	calls int32
}

func (b *blockingFetcher) Fetch(ctx context.Context, _, _ string) ([]byte, error) {
	atomic.AddInt32(&b.calls, 1)
	<-ctx.Done()
	return nil, ctx.Err()
}

func (b *blockingFetcher) FetchObject(ctx context.Context, _ string) ([]byte, error) {
	atomic.AddInt32(&b.calls, 1)
	<-ctx.Done()
	return nil, ctx.Err()
}

func (b *blockingFetcher) FetchType(ctx context.Context, _ string) ([]byte, error) {
	atomic.AddInt32(&b.calls, 1)
	<-ctx.Done()
	return nil, ctx.Err()
}

// TestEnrichInlinePreviews_DedupesOIDAcrossRounds — a self-referencing
// object whose decoded body re-exposes its own OID as a ref must not
// cause a second fetch in round 2.
func TestEnrichInlinePreviews_DedupesOIDAcrossRounds(t *testing.T) {
	t.Parallel()

	// :selfRef points to a stored object whose decoded body contains a
	// Pointer-style ref back to itself. The orchestrator must not fetch
	// :selfRef twice.
	const oid = "ffffffffffffffffffffffffffffffffffffffff:1"
	// Build an object whose value is a struct containing a single field
	// whose typed value is a RefValue pointing back at the same OID.
	selfRefBody := []byte(fmt.Sprintf(`{
		"objectid": %q,
		"value": {
			"@type": "/gno.StructValue",
			"Fields": [
				{"T":{"@type":"/gno.PointerType","Elt":{"@type":"/gno.StructType","Fields":[]}},
				 "V":{"@type":"/gno.RefValue","ObjectID":%q,"Hash":"","PkgPath":""}}
			]
		}
	}`, oid, oid))
	fetcher := &fakeObjectFetcher{bodies: map[string][]byte{oid: selfRefBody}}

	nodes := []StateNode{{
		Name: "root", Kind: "ref", Expandable: true, ObjectID: oid,
	}}

	EnrichInlinePreviews(context.Background(), nil, nodes, fetcher, nil)

	assert.Equal(t, int32(1), atomic.LoadInt32(&fetcher.calls),
		"cycle must not trigger a second fetch of the same OID in round 2")
}

// peakFetcher records the maximum number of overlapping fetch calls
// across both FetchObject and FetchType — used to assert the shared
// preview-pool cap.
type peakFetcher struct {
	delay   time.Duration
	body    []byte
	current int32
	peak    int32
}

func (p *peakFetcher) trackAndDelay() {
	cur := atomic.AddInt32(&p.current, 1)
	defer atomic.AddInt32(&p.current, -1)
	for {
		old := atomic.LoadInt32(&p.peak)
		if cur <= old || atomic.CompareAndSwapInt32(&p.peak, old, cur) {
			break
		}
	}
	time.Sleep(p.delay)
}

func (p *peakFetcher) FetchObject(_ context.Context, _ string) ([]byte, error) {
	p.trackAndDelay()
	return p.body, nil
}

func (p *peakFetcher) FetchType(_ context.Context, _ string) ([]byte, error) {
	p.trackAndDelay()
	return p.body, nil
}

// TestEnrichInlinePreviews_PeakConcurrencyBoundedByPool — the documented
// maxConcurrentObjectFetches cap is the TOTAL of obj+type fetches in flight,
// not per-pool. Spawn 2 * cap of each kind and assert the shared semaphore
// keeps the peak ≤ cap.
func TestEnrichInlinePreviews_PeakConcurrencyBoundedByPool(t *testing.T) {
	t.Parallel()

	const each = maxConcurrentObjectFetches * 2
	nodes := make([]StateNode, each)
	for i := range nodes {
		nodes[i] = StateNode{
			Name: fmt.Sprintf("R%d", i), Kind: "ref", Expandable: true,
			ObjectID: fmt.Sprintf("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:%d", i+1),
			TypeID:   fmt.Sprintf("gno.land/r/x.T%d", i),
		}
	}
	pf := &peakFetcher{delay: 25 * time.Millisecond, body: []byte(`{}`)}

	EnrichInlinePreviews(context.Background(), nil, nodes, pf, pf)

	peak := atomic.LoadInt32(&pf.peak)
	assert.LessOrEqual(t, int(peak), maxConcurrentObjectFetches,
		"obj+type fetches share one pool — peak in-flight must not exceed the documented cap")
	// Intentionally no `peak > 1` floor: GOMAXPROCS=1 CI runners can
	// observe a serialized peak even though the goroutines were eligible
	// to run in parallel. The upper-bound assertion is the load-bearing one.
}

// TestEnrich_AbortsOnCanceledContext — a pre-canceled ctx must short-circuit
// the file-fetch path so we don't drive RPC calls past the deadline.
func TestEnrich_AbortsOnCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	const N = 32 // > maxConcurrentFileFetches so the sem-bound branch matters
	nodes := make([]StateNode, N)
	for i := range nodes {
		nodes[i] = StateNode{
			Name: "f", Kind: "closure",
			Source: &SourceLocation{File: fmt.Sprintf("f%d.gno", i), StartLine: 1, EndLine: 1},
		}
	}
	bf := &blockingFetcher{}

	Enrich(ctx, nil, nodes, "/r/demo/foo", 0, bf, &fakeHighlighter{})

	for i := range nodes {
		assert.Empty(t, nodes[i].SourceHTML, "no fetch may complete under canceled ctx (node %d)", i)
	}
}

// TestEnrichInlinePreviews_AbortsOnCanceledContext — same invariant for the
// object/type pools.
func TestEnrichInlinePreviews_AbortsOnCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	const N = 32
	nodes := make([]StateNode, N)
	for i := range nodes {
		nodes[i] = StateNode{
			Name: fmt.Sprintf("R%d", i), Kind: "ref", Expandable: true,
			ObjectID: fmt.Sprintf("ffffffffffffffffffffffffffffffffffffffff:%d", i+1),
		}
	}
	bf := &blockingFetcher{}

	EnrichInlinePreviews(ctx, nil, nodes, bf, bf)

	for i := range nodes {
		assert.Empty(t, nodes[i].Children, "no preview may attach under canceled ctx (node %d)", i)
	}
}

// TestEnrichInlinePreviews_RecoversFromTypeFetcherPanic — same invariant
// for the type-fetch pool.
func TestEnrichInlinePreviews_RecoversFromTypeFetcherPanic(t *testing.T) {
	t.Parallel()

	const oid = "ffffffffffffffffffffffffffffffffffffffff:1"
	const tid = "gno.land/r/x.T"
	nodes := []StateNode{{
		Name: "R", Kind: "ref", Expandable: true,
		ObjectID: oid, TypeID: tid,
	}}
	objFetcher := &fakeObjectFetcher{
		bodies: map[string][]byte{oid: fakeStructResponse(oid, 1, 2)},
	}

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	require.NotPanics(t, func() {
		EnrichInlinePreviews(context.Background(), logger, nodes, objFetcher, panicTypeFetcher{})
	}, "panic in StateTypeFetcher.FetchType must be contained")

	log := buf.String()
	assert.Contains(t, log, "boom from type fetcher", "recovered panic value must appear in the log")
	assert.Contains(t, log, tid, "log must identify which TID fetch panicked")
}
