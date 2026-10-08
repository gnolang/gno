package gnoweb_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/stretchr/testify/require"
)

func newSearchPageHandler(t *testing.T) *gnoweb.HTTPHandler {
	t.Helper()

	config := newTestHandlerConfig(t, gnoweb.NewMockClient(&gnoweb.MockPackage{
		Domain: "example.com",
		Path:   "/r/mock/path",
		Files: map[string]string{
			"render.gno": `package main; func Render(path string) string { return "body" }`,
		},
	}))
	logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{}))
	handler, err := gnoweb.NewHTTPHandler(logger, config)
	require.NoError(t, err)

	return handler
}

// A result page is an unbounded URL space where each URL costs a path
// listing. It must not invite crawlers, and it must be briefly cacheable —
// while ordinary realm pages keep their indexable default.
func TestSearchPageIsNotCrawlableAndIsCacheable(t *testing.T) {
	t.Parallel()

	handler := newSearchPageHandler(t)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/r/mock/path$search&q=mock", nil))

	require.Contains(t, rr.Body.String(), `content="noindex, nofollow"`,
		"the search result space must not be crawled")
	require.Equal(t, "max-age=5", rr.Header().Get("Cache-Control"))

	plain := httptest.NewRecorder()
	handler.ServeHTTP(plain, httptest.NewRequest(http.MethodGet, "/r/mock/path", nil))
	require.Contains(t, plain.Body.String(), `content="index, follow"`,
		"noindex must be scoped to search pages, not leak onto realm pages")
}

// The omnibar is prefilled with the current path, so a path carrying realm
// arguments reaches the parser. A naive "token with a colon is a qualifier"
// rule turned pressing Enter into a failed search instead of a navigation.
func TestPathWithArgsIsSearchedAsTextNotRejected(t *testing.T) {
	t.Parallel()

	handler := newSearchPageHandler(t)
	target := "/r/mock/path$search&q=" + url.QueryEscape("/r/gnoland/pages:p/about")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, target, nil))

	require.NotContains(t, rr.Body.String(), "is not a search qualifier",
		"a gno path with arguments must be searched as text, not rejected as a bad qualifier")
}

// The no-JavaScript form submits `?q=`, which is all a GET form can produce.
// gnoweb's own links carry `$search&q=`. Both must reach the same handler.
func TestSearchPageAcceptsAPlainFormSubmission(t *testing.T) {
	t.Parallel()

	handler := newSearchPageHandler(t)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/r/mock/path$search?q=mock", nil))

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), `value="mock"`,
		"the query must round-trip into the form, or the no-JS path silently searched for nothing")
}

// Chain-wide results name no package, so the header must not offer the
// realm tabs: State, Source and Actions for "/" are all error pages.
func TestChainWideSearchHasNoPackageTabs(t *testing.T) {
	t.Parallel()

	handler := newSearchPageHandler(t)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/$search?q=blog", nil))
	require.Equal(t, http.StatusOK, rr.Code)
	for _, tab := range []string{"$state", "$source", "$help"} {
		require.NotContains(t, rr.Body.String(), tab,
			"chain-wide results must not link a %s tab", tab)
	}

	realm := httptest.NewRecorder()
	handler.ServeHTTP(realm, httptest.NewRequest(http.MethodGet, "/r/mock/path$search?q=blog", nil))
	require.Contains(t, realm.Body.String(), "$source",
		"a realm-scoped search keeps the realm tabs")
}

var headerAction = regexp.MustCompile(`id="header-searchbar"[^>]*?\saction="([^"]*)"`)

// On an alias page the header form must search what the JavaScript omnibar
// searches, from the path the reader asked for: on "/" the whole chain, not
// the realm the alias renders.
func TestAliasHeaderSearchScope(t *testing.T) {
	cfg := newTestHandlerConfig(t, gnoweb.NewMockClient(&gnoweb.MockPackage{
		Domain: "example.com",
		Path:   "/r/mock/path",
		Files: map[string]string{
			"render.gno": `package main; func Render(path string) string { return "body" }`,
		},
	}))
	cfg.Aliases = map[string]gnoweb.AliasTarget{
		"/":      {Value: "/r/mock/path", Kind: gnoweb.GnowebPath},
		"/about": {Value: "/r/mock/path:p/about", Kind: gnoweb.GnowebPath},
	}
	logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{}))
	h, err := gnoweb.NewHTTPHandler(logger, cfg)
	if err != nil {
		t.Fatal(err)
	}
	get := func(target string) string {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, target, nil))
		return rr.Body.String()
	}
	for _, tc := range []struct{ page, jsBase string }{
		{"/", "/"},
		{"/about", "/about"},
	} {
		m := headerAction.FindStringSubmatch(get(tc.page))
		if m == nil {
			t.Fatalf("%s: no header form action", tc.page)
		}
		noJS := m[1]
		js := tc.jsBase + "$search"
		noJSScoped := strings.Contains(get(noJS+"?q=blog"), "Scoped to")
		jsScoped := strings.Contains(get(js+"&q=blog"), "Scoped to")
		t.Logf("page %-7s no-JS action %-22q scoped=%v | JS href %-16q scoped=%v", tc.page, noJS, noJSScoped, js, jsScoped)
		if noJS != js || noJSScoped != jsScoped {
			t.Errorf("page %s: no-JS form searches %q (scoped=%v), JS searches %q (scoped=%v)", tc.page, noJS, noJSScoped, js, jsScoped)
		}
	}
}
