package checker

import (
	"fmt"
	"go/ast"
	"math"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestStatementCoverage(t *testing.T) {
	tests := []struct {
		name             string
		files            []*std.MemFile
		executeLines     []int
		expectedCoverage float64
	}{
		{
			name: "Simple function",
			files: []*std.MemFile{
				{
					Name: "test.go",
					Body: `package main // 1
                                        // 2
func test() {                           // 3
	a := 1                              // 4
	b := 2                              // 5
	c := a + b                          // 6
}                                       // 7
`,
				},
			},
			executeLines:     []int{4, 5, 6},
			expectedCoverage: 1.0,
		},
		{
			name: "Function with if statement",
			files: []*std.MemFile{
				{
					Name: "test.go",
					Body: `package main // 1
                                        // 2
func test(x int) int {                  // 3
	if x > 0 {                          // 4
		return x                        // 5
	}                                   // 6
	return -x                           // 7
}                                       // 8
`,
				},
			},
			executeLines:     []int{4, 5},
			expectedCoverage: 0.67,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := NewStatementCoverage(tt.files)

			// Instrument files
			for i, file := range tt.files {
				tt.files[i] = sc.Instrument(file)
			}

			// Simulate execution by marking covered lines
			for _, line := range tt.executeLines {
				for _, file := range sc.files {
					ast.Inspect(file, func(n ast.Node) bool {
						if stmt, ok := n.(ast.Stmt); ok {
							pos := sc.fset.Position(stmt.Pos())
							if pos.Line == line {
								sc.MarkCovered(stmt.Pos())
								fmt.Printf("Marked line %d in file %s\n", line, file.Name)
								return false
							}
						}
						return true
					})
				}
			}

			coverage := sc.CalculateCoverage()
			if !almostEqual(coverage, tt.expectedCoverage, 0.01) {
				t.Errorf("Expected coverage %f, but got %f", tt.expectedCoverage, coverage)
			}
		})
	}
}

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}
