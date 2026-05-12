// Three-var cycle: A -> B -> C -> A must be rejected.
package main

var (
	A = B
	B = C
	C = A
)

func main() {
	println(A)
}

// Error:
// main/var43.gno:5:2-7: invalid recursive value: A -> B -> C -> A

// TypeCheckError:
// main/var43.gno:5:2: initialization cycle for A; main/var43.gno:5:2: 	A refers to B; main/var43.gno:6:2: 	B refers to C; main/var43.gno:7:2: 	C refers to A
