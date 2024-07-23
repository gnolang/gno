package checker

import (
	"strconv"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestBranchCoverage(t *testing.T) {
	t.Skip("TODO")
	tests := []struct {
		name             string
		files            []*std.MemFile
		executeBranches  []int
		expectedCoverage float64
	}{
		{
			name: "Simple if statement",
			files: []*std.MemFile{
				{
					Name: "test.go",
					Body: `
package main

func test(x int) int {
	if x > 0 {
		return x
	}
	return -x
}
`,
				},
			},
			executeBranches:  []int{5},
			expectedCoverage: 0.5, // 1 out of 2 branches covered
		},
		// TODO: add more tests
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := NewBranchCoverage(tt.files)

			for i, file := range tt.files {
				tt.files[i] = bc.Instrument(file)
			}

			coverage := bc.CalculateCoverage()
			if !almostEqual(coverage, tt.expectedCoverage, 0.01) {
				t.Errorf("Expected coverage %f, but got %f", tt.expectedCoverage, coverage)
			}
		})
	}
}

func stringToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
