package main

// Tests that method bodies are analysed for package-level variable dependencies.
// A = T{}.GetB() has no direct syntactic dependency on B, but GetB's body
// reads B. The init-order analysis must trace into the method body (via
// selectorRecvTypeName + findDependentNames recursion) to discover that A
// transitively depends on B.
// Init order: T (type, no deps), B (no deps), A (deps T and B via GetB).

type T struct{}

func (T) GetB() int { return B }

var (
	A = T{}.GetB()
	B = 42
)

func main() { println(A) }

// Output:
// 42
