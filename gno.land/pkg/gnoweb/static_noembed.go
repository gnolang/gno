//go:build noembed

package gnoweb

import (
	"net/http"
	"os"
	"time"
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

// noembedAssetsVersion is fixed for the life of the process; assets are read
// from disk per request here, so there is no digest of them to derive it from.
var noembedAssetsVersion = time.Now().Format("20060102150405")

// AssetsVersion returns the token stamped on asset URLs. Assets served from
// disk can change while the process runs, which is why DefaultCacheAssetsHandler
// disables caching outright in this build; keeping those assets fresh is its job
// rather than this token's.
func AssetsVersion() string { return noembedAssetsVersion }
