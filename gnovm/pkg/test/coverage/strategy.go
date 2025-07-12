package coverage

import (
	"go/ast"
)

// branching strategy (Control Flow Branching)
var _ BranchingStrategy = (*DefaultBranchingStrategy)(nil)

// DefaultBranchingStrategy implements the default branching rules from Axiom A3
type DefaultBranchingStrategy struct{}

func (s *DefaultBranchingStrategy) ShouldInstrumentEntry(node ast.Node) bool {
	switch node.(type) {
	case *ast.FuncDecl, *ast.FuncLit:
		return true // All function entries
	case *ast.SwitchStmt, *ast.SelectStmt:
		return true // Rule R4: Switch/Select entry instrumentation
	default:
		return false
	}
}

func (s *DefaultBranchingStrategy) ShouldInstrumentBranches(node ast.Node) bool {
	switch node.(type) {
	case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt:
		return true
	case *ast.SwitchStmt, *ast.SelectStmt:
		return true // Each case is a branch
	default:
		return false
	}
}

func (s *DefaultBranchingStrategy) GetBranches(node ast.Node) []ast.Node {
	switch n := node.(type) {
	case *ast.IfStmt:
		branches := []ast.Node{n.Body}
		if n.Else != nil {
			branches = append(branches, n.Else)
		}
		return branches
	case *ast.SwitchStmt:
		var branches []ast.Node
		for _, stmt := range n.Body.List {
			if caseClause, ok := stmt.(*ast.CaseClause); ok {
				branches = append(branches, caseClause)
			}
		}
		return branches
	case *ast.SelectStmt:
		var branches []ast.Node
		for _, stmt := range n.Body.List {
			if commClause, ok := stmt.(*ast.CommClause); ok {
				branches = append(branches, commClause)
			}
		}
		return branches
	default:
		return nil
	}
}
