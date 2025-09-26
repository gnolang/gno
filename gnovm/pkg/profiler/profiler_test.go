package profiler

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestProfilerIntegration(t *testing.T) {
	t.Run("CPU profiling workflow", func(t *testing.T) {
		profiler := NewProfiler(ProfileCPU, 1)

		// Start profiling
		opts := Options{
			Type:       ProfileCPU,
			SampleRate: 1,
		}
		machine := &mockMachineInfo{
			cycles: 1000,
			frames: []FrameInfo{
				&mockFrameInfo{
					isCall:   true,
					funcName: "main",
					fileName: "main.gno",
					pkgPath:  "example.com/pkg",
					source:   &mockSourceInfo{line: 10, column: 5},
				},
			},
		}

		profiler.StartProfiling(machine, opts)

		// Simulate function execution
		profiler.RecordFuncEnter(machine, "compute")
		machine.cycles = 2000
		profiler.RecordSample(machine)
		profiler.RecordFuncExit(machine, "compute", 1000)

		// Stop profiling and get results
		profile := profiler.StopProfiling()

		if profile == nil {
			t.Fatal("Expected profile to be non-nil")
		}

		if profile.Type != ProfileCPU {
			t.Errorf("Expected profile type %d, got %d", ProfileCPU, profile.Type)
		}

		if len(profile.Samples) == 0 {
			t.Error("Expected at least one sample")
		}

		// Verify profile can be written
		var buf bytes.Buffer
		n, err := profile.WriteTo(&buf)
		if err != nil {
			t.Fatalf("Failed to write profile: %v", err)
		}
		if n == 0 {
			t.Error("Expected non-zero bytes written")
		}

		output := buf.String()
		if !strings.Contains(output, "Profile Type: CPU") {
			t.Error("Expected profile output to contain type")
		}
	})

	t.Run("Gas profiling workflow", func(t *testing.T) {
		profiler := NewProfiler(ProfileGas, 1)

		opts := Options{
			Type:       ProfileGas,
			SampleRate: 1,
		}
		machine := &mockMachineInfo{
			cycles:  1000,
			gasUsed: 500,
			frames: []FrameInfo{
				&mockFrameInfo{
					isCall:   true,
					funcName: "transfer",
					fileName: "token.gno",
					pkgPath:  "gno.land/p/demo/token",
					source:   &mockSourceInfo{line: 25, column: 3},
				},
			},
		}

		profiler.StartProfiling(machine, opts)

		// Record gas consumption
		profiler.RecordFuncEnter(machine, "transfer")
		machine.gasUsed = 1500
		profiler.RecordSample(machine)
		profiler.RecordFuncExit(machine, "transfer", 1000)

		profile := profiler.StopProfiling()

		if profile == nil {
			t.Fatal("Expected profile to be non-nil")
		}

		if profile.Type != ProfileGas {
			t.Errorf("Expected profile type %d, got %d", ProfileGas, profile.Type)
		}

		// Verify gas data in samples
		found := false
		for _, sample := range profile.Samples {
			if gas, ok := sample.NumLabel["gas"]; ok && len(gas) > 0 {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected to find gas data in samples")
		}
	})

	t.Run("Memory profiling workflow", func(t *testing.T) {
		profiler := NewProfiler(ProfileMemory, 1)

		opts := Options{
			Type:       ProfileMemory,
			SampleRate: 1,
		}
		machine := &mockMachineInfo{
			cycles: 1000,
			frames: []FrameInfo{
				&mockFrameInfo{
					isCall:   true,
					funcName: "allocateBuffer",
					fileName: "buffer.gno",
					pkgPath:  "example.com/buffer",
					source:   &mockSourceInfo{line: 15, column: 10},
				},
			},
		}

		profiler.StartProfiling(machine, opts)

		// Record memory allocations
		profiler.RecordAlloc(machine, 1024, 1, "[]byte")
		profiler.RecordAlloc(machine, 2048, 1, "struct")

		profile := profiler.StopProfiling()

		if profile == nil {
			t.Fatal("Expected profile to be non-nil")
		}

		if profile.Type != ProfileMemory {
			t.Errorf("Expected profile type %d, got %d", ProfileMemory, profile.Type)
		}

		// Verify memory allocation data
		// The profiler creates both individual allocation samples and function summary samples
		// We only want to count the actual allocation samples, not the summaries
		totalBytes := int64(0)
		allocCount := 0
		for _, sample := range profile.Samples {
			// Check if this is a direct allocation sample (has "type" label)
			if _, hasType := sample.Label["type"]; hasType && sample.SampleType == ProfileMemory {
				if bytes, ok := sample.NumLabel["bytes"]; ok && len(bytes) > 0 {
					totalBytes += bytes[0]
					allocCount++
				}
			}
		}
		if allocCount != 2 {
			t.Errorf("Expected 2 allocation samples, got %d", allocCount)
		}
		if totalBytes != 3072 {
			t.Errorf("Expected total bytes to be 3072, got %d", totalBytes)
		}
	})

	t.Run("Line-level profiling", func(t *testing.T) {
		profiler := NewProfiler(ProfileCPU, 1)
		profiler.EnableLineProfiling()

		if !profiler.IsLineProfilingEnabled() {
			t.Error("Expected line profiling to be enabled")
		}

		opts := Options{
			Type:       ProfileCPU,
			SampleRate: 1,
		}
		machine := &mockMachineInfo{
			cycles: 1000,
			frames: []FrameInfo{
				&mockFrameInfo{
					isCall:   true,
					funcName: "calculate",
					fileName: "calc.gno",
					pkgPath:  "example.com/calc",
					source:   &mockSourceInfo{line: 20, column: 5},
				},
			},
		}

		profiler.StartProfiling(machine, opts)

		// Record line samples
		profiler.RecordLineSample("calculate", "calc.gno", 20, 100)
		profiler.RecordLineSample("calculate", "calc.gno", 21, 200)
		profiler.RecordLineSample("calculate", "calc.gno", 20, 150)

		profile := profiler.StopProfiling()

		if profile == nil {
			t.Fatal("Expected profile to be non-nil")
		}

		// Get line statistics
		stats := profiler.LineStats("calc.gno")
		if stats == nil {
			t.Fatal("Expected line stats to be non-nil")
		}

		// Verify we have stats for lines 20 and 21
		if len(stats) != 2 {
			t.Errorf("Expected 2 line stats, got %d", len(stats))
		}

		// Verify line 20 stats
		if line20Stats, ok := stats[20]; ok {
			if line20Stats.Count() != 2 {
				t.Errorf("Expected line 20 count to be 2, got %d", line20Stats.Count())
			}
			if line20Stats.Cycles() != 250 {
				t.Errorf("Expected line 20 cycles to be 250, got %d", line20Stats.Cycles())
			}
		} else {
			t.Error("Expected stats for line 20")
		}

		// Verify line 21 stats
		if line21Stats, ok := stats[21]; ok {
			if line21Stats.Count() != 1 {
				t.Errorf("Expected line 21 count to be 1, got %d", line21Stats.Count())
			}
			if line21Stats.Cycles() != 200 {
				t.Errorf("Expected line 21 cycles to be 200, got %d", line21Stats.Cycles())
			}
		} else {
			t.Error("Expected stats for line 21")
		}
	})

	t.Run("JSON output format", func(t *testing.T) {
		profiler := NewProfiler(ProfileCPU, 1)

		opts := Options{
			Type:       ProfileCPU,
			SampleRate: 1,
		}
		machine := &mockMachineInfo{
			cycles: 1000,
			frames: []FrameInfo{
				&mockFrameInfo{
					isCall:   true,
					funcName: "process",
					fileName: "process.gno",
					pkgPath:  "example.com/proc",
					source:   &mockSourceInfo{line: 30, column: 1},
				},
			},
		}

		profiler.StartProfiling(machine, opts)
		profiler.RecordSample(machine)
		profile := profiler.StopProfiling()

		// Test JSON output
		var buf bytes.Buffer
		err := profile.WriteJSON(&buf)
		if err != nil {
			t.Fatalf("Failed to write JSON: %v", err)
		}

		// Verify JSON is valid
		var decoded Profile
		err = json.Unmarshal(buf.Bytes(), &decoded)
		if err != nil {
			t.Fatalf("Failed to decode JSON: %v", err)
		}

		if decoded.Type != ProfileCPU {
			t.Errorf("Expected decoded type to be %d, got %d", ProfileCPU, decoded.Type)
		}
	})

	t.Run("Test file filtering", func(t *testing.T) {
		profiler := NewProfiler(ProfileCPU, 1)

		opts := Options{
			Type:       ProfileCPU,
			SampleRate: 1,
		}

		// Create frames with test and non-test files
		testMachine := &mockMachineInfo{
			cycles: 1000,
			frames: []FrameInfo{
				&mockFrameInfo{
					isCall:   true,
					funcName: "TestFunction",
					fileName: "example_test.gno",
					pkgPath:  "example.com/pkg",
					source:   &mockSourceInfo{line: 10, column: 1},
				},
			},
		}

		normalMachine := &mockMachineInfo{
			cycles: 2000,
			frames: []FrameInfo{
				&mockFrameInfo{
					isCall:   true,
					funcName: "NormalFunction",
					fileName: "example.gno",
					pkgPath:  "example.com/pkg",
					source:   &mockSourceInfo{line: 20, column: 1},
				},
			},
		}

		profiler.StartProfiling(testMachine, opts)

		// Record samples from both test and non-test files
		profiler.RecordSample(testMachine)
		profiler.RecordSample(normalMachine)

		profile := profiler.StopProfiling()

		// Verify test files are excluded from output
		var buf bytes.Buffer
		profile.WriteTo(&buf)
		output := buf.String()

		if strings.Contains(output, "TestFunction") {
			t.Error("Expected test function to be filtered out")
		}
		if !strings.Contains(output, "NormalFunction") {
			t.Error("Expected normal function to be present")
		}
	})
}
