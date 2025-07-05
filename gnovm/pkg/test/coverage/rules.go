package coverage

import (
	"go/ast"
)

// instrumentation rules
var (
	_ InstrumentationRule = (*FunctionRule)(nil)
	_ InstrumentationRule = (*ConditionalRule)(nil)
	_ InstrumentationRule = (*LoopRule)(nil)
	_ InstrumentationRule = (*SwitchSelectRule)(nil)
	_ InstrumentationRule = (*CaseRule)(nil)
	_ InstrumentationRule = (*DeferRule)(nil)
	_ InstrumentationRule = (*StatementRule)(nil)
	_ InstrumentationRule = (*BranchRule)(nil)
)

// FunctionRule implements Rule R1: Function instrumentation
type FunctionRule struct{}

func (r *FunctionRule) Applies(node ast.Node) bool {
	switch node.(type) {
	case *ast.FuncDecl, *ast.FuncLit:
		return true
	default:
		return false
	}
}

func (r *FunctionRule) Apply(node ast.Node, engine *InstrumentationEngine) error {
	switch n := node.(type) {
	case *ast.FuncDecl:
		if n.Body != nil {
			line := engine.getLine(n.Body.Lbrace)
			engine.registerAndInstrument(n.Body, line)
		}
	case *ast.FuncLit:
		if n.Body != nil {
			line := engine.getLine(n.Body.Lbrace)
			engine.registerAndInstrument(n.Body, line)
		}
	}
	return nil
}

func (r *FunctionRule) GetRuleName() string { return "R1-Function" }

// ConditionalRule implements Rule R2: Conditional instrumentation with else if support
type ConditionalRule struct{}

func (r *ConditionalRule) Applies(node ast.Node) bool {
	_, ok := node.(*ast.IfStmt)
	return ok
}

func (r *ConditionalRule) Apply(node ast.Node, engine *InstrumentationEngine) error {
	ifStmt := node.(*ast.IfStmt)

	// Instrument true branch
	if ifStmt.Cond != nil {
		condLine := engine.getLine(ifStmt.Cond.Pos())
		engine.registerAndInstrument(ifStmt.Body, condLine)
	}

	// Handle else clause (including else if)
	if ifStmt.Else != nil {
		if elseBlock, ok := ifStmt.Else.(*ast.BlockStmt); ok {
			// Regular else block
			elseLine := engine.getLine(elseBlock.Lbrace)
			engine.registerAndInstrument(elseBlock, elseLine)
		} else if elseIf, ok := ifStmt.Else.(*ast.IfStmt); ok {
			// will be handled by recursive application of this rule
			_ = elseIf // The AST walker will visit this automatically
		}
	}

	return nil
}

func (r *ConditionalRule) GetRuleName() string { return "R2-Conditional" }

// LoopRule implements Rule R3: Loop instrumentation
type LoopRule struct{}

func (r *LoopRule) Applies(node ast.Node) bool {
	switch node.(type) {
	case *ast.ForStmt, *ast.RangeStmt:
		return true
	default:
		return false
	}
}

func (r *LoopRule) Apply(node ast.Node, engine *InstrumentationEngine) error {
	switch n := node.(type) {
	case *ast.ForStmt:
		var line int
		if n.Cond != nil {
			line = engine.getLine(n.Cond.Pos())
		} else {
			line = engine.getLine(n.Pos())
		}
		engine.registerAndInstrument(n.Body, line)
	case *ast.RangeStmt:
		line := engine.getLine(n.Pos())
		engine.registerAndInstrument(n.Body, line)
	}
	return nil
}

func (r *LoopRule) GetRuleName() string { return "R3-Loop" }

// SwitchSelectRule implements Rule R4: Switch/Select instrumentation with entry tracking
type SwitchSelectRule struct{}

func (r *SwitchSelectRule) Applies(node ast.Node) bool {
	switch node.(type) {
	case *ast.SwitchStmt, *ast.SelectStmt:
		return true
	default:
		return false
	}
}

func (r *SwitchSelectRule) Apply(node ast.Node, engine *InstrumentationEngine) error {
	switch n := node.(type) {
	case *ast.SwitchStmt:
		line := engine.getLine(n.Pos())
		engine.tracker.RegisterExecutableLine(engine.filename, line)
		// Create dummy block for switch entry instrumentation
		switchBlock := &ast.BlockStmt{
			Lbrace: n.Pos(),
			List:   []ast.Stmt{},
			Rbrace: n.End(),
		}
		engine.instrumentBlockStmt(switchBlock, line)
	case *ast.SelectStmt:
		line := engine.getLine(n.Pos())
		engine.tracker.RegisterExecutableLine(engine.filename, line)
		// Create dummy block for select entry instrumentation
		selectBlock := &ast.BlockStmt{
			Lbrace: n.Pos(),
			List:   []ast.Stmt{},
			Rbrace: n.End(),
		}
		engine.instrumentBlockStmt(selectBlock, line)
	}
	return nil
}

func (r *SwitchSelectRule) GetRuleName() string { return "R4-SwitchSelect" }

// CaseRule implements Rule R5: Case clause instrumentation
type CaseRule struct{}

func (r *CaseRule) Applies(node ast.Node) bool {
	switch node.(type) {
	case *ast.CaseClause, *ast.CommClause:
		return true
	default:
		return false
	}
}

func (r *CaseRule) Apply(node ast.Node, engine *InstrumentationEngine) error {
	switch n := node.(type) {
	case *ast.CaseClause:
		line := engine.getLine(n.Pos())
		engine.tracker.RegisterExecutableLine(engine.filename, line)
		n.Body = engine.instrumentCaseStmts(n.Body, line)
	case *ast.CommClause:
		line := engine.getLine(n.Pos())
		engine.tracker.RegisterExecutableLine(engine.filename, line)
		n.Body = engine.instrumentCaseStmts(n.Body, line)
	}
	return nil
}

func (r *CaseRule) GetRuleName() string { return "R5-Case" }

// DeferRule implements Rule R6: Defer instrumentation
type DeferRule struct{}

func (r *DeferRule) Applies(node ast.Node) bool {
	_, ok := node.(*ast.DeferStmt)
	return ok
}

func (r *DeferRule) Apply(node ast.Node, engine *InstrumentationEngine) error {
	deferStmt := node.(*ast.DeferStmt)
	line := engine.getLine(deferStmt.Pos())
	engine.tracker.RegisterExecutableLine(engine.filename, line)
	// Defer registration is tracked, actual execution happens at function exit
	return nil
}

func (r *DeferRule) GetRuleName() string { return "R6-Defer" }

// StatementRule implements statement-level instrumentation for assignments and expressions
type StatementRule struct{}

func (r *StatementRule) Applies(node ast.Node) bool {
	switch node.(type) {
	case *ast.AssignStmt, *ast.ExprStmt:
		return true
	default:
		return false
	}
}

func (r *StatementRule) Apply(node ast.Node, engine *InstrumentationEngine) error {
	var line int
	switch n := node.(type) {
	case *ast.AssignStmt:
		line = engine.getLine(n.Pos())
		engine.tracker.RegisterExecutableLine(engine.filename, line)
		if !engine.externalDetector.IsExternallyInstrumented(n) {
			markStmt := engine.createMarkLineStmt(engine.filename, line)
			engine.addStatementInstrumentation(n, markStmt)
		}
	case *ast.ExprStmt:
		line = engine.getLine(n.Pos())
		engine.tracker.RegisterExecutableLine(engine.filename, line)
		if !engine.externalDetector.IsExternallyInstrumented(n) {
			markStmt := engine.createMarkLineStmt(engine.filename, line)
			engine.addStatementInstrumentation(n, markStmt)
		}
	}
	return nil
}

func (r *StatementRule) GetRuleName() string { return "Statement-Level" }

// BranchRule implements branch statement instrumentation (break, continue, etc.)
type BranchRule struct{}

func (r *BranchRule) Applies(node ast.Node) bool {
	switch node.(type) {
	case *ast.ReturnStmt, *ast.BranchStmt:
		return true
	default:
		return false
	}
}

func (r *BranchRule) Apply(node ast.Node, engine *InstrumentationEngine) error {
	var line int
	switch n := node.(type) {
	case *ast.ReturnStmt:
		line = engine.getLine(n.Pos())
		engine.tracker.RegisterExecutableLine(engine.filename, line)
		// Return statements are covered by containing block instrumentation
	case *ast.BranchStmt:
		line = engine.getLine(n.Pos())
		engine.tracker.RegisterExecutableLine(engine.filename, line)
		// Branch statements affect control flow but don't need separate instrumentation
	}
	return nil
}

func (r *BranchRule) GetRuleName() string { return "Branch" }
