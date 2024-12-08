package gnoweb

import (
	"embed"
	"net/http"
)

const publicAssetsDir = "public"

//go:embed public/*
var assets embed.FS

func disableCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func AssetHandler(cache bool) http.Handler {
	if cache {
		return http.FileServer(http.FS(assets))
	}

	handler := http.StripPrefix(publicAssetsDir, http.FileServer(http.Dir(publicAssetsDir)))
	return disableCache(handler)
}
