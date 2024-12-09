package coverage

import (
	"go/ast"
	"go/parser"
	"go/token"
)

// detectExecutableLines analyzes the given source code content and returns a map
// of line numbers to boolean values indicating whether each line is executable.
func DetectExecutableLines(content string) (map[int]bool, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	executableLines := make(map[int]bool)

	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		if isExecutableLine(n) {
			line := fset.Position(n.Pos()).Line
			executableLines[line] = true
		}

		return true
	})

	return executableLines, nil
}

// countCodeLines counts the number of executable lines in the given source code content.
func CountCodeLines(content string) int {
	lines, err := DetectExecutableLines(content)
	if err != nil {
		return 0
	}

	return len(lines)
}

// isExecutableLine determines whether a given AST node represents an
// executable line of code for the purpose of code coverage measurement.
//
// It returns true for statement nodes that typically contain executable code,
// such as assignments, expressions, return statements, and control flow statements.
//
// It returns false for nodes that represent non-executable lines, such as
// declarations, blocks, and function definitions.
func isExecutableLine(node ast.Node) bool {
	switch n := node.(type) {
	case *ast.AssignStmt, *ast.ExprStmt, *ast.ReturnStmt, *ast.BranchStmt,
		*ast.IncDecStmt, *ast.GoStmt, *ast.DeferStmt, *ast.SendStmt:
		return true
	case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt,
		*ast.TypeSwitchStmt, *ast.SelectStmt:
		return true
	case *ast.CaseClause:
		// Even if a `case` condition (e.g., `case 1:`) in a `switch` statement is executed,
		// the condition itself is not included in the coverage; coverage only recorded for the
		// code block inside the corresponding `case` clause.
		return false
	case *ast.LabeledStmt:
		return isExecutableLine(n.Stmt)
	case *ast.FuncDecl:
		return false
	case *ast.BlockStmt:
		return false
	case *ast.DeclStmt:
		// check inner declarations in the DeclStmt (e.g. `var a, b = 1, 2`)
		// if there is a value initialization, then the line is executable
		genDecl, ok := n.Decl.(*ast.GenDecl)
		if ok && (genDecl.Tok == token.VAR || genDecl.Tok == token.CONST) {
			for _, spec := range genDecl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if ok && len(valueSpec.Values) > 0 {
					return true
				}
			}
		}
		return false
	case *ast.ImportSpec, *ast.TypeSpec, *ast.ValueSpec:
		return false
	case *ast.InterfaceType:
		return false
	case *ast.GenDecl:
		switch n.Tok {
		case token.VAR, token.CONST:
			for _, spec := range n.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if ok && len(valueSpec.Values) > 0 {
					return true
				}
			}
			return false
		case token.TYPE, token.IMPORT:
			return false
		default:
			return true
		}
	default:
		return false
	}
}
