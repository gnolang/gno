package test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestCache(t *testing.T) {
	t.Run("getCacheDir", func(t *testing.T) {
		dir, err := cacheDir()
		require.NoError(t, err)
		require.NotEmpty(t, dir)

		info, err := os.Stat(dir)
		require.NoError(t, err)
		require.True(t, info.IsDir())
	})

	t.Run("cache operations", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.gno")
		testContent := []byte("package main\n\nfunc main() {}")
		err := os.WriteFile(testFile, testContent, 0o644)
		require.NoError(t, err)

		output := "test output"
		duration := time.Second
		err = saveTestCache(testFile, testContent, output, duration)
		require.NoError(t, err)

		cache, err := loadTestCache(testFile, testContent)
		require.NoError(t, err)
		require.NotNil(t, cache)

		assert.Equal(t, output, cache.Output)
		assert.Equal(t, duration, cache.Duration)
		assert.NotEmpty(t, cache.Key.FileHash)
		assert.NotZero(t, cache.Timestamp)

		modifiedContent := []byte("package main\n\nfunc main() { println() }")
		cache, err = loadTestCache(testFile, modifiedContent)
		require.NoError(t, err)
		assert.Nil(t, cache)
	})

	t.Run("cache file path collision", func(t *testing.T) {
		tmpDir := t.TempDir()

		testFile1 := filepath.Join(tmpDir, "dir1", "test.gno")
		testFile2 := filepath.Join(tmpDir, "dir2", "test.gno")

		require.NoError(t, os.MkdirAll(filepath.Dir(testFile1), 0o755))
		require.NoError(t, os.MkdirAll(filepath.Dir(testFile2), 0o755))

		content1 := []byte("package main\n\nfunc test1() {}")
		content2 := []byte("package main\n\nfunc test2() {}")

		require.NoError(t, os.WriteFile(testFile1, content1, 0o644))
		require.NoError(t, os.WriteFile(testFile2, content2, 0o644))

		// make sure the cache file paths for each file are different
		path1, err := cacheFilePath(testFile1)
		require.NoError(t, err)

		path2, err := cacheFilePath(testFile2)
		require.NoError(t, err)

		assert.NotEqual(t, path1, path2)
	})

	t.Run("invalid cache operations", func(t *testing.T) {
		cache, err := loadTestCache("non_existent_file.gno", []byte{})
		require.NoError(t, err)
		assert.Nil(t, cache)

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "invalid.gno")
		err = os.WriteFile(testFile, []byte("invalid json"), 0o644)
		require.NoError(t, err)

		cacheFile, err := cacheFilePath(testFile)
		require.NoError(t, err)
		err = os.WriteFile(cacheFile, []byte("invalid json"), 0o644)
		require.NoError(t, err)

		cache, err = loadTestCache(testFile, []byte("test content"))
		assert.Error(t, err)
		assert.Nil(t, cache)
	})
}
