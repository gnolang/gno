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
// a 2xx is cached: an error, such as a 404, is never kept. Any other ?v=,
// such as the next release's asked of this one mid-deploy, is revalidated.
func CacheHandler(hash, version string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cache := "no-cache"
		if v := r.URL.Query().Get("v"); (v != "" && v == version) || isImmutableAsset(r.URL.Path) {
			cache = immutableCacheControl
		}

		// Return 304 if client's cached version matches
		if r.Header.Get("If-None-Match") == hash {
			w.Header().Set("ETag", hash)
			w.Header().Set("Cache-Control", cache)
			w.WriteHeader(http.StatusNotModified)
			return
		}

		next.ServeHTTP(&cacheWriter{ResponseWriter: w, hash: hash, cache: cache}, r)
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

// cacheWriter sets the caching headers once the status is known: ETag and
// cache on a 2xx, no caching on anything else.
type cacheWriter struct {
	http.ResponseWriter
	hash, cache string
	wroteHeader bool
}

func (w *cacheWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.wroteHeader = true
		if code >= 200 && code < 300 {
			w.Header().Set("ETag", w.hash)
			w.Header().Set("Cache-Control", w.cache)
		} else {
			w.Header().Set("Cache-Control", "no-store")
		}
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *cacheWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// NoCacheHandler always invalidates cache for all responses.
func NoCacheHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
