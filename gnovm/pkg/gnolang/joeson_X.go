package gnolang

import (
	"fmt"
	"strings"

	j "github.com/grepsuzette/joeson"
)

func grammar() *j.Grammar {
	if gm == nil {
		gm = j.GrammarFromLines(
			gnoRules,
			"GNO-grammar",
			// j.GrammarOptions{TraceOptions: j.Mute()},
		)
	}
	return gm
}

// rules and grammar for GNO

func i(a ...any) j.ILine                       { return j.I(a...) }
func o(a ...any) j.OLine                       { return j.O(a...) }
func rules(a ...j.Line) []j.Line               { return a }
func named(name string, thing any) j.NamedRule { return j.Named(name, thing) }

// Rewrite of X() with Joeson
func Xnew(x interface{}, args ...interface{}) Expr {
	switch cx := x.(type) {
	case Expr:
		return cx
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return X(fmt.Sprintf("%v", x))
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
	sexpr := x.(string)
	sexpr = fmt.Sprintf(sexpr, args...)
	sexpr = strings.TrimSpace(sexpr)
	ast := parseX(sexpr)
	if expr, ok := ast.(Expr); ok {
		return expr
	} else {
		panic("having " + ast.String() + " into an Expr requires some manual work")
	}
}

// Producing joeson.Ast, joeson.ParseError or gnolang.Node
func parseX(s string) j.Ast {
	return grammar().ParseString(s)
}

func StringWithRulenames(ast j.Ast) string {
	var b strings.Builder
	rule := ast.GetOrigin().RuleName
	if rule != "" {
		b.WriteString(j.BoldBlue("«" + ast.GetOrigin().RuleName + ""))
	}
	b.WriteString(j.BoldBlue("•"))
	// b.WriteString(j.Yellow(reflect.TypeOf(ast).String()))
	b.WriteString(ast.String())
	// b.WriteString(j.Red(ast.GetLocation().String()))
	if rule != "" {
		b.WriteString(j.BoldBlue("»"))
	}
	return b.String()
}
