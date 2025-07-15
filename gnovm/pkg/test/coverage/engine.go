package coverage

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"regexp"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

// InstrumentationEngine orchestrates the application of instrumentation rules
type InstrumentationEngine struct {
	fset                    *token.FileSet
	tracker                 *Tracker
	filename                string
	externalDetector        ExternalInstrumentationDetector
	branchingStrategy       BranchingStrategy
	rules                   []InstrumentationRule
	pendingInstrumentations map[ast.Stmt]ast.Stmt
}

func NewInstrumentationEngine(tracker *Tracker, filename string) *InstrumentationEngine {
	engine := &InstrumentationEngine{
		fset:                    token.NewFileSet(),
		tracker:                 tracker,
		filename:                filename,
		externalDetector:        NewCrossIdentifierDetector(),
		branchingStrategy:       &DefaultBranchingStrategy{},
		pendingInstrumentations: make(map[ast.Stmt]ast.Stmt),
	}

	// Register instrumentation rules (R1-R6)
	engine.rules = []InstrumentationRule{
		&FunctionRule{},     // R1: Function instrumentation
		&ConditionalRule{},  // R2: Conditional instrumentation (with else if support)
		&LoopRule{},         // R3: Loop instrumentation
		&SwitchSelectRule{}, // R4: Switch/Select instrumentation
		&CaseRule{},         // R5: Case clause instrumentation
		&DeferRule{},        // R6: Defer instrumentation
		&StatementRule{},    // Statement-level instrumentation
		&BranchRule{},       // Branch statement instrumentation
	}

	return engine
}

// InstrumentFile applies all instrumentation rules following the axiom system
func (engine *InstrumentationEngine) InstrumentFile(content []byte) ([]byte, error) {
	// Pre-process content to fix bare import statements
	contentStr := engine.preprocessBareImports(string(content))

	// Parse with comments preserved (Principle P2: Minimal Intrusion)
	f, err := parser.ParseFile(engine.fset, engine.filename, contentStr, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing failed: %w", err)
	}

	// Rule S1: Check for external instrumentation
	if engine.externalDetector.IsExternallyInstrumented(f) {
		engine.registerExecutableLines(f)
		return content, nil // Return original content
	}

	// First pass: Apply all instrumentation rules to see if we need testing import
	ast.Walk(engine, f)

	// Check if any instrumentations were added
	hasInstrumentations := len(engine.pendingInstrumentations) > 0

	// Check if file has any executable code (functions, methods)
	hasExecutableCode := false
	ast.Inspect(f, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.FuncDecl:
			hasExecutableCode = true
			return false
		}
		return true
	})

	// Only add testing import if we have instrumentations or executable code
	if hasInstrumentations || hasExecutableCode {
		if err := engine.ensureTestingImport(f); err != nil {
			return nil, err
		}

		// Apply statement-level instrumentations (2-phase approach)
		engine.applyStatementInstrumentations(f)
	}

	// Format output preserving comments (Principle P2)
	var buf bytes.Buffer
	if err := format.Node(&buf, engine.fset, f); err != nil {
		return nil, fmt.Errorf("code generation failed: %w", err)
	}

	return buf.Bytes(), nil
}

// Visit implements ast.Visitor - orchestrates rule application
func (engine *InstrumentationEngine) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	// Apply all applicable rules to the node
	for _, rule := range engine.rules {
		if rule.Applies(node) {
			if err := rule.Apply(node, engine); err != nil {
				// Log error but continue with other rules (Principle P2: Minimal Intrusion)
				fmt.Printf("Warning: Rule %s failed for node: %v\n", rule.GetRuleName(), err)
			}
		}
	}

	return engine
}

// helpers

// getLine returns the line number for a given position
func (engine *InstrumentationEngine) getLine(pos token.Pos) int {
	return engine.fset.Position(pos).Line
}

// registerAndInstrument registers a line as executable and instruments the block
func (engine *InstrumentationEngine) registerAndInstrument(block *ast.BlockStmt, line int) {
	engine.tracker.RegisterExecutableLine(engine.filename, line)
	engine.instrumentBlockStmt(block, line)
}

// createMarkLineStmt creates instrumentation call statement
func (engine *InstrumentationEngine) createMarkLineStmt(filename string, line int) ast.Stmt {
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.Ident{
					Name:    "testing",
					NamePos: token.NoPos,
				},
				Sel: &ast.Ident{
					Name:    "markLine",
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

// instrumentBlockStmt instruments a block statement following Axiom A2
func (engine *InstrumentationEngine) instrumentBlockStmt(block *ast.BlockStmt, line int) {
	if block == nil {
		return
	}

	// Check for external instrumentation
	if engine.externalDetector.IsExternallyInstrumented(block) {
		// Register but don't instrument
		for _, stmt := range block.List {
			if stmt != nil {
				stmtLine := engine.fset.Position(stmt.Pos()).Line
				engine.tracker.RegisterExecutableLine(engine.filename, stmtLine)
			}
		}
		return
	}

	// Insert MarkLine at block entry point (Axiom A2)
	markStmt := engine.createMarkLineStmt(engine.filename, line)
	block.List = append([]ast.Stmt{markStmt}, block.List...)
}

// instrumentCaseStmts instruments case statement bodies
func (engine *InstrumentationEngine) instrumentCaseStmts(body []ast.Stmt, line int) []ast.Stmt {
	markStmt := engine.createMarkLineStmt(engine.filename, line)
	return append([]ast.Stmt{markStmt}, body...)
}

// addStatementInstrumentation queues statement for instrumentation
func (engine *InstrumentationEngine) addStatementInstrumentation(stmt ast.Stmt, markStmt ast.Stmt) {
	engine.pendingInstrumentations[stmt] = markStmt
}

// applyStatementInstrumentations applies queued statement instrumentations
func (engine *InstrumentationEngine) applyStatementInstrumentations(f *ast.File) {
	if len(engine.pendingInstrumentations) == 0 {
		return
	}

	ast.Inspect(f, func(n ast.Node) bool {
		if block, ok := n.(*ast.BlockStmt); ok {
			engine.instrumentStatementsInBlock(block)
		}
		return true
	})
}

// instrumentStatementsInBlock applies statement-level instrumentation
func (engine *InstrumentationEngine) instrumentStatementsInBlock(block *ast.BlockStmt) {
	if block == nil {
		return
	}

	newStatements := make([]ast.Stmt, 0, len(block.List))

	for _, stmt := range block.List {
		// Check if this statement needs instrumentation
		if markStmt, needsInstrumentation := engine.pendingInstrumentations[stmt]; needsInstrumentation {
			// Add the mark statement before the original statement
			newStatements = append(newStatements, markStmt)
		}
		// Add the original statement
		newStatements = append(newStatements, stmt)
	}

	block.List = newStatements
}

// registerExecutableLines registers all executable lines without instrumenting
func (engine *InstrumentationEngine) registerExecutableLines(f *ast.File) {
	ast.Inspect(f, func(n ast.Node) bool {
		engine.registerNodeIfExecutable(n)
		return true
	})
}

// registerNodeIfExecutable registers a node's line if it's executable (Axiom A1)
func (engine *InstrumentationEngine) registerNodeIfExecutable(n ast.Node) {
	if n == nil {
		return
	}

	switch node := n.(type) {
	case *ast.FuncDecl:
		if node.Body != nil {
			funcLine := engine.fset.Position(node.Body.Lbrace).Line
			engine.tracker.RegisterExecutableLine(engine.filename, funcLine)
		}
	case *ast.FuncLit:
		if node.Body != nil {
			funcLine := engine.fset.Position(node.Body.Lbrace).Line
			engine.tracker.RegisterExecutableLine(engine.filename, funcLine)
		}
	case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.SelectStmt,
		*ast.CaseClause, *ast.CommClause, *ast.ReturnStmt, *ast.DeferStmt,
		*ast.BranchStmt, *ast.AssignStmt, *ast.ExprStmt:
		line := engine.fset.Position(node.Pos()).Line
		engine.tracker.RegisterExecutableLine(engine.filename, line)
	}
}

// ensureTestingImport adds testing import if not present
func (engine *InstrumentationEngine) ensureTestingImport(f *ast.File) error {
	// Use astutil.AddImport which handles all the complexity of adding imports correctly
	if !astutil.AddImport(engine.fset, f, "testing") {
		// Import was already present
		return nil
	}

	// Import was added successfully
	return nil
}

// preprocessBareImports fixes bare import statements (just "import" with no specs)
func (engine *InstrumentationEngine) preprocessBareImports(content string) string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))

	// Regular expression to match bare import statements
	bareImportRe := regexp.MustCompile(`^\s*import\s*$`)

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Check if this is a bare import line
		if bareImportRe.MatchString(line) {
			// Look ahead to see if the next line is a comment or empty
			hasValidImport := false
			for j := i + 1; j < len(lines); j++ {
				nextLine := strings.TrimSpace(lines[j])
				if nextLine == "" {
					continue
				}
				// Check if it's a comment
				if strings.HasPrefix(nextLine, "//") {
					continue
				}
				// Check if it's an import spec
				if strings.HasPrefix(nextLine, `"`) || strings.HasPrefix(nextLine, "(") {
					hasValidImport = true
				}
				break
			}

			// If no valid import follows, skip this bare import line
			if !hasValidImport {
				continue
			}
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}
