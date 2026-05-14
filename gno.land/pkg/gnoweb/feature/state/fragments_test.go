package state

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fragMockClient is a configurable ClientAdapter for fragment-handler tests.
// Distinct from json_test.go's mockJSONClient so per-method behavior — delays,
// canned bodies — can be tuned independently. The struct also captures
// per-method call counts so amplification regressions are observable.
//
// StateObject lookup tries objBodies[oid] first (per-OID dispatch for tests
// that exercise preview resolution), then falls back to objBytes (single
// canned payload for simple tests).
type fragMockClient struct {
	pkgBytes []byte
	pkgErr   error
	pkgDelay time.Duration

	objBytes   []byte
	objBodies  map[string][]byte
	objErr     error
	objMissErr error
	objDelay   time.Duration

	typBytes []byte
	typErr   error

	pkgCalls int32
	objCalls int32
	typCalls int32
}

func (m *fragMockClient) Realm(context.Context, string, string) ([]byte, error) {
	return nil, nil
}
func (m *fragMockClient) ListPaths(context.Context, string, int) ([]string, error) {
	return nil, nil
}
func (m *fragMockClient) Doc(context.Context, string, int64) (*doc.JSONDocumentation, error) {
	return nil, nil
}
func (m *fragMockClient) StatePkg(ctx context.Context, _ string, _ int64) ([]byte, error) {
	atomic.AddInt32(&m.pkgCalls, 1)
	if m.pkgDelay > 0 {
		select {
		case <-time.After(m.pkgDelay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return m.pkgBytes, m.pkgErr
}

func (m *fragMockClient) StateObject(ctx context.Context, oid string, _ int64) ([]byte, error) {
	atomic.AddInt32(&m.objCalls, 1)
	if m.objDelay > 0 {
		select {
		case <-time.After(m.objDelay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if m.objErr != nil {
		return nil, m.objErr
	}
	if m.objBodies != nil {
		if b, ok := m.objBodies[oid]; ok {
			return b, nil
		}
		if m.objMissErr != nil {
			return nil, m.objMissErr
		}
		return nil, errors.New("object not found")
	}
	return m.objBytes, nil
}

func (m *fragMockClient) StateType(context.Context, string, int64) ([]byte, error) {
	atomic.AddInt32(&m.typCalls, 1)
	return m.typBytes, m.typErr
}

// fragFileFetcher is a tiny components.FileFetcher used by frag=source tests.
type fragFileFetcher struct {
	body []byte
	err  error
}

func (f *fragFileFetcher) Fetch(_ context.Context, _, _ string) ([]byte, error) {
	return f.body, f.err
}

// echoHighlighter passes the source bytes straight through as template.HTML
// so tests can grep for content without parsing chroma output.
type echoHighlighter struct{}

func (echoHighlighter) Render(_ string, source []byte) (template.HTML, error) {
	return template.HTML("<pre>" + template.HTMLEscapeString(string(source)) + "</pre>"), nil
}

func newFragHandler(client *fragMockClient, ff *fragFileFetcher) *Handler {
	deps := Deps{
		Client:      client,
		Highlighter: echoHighlighter{},
	}
	if ff != nil {
		deps.FileFetcher = ff
	}
	return New(deps)
}

func serveFragReq(t *testing.T, h *Handler, q url.Values) *httptest.ResponseRecorder {
	t.Helper()
	if !q.Has("state") {
		q.Set("state", "")
	}
	u := &weburl.GnoURL{Path: "/r/demo", WebQuery: q}
	req := httptest.NewRequest(http.MethodGet, "/r/demo$state&"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	h.Handle(context.Background(), rec, req, u)
	return rec
}

// fragStructBody returns a minimal valid qobject_json struct payload with
// two int32 fields named for the test assertions.
func fragStructBody(oid string) []byte {
	return []byte(fmt.Sprintf(`{
		"objectid": %q,
		"value": {
			"@type": "/gno.StructValue",
			"Fields": [
				{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "AQAAAAAAAAA="},
				{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "AgAAAAAAAAA="}
			]
		}
	}`, oid))
}

const fragOID = "abcdef0123456789abcdef0123456789abcdef01:1"

// ---------- writeFragError ----------

func TestFragErrorReturnsHTTP200WithBody(t *testing.T) {
	rec := httptest.NewRecorder()
	status, view := writeFragError(rec, "boom", "retry hint")

	assert.Equal(t, http.StatusOK, status,
		"fragment errors return HTTP 200 so htmx swaps the body (ADR-004 §Decision §2)")
	assert.Nil(t, view, "writeFragError writes directly to w; view is always nil")
	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, `role="alert"`, "must surface as ARIA alert")
	assert.Contains(t, body, "b-state-frag-error", "must carry the b-state-frag-error class")
	assert.Contains(t, body, "boom", "message must appear in body")
	assert.Contains(t, body, "retry hint", "retry hint must appear when supplied")
	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"),
		"error fragments must not be cached")
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
}

// ---------- dispatch ----------

func TestServeFragmentUnknownFragType(t *testing.T) {
	h := newFragHandler(&fragMockClient{}, nil)
	rec := serveFragReq(t, h, url.Values{"frag": {"garbage"}})

	assert.Equal(t, http.StatusOK, rec.Code, "unknown frag → fragment-error (HTTP 200)")
	assert.Contains(t, rec.Body.String(), "b-state-frag-error")
}

// ---------- frag=node ----------

func TestFragNodeHappyPath(t *testing.T) {
	client := &fragMockClient{objBytes: fragStructBody(fragOID)}
	h := newFragHandler(client, nil)
	rec := serveFragReq(t, h, url.Values{
		"frag": {"node"},
		"oid":  {fragOID},
	})

	assert.Equal(t, http.StatusOK, rec.Code)
	ct := rec.Header().Get("Content-Type")
	assert.True(t, strings.HasPrefix(ct, "text/html"),
		"Content-Type = %q, want text/html...", ct)
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "noindex, nofollow", rec.Header().Get("X-Robots-Tag"),
		"fragment URLs must carry X-Robots-Tag per ADR-004 §URL contract")

	body := rec.Body.String()
	assert.Contains(t, body, "b-state-frag-node", "must render the fragNode template")
	assert.Contains(t, body, `data-shape="leaf"`,
		"happy path renders children via the shared state/node renderer")
	assert.Equal(t, int32(1), atomic.LoadInt32(&client.objCalls),
		"one StateObject call per frag=node request")
}

func TestFragNodeStampsHeight(t *testing.T) {
	// Set up a ref child so the template emits a nested hx-get URL we can grep.
	const innerOID = "ffffffffffffffffffffffffffffffffffffffff:9"
	body := []byte(fmt.Sprintf(`{
		"objectid": %q,
		"value": {
			"@type": "/gno.StructValue",
			"Fields": [
				{"T": {"@type": "/gno.RefType", "ID": "gno.land/r/x.User"},
				 "V": {"@type": "/gno.RefValue", "ObjectID": %q}}
			]
		}
	}`, fragOID, innerOID))
	client := &fragMockClient{objBodies: map[string][]byte{fragOID: body}}
	h := newFragHandler(client, nil)

	rec := serveFragReq(t, h, url.Values{
		"frag":   {"node"},
		"oid":    {fragOID},
		"height": {"12345"},
	})
	require.Equal(t, http.StatusOK, rec.Code, "got body=%q", rec.Body.String())

	out := rec.Body.String()
	assert.Contains(t, out, "height=12345",
		"every nested hx-get must inherit the parent page's height (stale-while-revalidate invariant)")

	assert.Equal(t, "public, max-age=86400, immutable", rec.Header().Get("Cache-Control"),
		"pinned-height fragments are immutable")
}

func TestFragNodeRefChildStaysLazilyExpandable(t *testing.T) {
	// A ref child whose target object IS resolvable. A tree-view frag=node
	// must STILL render it as a lazy <details> (b-state-lazy + hx-get) —
	// never flatten it into a pre-rendered branch via an eager fetch — so
	// the tree stays recursively drillable, one level (one StateObject RPC)
	// per click. view=tree because the lazy <details> is tree-view markup.
	const innerOID = "ffffffffffffffffffffffffffffffffffffffff:9"
	parent := []byte(fmt.Sprintf(`{
		"objectid": %q,
		"value": {
			"@type": "/gno.StructValue",
			"Fields": [
				{"T": {"@type": "/gno.RefType", "ID": "gno.land/r/x.User"},
				 "V": {"@type": "/gno.RefValue", "ObjectID": %q}}
			]
		}
	}`, fragOID, innerOID))
	client := &fragMockClient{objBodies: map[string][]byte{
		fragOID:  parent,
		innerOID: fragStructBody(innerOID), // resolvable — an eager fetch WOULD succeed
	}}
	h := newFragHandler(client, nil)

	rec := serveFragReq(t, h, url.Values{
		"frag": {"node"},
		"oid":  {fragOID},
		"view": {"tree"},
	})
	require.Equal(t, http.StatusOK, rec.Code, "got body=%q", rec.Body.String())

	out := rec.Body.String()
	assert.Contains(t, out, `class="b-state-lazy"`,
		"a ref child must render as a lazy <details> so the tree stays recursively expandable")
	assert.Contains(t, out, "hx-get=",
		"the lazy ref child must carry an hx-get to fetch its next level on click")
	assert.Equal(t, int32(1), atomic.LoadInt32(&client.objCalls),
		"frag=node loads only the clicked object — ref children are NOT eagerly fetched")
}

// fragMapBody returns a qobject_json MapValue payload with two string→int
// entries — the pretty-view fragment must render these as a fields table.
func fragMapBody(oid string) []byte {
	return []byte(fmt.Sprintf(`{
		"objectid": %q,
		"value": {
			"@type": "/gno.MapValue",
			"ObjectInfo": {"ID": %q},
			"List": {"List": [
				{"Key": {"T": {"@type": "/gno.PrimitiveType", "value": "16"}, "V": {"@type": "/gno.StringValue", "value": "alpha"}},
				 "Value": {"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "AQAAAAAAAAA="}},
				{"Key": {"T": {"@type": "/gno.PrimitiveType", "value": "16"}, "V": {"@type": "/gno.StringValue", "value": "beta"}},
				 "Value": {"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "AgAAAAAAAAA="}}
			]}
		}
	}`, oid, oid))
}

// Regression: a frag=node request for a pretty-view card (no view=tree) must
// render the pretty fields table — not the bare tree .row markup. With
// view=tree it must still render the tree format. The fragment learns the
// view mode from the request's `view` param.
func TestFragNodeRespectsViewMode(t *testing.T) {
	client := &fragMockClient{objBytes: fragMapBody(fragOID)}
	h := newFragHandler(client, nil)

	// Pretty (default): no view= param → pretty fields table.
	pretty := serveFragReq(t, h, url.Values{
		"frag": {"node"},
		"oid":  {fragOID},
	})
	require.Equal(t, http.StatusOK, pretty.Code, "got body=%q", pretty.Body.String())
	pb := pretty.Body.String()
	assert.Contains(t, pb, "fields-frame",
		"pretty-view fragment must render the state/decl-children fields table")
	assert.Contains(t, pb, "fields-head",
		"pretty-view fragment must render the fields-head row (map → Key heading)")
	assert.NotContains(t, pb, `class="row"`,
		"pretty-view fragment must NOT fall back to the bare tree .row markup")

	// Tree: view=tree → tree .row markup.
	tree := serveFragReq(t, h, url.Values{
		"frag": {"node"},
		"oid":  {fragOID},
		"view": {"tree"},
	})
	require.Equal(t, http.StatusOK, tree.Code, "got body=%q", tree.Body.String())
	tb := tree.Body.String()
	assert.Contains(t, tb, `class="row"`,
		"tree-view fragment must still render the tree state/node .row markup")
	assert.NotContains(t, tb, "fields-frame",
		"tree-view fragment must NOT render the pretty fields table")
}

func TestFragNodeCacheControlLatest(t *testing.T) {
	client := &fragMockClient{objBytes: fragStructBody(fragOID)}
	h := newFragHandler(client, nil)
	rec := serveFragReq(t, h, url.Values{
		"frag": {"node"},
		"oid":  {fragOID},
	})
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Cache-Control"), "max-age=1",
		"latest-mode fragments get the 1s freshness window")
}

// fragFuncBody returns a qobject_json payload whose value is a FuncValue
// carrying a Source RefNode spanning lines 2-4 of "f.gno".
func fragFuncBody(oid string) []byte {
	return []byte(fmt.Sprintf(`{
		"objectid": %q,
		"value": {
			"@type": "/gno.FuncValue",
			"Type": {"@type": "/gno.FuncType", "Params": [], "Results": []},
			"Name": "Render",
			"Source": {"@type": "/gno.RefNode",
				"Location": {"PkgPath": "gno.land/r/demo", "File": "f.gno",
					"Span": {"Pos": {"Line": "2", "Column": "1"}, "End": {"Line": "4", "Column": "1"}, "Num": "0"}}}}
	}`, oid))
}

// Regression: expanding a func/closure via frag=node must show the actual
// function body — the handler promotes the decoded FuncValue to the root
// and fetches+highlights its Source span. Without this the fragment was a
// useless "(function): func()" row (the original PR gap).
func TestFragNodeRendersFuncSource(t *testing.T) {
	client := &fragMockClient{objBytes: fragFuncBody(fragOID)}
	ff := &fragFileFetcher{body: []byte("line1\nfunc Render() {}\nbody\nlast\n")}
	h := newFragHandler(client, ff)
	rec := serveFragReq(t, h, url.Values{
		"frag": {"node"},
		"oid":  {fragOID},
	})

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, `<div class="source"`,
		"func frag must render the highlighted source body, not a bare child row")
	assert.Contains(t, body, "func Render() {}",
		"the fetched source span must appear in the fragment")
	assert.NotContains(t, body, "b-state-frag-empty",
		"a func with source is not 'empty'")
}

func TestFragNodeInvalidOID(t *testing.T) {
	h := newFragHandler(&fragMockClient{}, nil)
	rec := serveFragReq(t, h, url.Values{
		"frag": {"node"},
		"oid":  {"not-an-oid"},
	})

	assert.Equal(t, http.StatusOK, rec.Code, "validation failure surfaces via fragment-error (HTTP 200)")
	body := rec.Body.String()
	assert.Contains(t, body, "b-state-frag-error")
	assert.NotContains(t, body, "b-state-frag-children",
		"invalid input must not reach the success template")
}

func TestFragNodeNotFoundReturnsFragError(t *testing.T) {
	client := &fragMockClient{objErr: errors.New("object not found")}
	h := newFragHandler(client, nil)
	rec := serveFragReq(t, h, url.Values{
		"frag": {"node"},
		"oid":  {fragOID},
	})

	assert.Equal(t, http.StatusOK, rec.Code,
		"client errors map to fragment-error (HTTP 200), not real 4xx/5xx")
	assert.Contains(t, rec.Body.String(), "b-state-frag-error")
}

func TestFragNodeInternalErrorHidesDetail(t *testing.T) {
	client := &fragMockClient{objErr: errors.New("boom: chain blew up with secret")}
	h := newFragHandler(client, nil)
	rec := serveFragReq(t, h, url.Values{
		"frag": {"node"},
		"oid":  {fragOID},
	})

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "b-state-frag-error")
	assert.NotContains(t, body, "secret", "internal error detail must not leak to clients")
}

func TestFragNodeDepthCap(t *testing.T) {
	// Confirm DecodeObject uses the fragment depth bound. We don't observe the
	// (too deep) sentinel via the fragment template (which only renders the
	// immediate children, not their nested Children[]). Instead we call the
	// handler with a deeply-nested fixture and assert the decoded tree was
	// bounded by walking the in-memory root that DecodeObject produces — by
	// inspecting Handler behavior end-to-end we'd duplicate render_test.go.
	// Here the load-bearing assertion is that frag=node calls
	// DefaultFragmentRenderConfig (depth=3), which we verify via the public
	// API: DecodeObject on the same fixture with the same config must hit the
	// (too deep) sentinel within depth=3.
	const fixture = `{
		"objectid": "abcdef0123456789abcdef0123456789abcdef01:1",
		"value": {
			"@type": "/gno.StructValue",
			"Fields": [
				{"T": {"@type": "/gno.StructType", "Fields": []},
				 "V": {"@type": "/gno.StructValue",
				       "Fields": [
				         {"T": {"@type": "/gno.StructType", "Fields": []},
				          "V": {"@type": "/gno.StructValue",
				                "Fields": [
				                  {"T": {"@type": "/gno.StructType", "Fields": []},
				                   "V": {"@type": "/gno.StructValue",
				                         "Fields": [
				                           {"T": {"@type": "/gno.StructType", "Fields": []},
				                            "V": {"@type": "/gno.StructValue", "Fields": []}}
				                         ]}}
				                ]}}
				       ]}}
			]
		}
	}`
	root, err := DecodeObject(context.Background(), []byte(fixture), DefaultFragmentRenderConfig())
	require.NoError(t, err)

	// Walk down: at depth >= 3 we should hit a (too deep) sentinel rather
	// than a fully decoded struct. This pins the bound the fragment handler
	// applies via DefaultFragmentRenderConfig.
	cur := StateNode{Children: root.Children}
	hitSentinel := false
	for i := 0; i < 6; i++ {
		if cur.Kind == KindTruncated && cur.Type == "(too deep)" {
			hitSentinel = true
			break
		}
		if len(cur.Children) == 0 {
			break
		}
		cur = cur.Children[0]
	}
	assert.True(t, hitSentinel,
		"frag=node must use DefaultFragmentRenderConfig (depth=3) per ADR-004 §Resource bounds")

	// End-to-end smoke: the handler must still succeed on the deep fixture
	// (no panic, no error status).
	client := &fragMockClient{objBytes: []byte(fixture)}
	h := newFragHandler(client, nil)
	rec := serveFragReq(t, h, url.Values{
		"frag": {"node"},
		"oid":  {fragOID},
	})
	assert.Equal(t, http.StatusOK, rec.Code, "deeply nested input must not produce an error fragment")
}

func TestFragNodeTimeoutFiresWithin2s(t *testing.T) {
	// Mock client that sleeps longer than the per-fragment timeout. We use
	// a real ctx.WithTimeout in the handler (2s) but slow the mock enough
	// that the timeout fires deterministically. Use a 3s mock delay so we
	// hit the 2s bound; cap the test deadline at 4s as a watchdog.
	client := &fragMockClient{
		objBytes: fragStructBody(fragOID),
		objDelay: 3 * time.Second,
	}
	h := newFragHandler(client, nil)

	done := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		done <- serveFragReq(t, h, url.Values{
			"frag": {"node"},
			"oid":  {fragOID},
		})
	}()

	select {
	case rec := <-done:
		// Timeout path → mapClientError sees ctx.DeadlineExceeded; either
		// way the user sees a fragment-error (HTTP 200) — never a hang.
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "b-state-frag-error",
			"slow upstream → fragment-error, not a hung connection")
	case <-time.After(4 * time.Second):
		t.Fatal("frag=node must abort within the 2s per-fragment timeout")
	}
}

// ---------- frag=source ----------

func TestFragSourceHappyPath(t *testing.T) {
	src := []byte("line1\nline2\nline3\nline4\nline5\nline6\n")
	ff := &fragFileFetcher{body: src}
	h := newFragHandler(&fragMockClient{}, ff)

	rec := serveFragReq(t, h, url.Values{
		"frag": {"source"},
		"file": {"foo.gno"},
		"line": {"3"},
	})

	require.Equal(t, http.StatusOK, rec.Code, "got body=%q", rec.Body.String())
	body := rec.Body.String()
	assert.Contains(t, body, "b-state-frag-source", "must render fragSource template")
	assert.Contains(t, body, "foo.gno", "file name appears in the source-fragment header")
	assert.Contains(t, body, "line3", "the target line must appear in the highlighted slice")
}

func TestFragSourceRejectsLargeFile(t *testing.T) {
	// >256 KB body — the fragment must fall back to the link-only message.
	big := make([]byte, MaxFragmentFileSize+1)
	for i := range big {
		big[i] = 'a'
	}
	ff := &fragFileFetcher{body: big}
	h := newFragHandler(&fragMockClient{}, ff)

	rec := serveFragReq(t, h, url.Values{
		"frag": {"source"},
		"file": {"foo.gno"},
		"line": {"1"},
	})

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "b-state-frag-source",
		"oversize file still renders the source-fragment template (with a fallback message)")
	assert.NotContains(t, body, "aaaaaaaaaaaaaaaaa",
		"full content must not be inlined when >MaxFragmentFileSize")
}

func TestFragSourceRejectsInvalidFile(t *testing.T) {
	h := newFragHandler(&fragMockClient{}, &fragFileFetcher{body: []byte("ignored")})
	rec := serveFragReq(t, h, url.Values{
		"frag": {"source"},
		"file": {"../etc/passwd"},
		"line": {"1"},
	})

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "b-state-frag-error",
		"path-traversal must be caught by filePattern before any fetch")
	assert.NotContains(t, body, "ignored", "the fetcher must never be called")
}

// H7: the "See in code" permalink must use the routable $source webargs
// grammar with the full pkg path — not the dead relative ?source form.
func TestFragSourcePermalinkUsesWebargsGrammar(t *testing.T) {
	src := []byte("line1\nline2\nline3\nline4\n")
	ff := &fragFileFetcher{body: src}
	h := newFragHandler(&fragMockClient{}, ff)

	rec := serveFragReq(t, h, url.Values{
		"frag":   {"source"},
		"file":   {"foo.gno"},
		"line":   {"2"},
		"height": {"77"},
	})

	require.Equal(t, http.StatusOK, rec.Code, "got body=%q", rec.Body.String())
	body := rec.Body.String()
	assert.Contains(t, body, "/r/demo$", "permalink must use the $webargs grammar with the pkg path")
	assert.Contains(t, body, "source", "permalink must route to the full-source view")
	assert.NotContains(t, body, `href="?source`, "the dead relative ?source link must be gone")
	assert.Contains(t, body, "height=77", "permalink must carry the height param")
	assert.Contains(t, body, "#L2", "permalink must anchor at the target line")
}

// M4: a func/closure expanded via frag=node whose source file exceeds
// MaxFragmentFileSize must skip highlighting (no huge slice) and fall back
// to the lazy <details>/permalink — the handler must not crash.
func TestFragNodeFuncSourceOversizeFallsBack(t *testing.T) {
	client := &fragMockClient{objBytes: fragFuncBody(fragOID)}
	big := make([]byte, MaxFragmentFileSize+1)
	for i := range big {
		big[i] = 'a'
	}
	ff := &fragFileFetcher{body: big}
	h := newFragHandler(client, ff)
	rec := serveFragReq(t, h, url.Values{
		"frag": {"node"},
		"oid":  {fragOID},
	})

	require.Equal(t, http.StatusOK, rec.Code, "oversize func file must not crash the fragment")
	body := rec.Body.String()
	assert.NotContains(t, body, "aaaaaaaaaaaaaaaaa",
		"oversize source must not be inlined — highlighting is skipped")
	assert.NotContains(t, body, `<div class="source"`,
		"oversize func source falls back to the lazy <details>/permalink, not an inline slice")
}

// M8: the attacker-controlled depth param drives only a presentational
// --depth step-in; an absurd value must be clamped to maxFragmentDepth.
// view=tree because --depth is tree-view markup.
func TestFragNodeDepthClamped(t *testing.T) {
	client := &fragMockClient{objBytes: fragStructBody(fragOID)}
	h := newFragHandler(client, nil)
	rec := serveFragReq(t, h, url.Values{
		"frag":  {"node"},
		"oid":   {fragOID},
		"depth": {"999"},
		"view":  {"tree"},
	})

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, fmt.Sprintf("--depth: %d;", maxFragmentDepth+1),
		"depth=999 must clamp to maxFragmentDepth, children render at clamp+1")
	assert.NotContains(t, body, "--depth: 1000;", "the raw 999 must never reach the template")
}

func TestFragSourceRejectsInvalidLine(t *testing.T) {
	h := newFragHandler(&fragMockClient{}, &fragFileFetcher{body: []byte("x")})
	rec := serveFragReq(t, h, url.Values{
		"frag": {"source"},
		"file": {"foo.gno"},
		"line": {"0"},
	})

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "b-state-frag-error",
		"line<1 must be rejected by ValidateLine")
}
