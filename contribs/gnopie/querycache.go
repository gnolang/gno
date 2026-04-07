// Package-level query cache for read-only, static data.
// Only source code (vm/qfile) and function signatures (vm/qfuncs) are cached.
// State queries (vm/qstorage, vm/qeval, vm/qrender) are never cached.
package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const queryCacheMaxAge = 1 * time.Hour

// queryCacheDir returns the directory for cached query results.
func queryCacheDir(home string) string {
	return filepath.Join(home, "gnopie", "cache", "queries")
}

// queryCacheKey generates a cache key from the query path and data.
func queryCacheKey(queryPath, data string) string {
	h := sha256.Sum256([]byte(queryPath + "\x00" + data))
	return fmt.Sprintf("%x", h[:12])
}

// loadCachedQuery returns cached query result if fresh enough.
func loadCachedQuery(home, queryPath, data string) (string, bool) {
	dir := queryCacheDir(home)
	key := queryCacheKey(queryPath, data)
	path := filepath.Join(dir, key)

	info, err := os.Stat(path)
	if err != nil {
		return "", false
	}
	if time.Since(info.ModTime()) > queryCacheMaxAge {
		return "", false
	}

	result, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return string(result), true
}

// saveCachedQuery stores a query result in the cache.
func saveCachedQuery(home, queryPath, data, result string) {
	dir := queryCacheDir(home)
	_ = os.MkdirAll(dir, 0o755)
	key := queryCacheKey(queryPath, data)
	path := filepath.Join(dir, key)
	_ = os.WriteFile(path, []byte(result), 0o644)
}
