package main

// Tests a 3-level transitive dependency chain through alternating
// FuncDecl and ValueDecl nodes: A -> f() -> B -> g() -> C.
// resolveEffectiveDeps collapses FuncDecl edges; Kahn's algorithm
// handles ValueDecl→ValueDecl transitivity.

func f() int { return B }
func g() int { return C }

var (
	A = f()
	B = g()
	C = 42
)

func main() { println(A) }

// Output:
// 42
