package gnolang

import (
	j "github.com/grepsuzette/joeson"
)

type wrapped struct {
	expr Expr
	ast  j.Ast
}

// gnolang ast are gnolang.Expr
// joeson grammars must produce joeson.Ast
// So we need a wrapper.
func wrap(expr Expr, ast j.Ast) j.Ast {
	return wrapped{
		expr: expr,
		ast:  ast,
	}
}

func (w wrapped) String() string {
	return "<" + w.ast.GetLocation().RuleName + ":" + w.ast.String() + ">"
}

// these 2 are a bit fake, this is so wrapped also implements j.Ast
func (w wrapped) SetLocation(o j.Origin) {}
func (w wrapped) GetLocation() j.Origin  { return j.Origin{} }

// TODO - OLD STUFFS TO REMOVE
// function x() helps to quickly write a grammar.
// Calling x("foo") returns a callback `func(τ Ast) Ast`.
// Calling cb.ContentString() gives "<foo:" + τ.ContentString() + ">"
//
// For example:
//
// var rules_tokens = rules(
//
//	o(named("token", "( keyword | identifier | operator | punctuation | literal )"), x("token")),
//	i(named("keyword", "( 'break' | 'default' | 'func' | 'interface' | 'select' | 'case' | 'defer' | 'go' | 'map' | 'struct' | 'chan' | 'else' | 'goto' | 'package' | 'switch' | 'const' | 'fallthrough' | 'if' | 'range' | 'type' | 'continue' | 'for' | 'import' | 'return' | 'var' )"), x("keyword")),
//	i(named("identifier", "[a-zA-Z_][a-zA-Z0-9_]*"), x("identifier")), // letter { letter | unicode_digit } .   We rewrite it so to accelerate parsing
//	i(named("operator", "( '+' | '&' | '+=' | '&=' | '&&' | '==' | '!=' | '(' | ')' | '-' | '|' | '-=' | '|=' | '||' | '<' | '<=' | '[' |  ']' | '*' | '^' | '*=' | '^=' | '<-' | '>' | '>=' | '{' | '}' | '/' | '<<' | '/=' | '<<=' | '++' | '=' | ':=' | '%' | '>>' | '%=' | '>>=' | '--' | '!' | '...' | '&^' | '&^=' | '~' )"), x("operator")),
//
// ...
// )
//
// Here, whichever of keyword, identifier etc gets built,
// its ContentString() will be like "<token:keyword>", "<token:identifier>" etc.

// func x(typename string) func(j.Ast) j.Ast {
// 	return func(ast j.Ast) j.Ast {
// 		return dumb{typename, ast}
// 	}
// }

// func (ww w) ContentString() string { return ww.expr.String() }

// type dumb is used by x(). As the name hints, it's nothing too exciting
// type dumb struct {
// 	typename string
// 	ast      j.Ast
// }

// func (dumb dumb) String() string {
// 	return "<" + dumb.typename + ":" + dumb.ast.String() + ">"
// }
