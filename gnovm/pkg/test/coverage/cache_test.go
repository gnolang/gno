package coverage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCache_SaveAndLoad(t *testing.T) {
	// Create a unique cache directory for this test
	cacheDir := filepath.Join(t.TempDir(), "cache")
	cache, err := NewCacheWithDir(cacheDir)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test.gno")
	if err := os.WriteFile(testFile, []byte("package test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	coverage := map[string]map[string]map[int]int64{
		"test": {
			"test.gno": {
				1: 5,
				2: 3,
			},
		},
	}
	executableLines := map[string]map[string]map[int]bool{
		"test": {
			"test.gno": {
				1: true,
				2: true,
				3: true,
			},
		},
	}

	if err := cache.Save(testDir, coverage, executableLines); err != nil {
		t.Fatalf("Failed to save cache: %v", err)
	}

	entry, err := cache.Load(testDir)
	if err != nil {
		t.Fatalf("Failed to load cache: %v", err)
	}
	if entry == nil {
		t.Fatal("Expected cache entry, got nil")
	}

	if len(entry.Coverage) != len(coverage) {
		t.Errorf("Coverage mismatch: expected %d packages, got %d", len(coverage), len(entry.Coverage))
	}
	if entry.Coverage["test"]["test.gno"][1] != 5 {
		t.Errorf("Coverage count mismatch: expected 5, got %d", entry.Coverage["test"]["test.gno"][1])
	}
}

func TestCache_InvalidateOnFileChange(t *testing.T) {
	// Create a unique cache directory for this test
	cacheDir := filepath.Join(t.TempDir(), "cache")
	cache, err := NewCacheWithDir(cacheDir)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test.gno")
	if err := os.WriteFile(testFile, []byte("package test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	coverage := map[string]map[string]map[int]int64{
		"test": {"test.gno": {1: 1}},
	}
	executableLines := map[string]map[string]map[int]bool{
		"test": {"test.gno": {1: true}},
	}

	if err := cache.Save(testDir, coverage, executableLines); err != nil {
		t.Fatalf("Failed to save cache: %v", err)
	}

	// Modify file content after a small delay
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(testFile, []byte("package test\n// modified"), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	entry, err := cache.Load(testDir)
	if err != nil {
		t.Fatalf("Failed to load cache: %v", err)
	}
	if entry != nil {
		t.Error("Expected cache miss after file modification, got cache hit")
	}
}

func TestCache_ClearEntry(t *testing.T) {
	// Create a unique cache directory for this test
	cacheDir := filepath.Join(t.TempDir(), "cache")
	cache, err := NewCacheWithDir(cacheDir)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	testDir := t.TempDir()
	coverage := map[string]map[string]map[int]int64{
		"test": {"test.gno": {1: 1}},
	}
	executableLines := map[string]map[string]map[int]bool{
		"test": {"test.gno": {1: true}},
	}

	if err := cache.Save(testDir, coverage, executableLines); err != nil {
		t.Fatalf("Failed to save cache: %v", err)
	}

	if err := cache.ClearEntry(testDir); err != nil {
		t.Fatalf("Failed to clear cache entry: %v", err)
	}

	entry, err := cache.Load(testDir)
	if err != nil {
		t.Fatalf("Failed to load cache: %v", err)
	}
	if entry != nil {
		t.Error("Expected cache miss after clearing entry, got cache hit")
	}
}
