package gnolang

import (
	"fmt"
	"strings"

	j "github.com/grepsuzette/joeson"
)

// rules and grammar for GNO

func i(a ...any) j.ILine                       { return j.I(a...) }
func o(a ...any) j.OLine                       { return j.O(a...) }
func rules(a ...j.Line) []j.Line               { return a }
func named(name string, thing any) j.NamedRule { return j.Named(name, thing) }

// let's have Expr satisfy joeson.Ast
// func (e *Expr) ContentString() string { return "TODO switch and show, BinaryExpr etc. See nodes.go" }

// Rewrite of X() with Joeson
func X(x interface{}, args ...interface{}) Expr {
	switch cx := x.(type) {
	case Expr:
		return cx
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return Xold(fmt.Sprintf("%v", x))
	case string:
		if cx == "" {
			panic("input cannot be blank for X()")
		}
	case Name:
		if cx == "" {
			panic("input cannot be blank for X()")
		}
		x = string(cx)
	default:
		panic(fmt.Sprintf("unexpected input type for X(): %T", x))
	}
	expr := x.(string)
	expr = fmt.Sprintf(expr, args...)
	expr = strings.TrimSpace(expr)
	// first := expr[0]

	// return Xold(x, args...)
	//
	ast := grammar.ParseString(expr)
	if j.IsParseError(ast) {
		panic(ast.String())
	} else {
		return ast.(wrapped).expr
	}
}

// TODO find where to initialize it
func initGrammar() {
	grammar = j.GrammarFromLines(gnoRules, "GNO-grammar", j.GrammarOptions{TraceOptions: j.Mute()})
}
