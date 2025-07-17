package coverage

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// CacheEntry represents a cached coverage result.
type CacheEntry struct {
	Coverage        map[string]map[string]map[int]int64
	ExecutableLines map[string]map[string]map[int]bool
	Timestamp       time.Time
	Hash            string // Hash of the test files and source files
}

// Cache manages persistent storage of coverage data using LevelDB.
type Cache struct {
	db       *leveldb.DB
	cacheDir string
}

// NewCache creates a new coverage cache.
func NewCache() (*Cache, error) {
	cacheDir := filepath.Join(os.TempDir(), "gno-coverage-cache")
	return NewCacheWithDir(cacheDir)
}

// NewCacheWithDir creates a new coverage cache with a specific directory.
func NewCacheWithDir(cacheDir string) (*Cache, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	dbPath := filepath.Join(cacheDir, "coverage.db")
	db, err := leveldb.OpenFile(dbPath, &opt.Options{
		CompactionTableSize: 2 * opt.MiB,
		WriteBuffer:         4 * opt.MiB,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open leveldb: %w", err)
	}

	return &Cache{
		db:       db,
		cacheDir: cacheDir,
	}, nil
}

// Close closes the cache database.
func (c *Cache) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// generateCacheKey creates a unique key for the cache based on the target directory.
func (c *Cache) generateCacheKey(targetDir string) []byte {
	h := sha256.New()
	h.Write([]byte("coverage:"))
	h.Write([]byte(targetDir))
	return h.Sum(nil)[:16]
}

// computeHash calculates a hash of all relevant files to detect changes.
func (c *Cache) computeHash(targetDir string) (string, error) {
	h := sha256.New()

	// Walk through directory and hash relevant files
	err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only hash .gno files
		if !info.IsDir() && filepath.Ext(path) == ".gno" {
			// Include file path relative to targetDir
			relPath, err := filepath.Rel(targetDir, path)
			if err != nil {
				return err
			}
			h.Write([]byte(relPath))

			// Include modification time and size
			h.Write([]byte(info.ModTime().Format(time.RFC3339Nano)))
			h.Write([]byte(fmt.Sprintf("%d", info.Size())))
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// Load attempts to load cached coverage data for the given target directory.
func (c *Cache) Load(targetDir string) (*CacheEntry, error) {
	key := c.generateCacheKey(targetDir)

	data, err := c.db.Get(key, nil)
	if errors.Is(err, leveldb.ErrNotFound) {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cache entry: %w", err)
	}

	var entry CacheEntry
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&entry); err != nil {
		return nil, fmt.Errorf("failed to decode cache entry: %w", err)
	}

	// Verify hash to ensure cache is still valid
	currentHash, err := c.computeHash(targetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to compute hash: %w", err)
	}

	if entry.Hash != currentHash {
		// Cache is stale, remove it
		c.db.Delete(key, nil)
		return nil, nil
	}

	return &entry, nil
}

// Save stores coverage data in the cache.
func (c *Cache) Save(targetDir string, coverage map[string]map[string]map[int]int64, executableLines map[string]map[string]map[int]bool) error {
	key := c.generateCacheKey(targetDir)

	hash, err := c.computeHash(targetDir)
	if err != nil {
		return fmt.Errorf("failed to compute hash: %w", err)
	}

	entry := CacheEntry{
		Coverage:        coverage,
		ExecutableLines: executableLines,
		Timestamp:       time.Now(),
		Hash:            hash,
	}

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(entry); err != nil {
		return fmt.Errorf("failed to encode cache entry: %w", err)
	}

	if err := c.db.Put(key, buf.Bytes(), nil); err != nil {
		return fmt.Errorf("failed to save cache entry: %w", err)
	}

	return nil
}

// Clear removes all cached coverage data.
func (c *Cache) Clear() error {
	if c.db != nil {
		c.db.Close()
		c.db = nil
	}
	return os.RemoveAll(c.cacheDir)
}

// ClearEntry removes cached coverage data for a specific target directory.
func (c *Cache) ClearEntry(targetDir string) error {
	key := c.generateCacheKey(targetDir)
	return c.db.Delete(key, nil)
}
