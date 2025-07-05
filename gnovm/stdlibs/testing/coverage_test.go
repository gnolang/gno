package testing

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/test/coverage"
)

func TestMarkLine(t *testing.T) {
	coverageTracker = coverage.NewTracker()

	// Test marking a line
	X_markLine("test.gno", 42)

	// Check if the line was marked
	coverageData := coverageTracker.GetCoverage("test.gno")
	if count, ok := coverageData[42]; !ok || count != 1 {
		t.Errorf("Expected line 42 to be marked once, got %v", coverageData)
	}

	// Test marking the same line again
	X_markLine("test.gno", 42)

	// Check if the line count increased
	coverageData = coverageTracker.GetCoverage("test.gno")
	if count, ok := coverageData[42]; !ok || count != 2 {
		t.Errorf("Expected line 42 to be marked twice, got %v", coverageData)
	}

	// Test marking a different line
	X_markLine("test.gno", 100)

	// Check if both lines are marked
	coverageData = coverageTracker.GetCoverage("test.gno")
	if count, ok := coverageData[42]; !ok || count != 2 {
		t.Errorf("Expected line 42 to be marked twice, got %v", coverageData)
	}
	if count, ok := coverageData[100]; !ok || count != 1 {
		t.Errorf("Expected line 100 to be marked once, got %v", coverageData)
	}
}

func TestInstrumentCode(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		filename string
		want     string
	}{
		{
			name: "simple function",
			code: `package main

func test() int {
	return 42
}`,
			filename: "test.gno",
			want:     "testing.MarkLine",
		},
		{
			name: "if statement",
			code: `package main

func test(x int) int {
	if x > 0 {
		return 1
	}
	return 0
}`,
			filename: "test.gno",
			want:     "testing.MarkLine",
		},
		{
			name: "for loop",
			code: `package main

func test() int {
	for i := 0; i < 10; i++ {
		return i
	}
	return 0
}`,
			filename: "test.gno",
			want:     "testing.MarkLine",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := X_instrumentCode(tt.code, tt.filename)
			if !strings.Contains(got, tt.want) {
				t.Errorf("X_instrumentCode() = %v, want %v", got, tt.want)
			}
		})
	}
}
