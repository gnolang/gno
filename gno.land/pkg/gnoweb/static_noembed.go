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

// AssetHandler returns a handler to serve static files from the given assetsPath, with cache disabled.
func AssetHandler() http.Handler {
	adir := getAssetDir()
	return http.FileServer(http.Dir(adir))
}

// DefaultCacheAssetsHandler in noembed mode, always invlidate cache
var DefaultCacheAssetsHandler = NoCacheHandler
