package omnisearch

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/doc"
)

var tagPattern = regexp.MustCompile(`<[^>]+>`)

// renderText renders the results page and reduces it to its visible text.
func renderText(t *testing.T, data SearchData) string {
	t.Helper()
	var b strings.Builder
	if err := NewPageView(data).Render(&b); err != nil {
		t.Fatalf("render: %v", err)
	}
	return strings.Join(strings.Fields(tagPattern.ReplaceAllString(b.String(), " ")), " ")
}

// handlePage serves a results-page URL and returns its HTML.
func handlePage(t *testing.T, h *Handler, raw string) string {
	t.Helper()
	_, view := h.Handle(context.Background(), httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil), parseURL(t, raw))
	if view == nil {
		t.Fatalf("%s: no view", raw)
	}
	var b strings.Builder
	if err := view.Render(&b); err != nil {
		t.Fatalf("render: %v", err)
	}
	return b.String()
}

// An `in:` scope narrows the discovery search: the page headed "Scoped to"
// lists nothing outside it.
func TestDiscoveryRespectsScope(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{
		"/$search&q=blog+in:/r/alice/blog",
		"/r/bob/blog$search&q=blog+in:/r/alice/blog",
	} {
		h := newHandlerWithDir(t, newDiscoveryClient(), newDiscoveryDir(), nil)
		html := handlePage(t, h, raw)
		if !strings.Contains(html, `href="/r/alice/blog"`) {
			t.Errorf("%s: the scoped realm is missing", raw)
		}
		if strings.Contains(html, `href="/r/bob/blog"`) {
			t.Errorf("%s: lists /r/bob/blog, outside the scope", raw)
		}
		if !strings.Contains(html, `href="/$search?q=blog"`) {
			t.Errorf("%s: no link to the same search over the whole chain", raw)
		}
		// .header-info already spaces its items; a separator would sit
		// apart from both.
		if strings.Contains(html, "· <a") {
			t.Errorf("%s: the whole-chain link carries a separator", raw)
		}
	}
}

// The omnibar sends every query from the page path, so the page a discovery
// search was typed on is not a scope: `author:` and a bare word on a realm
// page search the whole chain, and the header says so.
func TestDiscoveryIgnoresThePageScope(t *testing.T) {
	t.Parallel()

	for raw, want := range map[string][]string{
		"/r/alice/blog$search&q=author:bob":    {"/r/bob/blog"},
		"/r/alice/blog$search&q=blog":          {"/r/alice/blog", "/r/bob/blog"},
		"/p/alice/util$search&q=is:realm+blog": {"/r/alice/blog", "/r/bob/blog"},
		"/r/alice/blog$search":                 nil,
	} {
		h := newHandlerWithDir(t, newDiscoveryClient(), newDiscoveryDir(), nil)
		html := handlePage(t, h, raw)
		for _, p := range want {
			if !strings.Contains(html, `href="`+p+`"`) {
				t.Errorf("%s: %s is missing", raw, p)
			}
		}
		text := strings.Join(strings.Fields(tagPattern.ReplaceAllString(html, " ")), " ")
		if !strings.Contains(text, "Whole chain") || strings.Contains(text, "Scoped to") {
			t.Errorf("%s: header does not say Whole chain:\n%s", raw, text)
		}
		if strings.Contains(html, "Search the whole chain") {
			t.Errorf("%s: offers to widen a search that is already chain-wide", raw)
		}
	}

	// The omnibar's JSON path answers the same, and does not echo the page
	// as the scope.
	h := newHandlerWithDir(t, newDiscoveryClient(), newDiscoveryDir(), nil)
	w := httptest.NewRecorder()
	h.Handle(context.Background(), w, httptest.NewRequest(http.MethodGet, "/", nil),
		parseURL(t, "/r/alice/blog$search&q=author:bob&json"))
	if body := w.Body.String(); !strings.Contains(body, `"/r/bob/blog"`) || strings.Contains(body, `"pkg_path"`) {
		t.Errorf("json: %s", body)
	}
}

// A sub-path sits inside its parent's scope; a longer sibling name does not.
func TestInScope(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		rel, scope string
		want       bool
	}{
		{"/r/alice/blog", "", true},
		{"/r/alice/blog", "/r/alice/blog", true},
		{"/r/alice/blog/v2", "/r/alice/blog", true},
		{"/r/alice/blogger", "/r/alice/blog", false},
		{"/r/bob/blog", "/r/alice/blog", false},
	} {
		if got := inScope(tc.rel, tc.scope); got != tc.want {
			t.Errorf("inScope(%q, %q) = %v, want %v", tc.rel, tc.scope, got, tc.want)
		}
	}
}

// A query discovery will not run says why, rather than "Nothing matched."
// while matching paths exist.
func TestDiscoveryExplainsWhatItWillNotRun(t *testing.T) {
	t.Parallel()

	for raw, want := range map[string]string{
		"a":            "at least 2 characters",
		"is:pkg util":  "is:pkg is not a kind",
		"author:b":     "at least 2 characters",
		"is:PACKAGE u": "",
	} {
		h := newHandlerWithDir(t, newDiscoveryClient(), newDiscoveryDir(), nil)
		out := renderText(t, h.build(context.Background(), mustQuery(t, h, raw, "")))
		if want == "" {
			continue
		}
		if !strings.Contains(out, want) {
			t.Errorf("%q: page does not say %q:\n%s", raw, want, out)
		}
		if strings.Contains(out, "Nothing matched.") {
			t.Errorf("%q: page says Nothing matched.", raw)
		}
	}
}

// A backend that could not answer is not "Nothing matched.".
func TestFailedGroupIsNotNothingMatched(t *testing.T) {
	t.Parallel()

	dir := newDiscoveryDir()
	dir.err = errors.New("dial tcp: connection refused")
	h := newHandlerWithDir(t, newDiscoveryClient(), dir, nil)
	out := renderText(t, h.build(context.Background(), mustQuery(t, h, "blog", "")))
	if !strings.Contains(out, "Could not answer") {
		t.Fatalf("no failure shown:\n%s", out)
	}
	if strings.Contains(out, "Nothing matched.") {
		t.Errorf("page says both Could not answer and Nothing matched.:\n%s", out)
	}
}

// `imports:json` narrows the imports rather than dropping its value.
func TestImportsValueNarrows(t *testing.T) {
	t.Parallel()

	c := &mockClient{doc: &doc.JSONDocumentation{Imports: []string{"encoding/json", "strings", "gno.land/p/nt/avl/v0"}}}
	h := newHandler(t, c, nil)
	groups, _ := h.Search(context.Background(), mustQuery(t, h, "imports:json", "/r/demo/boards"))
	if len(groups) != 1 || len(groups[0].Results) != 1 || groups[0].Results[0].Title != "encoding/json" {
		t.Fatalf("groups = %+v, want encoding/json only", groups)
	}
	groups, _ = h.Search(context.Background(), mustQuery(t, h, "imports", "/r/demo/boards"))
	if len(groups) != 1 || len(groups[0].Results) != 3 {
		t.Fatalf("groups = %+v, want every import", groups)
	}
}

// The provenance footer describes an answer the indexer gave. A group that
// failed, before or while asking, has none to describe.
func TestIndexerFooterOnlyOverAnIndexerAnswer(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, raw, scope string
		idx              *mockIndexer
		want             bool
	}{
		{"pre-flight error", "content:abc", "", &mockIndexer{}, false},
		{"needs a package", "activity", "", &mockIndexer{}, false},
		{"indexer down", "deploys", "/r/demo/boards", &mockIndexer{err: errors.New("dial tcp")}, false},
		{"answered with nothing", "deploys", "/r/demo/boards", &mockIndexer{}, true},
	} {
		h := newHandler(t, &mockClient{}, tc.idx)
		out := renderText(t, h.build(context.Background(), mustQuery(t, h, tc.raw, tc.scope)))
		if got := strings.Contains(out, "comes from an indexer"); got != tc.want {
			t.Errorf("%s: footer = %v, want %v:\n%s", tc.name, got, tc.want, out)
		}
	}
}

// An action is what the Actions page lists: an exported top-level function of
// a realm, crossing (a call) or not (a qeval query). Render has its own page.
func TestActionTagOnlyOnCallableFuncs(t *testing.T) {
	t.Parallel()

	jdoc := &doc.JSONDocumentation{Funcs: []*doc.JSONFunc{
		{Name: "helper", Signature: "func helper() int", File: "x.gno", Line: 3},
		{Name: "GetBoard", Signature: "func GetBoard(id int) string", File: "x.gno", Line: 7},
		{Name: "Render", Signature: "func Render(path string) string", File: "x.gno", Line: 9},
		{Name: "CreateBoard", Crossing: true, Signature: "func CreateBoard(cur realm, name string)", File: "x.gno", Line: 11},
		{Name: "cross", Crossing: true, Signature: "func cross(cur realm)", File: "x.gno", Line: 13},
	}}
	h := newHandler(t, &mockClient{doc: jdoc}, nil)
	for name, want := range map[string]bool{
		"helper": false, "GetBoard": true, "Render": false, "CreateBoard": true, "cross": false,
	} {
		groups, _ := h.Search(context.Background(), mustQuery(t, h, "func:"+name, "/r/demo/boards"))
		var r *Result
		for i := range groups[0].Results {
			if strings.HasPrefix(groups[0].Results[i].Title, "func "+name+"(") {
				r = &groups[0].Results[i]
			}
		}
		if r == nil {
			t.Fatalf("func:%s: no result", name)
		}
		if got := slices.Contains(r.Tags, "action"); got != want {
			t.Errorf("func:%s: action tag = %v, want %v", name, got, want)
		}
	}
}
