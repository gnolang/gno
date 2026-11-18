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
// type in files type-checked as own, running in realm pkgPath. A filetest is
// type-checked under its directory's path but runs under its PKGPATH.
func checkRealmValues(fset *token.FileSet, files []*ast.File, info *types.Info, own *types.Package, pkgPath string) []error {
	self := realmSelf{own, pkgPath}
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
			if call, ok := n.(*ast.CallExpr); ok && len(call.Args) == 1 && info.Types[call.Fun].IsType() {
				// a conversion to or from another realm's value type (or a
				// pointer to one) reinterprets that realm's data
				to, from := info.TypeOf(call), info.TypeOf(call.Args[0])
				isNil := from != nil && types.Identical(from, types.Typ[types.UntypedNil])
				if from != nil && !isNil && !types.Identical(derefType(to), derefType(from)) {
					t := foreignRealmValue(derefType(to), self)
					if t == nil {
						t = foreignRealmValue(derefType(from), self)
					}
					if t != nil {
						report(call.Pos(), t)
						return true
					}
				}
			}
			if call, ok := n.(*ast.CallExpr); ok {
				// an argument implicitly converted into another realm's value
				// parameter is a value of that type built here
				if sig, ok := types.Unalias(info.TypeOf(call.Fun)).(*types.Signature); ok {
					for i, arg := range call.Args {
						pt := paramType(sig, i, call.Ellipsis.IsValid())
						if pt == nil || types.Identical(pt, info.TypeOf(arg)) {
							continue
						}
						if t := foreignRealmValue(pt, self); t != nil {
							report(arg.Pos(), t)
						}
					}
				}
			}
			if x, ok := n.(ast.Expr); ok {
				tv, ok := info.Types[x]
				if !ok || !tv.IsValue() || exemptRealmValueUse(x, parent, info) {
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
func exemptRealmValueUse(x ast.Expr, parent ast.Node, info *types.Info) bool {
	switch p := parent.(type) {
	case *ast.SelectorExpr:
		return p.X == x
	case *ast.IndexExpr:
		return p.X == x // the element is judged at the index expression
	case *ast.RangeStmt:
		return p.X == x // bound key/value variables are judged themselves
	case *ast.CallExpr:
		// len(x) / cap(x) only read a length
		if id, ok := ast.Unparen(p.Fun).(*ast.Ident); ok && (id.Name == "len" || id.Name == "cap") {
			return info.Types[p.Fun].IsBuiltin()
		}
	case *ast.ExprStmt, *ast.DeferStmt, *ast.GoStmt:
		return true // a discarded call result
	case *ast.ParenExpr:
		return true // judged at the outer expression
	case *ast.UnaryExpr:
		return p.Op == token.AND // &T{} or &x: a pointer
	case *ast.CompositeLit:
		return p.Type == x // the type expression itself
	}
	return false
}

// realmSelf names the package being checked and the realm it runs in.
type realmSelf struct {
	pkg  *types.Package
	path string
}

// foreignRealmValue returns the first realm-declared struct/array type that
// t holds by value and that is not declared in self.
func foreignRealmValue(t types.Type, self realmSelf) types.Type {
	if t == nil {
		return nil // an expression go/types could not type
	}
	switch u := types.Unalias(t).(type) { // aliases: same verdict whatever gotypesalias says
	case *types.Named:
		if pkg := u.Obj().Pkg(); pkg != nil && pkg != self.pkg && pkg.Path() != self.path && IsRealmPath(pkg.Path()) {
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
	case *types.Tuple: // multi-value results and comma-ok forms
		for i := 0; i < u.Len(); i++ {
			if ft := foreignRealmValue(u.At(i).Type(), self); ft != nil {
				return ft
			}
		}
	}
	return nil
}

// derefType is the type a pointer (of any named pointer type) points at.
func derefType(t types.Type) types.Type {
	if t == nil {
		return nil
	}
	if p, ok := types.Unalias(t).Underlying().(*types.Pointer); ok {
		return p.Elem()
	}
	return t
}

// paramType is the type argument i is assigned to (nil if unknown).
func paramType(sig *types.Signature, i int, ellipsis bool) types.Type {
	ps := sig.Params()
	n := ps.Len()
	switch {
	case n == 0:
		return nil
	case sig.Variadic() && i >= n-1 && !ellipsis:
		if sl, ok := ps.At(n - 1).Type().(*types.Slice); ok {
			return sl.Elem()
		}
		return nil
	case i < n:
		return ps.At(i).Type()
	}
	return nil
}
