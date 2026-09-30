package gnoweb

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// Every deployment is indexable unless its operator passes -noindex. Testnets,
// staging and previews serve the same realms under the gno.land name, and
// search engines list them next to, or instead of, the real network; the
// opt-out keeps them out without ever risking the real one.

// normalizeCanonicalOrigin returns origin in the one spelling gnoweb copies
// into canonical tags, robots.txt and sitemap URLs: lowercase, no default port,
// no trailing slash. Harmless variants are fixed rather than refused, since a
// refusal stops gnoweb; anything that is not a bare http(s) origin (a path,
// query, credentials) is still an error.
func normalizeCanonicalOrigin(origin string) (string, error) {
	origin = strings.TrimRight(strings.TrimSpace(origin), "/")
	if origin == "" {
		return "", nil
	}
	u, err := url.Parse(origin)
	if err != nil {
		return "", fmt.Errorf("invalid canonical origin %q: %w", origin, err)
	}
	scheme, host := strings.ToLower(u.Scheme), strings.ToLower(u.Host)
	host = strings.TrimSuffix(host, map[string]string{"https": ":443", "http": ":80"}[scheme])
	// Rebuilding the URL from scheme and host alone catches everything else
	// at once, including an empty trailing "?" or "#" that Parse accepts.
	if (scheme != "http" && scheme != "https") || host == "" || strings.HasSuffix(host, ":") ||
		!strings.EqualFold((&url.URL{Scheme: u.Scheme, Host: u.Host}).String(), origin) {
		return "", fmt.Errorf("invalid canonical origin %q: want scheme://host[:port]", origin)
	}
	return scheme + "://" + host, nil
}

// noIndexMiddleware marks every response as not indexable. robots.txt still
// allows crawling: a crawler that may not fetch a page never sees its noindex,
// and keeps whatever it already indexed.
func noIndexMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Robots-Tag", "noindex, nofollow")
		next.ServeHTTP(w, r)
	})
}

// handlerRobotsTXT serves robots.txt. sitemapOrigin is the canonical origin
// when a sitemap is served, empty otherwise.
func handlerRobotsTXT(noindex bool, sitemapOrigin string) http.Handler {
	body := "User-agent: *\nDisallow: /search.json\nDisallow: /status.json\n"
	switch {
	case noindex:
		body = "# Not indexed: every response carries X-Robots-Tag: noindex.\nUser-agent: *\nAllow: /\n"
	case sitemapOrigin != "":
		body += "\nSitemap: " + sitemapOrigin + "/sitemap.xml\n"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		_, _ = w.Write([]byte(body))
	})
}

type sitemapURL struct {
	Loc string `xml:"loc"`
}

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
	URLs    []sitemapURL `xml:"url"`
}

// handlerSitemapXML lists the operator's alias pages: the curated entry
// points of the site. Realms are left out on purpose; most are demos or near
// empty, and crawlers reach the good ones through links. The directory is
// only read to drop a path alias whose realm this chain lacks, which would
// otherwise publish a 404. No lastmod: gnoweb has no honest source for one.
func handlerSitemapXML(logger *slog.Logger, origin string, aliases map[string]AliasTarget, dir RealmDirectory) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin == "" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		realms, _, err := dir.Paths(r.Context())
		if err != nil {
			logger.Error("sitemap: unable to list paths", "error", err)
			http.Error(w, "sitemap unavailable", http.StatusBadGateway)
			return
		}

		var set sitemapURLSet
		for _, p := range sitemapAliases(aliases, realms) {
			set.URLs = append(set.URLs, sitemapURL{Loc: origin + (&url.URL{Path: p}).EscapedPath()})
		}

		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		_, _ = w.Write([]byte(xml.Header))
		if err := xml.NewEncoder(w).Encode(set); err != nil {
			logger.Error("sitemap: unable to encode", "error", err)
		}
	})
}

// sitemapAliases returns, sorted, the alias paths worth listing: static pages
// always, and a path alias only when its realm exists on this chain.
func sitemapAliases(aliases map[string]AliasTarget, realms []string) []string {
	known := make(map[string]bool, len(realms))
	for _, p := range realms {
		known[p] = true
	}

	var paths []string
	for alias, target := range aliases {
		// Only keys gnoweb serves under that exact URL: a key it would
		// rewrite or redirect lists a URL that is not the page served.
		if _, redirected := Redirects[alias]; redirected || !isCleanWebPath(alias) {
			continue
		}
		switch target.Kind {
		case StaticMarkdown:
			paths = append(paths, alias)
		case GnowebPath:
			realm, _, _ := strings.Cut(target.Value, ":")
			realm, _, _ = strings.Cut(realm, "$")
			if known[realm] {
				paths = append(paths, alias)
			}
		}
	}
	slices.Sort(paths)
	return paths
}

// isCleanWebPath reports whether gnoweb parses p and encodes it back unchanged.
func isCleanWebPath(p string) bool {
	u, err := weburl.ParseFromURL(&url.URL{Path: p})
	return err == nil && u.EncodeWebURL() == p
}
