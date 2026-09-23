package omnisearch

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

func parseURL(t *testing.T, raw string) *weburl.GnoURL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("url.Parse(%q): %v", raw, err)
	}
	g, err := weburl.ParseFromURL(u)
	if err != nil {
		t.Fatalf("weburl.ParseFromURL(%q): %v", raw, err)
	}
	return g
}

func TestServeJSONShape(t *testing.T) {
	t.Parallel()

	h := newHandlerWithDir(t, newDiscoveryClient(), newDiscoveryDir(), nil)
	u := parseURL(t, "/r/alice/blog$search&q=blog&json")
	r := httptest.NewRequest(http.MethodGet, "/r/alice/blog", nil)
	w := httptest.NewRecorder()

	status, view := h.Handle(context.Background(), w, r, u)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	if view != nil {
		t.Fatal("view != nil on a json request — the body is written directly")
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q", ct)
	}
	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("missing X-Content-Type-Options: nosniff")
	}

	var resp jsonResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode body: %v\n%s", err, w.Body.String())
	}
	if resp.Query != "blog" {
		t.Errorf("query = %q, want blog", resp.Query)
	}
	// The selector list travels with every response: it is how the omnibar
	// learns what this deployment can answer.
	if len(resp.Selectors) == 0 {
		t.Error("selectors = empty, want the registered set")
	}
	for _, s := range resp.Selectors {
		if s.Source == SourceIndexer {
			t.Errorf("selector %q advertised without an indexer", s.Name)
		}
	}
}

func TestServeJSONRejectsAHostileQuery(t *testing.T) {
	t.Parallel()

	h := newHandlerWithDir(t, newDiscoveryClient(), newDiscoveryDir(), nil)
	u := parseURL(t, "/r/alice/blog$search&q="+url.QueryEscape(strings.Repeat("a", MaxQueryLen+1))+"&json")
	r := httptest.NewRequest(http.MethodGet, "/r/alice/blog", nil)
	w := httptest.NewRecorder()

	status, _ := h.Handle(context.Background(), w, r, u)
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", status)
	}
	var envelope struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil || envelope.Error == "" {
		t.Fatalf("want an {\"error\":…} envelope, got %q", w.Body.String())
	}
}

// The page path renders server-side, which is the whole reason the omnibar
// can stay additive: with JavaScript off the form still lands on a real page.
func TestServePageRendersWithoutJavaScript(t *testing.T) {
	t.Parallel()

	h := newHandlerWithDir(t, newDiscoveryClient(), newDiscoveryDir(), nil)
	u := parseURL(t, "/r/alice/blog$search&q=blog")
	r := httptest.NewRequest(http.MethodGet, "/r/alice/blog", nil)
	w := httptest.NewRecorder()

	status, view := h.Handle(context.Background(), w, r, u)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	if view == nil {
		t.Fatal("view = nil on the page path")
	}

	var body strings.Builder
	if err := view.Render(&body); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := body.String()
	for _, want := range []string{"b-omni", "/r/alice/blog", "Qualifiers"} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered page missing %q", want)
		}
	}
}

// An indexer-backed answer must carry its provenance into the markup, not
// just into the struct: the reader is the one who needs to know the row is
// not consensus data.
func TestIndexerResultsAreMarkedInTheMarkup(t *testing.T) {
	t.Parallel()

	h := newHandlerWithDir(t, newDiscoveryClient(), newDiscoveryDir(), &mockIndexer{})
	u := parseURL(t, "/r/alice/blog$search&q=account:g1abc")
	r := httptest.NewRequest(http.MethodGet, "/r/alice/blog", nil)
	w := httptest.NewRecorder()

	_, view := h.Handle(context.Background(), w, r, u)
	if view == nil {
		t.Fatal("view = nil")
	}
	var body strings.Builder
	if err := view.Render(&body); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := body.String()
	if !strings.Contains(out, "b-tag--indexer") {
		t.Error("indexer group is not tagged in the markup")
	}
	if !strings.Contains(out, "not consensus data") {
		t.Error("provenance footer missing from the markup")
	}
	if !strings.Contains(out, "last indexed block 185214") {
		t.Error("freshness stamp missing from the markup")
	}
}

type denyLimiter struct{}

func (denyLimiter) AllowRequest(*http.Request) bool { return false }

func TestRateLimitedRequestIsRefusedBeforeAnyFetch(t *testing.T) {
	t.Parallel()

	dir := newDiscoveryDir()
	h := New(Deps{
		Client:    newDiscoveryClient(),
		Directory: dir,
		Domain:    "gno.land",
		Limiter:   denyLimiter{},
	})

	u := parseURL(t, "/r/alice/blog$search&q=blog&json")
	r := httptest.NewRequest(http.MethodGet, "/r/alice/blog", nil)
	w := httptest.NewRecorder()

	status, _ := h.Handle(context.Background(), w, r, u)
	if status != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", status)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Error("missing Retry-After")
	}
	if dir.calls != 0 {
		t.Errorf("Directory.Paths called %d times on a throttled request, want 0", dir.calls)
	}
}

// `in:` overrides the URL's own scope so a package-scoped qualifier can be
// typed from anywhere, including a page that names no package.
func TestInQualifierSetsTheScope(t *testing.T) {
	t.Parallel()

	h := newHandlerWithDir(t, newDiscoveryClient(), newDiscoveryDir(), nil)
	q, err := ParseQuery("imports in:/r/demo/boards")
	if err != nil {
		t.Fatalf("ParseQuery: %v", err)
	}
	h.scope(q, parseURL(t, "/u/alice$search"))

	if q.PkgPath != "/r/demo/boards" {
		t.Errorf("PkgPath = %q, want /r/demo/boards", q.PkgPath)
	}
	if q.ChainPath != "gno.land/r/demo/boards" {
		t.Errorf("ChainPath = %q, want gno.land/r/demo/boards", q.ChainPath)
	}
}

// The omnibar's form submits `?q=`, because that is all a plain HTML GET form
// can produce. gnoweb's own links carry `$search&q=`. Both must reach the
// same handler, or the no-JavaScript path silently searches for nothing.
func TestQueryTermAcceptsBothGrammars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
	}{
		{"webarg, as gnoweb links build it", "/r/alice/blog$search&q=blog"},
		{"query string, as a GET form submits it", "/r/alice/blog$search?q=blog"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := queryTerm(parseURL(t, tt.url)); got != "blog" {
				t.Fatalf("queryTerm(%q) = %q, want %q", tt.url, got, "blog")
			}
		})
	}
}

// The results page's form must target the bare path plus `$search`. With
// action="" the browser re-submits the webargs already in the path, so each
// search would append another `q` to the URL.
func TestFormActionDoesNotAccumulateWebargs(t *testing.T) {
	t.Parallel()

	h := newHandlerWithDir(t, newDiscoveryClient(), newDiscoveryDir(), nil)
	q, err := ParseQuery("blog")
	if err != nil {
		t.Fatalf("ParseQuery: %v", err)
	}
	h.scope(q, parseURL(t, "/r/alice/blog$search&q=blog"))

	if q.formAction != "/r/alice/blog$search" {
		t.Fatalf("formAction = %q, want /r/alice/blog$search", q.formAction)
	}
}

// A fan-out selector must not run on a keystroke. The omnibar hits the JSON
// path per character; executing Render on eight realms there would be the
// amplification the resource bounds exist to prevent.
func TestFanoutSelectorIsRefusedOnTheOmnibarPath(t *testing.T) {
	t.Parallel()

	c := newDiscoveryClient()
	c.renders = map[string]string{"/r/alice/blog": "voting"}
	h := newHandlerWithDir(t, c, newDiscoveryDir(), nil)

	u := parseURL(t, "/r/alice/blog$search&q="+url.QueryEscape("render:voting author:alice")+"&json")
	r := httptest.NewRequest(http.MethodGet, "/r/alice/blog", nil)
	w := httptest.NewRecorder()

	if status, _ := h.Handle(context.Background(), w, r, u); status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	if c.realmCalls != 0 {
		t.Errorf("Realm called %d times on the omnibar path, want 0", c.realmCalls)
	}

	var resp jsonResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Groups) != 1 || resp.Groups[0].Error == "" {
		t.Fatalf("groups = %+v, want one group explaining it runs on the page", resp.Groups)
	}
}

// The omnibar re-asks the same question constantly. A matching If-None-Match
// must cost a 304, not a re-render.
func TestSearchJSONAnswers304OnAMatchingETag(t *testing.T) {
	t.Parallel()

	h := newHandlerWithDir(t, newDiscoveryClient(), newDiscoveryDir(), nil)
	u := parseURL(t, "/r/alice/blog$search&q=blog&json")

	first := httptest.NewRecorder()
	h.Handle(context.Background(), first, httptest.NewRequest(http.MethodGet, "/r/alice/blog", nil), u)

	etag := first.Header().Get("ETag")
	if etag == "" {
		t.Fatal("no ETag on the first response")
	}

	req := httptest.NewRequest(http.MethodGet, "/r/alice/blog", nil)
	req.Header.Set("If-None-Match", etag)
	second := httptest.NewRecorder()
	status, _ := h.Handle(context.Background(), second, req, u)

	if status != http.StatusNotModified {
		t.Fatalf("status = %d, want 304", status)
	}
	if second.Body.Len() != 0 {
		t.Fatalf("body = %d bytes, want none on a 304", second.Body.Len())
	}
}

// A bearer token is only sent when one is configured; most indexers are public.
func TestIndexerTokenIsOptional(t *testing.T) {
	t.Parallel()

	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"latestBlockHeight":1}}`))
	}))
	defer srv.Close()

	if _, err := indexer.New(srv.URL, "").LatestBlockHeight(context.Background()); err != nil {
		t.Fatalf("no token: %v", err)
	}
	if _, err := indexer.New(srv.URL, "s3cret").LatestBlockHeight(context.Background()); err != nil {
		t.Fatalf("with token: %v", err)
	}

	if len(seen) != 2 || seen[0] != "" || seen[1] != "Bearer s3cret" {
		t.Fatalf("Authorization headers = %q, want [\"\", \"Bearer s3cret\"]", seen)
	}
}
