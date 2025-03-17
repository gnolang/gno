package test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

const (
	// TestCacheDirName is the name of the directory where test cache is stored
	TestCacheDirName = ".gno-test-cache"
)

type TestCache struct {
	Key TestCacheKey
	// Test result output
	Output string `json:"output"`
	// Test execution time
	Duration time.Duration `json:"duration"`
	// Timestamp when the cache was created
	Timestamp time.Time `json:"timestamp"`
}

// TestCacheKey represents the key information for test cache validation
type TestCacheKey struct {
	// Hash of test file content
	FileHash string `json:"fileHash"`
	// Package dependency information
	Dependencies map[string]PackageInfo `json:"dependencies"`
}

// PackageInfo contains information about a package's state
type PackageInfo struct {
	// Hash of all package files combined
	ContentHash string `json:"contentHash"`
	// Import path of the package
	Path string `json:"path"`
	// Package files and their hashes
	Files map[string]string `json:"files"`
}

// TestFuncCache represents the cache for a single test function
type TestFuncCache struct {
	Key TestCacheKey
	// Test function name
	TestName string `json:"testName"`
	// Test result output
	Output string `json:"output"`
	// Test execution time
	Duration time.Duration `json:"duration"`
	// Timestamp when the cache was created
	Timestamp time.Time `json:"timestamp"`
	// Test result
	Result report `json:"result"`
}

// TestFileCache represents the cache for a test file containing multiple test functions
type TestFileCache struct {
	// Map of test function name to its cache
	Tests map[string]*TestFuncCache `json:"tests"`
}

// cacheDir returns the path to the test cache directory
func cacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	cacheDir := filepath.Join(homeDir, TestCacheDirName)
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}
	return cacheDir, nil
}

// cacheFilePath returns the path to the cache file for a given test file
func cacheFilePath(testFile string) (string, error) {
	cacheDir, err := cacheDir()
	if err != nil {
		return "", err
	}
	// Use hash of absolute path to avoid file name collisions
	absPath, err := filepath.Abs(testFile)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(absPath)))
	return filepath.Join(cacheDir, hash+".json"), nil
}

// loadTestCache loads the cached test result for a given test file
func loadTestCache(testFile string, content []byte) (*TestCache, error) {
	cacheFile, err := cacheFilePath(testFile)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache TestCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache: %w", err)
	}

	if !isValidCache(&cache, testFile, content) {
		return nil, nil
	}

	return &cache, nil
}

// saveTestCache saves the test result to cache
func saveTestCache(testFile string, content []byte, output string, duration time.Duration) error {
	cacheFile, err := cacheFilePath(testFile)
	if err != nil {
		return err
	}

	key, err := computeCacheKey(testFile, content)
	if err != nil {
		return fmt.Errorf("failed to compute cache key: %w", err)
	}

	cache := TestCache{
		Key:       key,
		Output:    output,
		Duration:  duration,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	if err := os.WriteFile(cacheFile, data, 0o644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// computeCacheKey generates a cache key for the given test file
func computeCacheKey(testFile string, content []byte) (TestCacheKey, error) {
	key := TestCacheKey{
		FileHash:     fmt.Sprintf("%x", sha256.Sum256(content)),
		Dependencies: make(map[string]PackageInfo),
	}

	fileNode, err := gno.ParseFile(testFile, string(content))
	if err != nil {
		return key, fmt.Errorf("failed to parse file: %w", err)
	}

	// process each import declaration
	decls := fileNode.Decls
	for _, decl := range decls {
		if decl == nil {
			continue
		}
		if importDecl, ok := decl.(*gno.ImportDecl); ok {
			if importDecl.PkgPath == "" {
				continue
			}
			key.Dependencies[importDecl.PkgPath] = PackageInfo{
				Path: importDecl.PkgPath,
			}
		}
	}

	return key, nil
}

// isValidCache checks if the cached result is still valid
func isValidCache(cache *TestCache, testFile string, content []byte) bool {
	currentKey, err := computeCacheKey(testFile, content)
	if err != nil {
		return false
	}

	// Check file content hash
	if cache.Key.FileHash != currentKey.FileHash {
		return false
	}

	// Check all dependencies
	for path, currentInfo := range currentKey.Dependencies {
		cachedInfo, exists := cache.Key.Dependencies[path]
		if !exists {
			return false
		}

		// Compare package content hash
		if cachedInfo.ContentHash != currentInfo.ContentHash {
			return false
		}

		// Compare individual files
		if !reflect.DeepEqual(cachedInfo.Files, currentInfo.Files) {
			return false
		}
	}

	// Check if any cached dependency no longer exists
	for path := range cache.Key.Dependencies {
		if _, exists := currentKey.Dependencies[path]; !exists {
			return false
		}
	}

	return true
}
