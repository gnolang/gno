// Multi-value decl with an extra level of indirection.
// A gets C, B gets D; C depends on B (cross-dependency via arithmetic).
package main

var (
	A, B = C, D
	C    = B + 1
	D    = 2
)

func main() {
	println(A, B, C) // 3 2 3
}

// Output:
// 3 2 3
