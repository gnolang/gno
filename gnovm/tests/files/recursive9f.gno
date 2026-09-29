package main

// An alias chain that loops back on itself, with no indirection and no
// defined type to resolve to.

type (
	A = B
	B = A
)

func main() {}

// Error:
// main/recursive9f.gno:7:2-7: invalid recursive type: A -> B -> A

// TypeCheckError:
// main/recursive9f.gno:7:2: invalid recursive type A; main/recursive9f.gno:7:2: 	A refers to B; main/recursive9f.gno:8:2: 	B refers to A
