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

// Each TestUXPromise* test pins one promise from PR #5649's body so a
// future refactor cannot silently break the documented UX contract.

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
	// X-Robots-Tag: object pages are not crawled (ADR-004 §URL contract)
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
// data so fragments can hydrate doc-comments client-side. This pins
// the §8 contract even though the controller-state.ts hydration is
// currently not implemented (flagged MAJOR in the review report).
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

// Promise: OID searchbar navigation (shared controller-searchbar still
// works). The page must include the search form data-controller hook so
// the existing TS controller-searchbar can bind.
func TestUXPromiseSearchBarPresent(t *testing.T) {
	client := &pageMockClient{pkgBytes: []byte(pageFixturePkg)}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{}, "/r/demo")

	body := rec.Body.String()
	if !strings.Contains(body, `data-controller="search"`) {
		t.Errorf("declaration search bar (controller-search) missing")
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

// Promise: Copy package JSON button — the button must exist and carry
// the relabeled text from ADR-004 §Consequences §Negative.
func TestUXPromiseCopyPackageJSONButton(t *testing.T) {
	client := &pageMockClient{pkgBytes: []byte(pageFixturePkg)}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{}, "/r/demo")

	body := rec.Body.String()
	if !strings.Contains(body, "Copy package JSON") {
		t.Errorf(`"Copy package JSON" button missing/relabeled differently`)
	}
}

// Promise: htmx config hardening flags reach the rendered <meta>. ADR-
// 004 §3 specifies allowEval=false, allowScriptTags=false,
// selfRequestsOnly=true, includeIndicatorStyles=false, historyCacheSize=0
// and explicitly says useTemplateFragments must NOT be set (removed in
// htmx 2.x).
func TestUXPromiseHtmxConfigHardening(t *testing.T) {
	client := &pageMockClient{pkgBytes: []byte(pageFixturePkg)}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{}, "/r/demo")

	body := rec.Body.String()
	required := []string{
		`"allowEval":false`,
		`"allowScriptTags":false`,
		`"selfRequestsOnly":true`,
		`"includeIndicatorStyles":false`,
		`"historyCacheSize":0`,
	}
	for _, r := range required {
		if !strings.Contains(body, r) {
			t.Errorf("htmx-config missing flag %q", r)
		}
	}
	if strings.Contains(body, `useTemplateFragments`) && !strings.Contains(body, `// useTemplateFragments`) {
		// Allow it inside a Go template comment ({{/* … */}}) which is
		// stripped at render — but never appearing as a real config key.
		if strings.Contains(body, `"useTemplateFragments"`) {
			t.Errorf("htmx-config must NOT set useTemplateFragments — flag removed in htmx 2.x")
		}
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

// Promise: JSON API endpoints stable (ADR-003 surface unchanged). The
// envelope, headers, and status codes must match ADR-003 baseline.
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
// cache header (ADR-004 §URL contract row 6).
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

// Promise: ADR-004 §8 doc-index hydration — an expandable tree-view
// frag=node child (the kind that can reference a named, documented
// declaration) MUST carry `[data-name]` + an empty `[data-doc-slot]`
// placeholder so controller-state.ts can project docs onto lazy-loaded
// fragments. The doc-slot pair is tree-view markup, hence view=tree.
func TestUXPromiseFragmentDocSlotPlaceholder(t *testing.T) {
	const oid = "abcdef0123456789abcdef0123456789abcdef01:8"
	// One field that is itself a nested struct → renders as a <details>
	// branch row, which carries the data-name + data-doc-slot pair.
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
	client := &fragMockClient{objBytes: body}
	h := newFragHandler(client, nil)
	rec := serveFragReq(t, h, url.Values{"frag": {"node"}, "oid": {oid}, "view": {"tree"}})

	if rec.Code != http.StatusOK {
		t.Fatalf("frag=node: status = %d, want 200", rec.Code)
	}
	bodyStr := rec.Body.String()
	if !strings.Contains(bodyStr, `data-name=`) {
		t.Errorf("expandable frag children must carry data-name for doc-index hydration")
	}
	if !strings.Contains(bodyStr, `data-doc-slot`) {
		t.Errorf("expandable frag children must carry an empty data-doc-slot placeholder")
	}
	if !strings.Contains(bodyStr, `class="b-state-doc"`) {
		t.Errorf("doc-slot must use the b-state-doc class for styling")
	}
}

// Regression: the htmx-lazy permalink ↗ lives INSIDE its parent
// <details>'s <summary> (ADR-004 §2 example). The <a> is interactive
// content, so a click on it navigates without toggling the <details>;
// a click anywhere else in the summary toggles + fires the hx-get. This
// pins the structure so a refactor can't reintroduce the wrapper-div
// detour that pushed the link to a confusing, faint right-margin slot.
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
		t.Errorf("%d b-state-permalink total but only %d inside <summary> — all must be nested in the summary (ADR-004 §2)", permalinks, insideSummary)
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

// R1's regression test lives in gno.land/pkg/gnoweb/handler_http_test.go
// (TestHTTPHandler_StatePageHeaderData) — it exercises the full Get()
// wire-in to verify the global header renders with realm-aware data.
// The feature/state package can't reach IndexLayout, so the test here
// would be tautological.

// Regression: the sidebar "On this page" label must render the
// b-expend-btn affordance (label + checkbox + chevron SVG) so the
// mobile-collapse interaction works. Earlier audits accidentally
// reduced it to a static <div class="expend-label">, dropping the
// tap-to-toggle control on small viewports. The state feature template
// is isolated and can't reuse components/layouts/aside.html's
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

// Regression: when ViewMode=tree, every server-rendered permalink must
// carry `&view=tree` so navigation preserves the user's selection across
// hops. Without this, a tree-view user clicking a ↗ permalink lands on
// the destination page in pretty mode, the JS controller has to flip the
// radio mid-render, and the URL share-ability promise is broken.
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
	rec := servePageReq(t, h, url.Values{"view": {"tree"}}, "/r/demo")
	body := rec.Body.String()

	// `&` in attribute-context interpolation becomes `&amp;` after html/template
	// escapes the template.URL value.
	if !strings.Contains(body, `&amp;view=tree`) {
		t.Errorf("tree-view permalinks missing &view=tree propagation; body head=%s", head(body, 600))
	}

	// Pretty mode (default) must NOT include &view= — keeps the canonical
	// URL short and the nginx cache key minimal.
	rec2 := servePageReq(t, h, url.Values{}, "/r/demo")
	if strings.Contains(rec2.Body.String(), `view=pretty`) || strings.Contains(rec2.Body.String(), `view=tree`) {
		t.Errorf("pretty-mode page unexpectedly stamps view= into permalinks")
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
