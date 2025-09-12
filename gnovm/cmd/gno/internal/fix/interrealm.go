package fix

import (
	"fmt"
	"go/ast"
	"go/token"

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
   custom patchces; and potential p code will first to be tested in
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

func interrealm(f *ast.File) (fixed bool) {
	// cross(fn)(args) -> fn(cross, args)
	// func(...) (...) { crossing(); ... -> func(cur realm, ...) (...) { ...
	// for transformed fn in same package:
	//     fn(args) -> fn(cur, args)
	// change reserved names into <name>_ as available
	reservedNames := [...]string{
		"cross",
		"realm",
		"gnocoin",
		"gnocoins",
		"istypednil",
		"revive",
		"address",
	}

	apply(
		f,
		func(c *astutil.Cursor, sc scopes) bool {
			return true
		},
		func(c *astutil.Cursor, sc scopes) bool {
			if isBlockNode(c.Node()) {
				// if popping out of a block, rename any reserved names.
				last := sc[len(sc)-1]
				for _, rn := range reservedNames {
					// When popping out of the last block, there are also names
					// without a corresponding definition. Ignore them, as we
					// cannot rename the matching definition.
					if du := last[rn]; du != nil && du.def != nil {
						newName := rn + "_"
						for last[newName] != nil {
							newName += "_"
						}
						du.rename(newName)
						fixed = true
					}
				}
			}
			switch n := c.Node().(type) {
			case *ast.FuncLit:
				fixed = convertCrossing(sc, n.Body, n.Type) || fixed
			case *ast.FuncDecl:
				fixed = convertCrossing(sc, n.Body, n.Type) || fixed
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
					fixed = true
				}
			}
			return true
		},
	)
	return
}

func convertCrossing(sc scopes, body *ast.BlockStmt, tp *ast.FuncType) bool {
	if !hasCrossingStatement(body) {
		return false
	}

	addRealmArg(sc, tp)
	body.Lbrace = body.List[0].End()
	body.List = body.List[1:]
	return true
}

func hasCrossingStatement(bs *ast.BlockStmt) bool {
	// This function will panic only for nil ptr derefences or invalid
	// assertions. Rather than handling each of them, simply no-op recover when
	// they happen and return false.
	defer func() {
		recover()
	}()
	cx := bs.List[0].(*ast.ExprStmt).X.(*ast.CallExpr)
	return cx.Fun.(*ast.Ident).Name == "crossing" && len(cx.Args) == 0
}

func addRealmArg(sc scopes, ft *ast.FuncType) {
	var names []*ast.Ident
	hasName := len(ft.Params.List) == 0 || len(ft.Params.List[0].Names) > 0
	if hasName {
		id := "cur"
		for sc.lookup(id) != nil {
			id += "_"
		}
		names = []*ast.Ident{ast.NewIdent(id)}
	}
	end := token.NoPos
	if len(ft.Params.List) == 0 {
		end = ft.Params.End() - 1
	}
	ft.Params.List = append([]*ast.Field{{
		Names: names,
		Type: &ast.Ident{
			NamePos: end,
			Name:    "realm",
		},
	}}, ft.Params.List...)
}
