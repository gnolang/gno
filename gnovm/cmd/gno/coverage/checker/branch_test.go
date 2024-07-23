package checker

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestBranchCoverage(t *testing.T) {
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
			// total branch added offset: 39, 63
			executeBranches:  []int{39},
			expectedCoverage: 0.5, // 1/2
		},
		{
			name: "If statement with else",
			files: []*std.MemFile{
				{
					Name: "test.go",
					Body: `
package main

func test(x int) int {
	if x > 0 {
		return x
	} else {
		return -x
	}
}`,
				},
			},
			executeBranches:  []int{39, 69},
			expectedCoverage: 1.0,
		},
		{
			name: "Nested if statement",
			files: []*std.MemFile{
				{
					Name: "test.go",
					Body: `
package main

func test(x int) int {
	if x > 0 {
		if x < 10 {
			return x
		} else {
			return 10
		}
	}
	return -x
}
`,
				},
			},
			// total branch added offset: 39, 106, 52, 85
			executeBranches:  []int{39, 52, 85},
			expectedCoverage: 0.75, // 3/4
		},
		{
			name: "Multiple conditions",
			files: []*std.MemFile{
				{
					Name: "test.go",
					Body: `
package main

func test(x int, y int) int {
	if x > 0 && y > 0 {
		return x + y
	}
	return -x
}
`,
				},
			},
			// total branch added offset: 46, 83
			executeBranches:  []int{46},
			expectedCoverage: 0.5, // 1/2
		},
		{
			name: "Switch statement",
			files: []*std.MemFile{
				{
					Name: "test.go",
					Body: `
package main

func test(x int) int {
	switch x {
	case 1:
		return 1
	case 2:
		return 2
	default:
		return 0
	}
}
`,
				},
			},
			executeBranches:  []int{51, 71, 91},
			expectedCoverage: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := NewBranchCoverage(tt.files)

			for i, file := range tt.files {
				tt.files[i] = bc.Instrument(file)
			}

			for _, offset := range tt.executeBranches {
				bc.MarkBranchTaken(offset)
			}

			coverage := bc.CalculateCoverage()
			if !almostEqual(coverage, tt.expectedCoverage, 0.01) {
				t.Errorf("Expected coverage %f, but got %f", tt.expectedCoverage, coverage)
			}
		})
	}
}
