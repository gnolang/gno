package gnolang

import (
	// "strconv"
	"strings"

	j "github.com/grepsuzette/joeson"
)

func fInt(it j.Ast, ctx *j.ParseContext) j.Ast {
	s := it.(j.NativeString).Str
	if strings.Contains(s, "__") {
		return j.NewParseError(ctx, "invalid: only one _ at a time")
	} else if strings.HasSuffix(s, "_") {
		return j.NewParseError(ctx, "invalid: _ must separate successive digits")
	} else {
		return wrap(&BasicLitExpr{
			Kind:  INT,
			Value: s,
		}, it)
	}
}

// func fBinaryExpr(it j.Ast) j.Ast {
// 	m := it.(j.NativeMap)
// 	lhs, b1 := m.GetExists("l")
// 	op_, b2 := m.GetStringExists("op")
// 	rhs, b3 := m.GetExists("r")
// 	if b1 && b2 && b3 {
// 		return expr2Ast(newBx(lhs.(w).expr, op_, rhs.(w).expr))
// 	} else {
// 		panic("assert")
// 	}
// }

func fExpression(it j.Ast) j.Ast {
	if m, ok := it.(j.NativeMap); ok {
		a := m.GetOrPanic("bx").(*j.NativeArray).Array
		if j.IsParseError(a[0]) {
			return a[0]
		} else if j.IsParseError(a[1]) {
			return a[1]
		} else {
			lh := a[0].(wrapped).expr
			op := Op2Word(a[1].(j.NativeString).Str)
			rh := a[2].(wrapped).expr
			return wrap(newBx(lh, op, rh), it)
		}
	} else {
		return it // Unary
	}
}

func fUnaryExpr(it j.Ast) j.Ast {
	if m, ok := it.(j.NativeMap); ok {
		a := m.GetOrPanic("ux").(*j.NativeArray).Array
		op := a[0].(j.NativeString).Str
		arg := a[1].(wrapped).expr
		switch op {
		case "*":
			return wrap(&StarExpr{X: arg}, it)
		case "&":
			return wrap(&RefExpr{X: arg}, it)
		case "+", "-", "!", "^", "<-":
			return wrap(&UnaryExpr{Op: Op2Word(op), X: arg}, it)
		default:
			panic("assert")
		}
	} else {
		return it // PrimaryExpr
	}
}
