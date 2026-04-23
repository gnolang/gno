// Three-variable multi-value decl where each depends on a separate chain.
// A gets D, B gets E, C gets F — D -> E -> F creates a linear dependency.
package main

var (
	A, B, C = D, E, F
	D        = 1
	E        = D
	F        = E
)

func main() {
	println(A, B, C) // 1 1 1
}

// Output:
// 1 1 1
