package gnolang

import (
	"fmt"
	"go/ast"
	"go/token"
)

// checkNoGenerics rejects go1.18 generics syntax before go/types runs.
// Gno targets go1.17 and does not support generics, but modern go/types
// accepts them — without this guard a package declaring (but not
// instantiating) generics deploys silently and its declarations are
// dead weight with undefined Gno semantics. Rejected syntactically:
//   - type parameters, on a type or func declaration (`type W[P any] ...`);
//   - interface type-set terms: unions (`A | B`) and approximations (`~T`).
//
// Bare non-interface type-set elements (`interface{ int }`) are NOT
// detectable syntactically (indistinguishable from a legal embedded
// interface ident) and are left to a future preprocess check.
//
// It reports the earliest-positioned offending construct, so the error is
// deterministic (the message can reach consensus-visible tx results).
func checkNoGenerics(fset *token.FileSet, gofs []*ast.File) error {
	var (
		off  token.Pos
		what string
	)
	note := func(pos token.Pos, kind string) {
		if !off.IsValid() || pos < off {
			off, what = pos, kind
		}
	}
	for _, gof := range gofs {
		ast.Inspect(gof, func(n ast.Node) bool {
			switch t := n.(type) {
			case *ast.TypeSpec:
				if t.TypeParams != nil {
					note(t.Name.Pos(), "generic type declarations")
				}
			case *ast.FuncType:
				if t.TypeParams != nil {
					note(t.Pos(), "generic functions")
				}
			case *ast.InterfaceType:
				for _, f := range t.Methods.List {
					if len(f.Names) != 0 {
						continue // a method, not a type-set term
					}
					switch e := f.Type.(type) {
					case *ast.BinaryExpr:
						// `|` in this position is only a type union
						// (elsewhere it is bitwise-or).
						if e.Op == token.OR {
							note(e.Pos(), "interface type unions")
						}
					case *ast.UnaryExpr:
						// `~` is exclusively type-approximation.
						if e.Op == token.TILDE {
							note(e.Pos(), "interface approximation (~) terms")
						}
					}
				}
			}
			return true
		})
	}
	if off.IsValid() {
		return fmt.Errorf("%s: %s are not supported (Gno targets go1.17)",
			fset.Position(off), what)
	}
	return nil
}
