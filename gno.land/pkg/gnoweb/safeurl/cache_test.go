package safeurl

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := NewCache(100, time.Hour)

	result := ScanResult{
		URL:       "https://example.com",
		Status:    StatusSafe,
		Verdict:   "safe",
		ScannedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	cache.Set("https://example.com", result)

	got, ok := cache.Get("https://example.com")
	if !ok {
		t.Fatal("expected to find cached result")
	}
	if got.URL != result.URL {
		t.Errorf("got URL %q, want %q", got.URL, result.URL)
	}
	if got.Status != result.Status {
		t.Errorf("got Status %v, want %v", got.Status, result.Status)
	}
}

func TestCache_GetMissing(t *testing.T) {
	cache := NewCache(100, time.Hour)

	_, ok := cache.Get("https://missing.com")
	if ok {
		t.Error("expected not to find missing URL")
	}
}

func TestCache_Expiration(t *testing.T) {
	cache := NewCache(100, time.Hour)

	// Set with already expired time
	result := ScanResult{
		URL:       "https://expired.com",
		Status:    StatusSafe,
		ExpiresAt: time.Now().Add(-time.Minute), // Already expired
	}

	cache.Set("https://expired.com", result)

	// Should not find expired entry
	_, ok := cache.Get("https://expired.com")
	if ok {
		t.Error("expected not to find expired entry")
	}
}

func TestCache_LRUEviction(t *testing.T) {
	// Small cache that can only hold 3 entries
	cache := NewCache(3, time.Hour)

	// Add 4 entries
	for i := 0; i < 4; i++ {
		cache.Set(string(rune('a'+i)), ScanResult{
			URL:       string(rune('a' + i)),
			Status:    StatusSafe,
			ExpiresAt: time.Now().Add(time.Hour),
		})
	}

	// First entry should be evicted
	_, ok := cache.Get("a")
	if ok {
		t.Error("expected first entry to be evicted")
	}

	// Last 3 entries should still be present
	for i := 1; i < 4; i++ {
		_, ok := cache.Get(string(rune('a' + i)))
		if !ok {
			t.Errorf("expected entry %q to be present", string(rune('a'+i)))
		}
	}
}

func TestCache_GetMulti(t *testing.T) {
	cache := NewCache(100, time.Hour)

	// Add some entries
	cache.Set("url1", ScanResult{URL: "url1", Status: StatusSafe, ExpiresAt: time.Now().Add(time.Hour)})
	cache.Set("url2", ScanResult{URL: "url2", Status: StatusUnsafe, ExpiresAt: time.Now().Add(time.Hour)})

	// Query multiple URLs
	found, missing := cache.GetMulti([]string{"url1", "url2", "url3"})

	if len(found) != 2 {
		t.Errorf("expected 2 found, got %d", len(found))
	}
	if len(missing) != 1 {
		t.Errorf("expected 1 missing, got %d", len(missing))
	}
	if missing[0] != "url3" {
		t.Errorf("expected missing[0] to be url3, got %q", missing[0])
	}
}

func TestCache_Len(t *testing.T) {
	cache := NewCache(100, time.Hour)

	if cache.Len() != 0 {
		t.Errorf("expected empty cache, got %d", cache.Len())
	}

	cache.Set("url1", ScanResult{URL: "url1", ExpiresAt: time.Now().Add(time.Hour)})
	cache.Set("url2", ScanResult{URL: "url2", ExpiresAt: time.Now().Add(time.Hour)})

	if cache.Len() != 2 {
		t.Errorf("expected 2 entries, got %d", cache.Len())
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(100, time.Hour)

	cache.Set("url1", ScanResult{URL: "url1", ExpiresAt: time.Now().Add(time.Hour)})
	cache.Set("url2", ScanResult{URL: "url2", ExpiresAt: time.Now().Add(time.Hour)})

	cache.Clear()

	if cache.Len() != 0 {
		t.Errorf("expected empty cache after clear, got %d", cache.Len())
	}
}

func TestCache_UpdateExisting(t *testing.T) {
	cache := NewCache(100, time.Hour)

	// Set initial value
	cache.Set("url1", ScanResult{URL: "url1", Status: StatusSafe, ExpiresAt: time.Now().Add(time.Hour)})

	// Update with new value
	cache.Set("url1", ScanResult{URL: "url1", Status: StatusUnsafe, ExpiresAt: time.Now().Add(time.Hour)})

	got, ok := cache.Get("url1")
	if !ok {
		t.Fatal("expected to find entry")
	}
	if got.Status != StatusUnsafe {
		t.Errorf("expected StatusUnsafe, got %v", got.Status)
	}

	// Should still have only 1 entry
	if cache.Len() != 1 {
		t.Errorf("expected 1 entry, got %d", cache.Len())
	}
}
