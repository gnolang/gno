package gnoweb

import "net/http"

// CacheHandler adds ETag and Cache-Control headers to all asset responses for caching.
func CacheHandler(hash string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set global ETag for all assets
		w.Header().Set("ETag", hash)
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

		// Return 304 if client's cached version matches
		if r.Header.Get("If-None-Match") == hash {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// NoCacheHandler always invalidates cache for all responses.
func NoCacheHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
