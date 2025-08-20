package gnolang

import (
	"sync"
	"testing"
)

func TestCoverageData(t *testing.T) {
	// Test NewCoverageData
	coverage := NewCoverageData(CoverageModeCount, "test/package")
	if coverage.Mode != CoverageModeCount {
		t.Errorf("Expected mode %v, got %v", CoverageModeCount, coverage.Mode)
	}
	if coverage.PkgPath != "test/package" {
		t.Errorf("Expected package path 'test/package', got '%s'", coverage.PkgPath)
	}
}

func TestCoverageBlock(t *testing.T) {
	coverage := NewCoverageData(CoverageModeCount, "test")
	
	// Test AddBlock
	blockIndex := coverage.AddBlock("test.gno", 1, 1, 5, 10, 3)
	if blockIndex != 0 {
		t.Errorf("Expected first block index to be 0, got %d", blockIndex)
	}
	
	if len(coverage.Blocks) != 1 {
		t.Errorf("Expected 1 block, got %d", len(coverage.Blocks))
	}
	
	block := coverage.Blocks[0]
	if block.StartLine != 1 || block.StartCol != 1 || block.EndLine != 5 || block.EndCol != 10 || block.NumStmt != 3 {
		t.Errorf("Block properties incorrect: %+v", block)
	}
}

func TestCoverageIncrement(t *testing.T) {
	testCases := []struct {
		mode     CoverageMode
		name     string
		expected uint64
	}{
		{CoverageModeSet, "set", 1},
		{CoverageModeCount, "count", 3},
		{CoverageModeAtomic, "atomic", 3},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			coverage := NewCoverageData(tc.mode, "test")
			blockIndex := coverage.AddBlock("test.gno", 1, 1, 1, 10, 1)
			
			// Increment the block multiple times
			coverage.IncrementBlock(blockIndex)
			coverage.IncrementBlock(blockIndex)
			coverage.IncrementBlock(blockIndex)
			
			block := coverage.Blocks[blockIndex]
			if tc.mode == CoverageModeSet {
				// Set mode should only be 1 regardless of increments
				if block.Count != 1 {
					t.Errorf("Expected count 1 for set mode, got %d", block.Count)
				}
			} else {
				// Count and atomic modes should increment
				if block.Count != tc.expected {
					t.Errorf("Expected count %d for %s mode, got %d", tc.expected, tc.name, block.Count)
				}
			}
		})
	}
}

func TestCoverageGetCoverage(t *testing.T) {
	coverage := NewCoverageData(CoverageModeCount, "test")
	
	// Add some blocks
	block1 := coverage.AddBlock("test.gno", 1, 1, 1, 10, 2)
	_ = coverage.AddBlock("test.gno", 2, 1, 2, 10, 3)
	block3 := coverage.AddBlock("test.gno", 3, 1, 3, 10, 1)
	
	// Initially no coverage
	covered, total := coverage.GetCoverage()
	if covered != 0 || total != 6 {
		t.Errorf("Expected covered=0, total=6, got covered=%d, total=%d", covered, total)
	}
	
	// Increment some blocks
	coverage.IncrementBlock(block1)
	coverage.IncrementBlock(block3)
	
	// Now should have coverage
	covered, total = coverage.GetCoverage()
	if covered != 3 || total != 6 {
		t.Errorf("Expected covered=3, total=6, got covered=%d, total=%d", covered, total)
	}
}

func TestCoverageThreadSafety(t *testing.T) {
	coverage := NewCoverageData(CoverageModeAtomic, "test")
	blockIndex := coverage.AddBlock("test.gno", 1, 1, 1, 10, 1)
	
	// Run concurrent increments
	var wg sync.WaitGroup
	numGoroutines := 100
	incrementsPerGoroutine := 10
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				coverage.IncrementBlock(blockIndex)
			}
		}()
	}
	
	wg.Wait()
	
	// Check that all increments were counted
	block := coverage.Blocks[blockIndex]
	expected := uint64(numGoroutines * incrementsPerGoroutine)
	if block.Count != expected {
		t.Errorf("Expected count %d, got %d", expected, block.Count)
	}
}

func TestCoverageModeString(t *testing.T) {
	testCases := []struct {
		mode     CoverageMode
		expected string
	}{
		{CoverageModeSet, "set"},
		{CoverageModeCount, "count"},
		{CoverageModeAtomic, "atomic"},
	}
	
	for _, tc := range testCases {
		if tc.mode.String() != tc.expected {
			t.Errorf("Expected %s, got %s", tc.expected, tc.mode.String())
		}
	}
}
