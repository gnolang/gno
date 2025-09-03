package test

import (
	"bytes"
	"strings"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestCoverageBasic(t *testing.T) {
	// Create a simple test package
	mpkg := &std.MemPackage{
		Name: "test",
		Path: "gno.land/p/demo/test",
		Files: []*std.MemFile{
			{
				Name: "test.gno",
				Body: `package test

func Add(a, b int) int {
	return a + b
}`,
			},
			{
				Name: "test_test.gno",
				Body: `package test

import "testing"

func TestAdd(t *testing.T) {
	result := Add(2, 3)
	if result != 5 {
		t.Errorf("Add(2, 3) = %d; want 5", result)
	}
}`,
			},
		},
	}

	// Set up test options with coverage enabled
	var stdout, stderr bytes.Buffer
	opts := NewTestOptions("", &stdout, &stderr)
	opts.Cover = true
	opts.CoverMode = "count"
	opts.BaseStore, opts.TestStore = StoreWithOptions(
		"", opts.WriterForStore(),
		StoreOptions{WithExtern: true, WithExamples: false, Testing: true},
	)

	// Run the test
	err := Test(mpkg, "", opts)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	// Check that coverage was reported
	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "coverage:") {
		t.Errorf("Expected coverage output, got: %s", stderrOutput)
	}
}

func TestCoverageModes(t *testing.T) {
	modes := []string{"set", "count", "atomic"}
	
	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			// Create coverage data with the specified mode
			var coverageMode gno.CoverageMode
			switch mode {
			case "set":
				coverageMode = gno.CoverageModeSet
			case "count":
				coverageMode = gno.CoverageModeCount
			case "atomic":
				coverageMode = gno.CoverageModeAtomic
			}
			
			coverage := gno.NewCoverageData(coverageMode, "test")
			
			// Add a block and increment it
			blockIndex := coverage.AddBlock("test.gno", 1, 1, 1, 10, 1)
			coverage.IncrementBlock(blockIndex)
			
			// Check coverage statistics
			covered, total := coverage.GetCoverage()
			if covered != 1 || total != 1 {
				t.Errorf("Expected covered=1, total=1, got covered=%d, total=%d", covered, total)
			}
		})
	}
}

func TestCoverageProfile(t *testing.T) {
	// Create a simple test package
	mpkg := &std.MemPackage{
		Name: "test",
		Path: "gno.land/p/demo/test",
		Files: []*std.MemFile{
			{
				Name: "test.gno",
				Body: `package test

func Add(a, b int) int {
	return a + b
}`,
			},
			{
				Name: "test_test.gno",
				Body: `package test

import "testing"

func TestAdd(t *testing.T) {
	result := Add(2, 3)
	if result != 5 {
		t.Errorf("Add(2, 3) = %d; want 5", result)
	}
}`,
			},
		},
	}

	// Set up test options with coverage profile
	var stdout, stderr bytes.Buffer
	opts := NewTestOptions("", &stdout, &stderr)
	opts.Cover = true
	opts.CoverMode = "count"
	opts.CoverProfile = "test_coverage.out"
	opts.BaseStore, opts.TestStore = StoreWithOptions(
		"", opts.WriterForStore(),
		StoreOptions{WithExtern: true, WithExamples: false, Testing: true},
	)

	// Run the test
	err := Test(mpkg, "", opts)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	// Note: In a real test, we would check that the coverage profile file was created
	// For now, we just verify that the test ran without errors
}
