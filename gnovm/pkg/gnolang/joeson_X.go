package gnolang

import (
	"fmt"
	"strings"

	j "github.com/grepsuzette/joeson"
)

func grammar() *j.Grammar {
	if gm == nil {
		gm = j.GrammarFromLines(
			"GNO-grammar",
			gnoRules,
			// j.GrammarOptions{TraceOptions: j.Mute()},
		)
	}
	return gm
}

// rules and grammar for GNO
// This is not the place to explain it in detail, but
// we will integrate some short explaination anyway.

// A convenient way to make list of subrules.
// rules() is similar to []j.Line{a,b,c,...}.
func rules(a ...j.Line) []j.Line { return a }

// "Inline" line of rule AKA ILine. Inline rules are always named().
// An inline rule can be referenced by its name, but when it isn't
// it is totally passive.
func i(a ...any) j.ILine { return j.I(a...) }

// "OR" rule. Inside a rank, "OR" rules (AKA OLine) are parsed one after the
// other until one returns something other than nil. Some of them are named,
// but they usually aren't, as it's more the point of an ILine to be
// referenced. If and when OLine are named it is just to clarify or put a name
// on what they are supposed to parse.
func o(a ...any) j.OLine { return j.O(a...) }

// A Key-value pair, where Key is the name.
// This is exclusively used with joeson ILine and OLine to name things.
func named(name string, thing interface{}) j.NamedRule { return j.Named(name, thing) }

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
// When a ParseError panic happens, it just returns that ParseError,
// allowing it to short-circuit the grammar.
func parseX(s string) (result j.Ast) {
	defer func() {
		if e := recover(); e != nil {
			if pe, ok := e.(j.ParseError); ok {
				result = pe
			} else {
				panic(e)
			}
		}
	}()
	if tokens, e := j.TokenStreamFromGoCode(s); e != nil {
		result = j.NewParseError(nil, e.Error())
	} else {
		result = grammar().ParseTokens(tokens)
	}
	return
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
