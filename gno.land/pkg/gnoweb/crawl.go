package gnoweb

import (
	"context"
	"encoding/xml"
	"errors"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// noIndexMiddleware marks every response noindex. robots.txt must still allow
// crawling, or crawlers never see it.
func noIndexMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Robots-Tag", noIndexNoFollow.String())
		next.ServeHTTP(w, r)
	})
}

// robotsDisallow keeps crawlers off the routes that query the node and are
// not pages, on every deployment. A page must never be listed: blocking it
// hides its noindex, and a URL already indexed then stays listed.
const robotsDisallow = "User-agent: *\nDisallow: /search.json\nDisallow: /status.json\nDisallow: /ready\n"

// crawlCacheControl lets a proxy hold robots.txt and the sitemap for an hour.
const crawlCacheControl = "public, max-age=3600"

// handlerRobotsTXT serves robots.txt; sitemapOrigin is empty without a sitemap.
func handlerRobotsTXT(noIndex bool, sitemapOrigin string) http.Handler {
	body := robotsDisallow
	if noIndex {
		body = "# Not indexed: every response carries X-Robots-Tag: noindex, nofollow.\n" + body
	}
	if sitemapOrigin != "" {
		body += "\nSitemap: " + sitemapOrigin + "/sitemap.xml\n"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			methodNotAllowed(w)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", crawlCacheControl)
		_, _ = w.Write([]byte(body))
	})
}

func methodNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Allow", "GET, HEAD")
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

type sitemapURL struct {
	Loc string `xml:"loc"`
}

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
	URLs    []sitemapURL `xml:"url"`
}

// handlerSitemapXML lists the alias pages only: most realms are demos or near
// empty, and crawlers reach the rest through links. served reports whether h
// answers a path, since a route of the router's own shadows an alias. The
// router mounts it only with a canonical origin and without -no-index.
func handlerSitemapXML(h *HTTPHandler, served func(path string) bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			methodNotAllowed(w)
			return
		}
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		if r.Method == http.MethodHead {
			return // the headers alone are not worth a node query
		}

		ctx, cancel := context.WithTimeout(r.Context(), h.requestTimeout())
		defer cancel()
		locs, complete := h.sitemapLocs(ctx, served)
		// A list cut short by a node failure must not be held for an hour.
		if complete {
			w.Header().Set("Cache-Control", crawlCacheControl)
		} else {
			w.Header().Set("Cache-Control", "no-store")
		}
		var set sitemapURLSet
		for _, loc := range locs {
			set.URLs = append(set.URLs, sitemapURL{Loc: loc})
		}
		_, _ = w.Write([]byte(xml.Header))
		if err := xml.NewEncoder(w).Encode(set); err != nil {
			h.Logger.Error("sitemap: unable to encode", "error", err)
		}
	})
}

// sitemapLocs returns the sorted canonical URLs of the alias pages the robots
// policy indexes and that name themselves as canonical, and whether every
// target could be checked. A realm alias is listed only if its target has
// something to show; a page a realm renders as missing still is.
// Classification and naming follow Get: keep them in step.
func (h *HTTPHandler) sitemapLocs(ctx context.Context, served func(string) bool) ([]string, bool) {
	var locs, failed []string
	exists := map[string]bool{} // by target path: aliases share realms
	for key, target := range h.Aliases {
		if _, redirected := Redirects[key]; redirected || !strings.HasPrefix(key, "/") || !served(key) {
			continue
		}
		// As in Get: a key gnoweb cannot parse is a 404; a static page is
		// classified and named by its key, a realm alias classified by its
		// target and named by its key.
		named, err := weburl.ParseFromURL(&url.URL{Path: key})
		if err != nil {
			continue
		}
		page := named
		if target.Kind == StaticMarkdown {
			page = &weburl.GnoURL{Path: key, WebQuery: named.WebQuery}
			named = page
		} else if page, err = weburl.ParseFromURL(&url.URL{Path: target.Value}); err != nil {
			continue
		} else if named.IsFile() { // as in Get: a key such as /x.md is no file under "/"
			named.Path, named.File = path.Join(named.Path, named.File), ""
		}
		// Only a page whose canonical is its own URL is listed.
		loc := h.canonicalURL(named)
		if loc != h.Static.CanonicalOrigin+(&url.URL{Path: key}).EscapedPath() ||
			h.policy.robots(h.policy.kind(page), page) != indexFollow {
			continue
		}
		if target.Kind == GnowebPath {
			ok, checked := exists[page.Path]
			if !checked {
				ok, err = h.sitemapTargetExists(ctx, page)
				if err != nil {
					failed = append(failed, page.Path)
				}
				exists[page.Path] = ok
			}
			if !ok {
				continue
			}
		}
		locs = append(locs, loc)
	}
	if len(failed) > 0 {
		// The client logs each failed query; one line here says what the
		// sitemap left out.
		h.Logger.Warn("sitemap: unable to check alias targets", "targets", failed)
	}
	slices.Sort(locs)
	return locs, len(failed) == 0
}

// sitemapTargetExists reports whether a realm alias's target has something to
// show: its package exists, or for a user page, its namespace holds one (the
// page also lists the address half of a registered name; a name whose
// packages all sit there is left out). Each is one query: the full listing is
// capped at 1000 entries, and r/gnoland/* sorts after every r/g1... user
// realm. A missing target is no error.
func (h *HTTPHandler) sitemapTargetExists(ctx context.Context, u *weburl.GnoURL) (bool, error) {
	switch _, isPackage := packagePath(u); {
	case !isPackage:
		// A listing such as /r/, or another alias, is left out.
		return false, nil
	case u.IsUser():
		paths, err := h.Client.ListPaths(ctx, "@"+u.Username(), 1)
		// An empty listing is a single blank line.
		return slices.ContainsFunc(paths, func(p string) bool { return p != "" }), err
	default:
		_, err := h.Client.ListFiles(ctx, u.Path, 0)
		// A target with a file-like last segment is a bad alias, not an outage.
		if errors.Is(err, ErrClientPackageNotFound) || errors.Is(err, ErrClientFileNotFound) {
			return false, nil
		}
		return err == nil, err
	}
}
