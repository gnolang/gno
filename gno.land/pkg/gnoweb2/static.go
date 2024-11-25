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

func AssetDevHandler() http.Handler {
	return disableCache(http.StripPrefix("/public", http.FileServer(http.Dir("public"))))
}

func AssetHandler() http.Handler {
	return http.FileServer(http.FS(assets))
}
