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

// normalizeCanonicalOrigin lowercases origin and drops a default port and
// trailing slashes. Anything but a bare http(s) origin is an error.
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
	// Rebuilding from scheme and host rejects a path, query, credentials, and
	// the empty trailing "?" or "#" that Parse accepts.
	if (scheme != "http" && scheme != "https") || host == "" || strings.HasSuffix(host, ":") ||
		!strings.EqualFold((&url.URL{Scheme: u.Scheme, Host: u.Host}).String(), origin) {
		return "", fmt.Errorf("invalid canonical origin %q: want scheme://host[:port]", origin)
	}
	return scheme + "://" + host, nil
}

// noIndexMiddleware marks every response noindex. robots.txt must still allow
// crawling, or crawlers never see it.
func noIndexMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Robots-Tag", "noindex, nofollow")
		next.ServeHTTP(w, r)
	})
}

// handlerRobotsTXT serves robots.txt; sitemapOrigin is empty without a sitemap.
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

// handlerSitemapXML lists the alias pages only: most realms are demos or near
// empty, and crawlers reach the rest through links.
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

// sitemapAliases returns the sorted aliases to list: static pages, and path
// aliases whose realm exists on this chain, so no 404 is published.
func sitemapAliases(aliases map[string]AliasTarget, realms []string) []string {
	known := make(map[string]bool, len(realms))
	for _, p := range realms {
		known[p] = true
	}

	var paths []string
	for alias, target := range aliases {
		// A key gnoweb redirects or rewrites is not the URL of the page served.
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
