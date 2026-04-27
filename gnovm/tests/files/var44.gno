// Multi-value decl where the two names swap — a genuine cycle.
package main

var (
	A, B = B, A
)

func main() {
	println(A, B)
}

// Error:
// main/var44.gno:2:1-10:2: invalid recursive value: A -> B -> A

// TypeCheckError:
// main/var44.gno:5:2: initialization cycle for A; main/var44.gno:5:2: 	A refers to B; main/var44.gno:5:5: 	B refers to A
