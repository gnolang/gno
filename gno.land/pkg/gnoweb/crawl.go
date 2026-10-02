package gnoweb

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

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

// fileLister is the subset of ClientAdapter the sitemap depends on.
type fileLister interface {
	ListFiles(ctx context.Context, path string, height int64) ([]string, error)
}

// handlerSitemapXML lists the alias pages only: most realms are demos or near
// empty, and crawlers reach the rest through links.
func handlerSitemapXML(logger *slog.Logger, origin string, aliases map[string]AliasTarget, client fileLister) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin == "" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		paths, err := sitemapAliases(r.Context(), aliases, client)
		if err != nil {
			logger.Error("sitemap: unable to check alias realms", "error", err)
			http.Error(w, "sitemap unavailable", http.StatusBadGateway)
			return
		}

		var set sitemapURLSet
		for _, p := range paths {
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
// aliases whose realm exists on this chain, so no 404 is published. Each
// target realm is checked with its own query: the full listing is capped at
// 1000 entries, and r/gnoland/* sorts after every r/g1... user realm. vm/qfile
// is used over vm/qpkgmeta_json because every node version answers it.
func sitemapAliases(ctx context.Context, aliases map[string]AliasTarget, client fileLister) ([]string, error) {
	var paths []string
	byRealm := map[string][]string{}
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
			if strings.HasPrefix(realm, "/r/") {
				byRealm[realm] = append(byRealm[realm], alias)
			}
		}
	}

	for realm, realmAliases := range byRealm {
		_, err := client.ListFiles(ctx, realm, 0)
		switch {
		case err == nil:
			paths = append(paths, realmAliases...)
		// A target with a file-like last segment is a bad alias, not an outage.
		case !errors.Is(err, ErrClientPackageNotFound) && !errors.Is(err, ErrClientFileNotFound):
			return nil, fmt.Errorf("%s: %w", realm, err)
		}
	}
	slices.Sort(paths)
	return paths, nil
}

// isCleanWebPath reports whether gnoweb parses p and encodes it back unchanged.
func isCleanWebPath(p string) bool {
	u, err := weburl.ParseFromURL(&url.URL{Path: p})
	return err == nil && u.EncodeWebURL() == p
}
