package coverage

import (
	"fmt"
	"go/ast"
	"go/format"
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
	// Cache for checking external instrumentation to avoid repeated AST traversals
	containsExternalCache map[ast.Node]bool
	// Track statements that need individual instrumentation
	pendingInstrumentations map[ast.Stmt]ast.Stmt
	// Track the current block context for statement instrumentation
	currentBlock *ast.BlockStmt
}

// NewCoverageInstrumenter create a new CoverageInstrumenter
func NewCoverageInstrumenter(tracker *CoverageTracker, filename string) *CoverageInstrumenter {
	return &CoverageInstrumenter{
		fset:                    token.NewFileSet(),
		tracker:                 tracker,
		filename:                filename,
		containsExternalCache:   make(map[ast.Node]bool),
		pendingInstrumentations: make(map[ast.Stmt]ast.Stmt),
	}
}

// InstrumentFile instrument the file by adding coverage tracking calls.
// It uses `parser.ParseComments` to preserve all comments including multiline `/* */` comments,
// and `format.Node` to properly format the output, ensuring comments remain in their correct positions.
func (ci *CoverageInstrumenter) InstrumentFile(content []byte) ([]byte, error) {
	// parse the file with comments preserved
	f, err := parser.ParseFile(ci.fset, ci.filename, string(content), parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing failed: %w", err)
	}

	// Check if the file is externally instrumented
	if ci.isExternallyInstrumented(f) {
		// If the file is externally instrumented, skip instrumentation but register all executable lines
		// This avoids preprocessing issues with external instrumentation systems
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

	// Post-process to apply statement-level instrumentation
	ci.applyStatementInstrumentations(f)

	// convert the modified AST to code
	// Use `format.Node` instead of `printer.Fprint` to handle comments properly.
	// `format.Node` preserves the original comment positions and applies proper formatting,
	// preventing syntax errors that can occur when multiline `/* */` comments are present.
	var buf strings.Builder
	if err := format.Node(&buf, ci.fset, f); err != nil {
		return nil, fmt.Errorf("code generation failed: %w", err)
	}

	return []byte(buf.String()), nil
}

// createMarkLineStmt creates an ast.ExprStmt that records the given filename and line number
func (ci *CoverageInstrumenter) createMarkLineStmt(filename string, line int) ast.Stmt {
	// Create proper token positions to avoid comment interference
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.Ident{
					Name:    "testing",
					NamePos: token.NoPos,
				},
				Sel: &ast.Ident{
					Name:    "MarkLine",
					NamePos: token.NoPos,
				},
			},
			Args: []ast.Expr{
				&ast.BasicLit{
					Kind:     token.STRING,
					Value:    fmt.Sprintf("%q", filename),
					ValuePos: token.NoPos,
				},
				&ast.BasicLit{
					Kind:     token.INT,
					Value:    fmt.Sprintf("%d", line),
					ValuePos: token.NoPos,
				},
			},
			Lparen: token.NoPos,
			Rparen: token.NoPos,
		},
	}
}

// instrumentBlockStmt adds coverage tracking to a block statement.
// It inserts a MarkLine call at the beginning of the block unless
// the block is externally instrumented.
func (ci *CoverageInstrumenter) instrumentBlockStmt(block *ast.BlockStmt, line int) {
	if block == nil {
		return
	}

	// Check if block is externally instrumented
	if ci.statementContainsExternal(block) {
		// Register all lines as executable but don't instrument
		// to avoid conflicts with external instrumentation
		for _, stmt := range block.List {
			if stmt != nil {
				stmtLine := ci.fset.Position(stmt.Pos()).Line
				ci.tracker.RegisterExecutableLine(ci.filename, stmtLine)
			}
		}
		return
	}

	// Insert MarkLine at the beginning of the block (even if empty)
	markStmt := ci.createMarkLineStmt(ci.filename, line)
	block.List = append([]ast.Stmt{markStmt}, block.List...)
}

// isExternallyInstrumented checks if the file is externally instrumented
// Currently checks for 'cross' identifier as a marker for external instrumentation
func (ci *CoverageInstrumenter) isExternallyInstrumented(f *ast.File) bool {
	var containsExternal bool
	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "cross" {
			containsExternal = true
			return false
		}
		return true
	})
	return containsExternal
}

// statementContainsExternal checks if a statement contains external instrumentation markers
func (ci *CoverageInstrumenter) statementContainsExternal(stmt ast.Node) bool {
	if cached, ok := ci.containsExternalCache[stmt]; ok {
		return cached
	}

	var containsExternal bool
	ast.Inspect(stmt, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "cross" {
			containsExternal = true
			return false
		}
		return true
	})

	ci.containsExternalCache[stmt] = containsExternal
	return containsExternal
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
	case *ast.FuncLit:
		// Anonymous functions are also executable
		if node.Body != nil {
			funcLine := ci.fset.Position(node.Body.Lbrace).Line
			ci.tracker.RegisterExecutableLine(ci.filename, funcLine)
		}
	case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.SelectStmt,
		*ast.CaseClause, *ast.CommClause, *ast.ReturnStmt, *ast.DeferStmt,
		*ast.GoStmt, *ast.BranchStmt:
		line := ci.fset.Position(node.Pos()).Line
		ci.tracker.RegisterExecutableLine(ci.filename, line)
	case *ast.AssignStmt, *ast.ExprStmt:
		// Assignment and expression statements with potential side effects
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
	case *ast.FuncLit:
		ci.instrumentFuncLit(n)
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
	case *ast.DeferStmt:
		ci.instrumentDeferStmt(n)
	case *ast.BranchStmt:
		ci.instrumentBranchStmt(n)
	case *ast.AssignStmt:
		ci.instrumentAssignStmt(n)
	case *ast.ExprStmt:
		ci.instrumentExprStmt(n)
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

// instrumentFuncLit instruments an anonymous function literal
func (ci *CoverageInstrumenter) instrumentFuncLit(n *ast.FuncLit) {
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

	// Handle else clause
	if n.Else != nil {
		if elseBlock, ok := n.Else.(*ast.BlockStmt); ok {
			// Regular else block
			elseLine := ci.getLine(elseBlock.Lbrace)
			ci.registerAndInstrument(elseBlock, elseLine)
		} else if elseIf, ok := n.Else.(*ast.IfStmt); ok {
			// else if - treat as separate if statement
			ci.instrumentIfStmt(elseIf)
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

	// Create a dummy block to instrument switch entry
	switchBlock := &ast.BlockStmt{
		Lbrace: n.Pos(),
		List:   []ast.Stmt{},
		Rbrace: n.End(),
	}
	ci.instrumentBlockStmt(switchBlock, line)
}

// instrumentSelectStmt instruments a select statement
func (ci *CoverageInstrumenter) instrumentSelectStmt(n *ast.SelectStmt) {
	line := ci.getLine(n.Pos())
	ci.tracker.RegisterExecutableLine(ci.filename, line)

	// Create a dummy block to instrument select entry
	selectBlock := &ast.BlockStmt{
		Lbrace: n.Pos(),
		List:   []ast.Stmt{},
		Rbrace: n.End(),
	}
	ci.instrumentBlockStmt(selectBlock, line)
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
	// Don't add separate instrumentation - covered by containing block
}

// instrumentDeferStmt instruments a defer statement
func (ci *CoverageInstrumenter) instrumentDeferStmt(n *ast.DeferStmt) {
	line := ci.getLine(n.Pos())
	ci.tracker.RegisterExecutableLine(ci.filename, line)
	// Note: actual execution happens at function exit, not here
}

// instrumentBranchStmt instruments branch statements (break, continue, goto, fallthrough)
func (ci *CoverageInstrumenter) instrumentBranchStmt(n *ast.BranchStmt) {
	line := ci.getLine(n.Pos())
	ci.tracker.RegisterExecutableLine(ci.filename, line)
	// These affect control flow but don't need block instrumentation
}

// instrumentAssignStmt instruments assignment statements
func (ci *CoverageInstrumenter) instrumentAssignStmt(n *ast.AssignStmt) {
	line := ci.getLine(n.Pos())
	ci.tracker.RegisterExecutableLine(ci.filename, line)

	// Check if this statement is externally instrumented
	if ci.statementContainsExternal(n) {
		return
	}

	// Create a wrapper block for the assignment to add instrumentation
	markStmt := ci.createMarkLineStmt(ci.filename, line)

	// Replace the assignment with a block containing mark + assignment
	// This is done by modifying the parent during AST walk
	ci.addStatementInstrumentation(n, markStmt)
}

// instrumentExprStmt instruments expression statements
func (ci *CoverageInstrumenter) instrumentExprStmt(n *ast.ExprStmt) {
	line := ci.getLine(n.Pos())
	ci.tracker.RegisterExecutableLine(ci.filename, line)

	// Check if this statement is externally instrumented
	if ci.statementContainsExternal(n) {
		return
	}

	// Create a wrapper block for the expression to add instrumentation
	markStmt := ci.createMarkLineStmt(ci.filename, line)

	// Replace the expression with a block containing mark + expression
	// This is done by modifying the parent during AST walk
	ci.addStatementInstrumentation(n, markStmt)
}

// addStatementInstrumentation marks a statement for instrumentation
// The actual instrumentation happens in a post-processing phase
func (ci *CoverageInstrumenter) addStatementInstrumentation(stmt ast.Stmt, markStmt ast.Stmt) {
	// Store instrumentation requests for post-processing
	if ci.pendingInstrumentations == nil {
		ci.pendingInstrumentations = make(map[ast.Stmt]ast.Stmt)
	}
	ci.pendingInstrumentations[stmt] = markStmt
}

// applyStatementInstrumentations applies statement-level instrumentation
// This runs after the AST walk to modify blocks with individual statement tracking
func (ci *CoverageInstrumenter) applyStatementInstrumentations(f *ast.File) {
	if len(ci.pendingInstrumentations) == 0 {
		return
	}

	// Walk through all blocks and apply statement instrumentations
	ast.Inspect(f, func(n ast.Node) bool {
		if block, ok := n.(*ast.BlockStmt); ok {
			ci.instrumentStatementsInBlock(block)
		}
		return true
	})
}

// instrumentStatementsInBlock instruments individual statements within a block
func (ci *CoverageInstrumenter) instrumentStatementsInBlock(block *ast.BlockStmt) {
	if block == nil {
		return
	}

	var newStatements []ast.Stmt

	for _, stmt := range block.List {
		// Check if this statement needs instrumentation
		if markStmt, needsInstrumentation := ci.pendingInstrumentations[stmt]; needsInstrumentation {
			// Add the mark statement before the original statement
			newStatements = append(newStatements, markStmt)
		}
		// Add the original statement
		newStatements = append(newStatements, stmt)
	}

	// Update the block with instrumented statements
	block.List = newStatements
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
