package gnoweb

import (
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeCanonicalOrigin(t *testing.T) {
	t.Parallel()

	// Harmless spellings are normalized rather than refused: a refusal stops
	// gnoweb from starting.
	for in, want := range map[string]string{
		"":                      "",
		"https://gno.land":      "https://gno.land",
		"http://localhost:8888": "http://localhost:8888",
		"https://gno.land/":     "https://gno.land",
		"https://gno.land//":    "https://gno.land",
		" https://gno.land \n":  "https://gno.land",
		"HTTPS://GNO.land":      "https://gno.land",
		"https://gno.land:443":  "https://gno.land",
		"http://gno.land:80":    "http://gno.land",
		"https://gno.land:8443": "https://gno.land:8443",
		"http://gno.land:443":   "http://gno.land:443",
	} {
		got, err := normalizeCanonicalOrigin(in)
		if assert.NoError(t, err, "origin %q", in) {
			assert.Equal(t, want, got, "origin %q", in)
		}
	}

	// Anything that is not a bare http(s) origin would be copied into every
	// canonical tag and sitemap URL, so it still stops startup.
	for _, in := range []string{
		"gno.land",
		"ftp://gno.land",
		"https://",
		"https://gno.land/r/demo",
		"https://gno.land?x=1",
		"https://gno.land#top",
		"https://user@gno.land",
		"javascript:alert(1)",
		"https://gno.land\n/evil",
		"https://gno.land?",
		"https://gno.land#",
		"https:gno.land",
		"https://gno.land:",
	} {
		_, err := normalizeCanonicalOrigin(in)
		assert.Error(t, err, "origin %q", in)
	}
}

func TestHandlerRobotsTXT(t *testing.T) {
	t.Parallel()

	get := func(noindex bool, origin string) *httptest.ResponseRecorder {
		rr := httptest.NewRecorder()
		handlerRobotsTXT(noindex, origin).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))
		return rr
	}

	t.Run("indexable with a sitemap", func(t *testing.T) {
		t.Parallel()
		rr := get(false, "https://gno.land")
		require.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "text/plain; charset=utf-8", rr.Header().Get("Content-Type"))
		assert.Equal(t, "public, max-age=3600", rr.Header().Get("Cache-Control"))
		assert.Equal(t, "User-agent: *\nDisallow: /search.json\nDisallow: /status.json\n\nSitemap: https://gno.land/sitemap.xml\n", rr.Body.String())
	})

	t.Run("indexable without a canonical origin", func(t *testing.T) {
		t.Parallel()
		rr := get(false, "")
		require.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "User-agent: *\nDisallow: /search.json\nDisallow: /status.json\n", rr.Body.String())
	})

	t.Run("noindex still allows crawling", func(t *testing.T) {
		t.Parallel()
		// A Disallow would hide the noindex header from crawlers.
		rr := get(true, "")
		require.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "# Not indexed: every response carries X-Robots-Tag: noindex.\nUser-agent: *\nAllow: /\n", rr.Body.String())
	})

	t.Run("method not allowed", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		handlerRobotsTXT(false, "").ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/robots.txt", nil))
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestHandlerSitemapXML(t *testing.T) {
	t.Parallel()

	aliases := map[string]AliasTarget{
		"/":              {Value: "/r/gnoland/home", Kind: GnowebPath},          // a whole realm
		"/about":         {Value: "/r/gnoland/pages:p/about", Kind: GnowebPath}, // one page of a realm
		"/events":        {Value: "/r/devrels/events", Kind: GnowebPath},        // realm absent here
		"/docs":          {Value: "/u/docs", Kind: GnowebPath},                  // not a realm
		"/terms":         {Value: "# Terms", Kind: StaticMarkdown},              // operator page
		"/foo bar":       {Value: "# Foo", Kind: StaticMarkdown},                // not a clean path
		"relative":       {Value: "# Relative", Kind: StaticMarkdown},           // no leading slash
		"/src":           {Value: "/r/demo/boards$source", Kind: GnowebPath},    // a view of a realm
		"/blog":          {Value: "# Blog", Kind: StaticMarkdown},               // redirected before aliases
		"/terms.md":      {Value: "# Terms", Kind: StaticMarkdown},              // gnoweb rewrites the extension
		"/r/demo/boards": {Value: "# Boards", Kind: StaticMarkdown},             // operator override of a realm URL
	}
	dir := stubDirectory{
		realms:   []string{"/r/gnoland/home", "/r/gnoland/pages", "/r/demo/boards"},
		packages: []string{"/p/nt/avl/v0"},
	}
	serve := func(origin, method string, d RealmDirectory) *httptest.ResponseRecorder {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(method, "/sitemap.xml", nil)
		req.Host = "evil.example"
		req.Header.Set("X-Forwarded-Host", "evil.example")
		handlerSitemapXML(newDiscardLogger(), origin, aliases, d).ServeHTTP(rr, req)
		return rr
	}

	t.Run("lists the curated alias pages under the canonical origin", func(t *testing.T) {
		t.Parallel()
		rr := serve("https://gno.land", http.MethodGet, dir)

		require.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/xml; charset=utf-8", rr.Header().Get("Content-Type"))
		assert.Equal(t, "public, max-age=3600", rr.Header().Get("Cache-Control"))
		body := rr.Body.String()
		assert.True(t, strings.HasPrefix(body, xml.Header+`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`), body)
		assert.NotContains(t, body, "evil.example")

		var set sitemapURLSet
		require.NoError(t, xml.Unmarshal(rr.Body.Bytes(), &set))
		var locs []string
		for _, u := range set.URLs {
			locs = append(locs, u.Loc)
		}
		// Realms and packages are not listed, only operator entry points.
		assert.Equal(t, []string{
			"https://gno.land/",
			"https://gno.land/about",
			"https://gno.land/r/demo/boards",
			"https://gno.land/src",
			"https://gno.land/terms",
		}, locs)
	})

	t.Run("not found without a canonical origin", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, http.StatusNotFound, serve("", http.MethodGet, dir).Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, http.StatusMethodNotAllowed, serve("https://gno.land", http.MethodPost, dir).Code)
	})

	t.Run("upstream error", func(t *testing.T) {
		t.Parallel()
		rr := serve("https://gno.land", http.MethodGet, stubDirectory{err: errors.New("rpc down")})
		assert.Equal(t, http.StatusBadGateway, rr.Code)
	})
}

func TestNewRouter_CrawlPolicy(t *testing.T) {
	t.Parallel()

	newRouter := func(t *testing.T, origin string, noindex bool) (http.Handler, error) {
		t.Helper()
		cfg := NewDefaultAppConfig()
		cfg.NodeRemote = "127.0.0.1:123456" // no node needed for these routes
		cfg.ChainID = "test"
		cfg.CanonicalOrigin = origin
		cfg.NoIndex = noindex
		// A static page renders without a node, so the page meta is testable here.
		cfg.Aliases = map[string]AliasTarget{"/static": NewStaticAlias("# Static\n\nBody.")}
		return NewRouter(log.NewTestingLogger(t), cfg)
	}
	get := func(h http.Handler, target string) *httptest.ResponseRecorder {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, target, nil))
		return rr
	}

	t.Run("the default stays indexable, with or without an origin", func(t *testing.T) {
		t.Parallel()
		// The safe direction: forgetting every flag must never take a
		// deployment out of search.
		for _, origin := range []string{"", "https://gno.land"} {
			router, err := newRouter(t, origin, false)
			require.NoError(t, err)
			for _, target := range []string{"/liveness", "/robots.txt"} {
				assert.Empty(t, get(router, target).Header().Get("X-Robots-Tag"), "%s with origin %q", target, origin)
			}
			assert.NotContains(t, get(router, "/robots.txt").Body.String(), "Allow: /\n")
			assert.Contains(t, get(router, "/static").Body.String(), `<meta name="robots" content="index, follow" />`)
		}
	})

	t.Run("an origin adds the sitemap, trailing slash trimmed", func(t *testing.T) {
		t.Parallel()
		router, err := newRouter(t, "https://gno.land/", false)
		require.NoError(t, err)
		assert.Contains(t, get(router, "/robots.txt").Body.String(), "Sitemap: https://gno.land/sitemap.xml\n")
	})

	t.Run("noindex marks every response and drops the sitemap", func(t *testing.T) {
		t.Parallel()
		router, err := newRouter(t, "https://gno.land", true)
		require.NoError(t, err)
		for _, target := range []string{"/liveness", "/robots.txt", "/sitemap.xml"} {
			assert.Equal(t, "noindex, nofollow", get(router, target).Header().Get("X-Robots-Tag"), target)
		}
		assert.NotContains(t, get(router, "/robots.txt").Body.String(), "Sitemap")
		assert.Equal(t, http.StatusNotFound, get(router, "/sitemap.xml").Code)
		assert.Contains(t, get(router, "/static").Body.String(), `<meta name="robots" content="noindex, nofollow" />`)
		// No page may point a canonical at the shared origin, including the
		// views that return before the page handler clears it.
		for _, target := range []string{"/static", "/r/demo/boards$state"} {
			body := get(router, target).Body.String()
			assert.NotContains(t, body, `rel="canonical"`, target)
			assert.NotContains(t, body, `content="https://gno.land`, target)
		}
	})

	t.Run("an invalid canonical origin stops startup", func(t *testing.T) {
		t.Parallel()
		_, err := newRouter(t, "https://gno.land/r/demo", false)
		require.Error(t, err)
	})
}
