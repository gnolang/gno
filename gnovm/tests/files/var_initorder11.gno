package main

// Tests that init order correctly handles transitive deps discovered through a
// pointer-receiver method body. A = (&T{}).Get() has no direct syntactic
// dependency on B, but Get's body reads B. This exercises the *PointerType
// branch in findDependentNames' SelectorExpr case:
//
// Contrast with var_initorder10.gno which uses a value receiver T{}.GetB().
// Init order: T (type, no deps), B (no deps), A (deps T and B via *T.Get).

type T struct{}

func (t *T) Get() int { return B }

var (
	A = (&T{}).Get()
	B = 42
)

func main() { println(A) }

// Output:
// 42
