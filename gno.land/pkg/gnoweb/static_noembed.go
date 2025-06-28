//go:build noembed
// +build noembed

package gnoweb

import (
	"net/http"
	"os"
)

// AssetDir is the directory to serve static assets from. It can be set at build time using -ldflags.
var AssetDir string

func getAssetDir() string {
	if len(AssetDir) > 0 {
		return AssetDir
	}

	if dir := os.Getenv("GNOWEB_ASSETDIR"); dir != "" {
		return dir
	}
	return "./public"
}

// AssetHandler returns an http.Handler to serve static files from the given assetsPath.
func AssetHandler() http.Handler {
	adir := getAssetDir()
	return http.FileServer(http.Dir(adir))
}

// DefaultCacheAssetsHandler in noembed mode always disables cache.
var DefaultCacheAssetsHandler = NoCacheHandler
