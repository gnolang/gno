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

	objCalls int32
}

func (m *pageMockClient) Realm(context.Context, string, string) ([]byte, error)    { return nil, nil }
func (m *pageMockClient) ListPaths(context.Context, string, int) ([]string, error) { return nil, nil }

func (m *pageMockClient) Doc(context.Context, string, int64) (*doc.JSONDocumentation, error) {
	return m.docResp, m.docErr
}

func (m *pageMockClient) StatePkg(context.Context, string, int64) ([]byte, error) {
	return m.pkgBytes, m.pkgErr
}

func (m *pageMockClient) StateObject(_ context.Context, oid string, _ int64) ([]byte, error) {
	atomic.AddInt32(&m.objCalls, 1)
	if m.objErr != nil {
		return nil, m.objErr
	}
	if b, ok := m.objBytes[oid]; ok {
		return b, nil
	}
	return nil, fmt.Errorf("not found: %s", oid)
}

func (m *pageMockClient) StateType(_ context.Context, tid string, _ int64) ([]byte, error) {
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
	heightCount := strings.Count(body, `&amp;height=12345`)
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

func TestServePagePreviewsResolved(t *testing.T) {
	// Three top-level refs with canned object bodies. ResolvePreviews must
	// fan out and inline at least three of them (well under the 15 cap).
	oid1 := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:1"
	oid2 := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb:2"
	oid3 := "cccccccccccccccccccccccccccccccccccccccc:3"
	pkg := fmt.Sprintf(`{
	  "names": ["R1", "R2", "R3"],
	  "values": [
	    {"T": {"@type": "/gno.RefType", "ID": "gno.land/r/demo.T"}, "V": {"@type": "/gno.RefValue", "ObjectID": %q}},
	    {"T": {"@type": "/gno.RefType", "ID": "gno.land/r/demo.T"}, "V": {"@type": "/gno.RefValue", "ObjectID": %q}},
	    {"T": {"@type": "/gno.RefType", "ID": "gno.land/r/demo.T"}, "V": {"@type": "/gno.RefValue", "ObjectID": %q}}
	  ]
	}`, oid1, oid2, oid3)

	client := &pageMockClient{
		pkgBytes: []byte(pkg),
		objBytes: map[string][]byte{
			oid1: previewStructBody(oid1, 1, 2),
			oid2: previewStructBody(oid2, 3, 4),
			oid3: previewStructBody(oid3, 5, 6),
		},
	}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{}, "/r/demo")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := atomic.LoadInt32(&client.objCalls); got < 3 {
		t.Errorf("objCalls = %d, want ≥ 3 (preview resolve), body=%s", got, head(rec.Body.String(), 400))
	}
}

func TestServePageNoCookieVary(t *testing.T) {
	client := &pageMockClient{pkgBytes: []byte(pageFixturePkg)}
	h := newPageHandler(client)
	rec := servePageReq(t, h, url.Values{}, "/r/demo")

	if v := rec.Header().Get("Vary"); strings.Contains(strings.ToLower(v), "cookie") {
		t.Errorf("Vary header unexpectedly contains Cookie: %q (ADR-004 dropped the cookie)", v)
	}
}

func TestServePageRawJSONEmbedded(t *testing.T) {
	// DecodePackage must succeed for the body to reach the Copy package
	// JSON hidden pre; the marker string round-trips through RawJSON.
	raw := []byte(`{
	  "names": ["v"],
	  "values": [
	    {"T": {"@type": "/gno.PrimitiveType", "value": "16"}, "V": {"@type": "/gno.StringValue", "value": "marker"}}
	  ]
	}`)
	h := newPageHandler(&pageMockClient{pkgBytes: raw})
	rec := servePageReq(t, h, url.Values{}, "/r/demo")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, head(rec.Body.String(), 400))
	}
	body := rec.Body.String()
	if !strings.Contains(body, `data-copy-target="state-raw-json"`) {
		t.Errorf("body missing data-copy-target=state-raw-json (hidden RawJSON pre)\n%s", head(body, 1200))
	}
	if !strings.Contains(body, "marker") {
		t.Errorf("body missing decoded marker string\n%s", head(body, 800))
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
	// ADR-004 §URL contract: canonical ?state is indexable — no X-Robots-Tag.
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
