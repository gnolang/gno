package lintrules

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

type AvlLimitRule struct{}

func (AvlLimitRule) Run(ctx *RuleContext, node gnolang.Node) error {
	call, ok := node.(*gnolang.CallExpr)
	if !ok {
		return nil
	}

	sel, ok := call.Func.(*gnolang.SelectorExpr)
	if !ok {
		return nil
	}

	m := string(sel.Sel)
	if m != "Iterate" && m != "ReverseIterate" {
		return nil
	}

	// --- Receiver type check
	recvT := gnolang.EvalStaticTypeOf(ctx.Store, ctx.File, sel.X)
	if !isAVLTree(recvT) {
		// DEBUG:
		// fmt.Printf("receiver is not AVL tree for %s\n", method)
		return nil
	}

	// --- Require at least 2 args (start, end)
	if len(call.Args) < 2 {
		return nil
	}

	// --- Check both args are "" literal
	if !isEmptyConstString(call.Args[0]) ||
		!isEmptyConstString(call.Args[1]) {
		return nil
	}

	// --- Check //nolint or //nolint:avl-limit on prev line
	if hasNoLintDirective(ctx, node.GetPos()) {
		return nil
	}

	return errors.New("avl tree error lolilol")
}

func isEmptyConstString(expr gnolang.Expr) bool {
	cs, ok := expr.(*gnolang.ConstExpr)
	if !ok {
		return false
	}
	if cs.T.Kind() != gnolang.StringKind {
		return false
	}
	return string(cs.V.(gnolang.StringValue)) == ""
}

func hasNoLintDirective(ctx *RuleContext, pos gnolang.Pos) bool {
	if ctx.Source == "" {
		return false
	}

	lines := strings.Split(ctx.Source, "\n")
	line := pos.Line - 1 // convert to 0-based index

	if line <= 0 || line > len(lines) {
		return false
	}

	prev := strings.TrimSpace(lines[line-1])

	if strings.HasPrefix(prev, "//nolint") {
		return true
	}
	return false
}

func isAVLTree(t gnolang.Type) bool {
	dt, ok := gnolang.UnwrapPointerType(t).(*gnolang.DeclaredType)
	if !ok {
		fmt.Printf("DEBUG: not declared type %T \n", t)
		return false
	}
	return dt.PkgPath == "gno.land/p/nt/avl" && dt.Name == "Tree"
}
