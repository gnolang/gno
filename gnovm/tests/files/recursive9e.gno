package main

// The slice breaks the cycle's infinite size, but every type in the cycle
// is an alias, so the chain never resolves to a concrete type.

type (
	A = []B
	B = A
)

func main() {}

// Error:
// main/recursive9e.gno:7:2-9: invalid recursive type: A -> B -> A

// TypeCheckError:
// main/recursive9e.gno:7:2: invalid recursive type A; main/recursive9e.gno:7:2: 	A refers to B; main/recursive9e.gno:8:2: 	B refers to A
