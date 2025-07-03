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

// CoverageTracker tracks the coverage data for multiple files.
// It maintains two maps: one for execution counts and one for all executable lines.
type CoverageTracker struct {
	data     map[string]map[int]int  // filename -> line number -> execution count
	allLines map[string]map[int]bool // filename -> line number -> is executable
}

func NewCoverageTracker() *CoverageTracker {
	return &CoverageTracker{
		data:     make(map[string]map[int]int),
		allLines: make(map[string]map[int]bool),
	}
}

// MarkLine marks the line as executed and increments its execution count.
// This method is called during test execution to track coverage.
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

// GetCoverage returns the coverage data for a specific file.
// Returns nil if no coverage data exists for the file.
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
	// Cache for checking cross identifier to avoid repeated AST traversals
	containsCrossCache map[ast.Node]bool
}

// NewCoverageInstrumenter create a new CoverageInstrumenter
func NewCoverageInstrumenter(tracker *CoverageTracker, filename string) *CoverageInstrumenter {
	return &CoverageInstrumenter{
		fset:               token.NewFileSet(),
		tracker:            tracker,
		filename:           filename,
		containsCrossCache: make(map[ast.Node]bool),
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
	if ci.fileContainsCross(f) {
		// If the file contains 'cross', skip instrumentation but register all executable lines
		// This avoids preprocessing issues with the special 'cross' identifier
		ci.registerExecutableLines(f)
		// Return original content without instrumentation
		return content, nil
	}

	// add testing import if not already present
	if err := ci.ensureTestingImport(f); err != nil {
		return nil, err
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

// instrumentBlockStmt adds coverage tracking to a block statement.
// It inserts a MarkLine call at the beginning of the block unless
// the block contains the special 'cross' identifier.
func (ci *CoverageInstrumenter) instrumentBlockStmt(block *ast.BlockStmt, line int) {
	if block == nil {
		return
	}

	// even if the block is empty, insert `MarkLine` at the beginning of the block
	if len(block.List) == 0 {
		markStmt := ci.createMarkLineStmt(ci.filename, line)
		block.List = append([]ast.Stmt{markStmt}, block.List...)
		return
	}

	// Check if first statement contains cross identifier
	if ci.statementContainsCross(block.List[0]) {
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
}

// fileContainsCross checks if the file contains the cross identifier
func (ci *CoverageInstrumenter) fileContainsCross(f *ast.File) bool {
	var containsCross bool
	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "cross" {
			containsCross = true
			return false
		}
		return true
	})
	return containsCross
}

// statementContainsCross checks if a statement contains the cross identifier
func (ci *CoverageInstrumenter) statementContainsCross(stmt ast.Stmt) bool {
	if cached, ok := ci.containsCrossCache[stmt]; ok {
		return cached
	}

	var containsCross bool
	ast.Inspect(stmt, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "cross" {
			containsCross = true
			return false
		}
		return true
	})

	ci.containsCrossCache[stmt] = containsCross
	return containsCross
}

// ensureTestingImport adds the testing import if not already present
func (ci *CoverageInstrumenter) ensureTestingImport(f *ast.File) error {
	for _, imp := range f.Imports {
		if imp.Path.Value == "\"testing\"" {
			return nil
		}
	}

	// Create a new import spec
	importSpec := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: "\"testing\"",
		},
	}

	// Find existing import declaration or create new one
	for _, decl := range f.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			genDecl.Specs = append(genDecl.Specs, importSpec)
			return nil
		}
	}

	// No existing import declaration, create a new one
	importDecl := &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: []ast.Spec{importSpec},
	}

	f.Decls = append([]ast.Decl{importDecl}, f.Decls...)
	return nil
}

// registerExecutableLines registers all executable lines without instrumenting
func (ci *CoverageInstrumenter) registerExecutableLines(f *ast.File) {
	ast.Inspect(f, func(n ast.Node) bool {
		ci.registerNodeIfExecutable(n)
		return true
	})
}

// registerNodeIfExecutable registers a node's line if it's an executable statement
func (ci *CoverageInstrumenter) registerNodeIfExecutable(n ast.Node) {
	if n == nil {
		return
	}

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
	case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.SelectStmt,
		*ast.CaseClause, *ast.CommClause, *ast.ReturnStmt:
		line := ci.fset.Position(node.Pos()).Line
		ci.tracker.RegisterExecutableLine(ci.filename, line)
	}
}

// instrumentCaseList adds a call to the markLine function to the case list
func (ci *CoverageInstrumenter) instrumentCaseStmts(body []ast.Stmt, line int) []ast.Stmt {
	markStmt := ci.createMarkLineStmt(ci.filename, line)
	return append([]ast.Stmt{markStmt}, body...)
}

// getLine returns the line number for a given position
func (ci *CoverageInstrumenter) getLine(pos token.Pos) int {
	return ci.fset.Position(pos).Line
}

// registerAndInstrument registers a line as executable and instruments the block
func (ci *CoverageInstrumenter) registerAndInstrument(block *ast.BlockStmt, line int) {
	ci.tracker.RegisterExecutableLine(ci.filename, line)
	ci.instrumentBlockStmt(block, line)
}

// Visit visit the AST node and add coverage
func (ci *CoverageInstrumenter) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	// only instrument executable nodes
	switch n := node.(type) {
	case *ast.FuncDecl:
		ci.instrumentFuncDecl(n)
	case *ast.IfStmt:
		ci.instrumentIfStmt(n)
	case *ast.ForStmt:
		ci.instrumentForStmt(n)
	case *ast.RangeStmt:
		ci.instrumentRangeStmt(n)
	case *ast.SwitchStmt:
		ci.instrumentSwitchStmt(n)
	case *ast.SelectStmt:
		ci.instrumentSelectStmt(n)
	case *ast.CaseClause:
		ci.instrumentCaseClause(n)
	case *ast.CommClause:
		ci.instrumentCommClause(n)
	case *ast.ReturnStmt:
		ci.instrumentReturnStmt(n)
	}

	return ci
}

// instrumentFuncDecl instruments a function declaration
func (ci *CoverageInstrumenter) instrumentFuncDecl(n *ast.FuncDecl) {
	if n.Body != nil {
		funcLine := ci.getLine(n.Body.Lbrace)
		ci.registerAndInstrument(n.Body, funcLine)
	}
}

// instrumentIfStmt instruments an if statement
func (ci *CoverageInstrumenter) instrumentIfStmt(n *ast.IfStmt) {
	if n.Cond != nil {
		condLine := ci.getLine(n.Cond.Pos())
		ci.registerAndInstrument(n.Body, condLine)
	}
	// Also instrument else block if present
	if n.Else != nil {
		if elseBlock, ok := n.Else.(*ast.BlockStmt); ok {
			elseLine := ci.getLine(elseBlock.Lbrace)
			ci.registerAndInstrument(elseBlock, elseLine)
		}
	}
}

// instrumentForStmt instruments a for statement
func (ci *CoverageInstrumenter) instrumentForStmt(n *ast.ForStmt) {
	var line int
	if n.Cond != nil {
		line = ci.getLine(n.Cond.Pos())
	} else {
		line = ci.getLine(n.Pos())
	}
	ci.registerAndInstrument(n.Body, line)
}

// instrumentRangeStmt instruments a range statement
func (ci *CoverageInstrumenter) instrumentRangeStmt(n *ast.RangeStmt) {
	line := ci.getLine(n.Pos())
	ci.registerAndInstrument(n.Body, line)
}

// instrumentSwitchStmt instruments a switch statement
func (ci *CoverageInstrumenter) instrumentSwitchStmt(n *ast.SwitchStmt) {
	line := ci.getLine(n.Pos())
	ci.tracker.RegisterExecutableLine(ci.filename, line)
	// Don't instrument the switch body directly - only instrument case clauses
}

// instrumentSelectStmt instruments a select statement
func (ci *CoverageInstrumenter) instrumentSelectStmt(n *ast.SelectStmt) {
	line := ci.getLine(n.Pos())
	ci.tracker.RegisterExecutableLine(ci.filename, line)
	// Don't instrument the select body directly - only instrument case clauses
}

// instrumentCaseClause instruments a case clause
func (ci *CoverageInstrumenter) instrumentCaseClause(n *ast.CaseClause) {
	line := ci.getLine(n.Pos())
	ci.tracker.RegisterExecutableLine(ci.filename, line)
	n.Body = ci.instrumentCaseStmts(n.Body, line)
}

// instrumentCommClause instruments a comm clause
func (ci *CoverageInstrumenter) instrumentCommClause(n *ast.CommClause) {
	line := ci.getLine(n.Pos())
	ci.tracker.RegisterExecutableLine(ci.filename, line)
	n.Body = ci.instrumentCaseStmts(n.Body, line)
}

// instrumentReturnStmt instruments a return statement
func (ci *CoverageInstrumenter) instrumentReturnStmt(n *ast.ReturnStmt) {
	line := ci.getLine(n.Pos())
	ci.tracker.RegisterExecutableLine(ci.filename, line)
}

// InstrumentPackage instruments all non-test .gno files in a package for coverage tracking.
// It skips test files and non-.gno files to avoid instrumentation conflicts.
func InstrumentPackage(pkg *std.MemPackage) error {
	if pkg == nil {
		return fmt.Errorf("package is nil")
	}

	for _, file := range pkg.Files {
		// skip test files and non-.gno files
		if !shouldInstrumentFile(file.Name) {
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

// shouldInstrumentFile determines if a file should be instrumented for coverage
func shouldInstrumentFile(filename string) bool {
	// Skip test files
	if strings.HasSuffix(filename, "_test.gno") || strings.HasSuffix(filename, "_filetest.gno") {
		return false
	}
	// Only instrument .gno files
	return strings.HasSuffix(filename, ".gno")
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
