package rules

import (
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/lint"
)

type RENDER001 struct{}

func init() {
	lint.MustRegister(&RENDER001{})
}

func (RENDER001) Info() lint.RuleInfo {
	return lint.RuleInfo{
		ID:       "RENDER001",
		Category: lint.CategoryRender,
		Name:     "invalid-render-signature",
		Severity: lint.SeverityError,
	}
}

func (RENDER001) Check(ctx *lint.RuleContext, node gnolang.Node) []lint.Issue {
	fn, ok := node.(*gnolang.FuncDecl)
	if !ok {
		return nil
	}

	if string(fn.Name) != "Render" {
		return nil
	}

	if fn.IsMethod {
		return nil
	}

	if !isFileLevelDecl(ctx.Parents) {
		return nil
	}

	if !gnolang.IsRealmPath(ctx.PkgPath) {
		return nil
	}

	if !isValidRenderSignature(fn) {
		return []lint.Issue{
			lint.NewIssue(
				"RENDER001",
				lint.SeverityError,
				"invalid Render function signature; must be func Render(string) string",
				ctx.File.FileName,
				fn.GetPos(),
			),
		}
	}

	return nil
}

func isValidRenderSignature(fn *gnolang.FuncDecl) bool {
	// After preprocessing, FuncTypeExpr has ATTR_TYPE_VALUE set to the resolved FuncType.
	if ft, ok := fn.Type.GetAttribute(gnolang.ATTR_TYPE_VALUE).(*gnolang.FuncType); ok {
		return isSingleStringType(ft.Params) && isSingleStringType(ft.Results)
	}
	// Fallback for unpreprocessed AST: check NameExpr directly.
	return isSingleStringNameExpr(fn.Type.Params) && isSingleStringNameExpr(fn.Type.Results)
}

func isSingleStringType(fields []gnolang.FieldType) bool {
	if len(fields) != 1 {
		return false
	}
	return fields[0].Type != nil && fields[0].Type.Kind() == gnolang.StringKind
}

func isSingleStringNameExpr(fields gnolang.FieldTypeExprs) bool {
	if len(fields) != 1 {
		return false
	}
	nx, ok := fields[0].Type.(*gnolang.NameExpr)
	if !ok {
		return false
	}
	return string(nx.Name) == "string"
}
