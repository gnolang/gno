package fix

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/ast/astutil"
)

/*
Translate Interrealm Spec 2 to Interrealm Spec 3 (Gno 0.9)

 - Interrealm Spec 1: Original; every realm function is (automatically)
 a crossing function. This was working for our examples and was
 conceptually simple, but several problems were identified late in
 development;

   1. p package code copied over to r realms would behave differently
   with respect to std.CurrentRealm() and std.PreviousRealm(). It will
   become typical after launch that p code gets copied to r code for
   cutstom patchces; and potential p code will first to be tested in
   more mutable r realms.

   2. a reentrancy issue exists where r realm's calls to some variable
   function/method `var A func(...)...` are usually of functions
   declared in external realms (such as callback functions expected to
   be provided by the external realm) but instead ends up being a
   function declared in the the same r realm, an expected realm
   boundary isn't there, and may lead to exploits.

 - Interrealm Spec 2: With explicit cross(fn)(...) and crossing()
 declarations. The previous problems were solved by explicit crossing()
 declarations in realm functions (solves 1), and explicit
 cross(fn)(...) calls (solves 2 for the most part). But more problems
 were identified after most of the migration was done for examples from
 spec 1 to spec 2:

   3. a reentrancy issue where if calls to r realm's function/method
   A() are usually expected to be done by external realms (creating a
   realm boundary), but the external caller does things to get the r
   realm to call its own A(), the expected realm boundary isn't created
   and may lead to exploits.

   3.A. As a more concrete example of problem 3, when a realm takes as
   parameter a callback function `cb func(...)...` that isn't expected
   to be a crossing function and thus not explicitly crossed into. An
   external user or realm can then craft a function literal expression
   that calls the aforementioned realm's crossing functions without an
   explicit cross(fn)(...) call, thereby again dissolving a realm
   function boundary where one should be.

   4. Users didn't like the cross(fn)(...) syntax.

 - Interrealm Spec 3: With @cross decorator and `cur realm` first
 argument type. Instead of declaring a crossing-function with
 `crossing()` as the first statement the @cross decorator is used for
 package/file level function/methods declarations. Function literals
 can likewise be declared crossing by being wrapped like
 cross(func(...)...{}). When calling from within the same realm
 (without creating a realm boundary), the `cur` value is passed through
 to the called function's via its first argument; but when a realm
 boundary is intended, `nil` is passed in instead. This resolves
 problem 3.A because a non-crossing function literal would not be
 declared with the `cur realm` first argument, and thus a non-crossing
 call of the same realm's crossing function would not be syntactically
 possible.

----------------------------------------

Also refer to the [Lint and Transpile ADR](./adr/pr4264_lint_transpile.md).
*/

func interrealm(opts Options, f *ast.File) bool {
	// cross(fn)(args) -> fn(cross, args)
	// func(...) (...) { crossing(); ... -> func(cur realm, ...) (...) { ...
	// for transformed fn in same package:
	//     fn(args) -> fn(cur, args)
	// identifiers renamed:
	//     cross realm gnocoin gnocoins istypednil

	apply(
		f,
		func(c *astutil.Cursor, sc scopes) bool {
			return true
		},
		func(c *astutil.Cursor, sc scopes) bool {
			switch n := c.Node().(type) {
			case *ast.FuncLit:
				if hasCrossingStatement(n.Body.List) {
					addRealmArg(n.Type, n)
					n.Body.List = n.Body.List[1:]
				}
			case *ast.FuncDecl:
				if hasCrossingStatement(n.Body.List) {
					addRealmArg(n.Type, n)
					n.Body.List = n.Body.List[1:]
				}
			case *ast.CallExpr:
				// Rewrite cross(fn)(args...) -> fn(cross, args...)
				cx, ok := n.Fun.(*ast.CallExpr)
				if !ok {
					break
				}
				id, ok := cx.Fun.(*ast.Ident)
				if ok && id.Name == "cross" {
					if len(cx.Args) != 1 {
						panic(fmt.Errorf("invalid cross(fn) call with %d args", len(cx.Args)))
					}
					n.Fun = cx.Args[0]
					n.Args = append([]ast.Expr{
						ast.NewIdent("cross"),
					}, n.Args...)
				}
			}
			return true
		},
	)
	// Assume fixed = true, there may be more transformations done afterwards.
	return true
}

func hasCrossingStatement(ss []ast.Stmt) bool {
	// This function will panic only for nil ptr derefences or invalid
	// assertions. Rather than handling each of them, simply no-op recover when
	// they happen and return false.
	defer func() {
		recover()
	}()
	cx := ss[0].(*ast.ExprStmt).X.(*ast.CallExpr)
	return cx.Fun.(*ast.Ident).Name == "crossing" && len(cx.Args) == 0
}

func addRealmArg(ft *ast.FuncType, n ast.Node) {
	var names []*ast.Ident
	hasName := len(ft.Params.List) > 0 && len(ft.Params.List[0].Names) > 0
	if hasName {
		used := findUsedNames(n)
		id := "cur"
		for {
			_, ok := used[id]
			if !ok {
				break
			}
			id += "_"
		}
		names = []*ast.Ident{ast.NewIdent(id)}
	}
	ft.Params.List = append([]*ast.Field{{
		Names: names,
		Type:  ast.NewIdent("realm"),
	}}, ft.Params.List...)
}

func findUsedNames(n ast.Node) map[string]struct{} {
	used := make(map[string]struct{}, 32)
	var visit func(ast.Node) bool
	visit = func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.SelectorExpr:
			// only care about n.X, ignore n.Sel.
			ast.Inspect(n.X, visit)
			return false
		case *ast.TypeSpec:
			// Type specs cannot use values within, so ignore.
			visit(n.Name)
			return false
		case *ast.Ident:
			used[n.Name] = struct{}{}
		}
		return true
	}
	ast.Inspect(n, visit)
	return used
}
