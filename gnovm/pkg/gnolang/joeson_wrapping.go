package gnolang

import (
	// "reflect"
	"strings"

	j "github.com/grepsuzette/joeson"
)

type wrapped struct {
	expr Expr
	ast  j.Ast
}

// gnolang ast are gnolang.Expr
// joeson grammars must produce joeson.Ast
// => wrapper
func wrap(expr Expr, ast j.Ast) j.Ast {
	return wrapped{
		expr: expr,
		ast:  ast,
	}
}

func StringWithRulenames(ast j.Ast) string {
	var b strings.Builder
	rule := ast.GetLocation().RuleName
	if rule != "" {
		b.WriteString(j.BoldBlue("«" + ast.GetLocation().RuleName + ""))
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

// wrapped String() is special in that it tries to show Rule names recursively
func (w wrapped) String() string {
	if w.ast == nil {
		return "Nil wrapped.ast"
	} else {
		return StringWithRulenames(w.ast)
	}
}

// implement j.Ast and Expr
func (w wrapped) assertExpr()                 {} // TODO it could be declared with the others
func (w wrapped) assertNode()                 {}
func (w wrapped) SetLocation(o j.Origin)      {}
func (w wrapped) GetLocation() j.Origin       { return w.ast.GetLocation() }
func (w wrapped) Copy() Node                  { panic("Copy not implemented for wrapped") }
func (w wrapped) GetLabel() Name              { return Name("") }
func (w wrapped) SetLabel(Name)               {}
func (w wrapped) GetLine() int                { return w.ast.GetLocation().Code.PosToLine(w.ast.GetLocation().Start) }
func (w wrapped) SetLine(int)                 { panic("Not authorized to SetLine() now") }
func (w wrapped) HasAttribute(key any) bool   { panic("Not implemented") }
func (w wrapped) GetAttribute(key any) any    { panic("Not implemented") }
func (w wrapped) SetAttribute(key, value any) { panic("Not implemented") }
