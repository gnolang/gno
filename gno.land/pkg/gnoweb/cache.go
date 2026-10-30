package gnoweb

import (
	"net/http"
	"path"
	"strings"
)

const immutableCacheControl = "public, max-age=31536000, immutable"

// CacheHandler adds ETag and Cache-Control headers to asset responses for
// caching. A response is cached as immutable when its URL cannot serve other
// bytes after a deploy: it carries this build's assets version (?v=), or
// names a file that never changes in place (a content-hashed chunk, a font). Other
// files, such as controller.js, which modules import without a version, are
// revalidated against the ETag, so a deploy never leaves a stale copy. Only
// a 2xx or a 304 keeps these headers: an error, such as a 404, is never
// cached. Any other ?v=, such as the next release's asked of this one
// mid-deploy, is revalidated.
func CacheHandler(hash, version string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cache := "no-cache"
		if v := r.URL.Query().Get("v"); (v != "" && v == version) || isImmutableAsset(r.URL.Path) {
			cache = immutableCacheControl
		}

		// Set for a 2xx or a 304 up front; cacheWriter replaces them on any
		// other status, which handlers always send through WriteHeader.
		w.Header().Set("ETag", hash)
		w.Header().Set("Cache-Control", cache)

		// Return 304 if client's cached version matches
		if r.Header.Get("If-None-Match") == hash {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		next.ServeHTTP(cacheWriter{w}, r)
	})
}

// isImmutableAsset reports whether an asset path names bytes that never
// change: esbuild's content-hashed chunk-<hash>.js, or a font.
func isImmutableAsset(p string) bool {
	name := path.Base(p)
	switch path.Ext(name) {
	case ".woff", ".woff2":
		return true
	case ".js":
		return strings.HasPrefix(name, "chunk-")
	}
	return false
}

// cacheWriter keeps an error, such as a 404, out of every cache: on a
// status other than 2xx or 304 it drops the ETag and caching headers
// CacheHandler set. http.FileServer sends its own 304 for a weak If-None-Match.
type cacheWriter struct{ http.ResponseWriter }

func (w cacheWriter) WriteHeader(code int) {
	if code != http.StatusNotModified && (code < 200 || code >= 300) {
		w.Header().Del("ETag")
		w.Header().Set("Cache-Control", "no-store")
	}
	w.ResponseWriter.WriteHeader(code)
}

// NoCacheHandler always invalidates cache for all responses.
func NoCacheHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
