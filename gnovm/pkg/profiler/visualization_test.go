package profiler

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// Helper function to create test profile data
func createTestProfile(t *testing.T) *Profile {
	t.Helper()
	return &Profile{
		Type:          ProfileCPU,
		TimeNanos:     time.Now().UnixNano(),
		DurationNanos: 1000000000, // 1 second
		Samples: []ProfileSample{
			{
				Location: []ProfileLocation{
					{Function: "main.fibonacci", File: "main.go", Line: 10},
				},
				Value: []int64{100, 50000}, // 100 calls, 50000 cycles
				NumLabel: map[string][]int64{
					"calls":       {100},
					"cycles":      {50000},
					"flat_cycles": {40000}, // 80% self time
					"cum_cycles":  {50000}, // 100% total time
				},
				SampleType: ProfileCPU,
			},
			{
				Location: []ProfileLocation{
					{Function: "main.calculate", File: "main.go", Line: 20},
				},
				Value: []int64{50, 30000}, // 50 calls, 30000 cycles
				NumLabel: map[string][]int64{
					"calls":       {50},
					"cycles":      {30000},
					"flat_cycles": {15000}, // 50% self time
					"cum_cycles":  {30000}, // 100% total time
				},
				SampleType: ProfileCPU,
			},
			{
				Location: []ProfileLocation{
					{Function: "main.helper", File: "main.go", Line: 30},
				},
				Value: []int64{200, 20000}, // 200 calls, 20000 cycles
				NumLabel: map[string][]int64{
					"calls":       {200},
					"cycles":      {20000},
					"flat_cycles": {20000}, // 100% self time (leaf function)
					"cum_cycles":  {20000}, // 100% total time
				},
				SampleType: ProfileCPU,
			},
		},
	}
}

func TestProfileWriteTo(t *testing.T) {
	profile := createTestProfile(t)

	var buf bytes.Buffer
	_, err := profile.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	output := buf.String()

	// Check header
	if !strings.Contains(output, "Profile Type: CPU") {
		t.Error("Missing profile type in output")
	}

	if !strings.Contains(output, "Duration: 1s") {
		t.Error("Missing duration in output")
	}

	// Check function entries
	if !strings.Contains(output, "main.fibonacci") {
		t.Error("Missing main.fibonacci in output")
	}

	if !strings.Contains(output, "50000") {
		t.Error("Missing cycle count for fibonacci")
	}
}

func TestProfileWriteTopList(t *testing.T) {
	profile := createTestProfile(t)

	var buf bytes.Buffer
	err := profile.WriteTopList(&buf)
	if err != nil {
		t.Fatalf("WriteTopList failed: %v", err)
	}

	output := buf.String()

	// Check header
	if !strings.Contains(output, "Top Functions by CPU Cycles") {
		t.Error("Missing header in top list output")
	}

	// Check for flat and cum columns
	if !strings.Contains(output, "Flat") || !strings.Contains(output, "Cum") {
		t.Error("Missing Flat/Cum column headers")
	}

	// Check that functions are sorted by cycles (fibonacci should be first)
	lines := strings.Split(output, "\n")
	fibonacciFound := false
	calculateFound := false

	for _, line := range lines {
		if strings.Contains(line, "main.fibonacci") {
			fibonacciFound = true
			// Check visual bar exists
			if !strings.Contains(line, "█") {
				t.Error("Missing visual bar for fibonacci")
			}
			// Check that the line contains percentage values
			if !strings.Contains(line, "%") {
				t.Error("Missing percentage values for fibonacci")
			}
		}
		if strings.Contains(line, "main.calculate") {
			calculateFound = true
			if !fibonacciFound {
				t.Error("Functions not sorted by cycles - calculate appeared before fibonacci")
			}
		}
	}

	if !fibonacciFound || !calculateFound {
		t.Error("Not all functions found in output")
	}

	// Check summary
	if !strings.Contains(output, "Total cycles: 100000") {
		t.Error("Missing or incorrect total cycles")
	}
}

func TestProfileWriteCallTree_Hierarchy(t *testing.T) {
	// Create profile with nested calls
	profile := &Profile{
		Type:          ProfileCPU,
		DurationNanos: 1000000000,
		Samples: []ProfileSample{
			{
				Location: []ProfileLocation{
					{Function: "main.root"},
					{Function: "main.branch"},
					{Function: "main.leaf1"},
				},
				Value:    []int64{5, 500},
				NumLabel: map[string][]int64{"calls": {5}, "cycles": {500}},
			},
			{
				Location: []ProfileLocation{
					{Function: "main.root"},
					{Function: "main.branch"},
					{Function: "main.leaf2"},
				},
				Value:    []int64{3, 300},
				NumLabel: map[string][]int64{"calls": {3}, "cycles": {300}},
			},
		},
	}

	var buf bytes.Buffer
	err := profile.WriteCallTree(&buf)
	if err != nil {
		t.Fatalf("WriteCallTree failed: %v", err)
	}

	output := buf.String()

	// Check hierarchy with indentation
	if !strings.Contains(output, "Call Tree") {
		t.Error("Missing call tree header")
	}

	// Debug output
	t.Logf("Call tree output:\n%s", output)

	// Check indentation levels
	lines := strings.Split(output, "\n")
	rootIndent := -1
	branchIndent := -1
	rootFound := false
	branchFound := false

	for _, line := range lines {
		if strings.Contains(line, "main.root") && !rootFound {
			// Count leading spaces/tabs/special chars up to function name
			indent := 0
			for i, ch := range line {
				if strings.HasPrefix(line[i:], "main.root") {
					break
				}
				if ch == ' ' || ch == '\t' || ch == '│' || ch == '├' || ch == '└' || ch == '─' {
					indent++
				}
			}
			rootIndent = indent
			rootFound = true
			t.Logf("Root line: %q, calculated indent: %d", line, indent)
		}
		if strings.Contains(line, "main.branch") && !branchFound {
			// Count leading spaces/tabs/special chars up to function name
			indent := 0
			for i, ch := range line {
				if strings.HasPrefix(line[i:], "main.branch") {
					break
				}
				if ch == ' ' || ch == '\t' || ch == '│' || ch == '├' || ch == '└' || ch == '─' {
					indent++
				}
			}
			branchIndent = indent
			branchFound = true
			t.Logf("Branch line: %q, calculated indent: %d", line, indent)
		}
	}

	if branchIndent <= rootIndent {
		t.Error("Branch should be indented more than root")
	}
}

func TestFlatCumulativeColumns(t *testing.T) {
	profile := &Profile{
		Type:          ProfileCPU,
		DurationNanos: 1000000000,
		Samples: []ProfileSample{
			{
				Location: []ProfileLocation{{Function: "main.parent"}},
				Value:    []int64{10, 10000},
				NumLabel: map[string][]int64{
					"calls":       {10},
					"flat_cycles": {2000},  // 20% self time
					"cum_cycles":  {10000}, // 100% total time
				},
			},
			{
				Location: []ProfileLocation{{Function: "main.child"}},
				Value:    []int64{10, 8000},
				NumLabel: map[string][]int64{
					"calls":       {10},
					"flat_cycles": {8000}, // 100% self time
					"cum_cycles":  {8000}, // 100% total time
				},
			},
		},
	}

	var buf bytes.Buffer
	err := profile.WriteTopList(&buf)
	if err != nil {
		t.Fatalf("WriteTopList failed: %v", err)
	}

	output := buf.String()

	// Check that parent function shows different flat vs cumulative
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "main.parent") {
			// Should show ~20% flat and 100% cumulative
			if !strings.Contains(line, "2000") {
				t.Error("Parent function should show 2000 flat cycles")
			}
			if !strings.Contains(line, "10000") {
				t.Error("Parent function should show 10000 cumulative cycles")
			}
		}
		if strings.Contains(line, "main.child") {
			// Should show 100% for both flat and cumulative
			if !strings.Contains(line, "8000") {
				t.Error("Child function should show 8000 cycles")
			}
		}
	}
}

func TestEmptyProfile(t *testing.T) {
	profile := &Profile{
		Type:          ProfileCPU,
		DurationNanos: 0,
		Samples:       []ProfileSample{},
	}

	// Test all formats with empty profile
	formats := []struct {
		name   string
		format ProfileFormat
	}{
		{"Text", FormatText},
		{"TopList", FormatTopList},
		{"CallTree", FormatCallTree},
	}

	for _, f := range formats {
		t.Run(f.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := profile.WriteFormat(&buf, f.format)
			if err != nil {
				t.Errorf("WriteFormat failed for %s: %v", f.name, err)
			}

			if buf.Len() == 0 {
				t.Errorf("Empty output for format %s", f.name)
			}
		})
	}
}

func TestLongFunctionNames_ShouldBeTruncated(t *testing.T) {
	longName := "github.com/very/long/package/path/with/many/segments/that/exceeds/normal/length.VeryLongFunctionNameThatShouldBeTruncated"

	profile := &Profile{
		Type:          ProfileCPU,
		DurationNanos: 1000000000,
		Samples: []ProfileSample{
			{
				Location: []ProfileLocation{{Function: longName}},
				Value:    []int64{1, 1000},
				NumLabel: map[string][]int64{"calls": {1}, "cycles": {1000}},
			},
		},
	}

	var buf bytes.Buffer
	err := profile.WriteTopList(&buf)
	if err != nil {
		t.Fatalf("WriteTopList failed: %v", err)
	}

	output := buf.String()

	// Check that long names are truncated with "..."
	if strings.Contains(output, longName) {
		t.Error("Long function name was not truncated")
	}

	if !strings.Contains(output, "...") {
		t.Error("Truncation indicator (...) not found")
	}
}
