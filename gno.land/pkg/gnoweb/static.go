//go:build !noembed
// +build !noembed

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

	// ETag is quoted per RFC 7232
	assetsHash = strconv.Quote(hex.EncodeToString(h.Sum(nil)))
}
