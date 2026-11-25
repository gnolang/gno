package gnoweb

import (
	"context"
	"encoding/xml"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerRobotsTXT(t *testing.T) {
	t.Parallel()

	get := func(noindex bool, origin string) *httptest.ResponseRecorder {
		rr := httptest.NewRecorder()
		handlerRobotsTXT(noindex, origin).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))
		return rr
	}
	const disallow = "User-agent: *\nDisallow: /search.json\nDisallow: /status.json\nDisallow: /ready\n"

	t.Run("indexable with a sitemap", func(t *testing.T) {
		t.Parallel()
		rr := get(false, "https://gno.land")
		require.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "text/plain; charset=utf-8", rr.Header().Get("Content-Type"))
		assert.Equal(t, "public, max-age=3600", rr.Header().Get("Cache-Control"))
		assert.Equal(t, disallow+"\nSitemap: https://gno.land/sitemap.xml\n", rr.Body.String())
	})

	t.Run("indexable without a canonical origin", func(t *testing.T) {
		t.Parallel()
		rr := get(false, "")
		require.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, disallow, rr.Body.String())
	})

	t.Run("noindex keeps pages crawlable", func(t *testing.T) {
		t.Parallel()
		// Only the node-backed routes are blocked: a Disallow on a page would
		// hide its noindex header from crawlers.
		rr := get(true, "")
		require.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "# Not indexed: every response carries X-Robots-Tag: noindex, nofollow.\n"+disallow, rr.Body.String())
	})

	t.Run("HEAD and other methods", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		handlerRobotsTXT(false, "").ServeHTTP(rr, httptest.NewRequest(http.MethodHead, "/robots.txt", nil))
		assert.Equal(t, http.StatusOK, rr.Code)
		rr = httptest.NewRecorder()
		handlerRobotsTXT(false, "").ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/robots.txt", nil))
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
		assert.Equal(t, "GET, HEAD", rr.Header().Get("Allow"))
	})
}

// sitemapClient records the package and namespace queries the sitemap makes,
// fails the one at fail with err, and leaves the one at hang unanswered until
// the request gives up. An empty namespace lists as one blank line, as the
// node answers.
type sitemapClient struct {
	ClientAdapter
	fail, hang string
	err        error
	calls      []string
	log        strings.Builder
}

func (c *sitemapClient) answer(ctx context.Context, path string) error {
	c.calls = append(c.calls, path)
	switch path {
	case c.fail:
		return c.err
	case c.hang:
		<-ctx.Done()
		return ctx.Err()
	}
	return nil
}

func (c *sitemapClient) ListFiles(ctx context.Context, path string, height int64) ([]string, error) {
	if err := c.answer(ctx, path); err != nil {
		return nil, err
	}
	return c.ClientAdapter.ListFiles(ctx, path, height)
}

func (c *sitemapClient) ListPaths(ctx context.Context, prefix string, limit int) ([]string, error) {
	if err := c.answer(ctx, prefix); err != nil {
		return nil, err
	}
	paths, err := c.ClientAdapter.ListPaths(ctx, prefix, limit)
	if err == nil && len(paths) == 0 {
		paths = []string{""}
	}
	return paths, err
}

func TestHandlerSitemapXML(t *testing.T) {
	t.Parallel()

	const addrRealm = "/r/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/app"
	aliases := map[string]AliasTarget{
		"/":              {Value: "/r/gnoland/home", Kind: GnowebPath},          // a whole realm
		"/home":          {Value: "/r/gnoland/home", Kind: GnowebPath},          // same target, a page of its own name
		"/news:latest":   {Value: "/r/gnoland/home:news", Kind: GnowebPath},     // a key with args names itself with them
		"/about":         {Value: "/r/gnoland/pages:p/about", Kind: GnowebPath}, // one page of a realm
		"/start":         {Value: "/r/gnoland/pages:p/start", Kind: GnowebPath}, // same realm, queried once
		"/events":        {Value: "/r/devrels/events", Kind: GnowebPath},        // parked: the node answers not found
		"/lib":           {Value: "/p/gnoland/lib", Kind: GnowebPath},           // a package
		"/oldlib":        {Value: "/p/gnoland/gone", Kind: GnowebPath},          // a package not on this chain
		"/realms":        {Value: "/r/", Kind: GnowebPath},                      // a listing: served by no alias, never queried
		"/docs":          {Value: "/u/docs", Kind: GnowebPath},                  // a user page whose namespace holds a package
		"/nobody":        {Value: "/u/nobody", Kind: GnowebPath},                // a user page with nothing to show
		"/nym":           {Value: "/r/nym/app", Kind: GnowebPath},               // a community realm, indexed under a registered name
		"/solo":          {Value: "/r/solo", Kind: GnowebPath},                  // a package at a namespace root
		"/ns":            {Value: "/r/nym", Kind: GnowebPath},                   // a namespace listing: no package there
		"/addr":          {Value: addrRealm, Kind: GnowebPath},                  // an address namespace: noindex, never queried
		"/src":           {Value: "/r/gnoland/home$source", Kind: GnowebPath},   // an indexed view, named by its key
		"/help":          {Value: "/r/gnoland/home$help", Kind: GnowebPath},     // a noindex view
		"/Hello.md":      {Value: "/r/gnoland/home", Kind: GnowebPath},          // a file-like key names itself
		"/terms":         {Value: "# Terms", Kind: StaticMarkdown},              // operator page
		"/license.md":    {Value: "# License", Kind: StaticMarkdown},            // a static key may read as a file
		"/foo bar":       {Value: "# Foo", Kind: StaticMarkdown},                // gnoweb cannot parse it: a 404
		"/ready":         {Value: "# Ready", Kind: StaticMarkdown},              // shadowed by the router's own route
		"relative":       {Value: "# Relative", Kind: StaticMarkdown},           // no leading slash: never requested
		"/blog":          {Value: "# Blog", Kind: StaticMarkdown},               // redirected before aliases
		"/r/demo/boards": {Value: "# Boards", Kind: StaticMarkdown},             // operator override of a realm URL
		"/a@b":           {Value: "# At", Kind: StaticMarkdown},                 // gnoweb cannot parse it: a 404
		"/x$source":      {Value: "# X", Kind: StaticMarkdown},                  // names x$source$source, not itself
		"/a$b":           {Value: "# Ab", Kind: StaticMarkdown},                 // a $ view: noindex
	}
	newClient := func() *sitemapClient {
		return &sitemapClient{ClientAdapter: NewMockClient(
			&MockPackage{Path: "/r/gnoland/home"},
			&MockPackage{Path: "/r/gnoland/pages"},
			&MockPackage{Path: "/p/gnoland/lib"},
			&MockPackage{Path: "/r/docs/home"},
			&MockPackage{Path: "/r/nym/app"},
			&MockPackage{Path: "/r/solo"},
			&MockPackage{Path: addrRealm},
			&MockPackage{Path: "/r/devrels/events", Inert: true},
		)}
	}
	serve := func(method string, c *sitemapClient) *httptest.ResponseRecorder {
		h := &HTTPHandler{
			Client:  c,
			Static:  StaticMetadata{CanonicalOrigin: "https://gno.land"},
			Aliases: aliases,
			Logger:  slog.New(slog.NewTextHandler(&c.log, nil)),
			policy:  newPagePolicy([]string{"gnoland", "docs"}, aliases, IndexRegisteredCommunity, false),
		}
		if c.hang != "" {
			h.Timeout = 100 * time.Millisecond
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(method, "/sitemap.xml", nil)
		req.Host = "evil.example"
		req.Header.Set("X-Forwarded-Host", "evil.example")
		handlerSitemapXML(h, func(p string) bool { return p != "/ready" }).ServeHTTP(rr, req)
		return rr
	}
	locs := func(t *testing.T, rr *httptest.ResponseRecorder) []string {
		t.Helper()
		var set sitemapURLSet
		require.NoError(t, xml.Unmarshal(rr.Body.Bytes(), &set))
		var locs []string
		for _, u := range set.URLs {
			locs = append(locs, u.Loc)
		}
		return locs
	}

	t.Run("lists the indexed alias pages under their canonical", func(t *testing.T) {
		t.Parallel()
		c := newClient()
		rr := serve(http.MethodGet, c)

		require.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/xml; charset=utf-8", rr.Header().Get("Content-Type"))
		assert.Equal(t, "public, max-age=3600", rr.Header().Get("Cache-Control"))
		body := rr.Body.String()
		assert.True(t, strings.HasPrefix(body, xml.Header+`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`), body)
		assert.NotContains(t, body, "evil.example")
		assert.Equal(t, []string{
			"https://gno.land/",
			"https://gno.land/Hello.md",
			"https://gno.land/about",
			"https://gno.land/docs",
			"https://gno.land/home",
			"https://gno.land/lib",
			"https://gno.land/license.md",
			"https://gno.land/news:latest",
			"https://gno.land/nym",
			"https://gno.land/r/demo/boards",
			"https://gno.land/solo",
			"https://gno.land/src",
			"https://gno.land/start",
			"https://gno.land/terms",
		}, locs(t, rr))
		// A missing target is no failure.
		assert.Empty(t, c.log.String())
	})

	t.Run("method not allowed", func(t *testing.T) {
		t.Parallel()
		rr := serve(http.MethodPost, newClient())
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
		assert.Equal(t, "GET, HEAD", rr.Header().Get("Allow"))
	})

	t.Run("HEAD queries no node", func(t *testing.T) {
		t.Parallel()
		c := newClient()
		rr := serve(http.MethodHead, c)
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/xml; charset=utf-8", rr.Header().Get("Content-Type"))
		assert.Empty(t, rr.Header().Get("Cache-Control"), "a GET may be no-store")
		assert.Empty(t, c.calls)
	})

	t.Run("queries each indexed target once, never the full listing", func(t *testing.T) {
		t.Parallel()
		// The full listing is capped at 1000 paths, so it cannot tell an
		// absent realm from one past the cap. A noindex target is not queried.
		c := newClient()
		require.Equal(t, http.StatusOK, serve(http.MethodGet, c).Code)
		assert.ElementsMatch(t, []string{
			"/r/gnoland/home", "/r/gnoland/pages", "/r/devrels/events", "/p/gnoland/lib", "/p/gnoland/gone",
			"/r/nym/app", "/r/solo", "/r/nym", "@docs", "@nobody",
		}, c.calls)
	})

	t.Run("a failed check drops its aliases, not the sitemap", func(t *testing.T) {
		t.Parallel()
		for _, tc := range []struct {
			fail    string
			err     error
			dropped []string
			failed  bool // a node failure, logged and not cached
		}{
			{"/r/gnoland/pages", errors.New("rpc down"), []string{"/about", "/start"}, true},
			{"@docs", errors.New("rpc down"), []string{"/docs"}, true},
			// A target with a file-like last segment is a bad alias.
			{"/r/gnoland/pages", ErrClientFileNotFound, []string{"/about", "/start"}, false},
		} {
			c := newClient()
			c.fail, c.err = tc.fail, tc.err
			rr := serve(http.MethodGet, c)
			require.Equal(t, http.StatusOK, rr.Code, tc.fail)
			got := locs(t, rr)
			assert.Contains(t, got, "https://gno.land/", tc.fail)
			for _, p := range tc.dropped {
				assert.NotContains(t, got, "https://gno.land"+p, tc.fail)
			}
			if tc.failed {
				assert.Equal(t, "no-store", rr.Header().Get("Cache-Control"), tc.fail)
				assert.Equal(t, 1, strings.Count(c.log.String(), "unable to check alias targets"), tc.fail)
			} else {
				assert.Equal(t, "public, max-age=3600", rr.Header().Get("Cache-Control"), tc.fail)
				assert.Empty(t, c.log.String(), tc.fail)
			}
		}
	})

	t.Run("a node that never answers costs one request timeout", func(t *testing.T) {
		t.Parallel()
		c := newClient()
		c.hang = "/r/gnoland/pages"
		done := make(chan *httptest.ResponseRecorder, 1)
		go func() { done <- serve(http.MethodGet, c) }()
		select {
		case rr := <-done:
			require.Equal(t, http.StatusOK, rr.Code)
			assert.NotContains(t, locs(t, rr), "https://gno.land/about")
		case <-time.After(5 * time.Second):
			t.Fatal("the sitemap waited on the node past its request timeout")
		}
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
		// A static page renders without a node. /ready is the router's own.
		cfg.Aliases = map[string]AliasTarget{
			"/static": NewStaticAlias("# Static\n\nBody."),
			"/ready":  NewStaticAlias("# Ready"),
			// mux redirects it to / before any handler.
			"/static/..": NewStaticAlias("# Unclean"),
		}
		return NewRouter(log.NewTestingLogger(t), cfg)
	}
	get := func(h http.Handler, target string) *httptest.ResponseRecorder {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, target, nil))
		return rr
	}

	t.Run("the default stays indexable, with or without an origin", func(t *testing.T) {
		t.Parallel()
		// Forgetting every flag must never take a deployment out of search.
		for _, origin := range []string{"", "https://gno.land"} {
			router, err := newRouter(t, origin, false)
			require.NoError(t, err)
			for _, target := range []string{"/liveness", "/robots.txt"} {
				assert.Empty(t, get(router, target).Result().Header.Get("X-Robots-Tag"), "%s with origin %q", target, origin)
			}
			if origin == "" {
				assert.Equal(t, http.StatusNotFound, get(router, "/sitemap.xml").Code)
			}
			assert.Contains(t, get(router, "/static").Body.String(), `<meta name="robots" content="index, follow" />`)
		}
	})

	t.Run("an origin adds the sitemap, trailing slash trimmed", func(t *testing.T) {
		t.Parallel()
		router, err := newRouter(t, "https://gno.land/", false)
		require.NoError(t, err)
		assert.Contains(t, get(router, "/robots.txt").Body.String(), "Sitemap: https://gno.land/sitemap.xml\n")
	})

	t.Run("the sitemap leaves out an alias a route shadows", func(t *testing.T) {
		t.Parallel()
		router, err := newRouter(t, "https://gno.land", false)
		require.NoError(t, err)
		rr := get(router, "/sitemap.xml")
		require.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "<loc>https://gno.land/static</loc>")
		assert.NotContains(t, rr.Body.String(), "/ready")
		assert.NotContains(t, rr.Body.String(), "/static/..")
	})

	t.Run("noindex marks every response and drops the sitemap", func(t *testing.T) {
		t.Parallel()
		router, err := newRouter(t, "https://gno.land", true)
		require.NoError(t, err)
		for _, target := range []string{"/liveness", "/robots.txt", "/sitemap.xml"} {
			assert.Equal(t, "noindex, nofollow", get(router, target).Result().Header.Get("X-Robots-Tag"), target)
		}
		assert.NotContains(t, get(router, "/robots.txt").Body.String(), "Sitemap")
		assert.Equal(t, http.StatusNotFound, get(router, "/sitemap.xml").Code)
		static := get(router, "/static").Body.String()
		assert.Contains(t, static, `<meta name="robots" content="noindex, nofollow" />`)
		assert.NotContains(t, static, `rel="canonical"`)
	})

	t.Run("an invalid canonical origin stops startup", func(t *testing.T) {
		t.Parallel()
		_, err := newRouter(t, "https://gno.land/r/demo", false)
		require.Error(t, err)
	})
}
