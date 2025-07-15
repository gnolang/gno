package coverage

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// Tracker tracks the coverage data for multiple files.
// Maintains invariants I1-I4 from the axiom system
type Tracker struct {
	data     map[string]map[int]int  // filename -> line number -> execution count (I2: non-negative)
	allLines map[string]map[int]bool // filename -> line number -> is executable (I1: inclusion relationship)
}

func NewTracker() *Tracker {
	return &Tracker{
		data:     make(map[string]map[int]int),
		allLines: make(map[string]map[int]bool),
	}
}

// MarkLine satisfies Invariant I2: non-negative execution counts
func (ct *Tracker) MarkLine(filename string, line int) {
	if _, ok := ct.data[filename]; !ok {
		ct.data[filename] = make(map[int]int)
	}
	ct.data[filename][line]++

	// Ensure Invariant I1: executed lines are in allLines
	ct.RegisterExecutableLine(filename, line)
}

// RegisterExecutableLine maintains Invariant I1: inclusion relationship
func (ct *Tracker) RegisterExecutableLine(filename string, line int) {
	if _, ok := ct.allLines[filename]; !ok {
		ct.allLines[filename] = make(map[int]bool)
	}
	ct.allLines[filename][line] = true
}

// RegisterFile parses a file and registers all executable lines without instrumentation
func (ct *Tracker) RegisterFile(filename string, content []byte) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parsing file %s: %w", filename, err)
	}

	// Walk the AST to find executable lines
	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		// Register lines for executable nodes
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Body != nil {
				line := fset.Position(node.Body.Lbrace).Line
				ct.RegisterExecutableLine(filename, line)
			}
		case *ast.IfStmt:
			line := fset.Position(node.If).Line
			ct.RegisterExecutableLine(filename, line)
		case *ast.ForStmt:
			line := fset.Position(node.For).Line
			ct.RegisterExecutableLine(filename, line)
		case *ast.RangeStmt:
			line := fset.Position(node.For).Line
			ct.RegisterExecutableLine(filename, line)
		case *ast.SwitchStmt:
			line := fset.Position(node.Switch).Line
			ct.RegisterExecutableLine(filename, line)
		case *ast.TypeSwitchStmt:
			line := fset.Position(node.Switch).Line
			ct.RegisterExecutableLine(filename, line)
		case *ast.SelectStmt:
			line := fset.Position(node.Select).Line
			ct.RegisterExecutableLine(filename, line)
		case *ast.CaseClause:
			if len(node.Body) > 0 {
				line := fset.Position(node.Body[0].Pos()).Line
				ct.RegisterExecutableLine(filename, line)
			}
		case *ast.AssignStmt, *ast.ExprStmt, *ast.ReturnStmt, *ast.BranchStmt:
			line := fset.Position(node.Pos()).Line
			ct.RegisterExecutableLine(filename, line)
		}
		return true
	})

	return nil
}

// ValidateInvariants checks that all coverage invariants hold
func (ct *Tracker) ValidateInvariants() error {
	for filename, executedLines := range ct.data {
		registeredLines, exists := ct.allLines[filename]
		if !exists {
			return fmt.Errorf("invariant I1 violated: executed lines exist but no registered lines for file %s", filename)
		}

		for line, count := range executedLines {
			// Invariant I2: non-negative counts
			if count < 0 {
				return fmt.Errorf("invariant I2 violated: negative execution count %d for line %d in file %s", count, line, filename)
			}

			// Invariant I1: executed lines must be registered
			if !registeredLines[line] {
				return fmt.Errorf("invariant I1 violated: line %d executed but not registered in file %s", line, filename)
			}
		}
	}
	return nil
}

// GetCoverage returns the coverage data for a specific file
func (ct *Tracker) GetCoverage(filename string) map[int]int {
	return ct.data[filename]
}

// GetCoverageData returns coverage data ensuring Invariant I3
func (ct *Tracker) GetCoverageData() map[string]*CoverageData {
	result := make(map[string]*CoverageData)

	for filename, executableLines := range ct.allLines {
		totalLines := len(executableLines)
		coveredLines := 0
		lineData := make(map[int]int)

		for line := range executableLines {
			if executedData, ok := ct.data[filename]; ok {
				if count, executed := executedData[line]; executed {
					lineData[line] = count
					if count > 0 {
						coveredLines++
					}
				} else {
					lineData[line] = 0
				}
			} else {
				lineData[line] = 0
			}
		}

		// Ensure Invariant I3: coverage ratio in [0, 100]
		coverageRatio := 0.0
		if totalLines > 0 {
			coverageRatio = float64(coveredLines) / float64(totalLines) * 100
			if coverageRatio < 0 {
				coverageRatio = 0
			}
			if coverageRatio > 100 {
				coverageRatio = 100
			}
		}

		result[filename] = &CoverageData{
			TotalLines:    totalLines,
			CoveredLines:  coveredLines,
			CoverageRatio: coverageRatio,
			LineData:      lineData,
		}
	}

	return result
}

// Reset resets the coverage data while maintaining invariants
func (ct *Tracker) Reset() {
	ct.data = make(map[string]map[int]int)
	ct.allLines = make(map[string]map[int]bool)
}

// PrintCoverage prints the coverage data to stdout
func (ct *Tracker) PrintCoverage() {
	coverageData := ct.GetCoverageData()

	var totalLines, totalCovered int
	for _, data := range coverageData {
		totalLines += data.TotalLines
		totalCovered += data.CoveredLines
	}

	overallCoverage := 0.0
	if totalLines > 0 {
		overallCoverage = float64(totalCovered) / float64(totalLines) * 100
	}

	fmt.Printf("\nCoverage Report:\n")
	fmt.Printf("Total Lines: %d\n", totalLines)
	fmt.Printf("Covered Lines: %d\n", totalCovered)
	fmt.Printf("Overall Coverage: %.2f%%\n\n", overallCoverage)

	for filename, data := range coverageData {
		fmt.Printf("File: %s\n", filename)
		fmt.Printf("  Total Lines: %d\n", data.TotalLines)
		fmt.Printf("  Covered Lines: %d\n", data.CoveredLines)
		fmt.Printf("  Coverage: %.2f%%\n", data.CoverageRatio)
		fmt.Println()
	}
}
