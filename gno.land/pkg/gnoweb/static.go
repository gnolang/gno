//go:build !noembed

package gnoweb

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io"
	"io/fs"
	"net/http"
	"sort"
	"strconv"
)

//go:embed public/*
var assets embed.FS

// AssetHandler returns an http.Handler to serve static assets from the embedded filesystem.
// Assets are always served from the embedded /public directory.
func AssetHandler() http.Handler {
	sub, err := fs.Sub(assets, "public")
	if err != nil {
		panic(err) // shouldn't fail if "public" exists
	}

	return http.FileServer(http.FS(sub))
}

// assetsHash stores a global ETag representing the content of all embedded files for cache validation.
var assetsHash string

// assetsVersion stores the token stamped on asset URLs. See AssetsVersion.
var assetsVersion string

// assetsVersionLen is how much of the asset digest that token carries. A prefix
// keeps the URL readable, and 48 bits is far more than telling two releases of
// the same asset set apart requires.
const assetsVersionLen = 12

// AssetsVersion returns the token stamped on asset URLs so a cache keyed on the
// URL refetches an asset once it changes. It is derived from the content of the
// embedded assets, so it is identical across restarts and across replicas
// running the same binary, and changes only when an asset changes.
func AssetsVersion() string { return assetsVersion }

var DefaultCacheAssetsHandler = func(next http.Handler) http.Handler {
	return CacheHandler(assetsHash, next)
}

func init() {
	// Collect file paths
	var paths []string
	fs.WalkDir(assets, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		paths = append(paths, p)
		return nil
	})
	sort.Strings(paths) // ensure deterministic order

	h := sha256.New()
	for _, p := range paths {
		f, err := assets.Open(p)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		io.Copy(h, f)
	}

	digest := hex.EncodeToString(h.Sum(nil))

	// ETag is quoted per RFC 7232
	assetsHash = strconv.Quote(digest)
	assetsVersion = digest[:assetsVersionLen]
}
