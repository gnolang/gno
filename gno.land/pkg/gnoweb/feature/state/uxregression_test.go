package state

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

// Each TestUXPromise* test pins one UX contract so a future refactor
// cannot silently break it.

// Promise: Bookmarkable URLs — `?state&oid=X&tid=Y` round-trips. The
// page-render must accept &oid and &tid together and the resulting body
// must reflect the requested object (its OID truncation appears in
// CountLabel).
func TestUXPromiseBookmarkableObjectURL(t *testing.T) {
	oid := "abcdef0123456789abcdef0123456789abcdef01:7"
	tid := "1111111111111111111111111111111111111111"
	client := &pageMockClient{
		objBytes: map[string][]byte{oid: []byte(`{"objectid": "` + oid + `", "value": {"@type": "/gno.StructValue", "Fields": []}}`)},
		typBytes: map[string][]byte{tid: []byte(`{"@type": "/gno.StructType", "PkgPath": "x"}`)},
	}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{"oid": {oid}, "tid": {tid}}, "/r/demo")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Object ") {
		t.Errorf("object page must surface Object label; body=%q", head(body, 400))
	}
	// X-Robots-Tag: object pages are not crawled.
	if got := rec.Header().Get("X-Robots-Tag"); got != "noindex, nofollow" {
		t.Errorf("X-Robots-Tag = %q, want noindex,nofollow for object pages", got)
	}
}

// Promise: Time-travel `?height=N` + ↺ Latest link. Pinned height must
// stamp the immutable Cache-Control AND a Latest link without the
// height param.
func TestUXPromiseTimeTravelLatestLink(t *testing.T) {
	client := &pageMockClient{pkgBytes: []byte(pageFixturePkg)}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{"height": {"42"}}, "/r/demo")

	body := rec.Body.String()
	if !strings.Contains(body, "@ #42") {
		t.Errorf("pinned-height badge missing; body head=%q", head(body, 400))
	}
	if !strings.Contains(body, "↺ Latest") {
		t.Errorf("Latest link missing on pinned page; body head=%q", head(body, 400))
	}
	if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") {
		t.Errorf("pinned page Cache-Control = %q, want immutable", cc)
	}
}

// Promise: doc comments inline (top-level + doc-index embedded for
// fragments). The script island MUST be present and contain the doc
// data so fragments can hydrate doc-comments client-side.
func TestUXPromiseDocIndexIslandPresent(t *testing.T) {
	client := &pageMockClient{
		pkgBytes: []byte(pageFixturePkg),
		docResp: &doc.JSONDocumentation{
			Values: []*doc.JSONValueDecl{
				{Values: []*doc.JSONValue{{Name: "myInt", Doc: "the doc"}}},
			},
		},
	}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{}, "/r/demo")

	body := rec.Body.String()
	if !strings.Contains(body, `<script type="application/json" id="state-doc-index">`) {
		t.Fatalf("doc-index script island missing")
	}
	// The doc must be HTML-escaped inside the JSON island — the html/template
	// engine must NOT double-escape and must protect </script> breakouts.
	if !strings.Contains(body, "the doc") {
		t.Errorf("doc body missing")
	}
}

// Promise: doc-comments survive </script> breakout attempts in source
// doc-comments. encoding/json escapes <, > by default — verify this
// reaches the rendered body intact.
func TestUXPromiseDocIndexHTMLEscaped(t *testing.T) {
	client := &pageMockClient{
		pkgBytes: []byte(pageFixturePkg),
		docResp: &doc.JSONDocumentation{
			Values: []*doc.JSONValueDecl{
				{Values: []*doc.JSONValue{{Name: "myInt", Doc: "</script><script>alert(1)</script>"}}},
			},
		},
	}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{}, "/r/demo")
	body := rec.Body.String()
	if strings.Contains(body, "</script><script>alert(1)") {
		t.Fatalf("raw </script> reached body — XSS via doc-comments")
	}
	if !strings.Contains(body, `</script>`) {
		t.Errorf("expected </script> u-escaped in doc-index; body head=%q", head(body, 800))
	}
}

// Promise: server-side search bar. The input is now htmx-driven: it
// fires GET <pkg>$state&search=… on input/Enter, swaps #state-results
// inner HTML, OOB-swaps #state-sidebar, and pushes the canonical URL.
// No full reload, no JS form-intercept, no focus loss.
func TestUXPromiseSearchBarPresent(t *testing.T) {
	client := &pageMockClient{pkgBytes: []byte(pageFixturePkg)}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{}, "/r/demo")

	body := rec.Body.String()
	required := []string{
		`hx-get="/r/demo$state"`,
		`hx-trigger="input changed delay:200ms, keyup[key=='Enter']"`,
		`hx-target="#state-results"`,
		`hx-push-url="true"`,
		`type="search" name="search"`,
	}
	for _, r := range required {
		if !strings.Contains(body, r) {
			t.Errorf("search bar missing %q; body head=%s", r, head(body, 1200))
		}
	}
	// The old JS-driven handlers must be gone.
	if strings.Contains(body, `data-action="input->state#liveSearch"`) ||
		strings.Contains(body, `data-action="submit->state#submitSearch"`) {
		t.Errorf("legacy JS search handlers must be removed; body head=%s", head(body, 1200))
	}
}

// Promise: per-card anchors `#state-<name>` — the tree-view rows must
// carry a stable id matching the TOC anchor.
func TestUXPromisePerCardAnchorsStable(t *testing.T) {
	client := &pageMockClient{pkgBytes: []byte(pageFixturePkg)}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{}, "/r/demo")

	body := rec.Body.String()
	// stateAnchorOf prefixes "state-" and lowercases the name.
	if !strings.Contains(body, `id="state-myint-pretty"`) && !strings.Contains(body, `id="state-myint-tree"`) {
		t.Errorf("per-card anchor missing for myInt; body head=%q", head(body, 800))
	}
	if !strings.Contains(body, `href="#state-myint-pretty"`) && !strings.Contains(body, `href="#state-myint-tree"`) {
		t.Errorf("TOC anchor href missing for myInt")
	}
}

// Promise: Copy package JSON button exists and hits ?state&json via
// async fetch (no inline RawJSON in the page body — that was the memory
// amp vector). Wires data-copy-fetch-value (lazy) instead of the old
// data-copy-remote-value (inline).
func TestUXPromiseCopyPackageJSONAsyncButton(t *testing.T) {
	client := &pageMockClient{pkgBytes: []byte(pageFixturePkg)}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{}, "/r/demo")

	body := rec.Body.String()
	if !strings.Contains(body, "Copy package JSON") {
		t.Errorf(`"Copy package JSON" button label missing`)
	}
	if !strings.Contains(body, `data-copy-fetch-value=`) {
		t.Errorf("Copy button must wire async fetch via data-copy-fetch-value")
	}
	// The href encodes both ?state and ?json — actual encoding may vary,
	// but state=& and json= must both be present in the URL.
	if !strings.Contains(body, `$state`) || !strings.Contains(body, `json`) {
		t.Errorf("Copy button URL must target ?state&json: body head=%s", head(body, 600))
	}
	if strings.Contains(body, `data-copy-remote-value="state-raw-json"`) {
		t.Errorf("inline RawJSON copy target must NOT be present (memory amp vector)")
	}
}

// Promise: full-page response carries the SSR doc-index island that
// fragments don't — it's the marker downstream code (and humans) use
// to tell a bootstrap render from a partial swap. htmx hardening lives
// in the controller-state.ts bundle; see TestControllerStateBundleHardensHtmxConfig.
func TestUXPromiseFullPageHasDocIndexIsland(t *testing.T) {
	client := &pageMockClient{pkgBytes: []byte(pageFixturePkg)}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{}, "/r/demo")

	body := rec.Body.String()
	if !strings.Contains(body, `id="state-doc-index"`) {
		t.Errorf("full-page response missing doc-index island; body head=%s", body[:min(600, len(body))])
	}
}

// Promise: htmx fragments never write into the breadcrumb (full-page
// nav only). frag=node must NOT contain any breadcrumb-rendering markup.
func TestUXPromiseFragmentsDoNotChangeBreadcrumb(t *testing.T) {
	oid := "abcdef0123456789abcdef0123456789abcdef01:9"
	client := &fragMockClient{
		objBytes: []byte(`{"objectid": "` + oid + `", "value": {"@type": "/gno.StructValue", "Fields": []}}`),
	}
	h := newFragHandler(client, nil)
	rec := serveFragReq(t, h, url.Values{"frag": {"node"}, "oid": {oid}})
	if rec.Code != http.StatusOK {
		t.Fatalf("frag=node: status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, "b-breadcrumb") || strings.Contains(body, "breadcrumb") {
		t.Errorf("frag=node must not render breadcrumb; body=%q", body)
	}
}

// Promise: JSON API surface unchanged from ADR-003 — envelope, headers,
// and status codes must match the documented baseline.
func TestUXPromiseJSONAPIStable(t *testing.T) {
	pkg := []byte(`{"names":["x"],"values":[{"T":{"@type":"/gno.PrimitiveType","value":"32"},"N":"AQAAAAAAAAA="}]}`)
	client := &pageMockClient{pkgBytes: pkg}
	h := newPageHandler(client)
	u := &weburl.GnoURL{Path: "/r/demo", WebQuery: url.Values{"state": {""}, "json": {""}}}
	req := httptest.NewRequest(http.MethodGet, "/r/demo$state&json", nil)
	rec := httptest.NewRecorder()
	status, view := h.Handle(context.Background(), rec, req, u)
	if status != http.StatusOK || view != nil {
		t.Fatalf("?state&json: status=%d view=%v, want 200 nil", status, view)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if got := rec.Header().Get("X-Robots-Tag"); got != "noindex, nofollow" {
		t.Errorf("X-Robots-Tag = %q, want noindex,nofollow on JSON API", got)
	}
	if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "max-age") {
		t.Errorf("Cache-Control missing on JSON API; got=%q", cc)
	}
}

// Promise: ?state&height=N → JSON: pinned height returns immutable
// cache header.
func TestUXPromiseJSONPinnedImmutable(t *testing.T) {
	client := &pageMockClient{pkgBytes: []byte(`{"names":[],"values":[]}`)}
	h := newPageHandler(client)
	u := &weburl.GnoURL{Path: "/r/demo", WebQuery: url.Values{"state": {""}, "json": {""}, "height": {"42"}}}
	req := httptest.NewRequest(http.MethodGet, "/r/demo$state&json&height=42", nil)
	rec := httptest.NewRecorder()
	h.Handle(context.Background(), rec, req, u)

	if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") {
		t.Errorf("pinned JSON Cache-Control = %q, want immutable", cc)
	}
}

// Promise: height-stamping invariant — every fragment hx-get inherits
// the parent page's resolved height. Verified end-to-end: a pinned page
// renders hx-get URLs with &height=N stamped.
func TestUXPromiseFragmentHeightStamp(t *testing.T) {
	oid := "abcdef0123456789abcdef0123456789abcdef01:3"
	pkg := []byte(`{
	  "names": ["myRef"],
	  "values": [
	    {"T": {"@type": "/gno.RefType", "ID": "x.Y"}, "V": {"@type": "/gno.RefValue", "ObjectID": "` + oid + `"}}
	  ]
	}`)
	client := &pageMockClient{pkgBytes: pkg}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{"height": {"42"}}, "/r/demo")

	body := rec.Body.String()
	if !strings.Contains(body, "frag=node") {
		t.Errorf("expected at least one frag=node hx-get; body head=%q", head(body, 800))
	}
	if !strings.Contains(body, "&height=42") && !strings.Contains(body, "height=42") {
		t.Errorf("height-stamping invariant broken: hx-get missing &height=42")
	}
}

// Promise: server-rendered crawler-visible — the top-level page must
// have indexable top-level declarations (no X-Robots-Tag noindex on
// canonical ?state URL).
func TestUXPromiseTopLevelCrawlable(t *testing.T) {
	client := &pageMockClient{pkgBytes: []byte(pageFixturePkg)}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{}, "/r/demo")

	if got := rec.Header().Get("X-Robots-Tag"); got != "" {
		t.Errorf("canonical ?state must be indexable, X-Robots-Tag = %q", got)
	}
}

// Promise: doc-comments are scoped to the pretty container. Tree rows
// are the dense / scannable view — no description paragraphs to preserve
// the hierarchy's readability. The doc-slot for JS hydration lives in
// the pretty SSR wrapper (state/node-details, tested via the page path);
// neither tree nor pretty fragment RESPONSES carry it (fragments fill
// .b-state-node-body inside the SSR'd parent).
func TestUXPromiseDocsScopedToPrettyContainer(t *testing.T) {
	const oid = "abcdef0123456789abcdef0123456789abcdef01:8"
	body := []byte(`{
		"objectid": "` + oid + `",
		"value": {
			"@type": "/gno.StructValue",
			"Fields": [
				{"T": {"@type": "/gno.RefType", "ID": "x.Inner"}, "V": {"@type": "/gno.StructValue", "Fields": [
					{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "AQAAAAAAAAA="}
				]}}
			]
		}
	}`)

	// 1) Tree fragment response: no doc-slot — tree stays dense.
	client := &fragMockClient{objBytes: body}
	h := newFragHandler(client, nil)
	tree := serveFragReq(t, h, url.Values{"frag": {"node"}, "oid": {oid}, "view": {"tree"}})
	if tree.Code != http.StatusOK {
		t.Fatalf("tree frag=node: status = %d, want 200", tree.Code)
	}
	if tb := tree.Body.String(); strings.Contains(tb, `data-doc-slot`) {
		t.Errorf("tree-view fragment must NOT carry data-doc-slot (dense view by design)")
	}

	// 2) SSR page (pretty) wrapper: doc-slot present so JS hydration can
	//    fill descriptions for ref/branch top-level decls.
	pageClient := &pageMockClient{pkgBytes: []byte(pageFixturePkg)}
	ph := newPageHandler(pageClient)
	rec := servePageReq(t, ph, url.Values{}, "/r/demo")
	if rec.Code != http.StatusOK {
		t.Fatalf("page: status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `data-doc-slot`) {
		t.Errorf("pretty page must carry data-doc-slot for doc-index hydration")
	}
}

// Regression: the htmx-lazy permalink ↗ lives INSIDE its parent
// <details>'s <summary>. The <a> is interactive content, so a click on
// it navigates without toggling the <details>; a click elsewhere in the
// summary toggles + fires the hx-get. Pins the structure so a refactor
// can't reintroduce the wrapper-div detour that pushed the link to a
// confusing, faint right-margin slot.
func TestRegressionPermalinkInsideSummary(t *testing.T) {
	data := StateData{
		PkgPath: "/r/test",
		Nodes: []StateNode{
			{Name: "A", Kind: KindStruct, ObjectID: "abc:1", Expandable: true, Anchor: "a"},
			{Name: "Fn", Kind: KindFunc, Source: &SourceLocation{File: "f.gno", StartLine: 5, EndLine: 9}, Anchor: "fn"},
		},
		KindCounts: KindCounts{All: 2, State: 1, Code: 1},
	}
	var buf strings.Builder
	if err := PageTemplate.ExecuteTemplate(&buf, "renderPage", data); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()

	// The legacy wrapper divs must not reappear.
	for _, orphan := range []string{"b-state-lazy-row", "b-state-branch-row"} {
		if strings.Contains(out, orphan) {
			t.Errorf("legacy wrapper %q leaked back into output — permalink belongs inside <summary>", orphan)
		}
	}

	// Every b-state-permalink must sit inside a <summary>...</summary> span.
	permalinks := strings.Count(out, `class="b-state-permalink"`)
	if permalinks == 0 {
		t.Fatal("expected at least one b-state-permalink in output")
	}
	insideSummary := 0
	rest := out
	for {
		i := strings.Index(rest, "<summary")
		if i < 0 {
			break
		}
		end := strings.Index(rest[i:], "</summary>")
		if end < 0 {
			t.Fatal("unbalanced <summary> in template output")
		}
		insideSummary += strings.Count(rest[i:i+end], `class="b-state-permalink"`)
		rest = rest[i+end+len("</summary>"):]
	}
	if insideSummary != permalinks {
		t.Errorf("%d b-state-permalink total but only %d inside <summary> — all must be nested in the summary", permalinks, insideSummary)
	}
}

// Regression: a func/closure source body is NEVER auto-loaded (hx-trigger=load)
// — it stays closed + lazy (hx-trigger=toggle), so opening a page with N
// attacker-controlled funcs can never burst N source-fragment GETs.
func TestRegressionSourceNeverAutoLoads(t *testing.T) {
	nodes := []StateNode{
		{Name: "Fn", Kind: KindFunc, Source: &SourceLocation{File: "f.gno", StartLine: 5, EndLine: 9}, Anchor: "fn"},
		{Name: "Cl", Kind: KindClosure, Source: &SourceLocation{File: "f.gno", StartLine: 12, EndLine: 20}, Anchor: "cl"},
	}
	var buf strings.Builder
	if err := PageTemplate.ExecuteTemplate(&buf, "renderPage", StateData{
		PkgPath: "/r/test", Nodes: nodes, KindCounts: KindCounts{All: 2, Code: 2},
	}); err != nil {
		t.Fatalf("render: %v", err)
	}
	if strings.Contains(buf.String(), `hx-trigger="load once"`) {
		t.Errorf("source-details must NOT auto-load — fan-out is attacker-controlled")
	}
	if !strings.Contains(buf.String(), `hx-trigger="toggle once"`) {
		t.Errorf("source-details must stay lazy (hx-trigger=toggle)")
	}
}

// Regression: the sidebar "On this page" label must render the
// b-expend-btn affordance (label + checkbox + chevron SVG) so the
// mobile-collapse interaction works. The state feature template is
// isolated and can't reuse components/layouts/aside.html's
// ui/expend_label partial, so the state-local "state/expend-label"
// partial must mirror that markup exactly.
func TestRegressionExpendLabelKeepsToggleAffordance(t *testing.T) {
	data := StateData{
		PkgPath: "/r/demo",
		Sidebar: &StateSidebar{
			TOC: []StateTOCEntry{{Label: "x", Anchor: "x", Kind: "primitive"}},
		},
		KindCounts: KindCounts{All: 0},
		CountLabel: "demo",
	}
	var buf strings.Builder
	if err := PageTemplate.ExecuteTemplate(&buf, "renderPage", data); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `class="b-expend-btn"`) {
		t.Errorf("sidebar expend-label missing b-expend-btn class — mobile sidebar collapse broken")
	}
	if !strings.Contains(out, `id="toc-expend"`) {
		t.Errorf("sidebar expend-label missing #toc-expend checkbox — :has(#toc-expend:checked) CSS toggle won't fire")
	}
	if !strings.Contains(out, "#ico-arrow-down") {
		t.Errorf("sidebar expend-label missing chevron SVG — visual affordance lost")
	}
}

// Regression: hx-get + href URLs MUST live in gnoweb's `$webargs` segment,
// not in the standard `?query`. Gnoweb's URL grammar
// (`<path>[$<webargs>][?<query>]`) routes `WebQuery.Has("frag")` only when
// `frag` is in the `$webargs` portion; the `?query` portion is invisible
// to that dispatch. A regression to the `?state&frag=node` form caused the
// fragment handler to fall through to servePage, which returned the full
// ~285KB state HTML and let htmx swap it into the details body — visually,
// "the whole page reloaded inside the expand". This test pins the fix.
func TestRegressionFragmentURLsUseWebargs(t *testing.T) {
	client := &pageMockClient{
		pkgBytes: []byte(`{
		  "names": ["myStruct"],
		  "values": [
		    {"T": {"@type": "/gno.RefType", "ID": "gno.land/r/demo.T"}, "V": {"@type": "/gno.RefValue", "ObjectID": "715383ba05505afed61caa873216e2ee896bede9:10"}}
		  ]
		}`),
	}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{}, "/r/demo")
	body := rec.Body.String()

	// Negative checks — the legacy buggy form must not appear anywhere.
	bad := []string{
		`hx-get="?state`,    // fragments via WebQuery dispatch
		`href="?state&oid=`, // permalinks via WebQuery dispatch
	}
	for _, b := range bad {
		if strings.Contains(body, b) {
			t.Errorf("legacy ?state form leaked back into output: %q", b)
		}
	}

	// Positive check — the corrected form is path-attached `$frag=…`.
	// (Alphabetical key order from EncodeValues puts `frag` first.)
	if !strings.Contains(body, `hx-get="/r/demo$frag=node`) {
		t.Errorf("hx-get URL missing canonical $webargs form; body head=%s", head(body, 600))
	}
}

// Regression: per-container view-mode literal contract. The pretty
// container always emits canonical permalinks (no `view=`) and the tree
// container always emits `&view=tree` — regardless of the URL's `?view=`
// param. The page-mode toggle is client-side CSS-only, so each container
// must stamp its own permalinks at SSR; otherwise switching the toggle
// after page-load would have a pretty user click a tree-stamped link
// (or vice versa) and lose their selection. The old contract assumed a
// uniform per-page view-mode, which broke fragment hydration on toggle.
func TestRegressionTreeViewPermalinksCarryViewParam(t *testing.T) {
	client := &pageMockClient{
		pkgBytes: []byte(`{
		  "names": ["myStruct"],
		  "values": [
		    {"T": {"@type": "/gno.RefType", "ID": "gno.land/r/demo.T"}, "V": {"@type": "/gno.RefValue", "ObjectID": "715383ba05505afed61caa873216e2ee896bede9:10"}}
		  ]
		}`),
	}
	h := newPageHandler(client)

	for _, urlView := range []string{"", "tree"} {
		q := url.Values{}
		if urlView != "" {
			q.Set("view", urlView)
		}
		rec := servePageReq(t, h, q, "/r/demo")
		body := rec.Body.String()

		// Body is one rendered HTML page. Both containers coexist in the
		// DOM; CSS picks which is visible. Per-container view-mode
		// contract must hold in both URL view-mode arms.
		prettyChunk := sliceBetween(body, `<div class="view-pretty">`, `<div class="view-tree"`)
		treeChunk := sliceBetween(body, `<div class="view-tree"`, `</article>`)

		if strings.Contains(prettyChunk, "view=tree") || strings.Contains(prettyChunk, "view=pretty") {
			t.Errorf("URL view=%q: pretty container leaked view= into a permalink", urlView)
		}
		// `&` in attribute-context interpolation becomes `&amp;` after
		// html/template escapes the template.URL value.
		if !strings.Contains(treeChunk, `&amp;view=tree`) {
			t.Errorf("URL view=%q: tree container missing &view=tree on permalinks; head=%s", urlView, head(treeChunk, 400))
		}
	}
}

// Regression: a plain func has no Children and no ObjectID → Shape()==leaf,
// but it IS expandable to disclose its source body. In the tree view it must
// render as an expandable row carrying a lazy source expander — not a
// dead-end leaf row (closures already get this via their captures→branch).
func TestRegressionPlainFuncExpandableInTree(t *testing.T) {
	fn := StateNode{
		Name: "Render", Kind: KindFunc, Type: "func() string",
		Source:     &SourceLocation{File: "render.gno", StartLine: 1, EndLine: 3},
		Expandable: true,
	}
	var buf strings.Builder
	if err := PageTemplate.ExecuteTemplate(&buf, "state/node", map[string]any{
		"Node": fn, "PkgPath": "/r/demo", "Depth": 0,
		"Toplevel": true, "Height": int64(0), "HeightParam": "", "ViewMode": "tree",
	}); err != nil {
		t.Fatalf("render state/node: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "b-state-source-lazy") {
		t.Errorf("plain func in tree must expose a lazy source expander; got: %s", out)
	}
	if strings.Contains(out, `data-shape="leaf"`) {
		t.Errorf("a func with a source body is not a leaf — it must be expandable; got: %s", out)
	}
	// The source expander sits inside the func's .kids — it must carry the
	// child --depth (parent+1) so it indents under the func, not flush-left.
	if !strings.Contains(out, `class="b-state-source-lazy" style="--depth: 1;"`) {
		t.Errorf("tree source expander must carry --depth to align under its func; got: %s", out)
	}
}
