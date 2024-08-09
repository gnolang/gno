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
			// total branch added offset: 36, 39, 63
			executeBranches:  []int{39},
			expectedCoverage: 0.33,
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
			// total branch added offset: 36, 39, 69
			executeBranches:  []int{39, 69},
			expectedCoverage: 0.67, // 2/3
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
			// total branch added offset: 36, 39, 106, 52, 85
			executeBranches:  []int{39, 52, 85},
			expectedCoverage: 0.60, // 3/5
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
			// total branch added offset: 43, 46, 49, 58 83
			executeBranches:  []int{46},
			expectedCoverage: 0.20, // 1/5
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
			// total branch added offset: 36, 51, 71, 91
			executeBranches:  []int{51, 71, 91},
			expectedCoverage: 0.75, // 3/4
		},
		{
			name: "Function coverage",
			files: []*std.MemFile{
				{
					Name: "test.go",
					Body: `
package main

func foo() int {
    return 1
}

func bar() int {
    return 2
}

func main() {
    foo()
}
`,
				},
			},
			// total branch added offset: 30, 63, 93
			executeBranches:  []int{30, 63},
			expectedCoverage: 0.67,
		},
		{
			name: "For loop",
			files: []*std.MemFile{
				{
					Name: "test.go",
					Body: `
package main

func test() int {
    sum := 0
    for i := 0; i < 5; i++ {
        sum += i
    }
    return sum
}
`,
				},
			},
			executeBranches:  []int{31, 62}, // 함수 시작과 for 루프 조건
			expectedCoverage: 1,
		},
		{
			name: "Range loop",
			files: []*std.MemFile{
				{
					Name: "test.go",
					Body: `
package main

func test() int {
    numbers := []int{1, 2, 3, 4, 5}
    sum := 0
    for _, num := range numbers {
        sum += num
    }
    return sum
}
`,
				},
			},
			executeBranches:  []int{31, 86},
			expectedCoverage: 1.0,
		},
		{
			name: "Defer statement",
			files: []*std.MemFile{
				{
					Name: "test.go",
					Body: `
package main

func test() {
    defer func() {
        recover()
    }()
    panic("test panic")
}
`,
				},
			},
			executeBranches:  []int{27, 33},
			expectedCoverage: 1.0,
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
