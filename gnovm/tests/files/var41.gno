// Genuinely cyclic var declarations must still be rejected.
package main

var (
	A = B
	B = A
)

func main() {
	println(A, B)
}

// Error:
// main/var41.gno:5:2-7: invalid recursive value: A -> B -> A

// TypeCheckError:
// main/var41.gno:5:2: initialization cycle for A; main/var41.gno:5:2: 	A refers to B; main/var41.gno:6:2: 	B refers to A
