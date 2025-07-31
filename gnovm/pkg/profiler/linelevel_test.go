package profiler

// import (
// 	"bytes"
// 	"fmt"
// 	"strings"
// 	"testing"
// )

// // Test for enhanced ProfileLocation with getters/setters
// func TestProfileLocation(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		setup    func() *profileLocation
// 		validate func(t *testing.T, loc *profileLocation)
// 	}{
// 		{
// 			name: "basic location creation",
// 			setup: func() *profileLocation {
// 				return newProfileLocation("main.foo", "/path/to/file.gno", 42, 10)
// 			},
// 			validate: func(t *testing.T, loc *profileLocation) {
// 				if loc.Function() != "main.foo" {
// 					t.Errorf("expected function 'main.foo', got '%s'", loc.Function())
// 				}
// 				if loc.File() != "/path/to/file.go" {
// 					t.Errorf("expected file '/path/to/file.go', got '%s'", loc.File())
// 				}
// 				if loc.Line() != 42 {
// 					t.Errorf("expected line 42, got %d", loc.Line())
// 				}
// 				if loc.Column() != 10 {
// 					t.Errorf("expected column 10, got %d", loc.Column())
// 				}
// 			},
// 		},
// 		{
// 			name: "location with PC",
// 			setup: func() *profileLocation {
// 				loc := newProfileLocation("test.bar", "test.go", 10, 5)
// 				loc.SetPC(0x1234)
// 				return loc
// 			},
// 			validate: func(t *testing.T, loc *profileLocation) {
// 				if loc.PC() != 0x1234 {
// 					t.Errorf("expected PC 0x1234, got 0x%x", loc.PC())
// 				}
// 			},
// 		},
// 		{
// 			name: "inline call location",
// 			setup: func() *profileLocation {
// 				loc := newProfileLocation("inline.func", "inline.go", 20, 0)
// 				loc.SetInlineCall(true)
// 				return loc
// 			},
// 			validate: func(t *testing.T, loc *profileLocation) {
// 				if !loc.IsInlineCall() {
// 					t.Error("expected inline call to be true")
// 				}
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			loc := tt.setup()
// 			tt.validate(t, loc)
// 		})
// 	}
// }

// // Test for LocationKey and caching
// func TestLocationCaching(t *testing.T) {
// 	cache := newLocationCache()

// 	// Test creating and caching a location
// 	key := LocationKey{
// 		PkgPath:  "gno.land/p/demo/test",
// 		Function: "TestFunc",
// 		File:     "test.gno",
// 		Line:     100,
// 	}

// 	loc1 := cache.getOrCreate(key)
// 	loc2 := cache.getOrCreate(key)

// 	// Should return the same instance
// 	if loc1 != loc2 {
// 		t.Error("expected same location instance from cache")
// 	}

// 	// Test cache size
// 	if cache.size() != 1 {
// 		t.Errorf("expected cache size 1, got %d", cache.size())
// 	}

// 	// Test different key
// 	key2 := LocationKey{
// 		PkgPath:  "gno.land/p/demo/test",
// 		Function: "TestFunc2",
// 		File:     "test.gno",
// 		Line:     200,
// 	}

// 	loc3 := cache.getOrCreate(key2)
// 	if loc3 == loc1 {
// 		t.Error("expected different location instance for different key")
// 	}

// 	if cache.size() != 2 {
// 		t.Errorf("expected cache size 2, got %d", cache.size())
// 	}
// }

// // Test for line-level statistics
// func TestLineStats(t *testing.T) {
// 	stats := newLineStats()

// 	// Test initial state
// 	if stats.GetCount() != 0 {
// 		t.Errorf("expected initial count 0, got %d", stats.GetCount())
// 	}
// 	if stats.GetCycles() != 0 {
// 		t.Errorf("expected initial cycles 0, got %d", stats.GetCycles())
// 	}

// 	// Test updating stats
// 	stats.addSample(100, 1, 1024)

// 	if stats.GetCount() != 1 {
// 		t.Errorf("expected count 1, got %d", stats.GetCount())
// 	}
// 	if stats.GetCycles() != 100 {
// 		t.Errorf("expected cycles 100, got %d", stats.GetCycles())
// 	}
// 	if stats.GetAllocations() != 1 {
// 		t.Errorf("expected allocations 1, got %d", stats.GetAllocations())
// 	}
// 	if stats.GetAllocBytes() != 1024 {
// 		t.Errorf("expected alloc bytes 1024, got %d", stats.GetAllocBytes())
// 	}

// 	// Test multiple samples
// 	stats.addSample(50, 2, 512)

// 	if stats.GetCount() != 2 {
// 		t.Errorf("expected count 2, got %d", stats.GetCount())
// 	}
// 	if stats.GetCycles() != 150 {
// 		t.Errorf("expected cycles 150, got %d", stats.GetCycles())
// 	}
// }

// // Test for enhanced profiler with line-level support
// func TestProfilerWithLineLevel(t *testing.T) {
// 	profiler := NewProfiler(ProfileCPU, 1) // Sample every operation

// 	// Enable line-level profiling
// 	profiler.EnableLineLevel(true)

// 	profiler.Start()

// 	// Simulate some operations with location info
// 	m := &Machine{Cycles: 0}

// 	// Mock location 1
// 	loc1 := newProfileLocation("test.func1", "test.gno", 10, 5)
// 	profiler.RecordLineLevel(m, loc1, 100)

// 	// Mock location 2 (same line, different column)
// 	loc2 := newProfileLocation("test.func1", "test.gno", 10, 15)
// 	profiler.RecordLineLevel(m, loc2, 50)

// 	// Mock location 3 (different line)
// 	loc3 := newProfileLocation("test.func2", "test.gno", 20, 1)
// 	profiler.RecordLineLevel(m, loc3, 200)

// 	_ = profiler.Stop()

// 	// Verify line stats were collected
// 	lineStats := profiler.GetLineStats("test.gno")
// 	if lineStats == nil {
// 		t.Fatal("expected line stats for test.gno")
// 	}

// 	// Check line 10 stats (should combine both columns)
// 	if stats, exists := lineStats[10]; exists {
// 		if stats.GetCount() != 2 {
// 			t.Errorf("expected 2 samples for line 10, got %d", stats.GetCount())
// 		}
// 		if stats.GetCycles() != 150 {
// 			t.Errorf("expected 150 cycles for line 10, got %d", stats.GetCycles())
// 		}
// 	} else {
// 		t.Error("expected stats for line 10")
// 	}

// 	// Check line 20 stats
// 	if stats, exists := lineStats[20]; exists {
// 		if stats.GetCount() != 1 {
// 			t.Errorf("expected 1 sample for line 20, got %d", stats.GetCount())
// 		}
// 		if stats.GetCycles() != 200 {
// 			t.Errorf("expected 200 cycles for line 20, got %d", stats.GetCycles())
// 		}
// 	} else {
// 		t.Error("expected stats for line 20")
// 	}
// }

// // Test for source code annotation
// func TestSourceAnnotation(t *testing.T) {
// 	profiler := NewProfiler(ProfileCPU, 1)
// 	profiler.EnableLineLevel(true)
// 	profiler.Start()

// 	// Add some line stats
// 	m := &Machine{Cycles: 0}
// 	profiler.RecordLineLevel(m, newProfileLocation("test", "mock.gno", 1, 0), 1000)
// 	profiler.RecordLineLevel(m, newProfileLocation("test", "mock.gno", 3, 0), 5000)
// 	profiler.RecordLineLevel(m, newProfileLocation("test", "mock.gno", 3, 0), 3000)
// 	profiler.RecordLineLevel(m, newProfileLocation("test", "mock.gno", 5, 0), 100)

// 	profile := profiler.Stop()

// 	// Mock source content
// 	mockSource := `package test

// func hotFunction() {  // This is a hot line
//     doWork()
//     return            // This line is cold
// }`

// 	// Test annotation output
// 	var buf bytes.Buffer
// 	err := profile.WriteSourceAnnotated(&buf, "mock.gno", strings.NewReader(mockSource))
// 	if err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}

// 	output := buf.String()

// 	// Verify output contains expected elements
// 	if !strings.Contains(output, "mock.gno") {
// 		t.Error("expected filename in output")
// 	}

// 	if !strings.Contains(output, "8000") {
// 		t.Error("expected cycle count 8000 for line 3")
// 	}

// 	if !strings.Contains(output, "HOT") {
// 		t.Error("expected HOT marker for high-cycle line")
// 	}

// 	// Line 5 should show minimal cycles
// 	if !strings.Contains(output, "100") {
// 		t.Error("expected cycle count 100 for line 5")
// 	}
// }

// // Test for memory pooling
// func TestMemoryPooling(t *testing.T) {
// 	profiler := NewProfiler(ProfileCPU, 1)
// 	profiler.EnableLineLevel(true)

// 	// Get location from pool
// 	loc1 := profiler.getLocationFromPool()
// 	if loc1 == nil {
// 		t.Fatal("expected non-nil location from pool")
// 	}

// 	// Set some values
// 	loc1.setValues("test", "file.go", 10, 5)

// 	// Return to pool
// 	profiler.putLocationToPool(loc1)

// 	// Get again - should be reset
// 	loc2 := profiler.getLocationFromPool()
// 	if loc2.Function() != "" {
// 		t.Error("expected location to be reset when retrieved from pool")
// 	}

// 	// Verify it's the same instance (pooling works)
// 	if loc1 != loc2 {
// 		t.Error("expected same instance from pool")
// 	}
// }

// // Benchmark for location caching
// func BenchmarkLocationCaching(b *testing.B) {
// 	cache := newLocationCache()
// 	keys := make([]LocationKey, 100)

// 	// Create diverse keys
// 	for i := 0; i < 100; i++ {
// 		keys[i] = LocationKey{
// 			PkgPath:  "gno.land/p/demo/test",
// 			Function: fmt.Sprintf("Func%d", i%10),
// 			File:     fmt.Sprintf("file%d.gno", i%5),
// 			Line:     i,
// 		}
// 	}

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		key := keys[i%100]
// 		_ = cache.getOrCreate(key)
// 	}
// }
