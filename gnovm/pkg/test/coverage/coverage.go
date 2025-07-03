package coverage

import (
	"fmt"
	"go/ast"
	"go/format" // Using format instead of printer to properly handle comments
	"go/parser"
	"go/token"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/std"
)

var globalTracker = NewCoverageTracker()

// CoverageTracker tracks the coverage data
type CoverageTracker struct {
	data     map[string]map[int]int  // filename -> line number -> count
	allLines map[string]map[int]bool // filename -> line number -> exists (all executable lines)
}

func NewCoverageTracker() *CoverageTracker {
	return &CoverageTracker{
		data:     make(map[string]map[int]int),
		allLines: make(map[string]map[int]bool),
	}
}

// MarkLine mark the line as executed
func (ct *CoverageTracker) MarkLine(filename string, line int) {
	if _, ok := ct.data[filename]; !ok {
		ct.data[filename] = make(map[int]int)
	}
	ct.data[filename][line]++
}

// RegisterExecutableLine registers a line as executable (for coverage calculation)
func (ct *CoverageTracker) RegisterExecutableLine(filename string, line int) {
	if _, ok := ct.allLines[filename]; !ok {
		ct.allLines[filename] = make(map[int]bool)
	}
	ct.allLines[filename][line] = true
}

// GetCoverage return the coverage data for a specific file
func (ct *CoverageTracker) GetCoverage(filename string) map[int]int {
	return ct.data[filename]
}

// CoverageData represents the coverage data for a file
type CoverageData struct {
	TotalLines    int
	CoveredLines  int
	CoverageRatio float64
	LineData      map[int]int // line number -> execution count
}

// GetCoverageData returns the coverage data for all files
func (ct *CoverageTracker) GetCoverageData() map[string]*CoverageData {
	result := make(map[string]*CoverageData)

	// Process all files that have executable lines
	for filename, executableLines := range ct.allLines {
		totalLines := len(executableLines)
		coveredLines := 0

		lineData := make(map[int]int)

		// Check coverage for each executable line
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

		coverageRatio := 0.0
		if totalLines > 0 {
			coverageRatio = float64(coveredLines) / float64(totalLines) * 100
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

// PrintCoverage prints the coverage data to stdout
func (ct *CoverageTracker) PrintCoverage() {
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

// CoverageInstrumenter instrument the AST to add coverage
type CoverageInstrumenter struct {
	fset     *token.FileSet
	tracker  *CoverageTracker
	filename string
}

// NewCoverageInstrumenter create a new CoverageInstrumenter
func NewCoverageInstrumenter(tracker *CoverageTracker, filename string) *CoverageInstrumenter {
	return &CoverageInstrumenter{
		fset:     token.NewFileSet(),
		tracker:  tracker,
		filename: filename,
	}
}

// InstrumentFile instrument the file by adding coverage tracking calls.
// It uses parser.ParseComments to preserve all comments including multiline /* */ comments,
// and format.Node to properly format the output, ensuring comments remain in their correct positions.
func (ci *CoverageInstrumenter) InstrumentFile(content []byte) ([]byte, error) {
	// parse the file with comments preserved
	f, err := parser.ParseFile(ci.fset, ci.filename, string(content), parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing failed: %w", err)
	}

	// Check if the file contains any usage of 'cross' identifier
	var containsCross bool
	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "cross" {
			containsCross = true
			return false
		}
		return true
	})

	if containsCross {
		// If the file contains 'cross', skip instrumentation but register all executable lines
		// This avoids preprocessing issues with the special 'cross' identifier
		ast.Inspect(f, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				if node.Body != nil {
					funcLine := ci.fset.Position(node.Body.Lbrace).Line
					ci.tracker.RegisterExecutableLine(ci.filename, funcLine)
					// Register all statement lines in the function
					for _, stmt := range node.Body.List {
						if stmt != nil {
							stmtLine := ci.fset.Position(stmt.Pos()).Line
							ci.tracker.RegisterExecutableLine(ci.filename, stmtLine)
						}
					}
				}
			case *ast.IfStmt:
				line := ci.fset.Position(node.Pos()).Line
				ci.tracker.RegisterExecutableLine(ci.filename, line)
			case *ast.ForStmt:
				line := ci.fset.Position(node.Pos()).Line
				ci.tracker.RegisterExecutableLine(ci.filename, line)
			case *ast.RangeStmt:
				line := ci.fset.Position(node.Pos()).Line
				ci.tracker.RegisterExecutableLine(ci.filename, line)
			case *ast.SwitchStmt:
				line := ci.fset.Position(node.Pos()).Line
				ci.tracker.RegisterExecutableLine(ci.filename, line)
			case *ast.CaseClause:
				line := ci.fset.Position(node.Pos()).Line
				ci.tracker.RegisterExecutableLine(ci.filename, line)
			case *ast.ReturnStmt:
				line := ci.fset.Position(node.Pos()).Line
				ci.tracker.RegisterExecutableLine(ci.filename, line)
			}
			return true
		})
		// Return original content without instrumentation
		return content, nil
	}

	// add testing import if not already present
	hasTestingImport := false
	for _, imp := range f.Imports {
		if imp.Path.Value == "\"testing\"" {
			hasTestingImport = true
			break
		}
	}
	if !hasTestingImport {
		// Create a new import declaration
		importDecl := &ast.GenDecl{
			Tok: token.IMPORT,
			Specs: []ast.Spec{
				&ast.ImportSpec{
					Path: &ast.BasicLit{
						Kind:  token.STRING,
						Value: "\"testing\"",
					},
				},
			},
		}

		// If there are existing imports, add testing to the first import declaration
		if len(f.Imports) > 0 {
			// Find the first import declaration
			for _, decl := range f.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
					// Add testing to the existing import declaration
					genDecl.Specs = append(genDecl.Specs, importDecl.Specs[0])
					break
				}
			}
		} else {
			// Add the import declaration to the beginning of the file
			if len(f.Decls) > 0 {
				f.Decls = append([]ast.Decl{importDecl}, f.Decls...)
			} else {
				f.Decls = []ast.Decl{importDecl}
			}
		}
	}

	// modify the AST
	ast.Walk(ci, f)

	// convert the modified AST to code
	// Use format.Node instead of printer.Fprint to handle comments properly.
	// format.Node preserves the original comment positions and applies proper Go formatting,
	// preventing syntax errors that can occur when multiline /* */ comments are present.
	var buf strings.Builder
	if err := format.Node(&buf, ci.fset, f); err != nil {
		return nil, fmt.Errorf("code generation failed: %w", err)
	}

	return []byte(buf.String()), nil
}

// createMarkLineStmt creates an ast.ExprStmt that records the given filename and line number
func (ci *CoverageInstrumenter) createMarkLineStmt(filename string, line int) ast.Stmt {
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "testing"},
				Sel: &ast.Ident{Name: "MarkLine"},
			},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", filename)},
				&ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", line)},
			},
		},
	}
}

// instrumentBlockStmt adds a call to the markLine function to the BlockStmt
func (ci *CoverageInstrumenter) instrumentBlockStmt(block *ast.BlockStmt, line int) {
	if block == nil {
		return
	}

	// Check if first statement contains cross identifier
	if len(block.List) > 0 && ci.statementContainsCross(block.List[0]) {
		// Register all lines as executable but don't instrument
		// to avoid preprocessing issues with the 'cross' keyword
		for _, stmt := range block.List {
			if stmt != nil {
				stmtLine := ci.fset.Position(stmt.Pos()).Line
				ci.tracker.RegisterExecutableLine(ci.filename, stmtLine)
			}
		}
		return
	}

	markStmt := ci.createMarkLineStmt(ci.filename, line)
	block.List = append([]ast.Stmt{markStmt}, block.List...)

	// Note: return statement handling is done in Visit method to avoid duplication
}

// statementContainsCross checks if a statement contains the cross identifier
func (ci *CoverageInstrumenter) statementContainsCross(stmt ast.Stmt) bool {
	var containsCross bool
	ast.Inspect(stmt, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "cross" {
			containsCross = true
			return false
		}
		return true
	})
	return containsCross
}

// instrumentCaseList adds a call to the markLine function to the case list
func (ci *CoverageInstrumenter) instrumentCaseStmts(body []ast.Stmt, line int) []ast.Stmt {
	markStmt := ci.createMarkLineStmt(ci.filename, line)
	return append([]ast.Stmt{markStmt}, body...)
}

// Visit visit the AST node and add coverage
func (ci *CoverageInstrumenter) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	// get the position information of the node
	pos := ci.fset.Position(node.Pos())
	line := pos.Line

	// only instrument executable nodes
	switch n := node.(type) {
	case *ast.FuncDecl:
		if n.Body != nil {
			// Register the function entry line
			funcLine := ci.fset.Position(n.Body.Lbrace).Line
			ci.tracker.RegisterExecutableLine(ci.filename, funcLine)
			ci.instrumentBlockStmt(n.Body, funcLine)
		}
	case *ast.IfStmt:
		if n.Cond != nil {
			condLine := ci.fset.Position(n.Cond.Pos()).Line
			ci.tracker.RegisterExecutableLine(ci.filename, condLine)
			ci.instrumentBlockStmt(n.Body, condLine)
		}
		// Also instrument else block if present
		if n.Else != nil {
			if elseBlock, ok := n.Else.(*ast.BlockStmt); ok {
				elseLine := ci.fset.Position(elseBlock.Lbrace).Line
				ci.tracker.RegisterExecutableLine(ci.filename, elseLine)
				ci.instrumentBlockStmt(elseBlock, elseLine)
			}
		}
	case *ast.ForStmt:
		if n.Cond != nil {
			condLine := ci.fset.Position(n.Cond.Pos()).Line
			ci.tracker.RegisterExecutableLine(ci.filename, condLine)
			ci.instrumentBlockStmt(n.Body, condLine)
		} else {
			// For loop without condition
			ci.tracker.RegisterExecutableLine(ci.filename, line)
			ci.instrumentBlockStmt(n.Body, line)
		}
	case *ast.RangeStmt:
		ci.tracker.RegisterExecutableLine(ci.filename, line)
		ci.instrumentBlockStmt(n.Body, line)
	case *ast.SwitchStmt:
		ci.tracker.RegisterExecutableLine(ci.filename, line)
		// Don't instrument the switch body directly - only instrument case clauses
	case *ast.SelectStmt:
		ci.tracker.RegisterExecutableLine(ci.filename, line)
		// Don't instrument the select body directly - only instrument case clauses
	case *ast.CaseClause:
		ci.tracker.RegisterExecutableLine(ci.filename, line)
		n.Body = ci.instrumentCaseStmts(n.Body, line)
	case *ast.CommClause:
		ci.tracker.RegisterExecutableLine(ci.filename, line)
		n.Body = ci.instrumentCaseStmts(n.Body, line)
	case *ast.ReturnStmt:
		// Just register the line as executable, don't modify the AST here
		// The return will be covered by the block instrumentation
		ci.tracker.RegisterExecutableLine(ci.filename, line)
	}

	return ci
}

// InstrumentPackage instrument the package
func InstrumentPackage(pkg *std.MemPackage) error {
	for _, file := range pkg.Files {
		// skip test files
		if strings.HasSuffix(file.Name, "_test.gno") {
			continue
		}

		if !strings.HasSuffix(file.Name, ".gno") {
			continue
		}

		instrumenter := NewCoverageInstrumenter(globalTracker, file.Name)
		instrumented, err := instrumenter.InstrumentFile([]byte(file.Body))
		if err != nil {
			return fmt.Errorf("failed to instrument file %s: %w", file.Name, err)
		}
		file.Body = string(instrumented)
	}
	return nil
}

// GetGlobalTracker returns the global coverage tracker
func GetGlobalTracker() *CoverageTracker {
	return globalTracker
}

// Reset resets the coverage data
func (ct *CoverageTracker) Reset() {
	ct.data = make(map[string]map[int]int)
	ct.allLines = make(map[string]map[int]bool)
}
