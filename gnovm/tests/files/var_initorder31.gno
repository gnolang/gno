// run

package main

// Regression: codaInitOrderDeps must unwrap *PointerType at the type
// level for method expressions on a pointer-aliased type. Here PT is
// a type alias for *T, so n.X is the name PT (no *StarExpr at AST
// level) and tv.Type is *PointerType{Elt: *DeclaredType{T}}. The
// type-level unwrap is what makes the T.M dep land.

type T int

type PT = *T

func (r *T) M() int { return a }

var dummy = PT.M(new(T))

var a = identity(1)

func identity(x int) int { return x }

func main() {
	if dummy != 1 {
		panic("dummy != 1")
	}
	println(dummy)
}

// Output:
// 1
