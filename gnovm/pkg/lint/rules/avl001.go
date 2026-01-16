package rules

import (
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/lint"
)

const (
	avlPkgPath  = "gno.land/p/nt/avl"
	avlTreeName = "Tree"
)

type AVL001 struct{}

func init() {
	lint.MustRegister(&AVL001{})
}

func (AVL001) Info() lint.RuleInfo {
	return lint.RuleInfo{
		ID:       "AVL001",
		Category: lint.CategoryAVL,
		Name:     "unbounded-iteration",
		Severity: lint.SeverityWarning,
	}
}

func (AVL001) Check(ctx *lint.RuleContext, node gnolang.Node) []lint.Issue {
	call, ok := node.(*gnolang.CallExpr)
	if !ok {
		return nil
	}

	sel, ok := call.Func.(*gnolang.SelectorExpr)
	if !ok {
		return nil
	}

	methodName := string(sel.Sel)
	if methodName != "Iterate" && methodName != "ReverseIterate" {
		return nil
	}

	recvType := getTypeOf(sel.X)
	if recvType == nil {
		return nil
	}

	if !isAVLTree(recvType) {
		return nil
	}

	if !hasEmptyStringBounds(call) {
		return nil
	}

	return []lint.Issue{
		lint.NewIssue(
			"AVL001",
			lint.SeverityWarning,
			"unbounded "+methodName+" on avl.Tree (both start and end are empty strings)",
			ctx.File.FileName,
			call.GetPos(),
		),
	}
}

func getTypeOf(x gnolang.Expr) gnolang.Type {
	t, _ := x.GetAttribute(gnolang.ATTR_TYPEOF_VALUE).(gnolang.Type)
	return t
}

func isAVLTree(t gnolang.Type) bool {
	if pt, ok := t.(*gnolang.PointerType); ok {
		t = pt.Elt
	}

	dt, ok := t.(*gnolang.DeclaredType)
	if !ok {
		return false
	}

	return dt.PkgPath == avlPkgPath && string(dt.Name) == avlTreeName
}

func hasEmptyStringBounds(call *gnolang.CallExpr) bool {
	if len(call.Args) < 2 {
		return false
	}

	return isEmptyStringLiteral(call.Args[0]) && isEmptyStringLiteral(call.Args[1])
}

func isEmptyStringLiteral(expr gnolang.Expr) bool {
	switch e := expr.(type) {
	case *gnolang.BasicLitExpr:
		return e.Kind == gnolang.STRING && e.Value == `""`
	case *gnolang.ConstExpr:
		if e.T.Kind() != gnolang.StringKind {
			return false
		}
		return string(e.V.(gnolang.StringValue)) == ""
	}
	return false
}
