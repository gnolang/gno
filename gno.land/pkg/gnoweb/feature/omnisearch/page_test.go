package omnisearch

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
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

// A scope, typed as `in:` or taken from the page, narrows the discovery
// search: the page headed "Scoped to" lists nothing outside it.
func TestDiscoveryRespectsScope(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{
		"/$search&q=blog+in:/r/alice/blog",
		"/r/alice/blog$search&q=blog",
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
