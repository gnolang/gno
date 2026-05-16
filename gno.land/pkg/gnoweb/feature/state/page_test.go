package state

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

// pageMockClient is a configurable ClientAdapter for servePage tests.
// objCalls counts StateObject hits so preview-resolution can be observed.
type pageMockClient struct {
	pkgBytes []byte
	pkgErr   error
	docResp  *doc.JSONDocumentation
	docErr   error
	objBytes map[string][]byte
	objErr   error
	typBytes map[string][]byte
	typErr   error

	// *Panic fields make the matching fetch panic — exercises the recover
	// guards in page.go's errgroup goroutines (errgroup doesn't recover).
	pkgPanic, docPanic, objPanic, typPanic bool

	// docHook lets a test observe the ctx Doc receives (e.g. assert that
	// errgroup.WithContext cancels Doc when StatePkg errors fast).
	docHook func(ctx context.Context)

	objCalls int32
}

func (m *pageMockClient) Realm(context.Context, string, string) ([]byte, error)    { return nil, nil }
func (m *pageMockClient) ListPaths(context.Context, string, int) ([]string, error) { return nil, nil }

func (m *pageMockClient) Doc(ctx context.Context, _ string, _ int64) (*doc.JSONDocumentation, error) {
	if m.docPanic {
		panic("doc fetch boom")
	}
	if m.docHook != nil {
		m.docHook(ctx)
	}
	return m.docResp, m.docErr
}

func (m *pageMockClient) StatePkg(context.Context, string, int64) ([]byte, error) {
	if m.pkgPanic {
		panic("statepkg fetch boom")
	}
	return m.pkgBytes, m.pkgErr
}

func (m *pageMockClient) StateObject(_ context.Context, oid string, _ int64) ([]byte, error) {
	atomic.AddInt32(&m.objCalls, 1)
	if m.objPanic {
		panic("stateobject fetch boom")
	}
	if m.objErr != nil {
		return nil, m.objErr
	}
	if b, ok := m.objBytes[oid]; ok {
		return b, nil
	}
	return nil, fmt.Errorf("not found: %s", oid)
}

func (m *pageMockClient) StateType(_ context.Context, tid string, _ int64) ([]byte, error) {
	if m.typPanic {
		panic("statetype fetch boom")
	}
	if m.typErr != nil {
		return nil, m.typErr
	}
	return m.typBytes[tid], nil
}

func newPageHandler(client *pageMockClient) *Handler {
	return New(Deps{
		Client:      client,
		Highlighter: stubHighlighter{},
	})
}

// servePageReq invokes servePage and, on a non-nil View return, renders
// it into the recorder body. Mirrors what gnoweb's Get dispatch does:
// servePage sets cache/SEO headers on w, returns (status, view), and
// the dispatch writes the body via IndexLayout. The test renders the
// View standalone so we can assert on body bytes.
func servePageReq(t *testing.T, h *Handler, query url.Values, path string) *httptest.ResponseRecorder {
	t.Helper()
	if path == "" {
		path = "/r/demo"
	}
	if !query.Has("state") {
		query.Set("state", "")
	}
	u := &weburl.GnoURL{Path: path, WebQuery: query}
	req := httptest.NewRequest(http.MethodGet, path+"$state", nil)
	rec := httptest.NewRecorder()
	status, view := h.servePage(context.Background(), rec, req, u)
	rec.Code = status
	if view != nil {
		if err := view.Render(rec.Body); err != nil {
			t.Fatalf("render view: %v", err)
		}
	}
	return rec
}

// pageFixturePkg returns a small qpkg_json payload with three top-level
// decls: an int primitive, a string primitive, and a struct ref.
const pageFixturePkg = `{
  "names": ["myInt", "myStr", "myStruct"],
  "values": [
    {"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "KgAAAAAAAAA="},
    {"T": {"@type": "/gno.PrimitiveType", "value": "16"}, "V": {"@type": "/gno.StringValue", "value": "hello"}},
    {"T": {"@type": "/gno.RefType", "ID": "gno.land/r/demo.MyStruct"}, "V": {"@type": "/gno.RefValue", "ObjectID": "715383ba05505afed61caa873216e2ee896bede9:10"}}
  ]
}`

// pageFixtureObj is a minimal qobject_json payload for the object page tests.
const pageFixtureObj = `{
	"objectid": "abcdef0123456789abcdef0123456789abcdef01:1",
	"value": {
		"@type": "/gno.StructValue",
		"ObjectInfo": {"ID": "abcdef0123456789abcdef0123456789abcdef01:1"},
		"Fields": [
			{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "AQAAAAAAAAA="},
			{"T": {"@type": "/gno.PrimitiveType", "value": "16"}, "V": {"@type": "/gno.StringValue", "value": "test"}}
		]
	}
}`

func TestServePagePackageHappyPath(t *testing.T) {
	client := &pageMockClient{
		pkgBytes: []byte(pageFixturePkg),
		docResp: &doc.JSONDocumentation{
			Values: []*doc.JSONValueDecl{
				{Values: []*doc.JSONValue{{Name: "myInt", Doc: "the answer"}}},
			},
		},
	}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{}, "/r/demo")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	must := []string{
		`<meta name="htmx-config"`,
		// Gnoweb URL grammar puts state args in $webargs, NOT ?query —
		// `?state&frag=node` would fall through the WebQuery dispatch
		// and return the full page, defeating htmx lazy expansion.
		`hx-get="/r/demo$frag=node`,
		"myInt",
		"myStr",
		"myStruct",
	}
	for _, m := range must {
		if !strings.Contains(body, m) {
			t.Errorf("body missing %q (body head: %s)", m, head(body, 800))
		}
	}
}

func TestServePageObjectHappyPath(t *testing.T) {
	client := &pageMockClient{
		objBytes: map[string][]byte{validOID: []byte(pageFixtureObj)},
	}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{"oid": {validOID}}, "/r/demo")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	// Identity panel surfaces the Object ID (truncated mid). Realm row links
	// back to ?state.
	if !strings.Contains(body, "Object ID") {
		t.Errorf("body missing Object ID label\n%s", head(body, 800))
	}
	if !strings.Contains(body, "Realm") {
		t.Errorf("body missing Realm sidebar label\n%s", head(body, 800))
	}
	// Field-level rows surfaced.
	if !strings.Contains(body, "test") {
		t.Errorf("body missing struct field value\n%s", head(body, 800))
	}
}

func TestServePageInvalidOID400(t *testing.T) {
	// servePage returns the status + view; writePage does not write because
	// the dispatch returns early. Assert on the function return value (the
	// gnoweb wire-in stamps the status from the return, not the recorder).
	h := newPageHandler(&pageMockClient{})
	u := &weburl.GnoURL{Path: "/r/demo", WebQuery: url.Values{"oid": {"garbage"}, "state": {""}}}
	req := httptest.NewRequest(http.MethodGet, "/r/demo$state", nil)
	status, view := h.servePage(context.Background(), httptest.NewRecorder(), req, u)
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
	}
	if view == nil {
		t.Fatalf("view is nil; want a status-error view")
	}
}

func TestServePageNotFound404(t *testing.T) {
	h := newPageHandler(&pageMockClient{pkgErr: errors.New("package not found")})
	u := &weburl.GnoURL{Path: "/r/demo", WebQuery: url.Values{"state": {""}}}
	req := httptest.NewRequest(http.MethodGet, "/r/demo$state", nil)
	status, view := h.servePage(context.Background(), httptest.NewRecorder(), req, u)
	if status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", status, http.StatusNotFound)
	}
	if view == nil {
		t.Fatalf("view is nil; want a status-error view")
	}
}

func TestServePagePinnedHeight(t *testing.T) {
	client := &pageMockClient{pkgBytes: []byte(pageFixturePkg)}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{"height": {"12345"}}, "/r/demo")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	cc := rec.Header().Get("Cache-Control")
	if !strings.Contains(cc, "max-age=86400") || !strings.Contains(cc, "immutable") {
		t.Fatalf("Cache-Control = %q, want max-age=86400 + immutable for pinned height", cc)
	}
	body := rec.Body.String()
	// Every hx-get URL must inherit the resolved height stamp. The
	// gnoweb $webargs grammar puts height inside the path-attached
	// webquery (encoded as `&amp;` by html/template in attribute contexts).
	hxGetCount := strings.Count(body, `hx-get="`)
	// Height can land in any position of the alpha-sorted webarg list, so
	// either prefix is valid: `$height=…` (first key) or `&amp;height=…`.
	heightCount := strings.Count(body, `height=12345`)
	if hxGetCount == 0 {
		t.Fatalf("expected at least one hx-get in body\n%s", head(body, 800))
	}
	if heightCount < hxGetCount {
		t.Errorf("hx-get=%d > height-stamp=%d, some fragments not stamped", hxGetCount, heightCount)
	}
}

func TestServePageLatestHeight(t *testing.T) {
	client := &pageMockClient{pkgBytes: []byte(pageFixturePkg)}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{}, "/r/demo")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	cc := rec.Header().Get("Cache-Control")
	if !strings.Contains(cc, "max-age=1") {
		t.Fatalf("Cache-Control = %q, want max-age=1 for latest height", cc)
	}
	body := rec.Body.String()
	// HeightParam empty → no &height= stamp in hx-get URLs.
	if strings.Contains(body, "&amp;height=") || strings.Contains(body, "&height=") {
		t.Errorf("latest mode unexpectedly stamps &height= into hx-get URLs\n%s", head(body, 600))
	}
}

func TestServePageEmbedsDocIndex(t *testing.T) {
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
		t.Fatalf("body missing doc-index script element\n%s", head(body, 1200))
	}
	// Doc map keyed by Name → Doc string.
	if !strings.Contains(body, "myInt") || !strings.Contains(body, "the doc") {
		t.Errorf("doc index payload missing expected entries\n%s", head(body, 1200))
	}
}

func TestServePageNoCookieVary(t *testing.T) {
	client := &pageMockClient{pkgBytes: []byte(pageFixturePkg)}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{}, "/r/demo")

	if v := rec.Header().Get("Vary"); strings.Contains(strings.ToLower(v), "cookie") {
		t.Errorf("Vary header unexpectedly contains Cookie: %q", v)
	}
}

// Pin the contract that the SSR response does NOT inline the raw
// qpkg_json payload — that would amplify every page-view by the full
// chain payload size for a copy button most viewers never click.
func TestServePageDoesNotInlineRawJSON(t *testing.T) {
	raw := []byte(`{
	  "names": ["v"],
	  "values": [
	    {"T": {"@type": "/gno.PrimitiveType", "value": "16"}, "V": {"@type": "/gno.StringValue", "value": "rawjson-canary"}}
	  ]
	}`)
	h := newPageHandler(&pageMockClient{pkgBytes: raw})
	rec := servePageReq(t, h, url.Values{}, "/r/demo")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, head(rec.Body.String(), 400))
	}
	body := rec.Body.String()
	if strings.Contains(body, `data-copy-target="state-raw-json"`) {
		t.Errorf("body must NOT contain state-raw-json copy target (RawJSON inlining was dropped)\n%s", head(body, 1200))
	}
}

func TestServePageSetsHTMLHeaders(t *testing.T) {
	client := &pageMockClient{pkgBytes: []byte(pageFixturePkg)}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{}, "/r/demo")
	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
		t.Errorf("Content-Type = %q, want text/html...", got)
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
	}
	// Canonical ?state is indexable — no X-Robots-Tag.
	if got := rec.Header().Get("X-Robots-Tag"); got != "" {
		t.Errorf("X-Robots-Tag = %q on canonical ?state page; want empty (indexable)", got)
	}
}

func TestServePageObjectSetsNoindex(t *testing.T) {
	client := &pageMockClient{
		objBytes: map[string][]byte{validOID: []byte(pageFixtureObj)},
	}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{"oid": {validOID}}, "/r/demo")
	if got := rec.Header().Get("X-Robots-Tag"); got != "noindex, nofollow" {
		t.Errorf("X-Robots-Tag = %q on ?state&oid= page; want noindex, nofollow", got)
	}
}

// errgroup does not recover panics and net/http only recovers the
// connection goroutine — so a panic from amino decoding attacker-
// controlled chain data inside a page-path fetch goroutine would crash
// the whole process. Fatal fetches (StatePkg/StateObject) must surface
// as a clean 500; non-fatal ones (Doc/StateType) must be swallowed so
// the page still renders.
func TestServePageRecoversFetchGoroutinePanics(t *testing.T) {
	const validOID = "abcdef0123456789abcdef0123456789abcdef01:7"
	const validTID = "1111111111111111111111111111111111111111"
	for _, tc := range []struct {
		name       string
		client     *pageMockClient
		query      url.Values
		wantStatus int
	}{
		{
			"StatePkg panic -> clean 500, not a process crash",
			&pageMockClient{pkgPanic: true},
			url.Values{},
			http.StatusInternalServerError,
		},
		{
			"Doc panic -> swallowed, package page still renders",
			&pageMockClient{pkgBytes: []byte(pageFixturePkg), docPanic: true},
			url.Values{},
			http.StatusOK,
		},
		{
			"StateObject panic -> clean 500, not a process crash",
			&pageMockClient{objPanic: true},
			url.Values{"oid": {validOID}},
			http.StatusInternalServerError,
		},
		{
			"StateType panic -> swallowed, object page still renders",
			&pageMockClient{
				objBytes: map[string][]byte{validOID: []byte(pageFixtureObj)},
				typPanic: true,
			},
			url.Values{"oid": {validOID}, "tid": {validTID}},
			http.StatusOK,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newPageHandler(tc.client)
			rec := servePageReq(t, h, tc.query, "/r/demo")
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
		})
	}
}

// When the primary fetch (StatePkg / StateObject) fails fast, the
// sibling goroutine (Doc / StateType) should see its ctx cancelled so
// it does not waste a chain RPC slot and a round trip after the page
// has already been doomed. Requires errgroup.WithContext.
func TestServePackagePageCancelsSiblingOnFatalFetch(t *testing.T) {
	docSawCancel := make(chan struct{}, 1)
	client := &pageMockClient{
		pkgErr: errors.New("package not found"),
		docHook: func(ctx context.Context) {
			select {
			case <-ctx.Done():
				docSawCancel <- struct{}{}
			case <-time.After(500 * time.Millisecond):
				// no cancel observed — sibling will be reported as not cancelled
			}
		},
	}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{}, "/r/demo")
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
	select {
	case <-docSawCancel:
		// pass
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Doc goroutine ctx was not cancelled after StatePkg failed; errgroup.WithContext missing?")
	}
}

// Pretty view must NOT burst preview RPCs at render time; refs
// hydrate via hx-trigger="revealed once" on viewport entry.
func TestServePackagePageDefersPreviewsViaRevealedTrigger(t *testing.T) {
	client := &pageMockClient{pkgBytes: refsPkgJSON(3)}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{}, "/r/demo")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := atomic.LoadInt32(&client.objCalls); got != 0 {
		t.Errorf("objCalls = %d, want 0 (no preview burst at render time)", got)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `hx-trigger="revealed once"`) {
		t.Errorf("expected hx-trigger=\"revealed once\" in skeleton; missing from body head:\n%s",
			head(body, 800))
	}
}

// TestServePackagePageFullSidebar — the full sidebar TOC must surface
// every realm decl (icon + label + href) regardless of which slice the
// current page renders. On-page entries get in-page anchors; off-page
// entries get the paginated cross-page URL. 12 decls + default page
// size of 5 → 7 off-page entries surface.
func TestServePackagePageFullSidebar(t *testing.T) {
	h := newPageHandler(&pageMockClient{pkgBytes: buildManyTopLevelDeclsFixture(12)})
	rec := servePageReq(t, h, url.Values{}, "/r/demo")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	// Every decl (v0..v11) appears in the TOC, on-page first then off-page.
	for i := 0; i < 12; i++ {
		name := fmt.Sprintf("v%d", i)
		if !strings.Contains(body, `data-name="`+name+`"`) {
			t.Errorf("body missing TOC entry for %q\n%s", name, head(body, 800))
		}
	}
	// Default page (offset=0, limit=5) → indices 0..4 land in-page; 5..11
	// land off-page. Off-page anchors point at the paginated URL.
	if !strings.Contains(body, `href="#state-v0-pretty"`) {
		t.Errorf("on-page entry must use in-page anchor; body head:\n%s", head(body, 1200))
	}
	// One of the off-page entries — paginated href + fragment.
	if !strings.Contains(body, `offset=5`) || !strings.Contains(body, `#state-v5-pretty`) {
		t.Errorf("off-page entry must use paginated href with fragment; body head:\n%s", head(body, 1200))
	}
	// Off-page <li> carries data-off-page="true" for CSS/aria.
	if !strings.Contains(body, `data-off-page="true"`) {
		t.Errorf("off-page <li> must carry data-off-page attr; body head:\n%s", head(body, 1200))
	}
}

// TestServePackagePageSearchFiltersAndPaginates — `?search=v1` filters
// both the rendered card set AND the sidebar TOC to names containing
// "v1" (v1, v10, v11). Non-matching entries are dropped from the sidebar
// so users see exactly what's on-page.
func TestServePackagePageSearchFiltersAndPaginates(t *testing.T) {
	h := newPageHandler(&pageMockClient{pkgBytes: buildManyTopLevelDeclsFixture(12)})
	rec := servePageReq(t, h, url.Values{"search": {"v1"}}, "/r/demo")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()

	// Exactly v1, v10, v11 land as decl cards.
	cardCount := strings.Count(body, `class="b-state-decl"`)
	if cardCount != 3 {
		t.Errorf("got %d b-state-decl cards, want 3 (v1, v10, v11); body head:\n%s",
			cardCount, head(body, 1200))
	}
	for _, name := range []string{"v1", "v10", "v11"} {
		if !strings.Contains(body, `data-name="`+name+`"`) {
			t.Errorf("filtered card for %q missing; body head:\n%s", name, head(body, 1200))
		}
	}

	// Sidebar TOC mirrors the filter — non-matching entries dropped.
	for _, name := range []string{"v0", "v2", "v3", "v4", "v5", "v6", "v7", "v8", "v9"} {
		if strings.Contains(body, `data-name="`+name+`"`) {
			t.Errorf("sidebar TOC kept non-matching %q under search filter; body head:\n%s",
				name, head(body, 1200))
		}
	}
}

// TestServePackagePageSearchInvalid400 — a control byte in search must
// surface as a clean 400, not a reflected banner.
func TestServePackagePageSearchInvalid400(t *testing.T) {
	h := newPageHandler(&pageMockClient{pkgBytes: buildManyTopLevelDeclsFixture(3)})
	u := &weburl.GnoURL{Path: "/r/demo", WebQuery: url.Values{"state": {""}, "search": {"bad\nquery"}}}
	req := httptest.NewRequest(http.MethodGet, "/r/demo$state", nil)
	status, view := h.servePage(context.Background(), httptest.NewRecorder(), req, u)
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
	}
	if view == nil {
		t.Fatalf("view is nil; want a status-error view")
	}
}

// TestSearchResetsOffsetPastFilteredTotal — if the user paginates deep
// and then submits a search whose result set is small, the carried-over
// offset would push them past the filtered total and render a blank
// page. The server resets offset to 0 in that case so the user always
// lands on the first page of matches.
func TestSearchResetsOffsetPastFilteredTotal(t *testing.T) {
	h := newPageHandler(&pageMockClient{pkgBytes: buildManyTopLevelDeclsFixture(12)})
	// "v0" matches exactly one decl. offset=10 (page 3 in the unfiltered
	// realm) is past the filtered total of 1.
	rec := servePageReq(t, h, url.Values{"search": {"v0"}, "offset": {"10"}}, "/r/demo")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `class="b-state-decl" id="state-v0-pretty"`) {
		t.Errorf("v0 card missing — offset wasn't reset to 0; body head:\n%s", head(body, 1500))
	}
}

// TestSearchEmptyStateKeepsSubheader — a no-match query keeps the
// subheader (search bar + kind tabs) and the sidebar shell visible so
// the user can keep typing / clear; only the cards list collapses to
// the empty-state copy.
func TestSearchEmptyStateKeepsSubheader(t *testing.T) {
	h := newPageHandler(&pageMockClient{pkgBytes: buildManyTopLevelDeclsFixture(3)})
	rec := servePageReq(t, h, url.Values{"search": {"xyz"}}, "/r/demo")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "No declarations match") {
		t.Errorf("body missing empty-state copy; body head:\n%s", head(body, 1200))
	}
	if !strings.Contains(body, `type="search" name="search"`) {
		t.Errorf("search bar disappeared on 0 results — should stay anchored; body head:\n%s", head(body, 1200))
	}
	if !strings.Contains(body, `class="b-sidebar`) {
		t.Errorf("sidebar shell disappeared on 0 results — should stay visible; body head:\n%s", head(body, 1200))
	}
}

// TestServePackagePageSearchFragmentOnHXRequest — htmx requests
// (identified by the HX-Request header) get the search fragment back,
// not the full page chrome. The body must contain the cards container,
// the OOB-marked sidebar, and NO <html> wrapper. Response carries
// HX-Push-Url so the address bar reflects the filtered URL.
func TestServePackagePageSearchFragmentOnHXRequest(t *testing.T) {
	h := newPageHandler(&pageMockClient{pkgBytes: buildManyTopLevelDeclsFixture(12)})
	u := &weburl.GnoURL{Path: "/r/demo", WebQuery: url.Values{"state": {""}, "search": {"v1"}}}
	req := httptest.NewRequest(http.MethodGet, "/r/demo$state&search=v1", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	status, view := h.servePage(context.Background(), rec, req, u)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	if view != nil {
		t.Fatalf("view = %v, want nil (body already written)", view)
	}
	body := rec.Body.String()

	// Chrome-less fragment — no full-page wrapper.
	if strings.Contains(body, "<html") || strings.Contains(body, "<!doctype") {
		t.Errorf("htmx fragment must NOT include page chrome; body head=%s", head(body, 400))
	}
	// Cards for v1 / v10 / v11 land in the fragment.
	for _, name := range []string{"v1", "v10", "v11"} {
		if !strings.Contains(body, `data-name="`+name+`"`) {
			t.Errorf("fragment missing card %q; body head=%s", name, head(body, 1200))
		}
	}
	// OOB sidebar present so the TOC stays in sync with the filtered set.
	if !strings.Contains(body, `id="state-sidebar"`) {
		t.Errorf("fragment missing OOB sidebar id; body head=%s", head(body, 1200))
	}
	if !strings.Contains(body, `hx-swap-oob="true"`) {
		t.Errorf("fragment sidebar missing hx-swap-oob=\"true\"; body head=%s", head(body, 1200))
	}
	// HX-Push-Url carries the canonical $webargs URL the address bar should show.
	pushURL := rec.Header().Get("HX-Push-Url")
	if pushURL == "" {
		t.Fatalf("HX-Push-Url header missing")
	}
	if !strings.Contains(pushURL, "$") || !strings.Contains(pushURL, "search=v1") {
		t.Errorf("HX-Push-Url = %q; want canonical $webargs URL with search=v1", pushURL)
	}
	// Same canonical cache header as the full page.
	if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "max-age") {
		t.Errorf("fragment Cache-Control = %q; want a max-age value", cc)
	}
}

// TestServePackagePageFullPageWithoutHXRequest — without the HX-Request
// header, the same URL returns the full page (chrome inside IndexLayout).
// Pins the dispatch contract so a missing HX header never accidentally
// triggers the fragment path.
func TestServePackagePageFullPageWithoutHXRequest(t *testing.T) {
	h := newPageHandler(&pageMockClient{pkgBytes: buildManyTopLevelDeclsFixture(12)})
	rec := servePageReq(t, h, url.Values{"search": {"v1"}}, "/r/demo")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Header().Get("HX-Push-Url") != "" {
		t.Errorf("HX-Push-Url unexpectedly set on full-page response")
	}
	body := rec.Body.String()
	// Full-page response contains the htmx-config meta — fragments don't.
	if !strings.Contains(body, `<meta name="htmx-config"`) {
		t.Errorf("full-page response missing htmx-config meta; body head=%s", head(body, 600))
	}
}
