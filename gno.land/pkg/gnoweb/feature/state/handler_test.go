package state

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

type stubHighlighter struct{}

func (stubHighlighter) Render(string, []byte) (template.HTML, error) { return "", nil }

// stubClient is a minimal ClientAdapter for handler tests — every method
// returns zero values so the stub Handle path can be exercised without
// importing the gnoweb package (which would create a test-time cycle).
type stubClient struct{}

func (stubClient) Realm(context.Context, string, string) ([]byte, error)    { return nil, nil }
func (stubClient) ListPaths(context.Context, string, int) ([]string, error) { return nil, nil }
func (stubClient) Doc(context.Context, string, int64) (*doc.JSONDocumentation, error) {
	return nil, nil
}
func (stubClient) StatePkg(context.Context, string, int64) ([]byte, error)    { return nil, nil }
func (stubClient) StateObject(context.Context, string, int64) ([]byte, error) { return nil, nil }
func (stubClient) StateType(context.Context, string, int64) ([]byte, error)   { return nil, nil }

func newTestHandler() *Handler {
	return New(Deps{
		Client:      stubClient{},
		Highlighter: stubHighlighter{},
	})
}

// TestHandleDispatchesByQuery pins the URL → handler routing so
// regressions in the Handle dispatch surface here rather than as
// downstream test failures. All three endpoints are wired; we assert
// only that Handle returns non-zero status (no panic, no fallthrough).
func TestHandleDispatchesByQuery(t *testing.T) {
	h := newTestHandler()
	cases := []struct {
		name  string
		query url.Values
	}{
		{"json", url.Values{"state": {""}, "json": {""}}},
		{"frag", url.Values{"state": {""}, "frag": {"node"}, "oid": {"abcdef0123456789abcdef0123456789abcdef01:1"}}},
		{"page", url.Values{"state": {""}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			u := &weburl.GnoURL{Path: "/r/demo", WebQuery: c.query}
			req := httptest.NewRequest(http.MethodGet, "/r/demo$state", nil)
			rec := httptest.NewRecorder()
			status, _ := h.Handle(context.Background(), rec, req, u)
			if status == 0 {
				t.Fatalf("Handle returned zero status — dispatch fell through")
			}
		})
	}
}

// TestHandleEndToEndFragmentFromTemplateURL exercises the full
// template→ParseFromURL→Handle path for a generated `hx-get` URL.
// Regression coverage for the `?state` vs `$state` bug: if the template
// emits the wrong grammar, ParseFromURL puts `frag` into Query (not
// WebQuery), Handle.dispatch's `WebQuery.Has("frag")` misses, and the
// fragment falls through to the full-page renderer. The test goes via
// the same code path the browser triggers — string the generated href
// through ParseFromURL, then assert the response is a fragment (small
// HTML body, no `<!doctype>`), not a full page.
func TestHandleEndToEndFragmentFromTemplateURL(t *testing.T) {
	// Build the URL the way the template does (stateFragNodeHref).
	oid := "abcdef0123456789abcdef0123456789abcdef01:1"
	href := string(stateFragNodeHref("/r/demo", oid, "", 0, ""))
	// Round-trip through net/url and ParseFromURL like a real request would.
	stdURL, err := url.Parse(href)
	if err != nil {
		t.Fatalf("parse generated href %q: %v", href, err)
	}
	u, err := weburl.ParseFromURL(stdURL)
	if err != nil {
		t.Fatalf("ParseFromURL %q: %v", href, err)
	}
	// Sanity: the parser placed `frag` into WebQuery (not Query).
	if u.WebQuery.Get("frag") != "node" {
		t.Fatalf("frag not in WebQuery: WebQuery=%v Query=%v", u.WebQuery, u.Query)
	}

	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, href, nil)
	rec := httptest.NewRecorder()
	status, view := h.Handle(context.Background(), rec, req, u)
	if status != http.StatusOK {
		t.Fatalf("Handle status = %d, want 200", status)
	}
	if view != nil {
		t.Fatalf("Handle returned non-nil view — fragment path must write body directly and return nil")
	}
	body := rec.Body.String()
	// The dispatch must reach the fragment branch — fragment success
	// (`b-state-frag-node`) OR fragment error (`b-state-frag-error`) both
	// confirm that. The bug we're guarding against is full-page fallback,
	// which would surface `<!doctype>` and the SSR doc-index island.
	if strings.Contains(strings.ToLower(body), "<!doctype") {
		t.Fatalf("Handle returned a FULL PAGE for a fragment request — URL grammar regression")
	}
	if !strings.Contains(body, "b-state-frag-") {
		t.Errorf("body lacks any fragment wrapper — dispatch fell through to a non-fragment branch; got: %s", body)
	}
}

// The per-IP rate limiter is the load-bearing security layer that bounds
// RPC fan-out. Calling Handle from one IP past the burst must return 429
// (plain) or HTTP 200 + fragment-error (htmx) — never silently pass through.
func TestHandleEnforcesRateLimit(t *testing.T) {
	now := time.Unix(0, 0)
	h := New(Deps{
		Client:      stubClient{},
		Highlighter: stubHighlighter{},
		RateLimit: RateLimitConfig{
			PerMinute: 1,
			Burst:     1,
			NowFunc:   func() time.Time { return now },
		},
	})
	u := &weburl.GnoURL{Path: "/r/demo", WebQuery: url.Values{"state": {""}}}

	r1 := httptest.NewRequest(http.MethodGet, "/r/demo$state", nil)
	r1.RemoteAddr = "1.1.1.1:9999"
	rec1 := httptest.NewRecorder()
	if status, _ := h.Handle(context.Background(), rec1, r1, u); status == 0 {
		t.Fatalf("first request: zero status")
	}

	r2 := httptest.NewRequest(http.MethodGet, "/r/demo$state", nil)
	r2.RemoteAddr = "1.1.1.1:9999"
	rec2 := httptest.NewRecorder()
	status, view := h.Handle(context.Background(), rec2, r2, u)
	if status != http.StatusTooManyRequests {
		t.Fatalf("over-limit (plain): status = %d, want 429", status)
	}
	if view != nil {
		t.Fatalf("over-limit: expected nil view, got %v", view)
	}
	if rec2.Header().Get("Retry-After") != "60" {
		t.Fatalf("over-limit: missing Retry-After header")
	}

	r3 := httptest.NewRequest(http.MethodGet, "/r/demo$state", nil)
	r3.RemoteAddr = "1.1.1.1:9999"
	r3.Header.Set("HX-Request", "true")
	rec3 := httptest.NewRecorder()
	status, view = h.Handle(context.Background(), rec3, r3, u)
	if status != http.StatusOK {
		t.Fatalf("over-limit (htmx): status = %d, want 200 (fragment-error)", status)
	}
	if view != nil {
		t.Fatalf("over-limit (htmx): expected nil view")
	}
	if !strings.Contains(rec3.Body.String(), "b-state-frag-error") {
		t.Fatalf("over-limit (htmx): missing fragment-error marker; body=%q", rec3.Body.String())
	}
}
