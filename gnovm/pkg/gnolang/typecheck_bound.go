package gnolang

import (
	"fmt"
	"go/ast"
	"go/token"
	"math"

	"github.com/gnolang/gno/tm2/pkg/overflow"
)

// typeExpansionBudget bounds the number of nodes that go/types' validType walk
// would visit for any single named type in a package.
//
// go/types validates that named types do not "expand" indefinitely
// (src/go/types/validtype.go). Its walk follows value-containment edges only —
// struct fields, array elements, interface embeddeds, union terms, and a named
// type's underlying RHS — and crucially it does NOT memoize visited types (the
// optimization is commented out as a workaround for golang/go#65711). As a
// result the walk is exponential in the worst case: a "doubling" chain such as
//
//	type T0 struct{ x int }
//	type T1 struct{ a, b [0]T0 } // references the previous level TWICE by value
//	type T2 struct{ a, b [0]T1 }
//	// ... depth ~40 => 2^40 node visits
//
// hangs the type checker. Because TypeCheckMemPackage runs unmetered at deploy
// time (VMKeeper.AddPackage / MsgRun), a ~40-line package is a consensus DoS.
//
// checkTypeExpansionBound computes that same node-visit count, but WITH the
// memoization validType lacks, making it linear, and rejects the package before
// go/types runs if any named type's count exceeds the budget. The budget is a
// deterministic count (not a wall-clock limit), so the check is consensus-safe.
//
// 1_000_000 leaves enormous headroom over honest packages (whose counts are in
// the tens to low thousands) while keeping validType's actual run at the bound
// well under ~10ms.
//
// MAINTENANCE: cost() below mirrors the exact set of edges validType recurses
// through. If a Go toolchain upgrade changes that set (adds a containment edge,
// or finally memoizes validType per golang/go#65711), revisit this file —
// under-counting a new edge would silently reopen the DoS.
const typeExpansionBudget = 1_000_000

// checkTypeExpansionBound rejects packages whose named types would cause
// go/types' validType walk to run super-linearly. See typeExpansionBudget.
func checkTypeExpansionBound(fset *token.FileSet, gofs []*ast.File) error {
	// Collect every type declaration (package-level and function-local) keyed by
	// name. Local types are validated by go/types too, so they are also a vector.
	// On name collision (invalid Go that go/types would reject later) we keep all
	// candidates and resolve an identifier to the highest-cost one — a safe
	// over-approximation that never under-counts.
	decls := make(map[string][]*ast.TypeSpec)
	for _, gof := range gofs {
		ast.Inspect(gof, func(n ast.Node) bool {
			if ts, ok := n.(*ast.TypeSpec); ok {
				decls[ts.Name.Name] = append(decls[ts.Name.Name], ts)
			}
			return true
		})
	}

	memo := make(map[ast.Expr]uint64)
	visiting := make(map[ast.Expr]bool)

	// cost returns the number of validType nodes visited for a type expression.
	var cost func(e ast.Expr) uint64
	// namedCost returns the count for a named type given its RHS expression:
	// one node for the named type itself plus the cost of its underlying type.
	namedCost := func(rhs ast.Expr) uint64 {
		if v, ok := memo[rhs]; ok {
			return v
		}
		if visiting[rhs] {
			// A value-containment cycle: an invalid recursive type that go/types
			// detects and reports itself. Return a finite count so we neither loop
			// nor pre-empt go/types' diagnostic.
			return 1
		}
		visiting[rhs] = true
		v := satAdd(1, cost(rhs))
		visiting[rhs] = false
		memo[rhs] = v
		return v
	}
	identCost := func(name string) uint64 {
		specs := decls[name]
		if len(specs) == 0 {
			// Builtin, type parameter, or otherwise unresolved within this package:
			// validType treats it as a leaf.
			return 1
		}
		var best uint64 = 1
		for _, ts := range specs {
			if c := namedCost(ts.Type); c > best {
				best = c
			}
		}
		return best
	}
	cost = func(e ast.Expr) uint64 {
		switch t := e.(type) {
		case *ast.ParenExpr:
			return cost(t.X)
		case *ast.StarExpr:
			return 1 // pointer: validType does not recurse
		case *ast.ArrayType:
			if t.Len == nil {
				return 1 // slice: not recursed
			}
			return satAdd(1, cost(t.Elt)) // array: recurse into element
		case *ast.MapType, *ast.ChanType, *ast.FuncType:
			return 1 // not recursed
		case *ast.StructType:
			total := uint64(1)
			for _, f := range t.Fields.List {
				mult := uint64(len(f.Names))
				if mult == 0 {
					mult = 1 // embedded field
				}
				total = satAdd(total, satMul(mult, cost(f.Type)))
				if total > typeExpansionBudget {
					return total
				}
			}
			return total
		case *ast.InterfaceType:
			total := uint64(1)
			for _, f := range t.Methods.List {
				if len(f.Names) != 0 {
					continue // method: a func signature, not recursed
				}
				total = satAdd(total, cost(f.Type)) // embedded type / type elem
				if total > typeExpansionBudget {
					return total
				}
			}
			return total
		case *ast.Ident:
			return identCost(t.Name)
		case *ast.SelectorExpr:
			return 1 // imported type: already validated in its own package
		case *ast.IndexExpr:
			return cost(t.X) // generic instantiation: bound by the base type
		case *ast.IndexListExpr:
			return cost(t.X)
		default:
			return 1
		}
	}

	// Report the earliest-declared offending type, so the error is deterministic
	// regardless of map iteration order (the message can reach consensus-visible
	// tx results).
	var off *ast.TypeSpec
	var offVal uint64
	for _, specs := range decls {
		for _, ts := range specs {
			if v := namedCost(ts.Type); v > typeExpansionBudget &&
				(off == nil || ts.Name.Pos() < off.Name.Pos()) {
				off, offVal = ts, v
			}
		}
	}
	if off != nil {
		return fmt.Errorf(
			"%s: type %s expands to at least %d nodes during type validation, "+
				"exceeding the limit of %d (possible denial-of-service vector)",
			fset.Position(off.Name.Pos()), off.Name.Name, offVal, typeExpansionBudget)
	}
	return nil
}

// satAdd returns a+b, saturating at math.MaxUint64 on overflow.
func satAdd(a, b uint64) uint64 {
	if s, ok := overflow.Add(a, b); ok {
		return s
	}
	return math.MaxUint64
}

// satMul returns a*b, saturating at math.MaxUint64 on overflow.
func satMul(a, b uint64) uint64 {
	if p, ok := overflow.Mul(a, b); ok {
		return p
	}
	return math.MaxUint64
}
