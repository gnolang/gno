package gnoweb

import (
	"embed"
	_ "embed"
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

func AssetHandler(basePath string, cache bool) http.Handler {
	if cache {
		return http.FileServer(http.FS(assets))
	}

	handler := http.StripPrefix(basePath, http.FileServer(http.Dir(basePath)))
	return disableCache(handler)

}
