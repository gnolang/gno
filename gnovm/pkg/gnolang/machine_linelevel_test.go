package gnolang

import (
	"strings"
	"testing"
)

// Test for Machine.getCurrentLocation
func TestMachineGetCurrentLocation(t *testing.T) {
	// Create a simple program to test location tracking

	// Create machine and parse code
	m := NewMachine("test", nil)
	m.Profiler = NewProfiler(ProfileCPU, 1)
	m.Profiler.EnableLineLevel(true)
	m.Profiler.Start()

	// For this test, we'll simulate location tracking
	// In real usage, this would be called during execution

	// Test statement location
	t.Run("statement location", func(t *testing.T) {
		// Simulate a statement node
		stmt := &AssignStmt{}
		stmt.SetSpan(Span{Pos: Pos{Line: 9, Column: 1}, End: Pos{Line: 9, Column: 10}})
		m.Stmts = append(m.Stmts, stmt)

		loc := m.getCurrentLocation()
		if loc == nil {
			t.Fatal("expected location, got nil")
		}

		if loc.Line() != 9 {
			t.Errorf("expected line 9, got %d", loc.Line())
		}

		// Clean up
		m.Stmts = m.Stmts[:0]
	})

	// Test expression location
	t.Run("expression location", func(t *testing.T) {
		// Simulate an expression node
		expr := &CallExpr{}
		expr.SetSpan(Span{Pos: Pos{Line: 11, Column: 1}, End: Pos{Line: 11, Column: 15}})
		m.Exprs = append(m.Exprs, expr)

		loc := m.getCurrentLocation()
		if loc == nil {
			t.Fatal("expected location, got nil")
		}

		if loc.Line() != 11 {
			t.Errorf("expected line 11, got %d", loc.Line())
		}

		// Clean up
		m.Exprs = m.Exprs[:0]
	})

	// Test frame source location
	t.Run("frame source location", func(t *testing.T) {
		// Simulate a frame with source
		call := &CallExpr{}
		call.SetSpan(Span{Pos: Pos{Line: 11, Column: 1}, End: Pos{Line: 11, Column: 15}})

		fv := &FuncValue{
			Name:    "add",
			PkgPath: "test",
		}

		m.PushFrameCall(call, fv, TypedValue{}, false)

		loc := m.getCurrentLocation()
		if loc == nil {
			t.Fatal("expected location, got nil")
		}

		if loc.Function() != "test.add" {
			t.Errorf("expected function 'test.add', got '%s'", loc.Function())
		}

		// Clean up
		m.Frames = m.Frames[:0]
	})
}

// Test for Machine.RecordCurrentLocation
func TestMachineRecordCurrentLocation(t *testing.T) {
	m := NewMachine("test", nil)
	m.Profiler = NewProfiler(ProfileCPU, 1)
	m.Profiler.EnableLineLevel(true)
	m.Profiler.Start()

	// Set up a mock location
	stmt := &AssignStmt{}
	stmt.SetSpan(Span{Pos: Pos{Line: 42, Column: 1}, End: Pos{Line: 42, Column: 10}})
	m.Stmts = append(m.Stmts, stmt)
	m.Package = &PackageValue{PkgPath: "test/pkg"}

	// Record location
	m.RecordCurrentLocation(1000)

	// Stop profiling and check results
	profile := m.Profiler.Stop()

	// Verify sample was recorded
	if len(profile.Samples) != 1 {
		t.Fatalf("expected 1 sample, got %d", len(profile.Samples))
	}

	sample := profile.Samples[0]
	if len(sample.Location) != 1 {
		t.Fatalf("expected 1 location, got %d", len(sample.Location))
	}

	loc := sample.Location[0]
	if loc.Line != 42 {
		t.Errorf("expected line 42, got %d", loc.Line)
	}

	if loc.File != "test/pkg" {
		t.Errorf("expected file 'test/pkg', got '%s'", loc.File)
	}

	// Check line stats
	lineStats := m.Profiler.GetLineStats("test/pkg")
	if lineStats == nil {
		t.Fatal("expected line stats")
	}

	if stats, exists := lineStats[42]; exists {
		if stats.GetCycles() != 1000 {
			t.Errorf("expected 1000 cycles, got %d", stats.GetCycles())
		}
		if stats.GetCount() != 1 {
			t.Errorf("expected count 1, got %d", stats.GetCount())
		}
	} else {
		t.Error("expected stats for line 42")
	}
}

// Integration test with actual code execution
func TestLineProfileIntegration(t *testing.T) {
	// Skip for now as it requires full machine execution setup
	t.Skip("Integration test requires full execution environment")

	// This would test actual execution with line-level profiling
	// including:
	// - Parsing and preprocessing
	// - Running with profiling enabled
	// - Collecting line-level statistics
	// - Generating annotated source output
}

// Test the complete workflow
func TestProfileWorkflow(t *testing.T) {
	// Create profiler
	profiler := NewProfiler(ProfileCPU, 1)
	profiler.EnableLineLevel(true)
	profiler.Start()

	// Simulate some execution
	m := &Machine{Profiler: profiler}

	// Record several locations
	locations := []struct {
		file   string
		line   int
		cycles int64
	}{
		{"main.gno", 10, 1000},
		{"main.gno", 10, 500}, // Same line, accumulate
		{"main.gno", 15, 3000},
		{"main.gno", 20, 100},
		{"helper.gno", 5, 2000},
	}

	for _, loc := range locations {
		ploc := newProfileLocation("test", loc.file, loc.line, 0)
		profiler.RecordLineLevel(m, ploc, loc.cycles)
	}

	profile := profiler.Stop()

	// Test annotated output for main.gno
	mockSource := `package main

import "fmt"

func helper() {
    // Some work
}

func main() {
    x := 42        // line 10
    y := x * 2
    
    for i := 0; i < 10; i++ {
        helper()   // line 15
    }
    
    fmt.Println(x, y)
    
    return         // line 20
}`

	var buf strings.Builder
	err := profile.WriteSourceAnnotated(&buf, "main.gno", strings.NewReader(mockSource))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Verify output contains expected elements
	checks := []string{
		"main.gno",
		"1500", // Total cycles for line 10
		"3000", // Cycles for line 15
		"HOT",  // Line 15 should be marked as hot
		"100",  // Cycles for line 20
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain '%s'", check)
		}
	}
}
