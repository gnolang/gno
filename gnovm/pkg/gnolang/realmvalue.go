package gnolang

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

// Realm value rule: outside realm a, no value of a's struct or array types
// may appear (variables, fields, parameters, results, expressions); only
// pointers and interfaces reach them. A value is always its holder's.

func newRealmValueInfo() *types.Info {
	return &types.Info{
		Types:     map[ast.Expr]types.TypeAndValue{},
		Defs:      map[*ast.Ident]types.Object{},
		Implicits: map[ast.Node]types.Object{},
	}
}

// checkRealmValues reports each value of another realm's struct or array
// type in files type-checked as pkgPath.
func checkRealmValues(fset *token.FileSet, files []*ast.File, info *types.Info, pkgPath string) []error {
	self := pkgPath
	var errs []error
	report := func(pos token.Pos, t types.Type) {
		errs = append(errs, types.Error{
			Fset: fset, Pos: pos,
			Msg: fmt.Sprintf("cannot hold a value of %s outside its realm (use a pointer)", t),
		})
	}
	for _, f := range files {
		var stack []ast.Node
		ast.Inspect(f, func(n ast.Node) bool {
			if n == nil {
				stack = stack[:len(stack)-1]
				return true
			}
			var parent ast.Node
			if len(stack) > 0 {
				parent = stack[len(stack)-1]
			}
			stack = append(stack, n)
			switch x := n.(type) {
			case *ast.FuncType:
				switch parent.(type) {
				case *ast.FuncDecl, *ast.FuncLit:
					// a signature: every parameter and result is a binding
					for _, fl := range []*ast.FieldList{x.Params, x.Results} {
						if fl == nil {
							continue
						}
						for _, fd := range fl.List {
							if t := foreignRealmValue(info.TypeOf(fd.Type), self); t != nil {
								report(fd.Type.Pos(), t)
							}
						}
					}
				}
				stack = stack[:len(stack)-1] // no nil callback for a skipped node
				return false                 // a type expression, or judged above
			case *ast.TypeSpec:
				// own type built on another realm's value type
				if t := foreignRealmValue(info.TypeOf(x.Type), self); t != nil {
					report(x.Name.Pos(), t)
				}
			case *ast.CaseClause:
				if obj := info.Implicits[x]; obj != nil {
					if t := foreignRealmValue(obj.Type(), self); t != nil {
						report(x.Pos(), t)
					}
				}
			}
			if id, ok := n.(*ast.Ident); ok && id.Name != "_" {
				if v, ok := info.Defs[id].(*types.Var); ok {
					if t := foreignRealmValue(v.Type(), self); t != nil {
						report(id.Pos(), t)
						return true
					}
				}
			}
			if x, ok := n.(ast.Expr); ok {
				tv, ok := info.Types[x]
				if !ok || !tv.IsValue() || exemptRealmValueUse(x, parent) {
					return true
				}
				if t := foreignRealmValue(tv.Type, self); t != nil {
					report(x.Pos(), t)
				}
			}
			return true
		})
	}
	return errs
}

// exemptRealmValueUse: reading a field of a temporary, or discarding it.
func exemptRealmValueUse(x ast.Expr, parent ast.Node) bool {
	switch p := parent.(type) {
	case *ast.SelectorExpr:
		return p.X == x
	case *ast.ExprStmt:
		return true
	case *ast.ParenExpr:
		return true // judged at the outer expression
	case *ast.UnaryExpr:
		return p.Op == token.AND // &T{} or &x: a pointer
	case *ast.CompositeLit:
		return p.Type == x // the type expression itself
	}
	return false
}

// foreignRealmValue returns the first realm-declared struct/array type that
// t holds by value and that is not declared in self.
func foreignRealmValue(t types.Type, self string) types.Type {
	switch u := t.(type) {
	case *types.Named:
		if pkg := u.Obj().Pkg(); pkg != nil && pkg.Path() != self && IsRealmPath(pkg.Path()) {
			switch u.Underlying().(type) {
			case *types.Struct, *types.Array:
				return u
			}
		}
		return nil // own or library named types: checked where declared
	case *types.Array:
		return foreignRealmValue(u.Elem(), self)
	case *types.Slice:
		return foreignRealmValue(u.Elem(), self)
	case *types.Map:
		if t := foreignRealmValue(u.Key(), self); t != nil {
			return t
		}
		return foreignRealmValue(u.Elem(), self)
	case *types.Chan:
		return foreignRealmValue(u.Elem(), self)
	case *types.Struct:
		for i := 0; i < u.NumFields(); i++ {
			if ft := foreignRealmValue(u.Field(i).Type(), self); ft != nil {
				return ft
			}
		}
	}
	return nil
}
