package gnoweb

import (
	"embed"
	"net/http"
)

//go:embed public/*
var assets embed.FS

func disableCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// AssetHandler returns the handler to serve static assets. If cache is true,
// these will be served using the static files embedded in the binary; otherwise
// they will served from the filesystem.
func AssetHandler() http.Handler {
	return http.FileServer(http.FS(assets))
}

func DevAssetHandler(path, dir string) http.Handler {
	handler := http.StripPrefix(path, http.FileServer(http.Dir(dir)))
	return disableCache(handler)
}
